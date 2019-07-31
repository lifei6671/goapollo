package goapollo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultConfigName = "app.properties"
	defaultNamespace  = "application"

	defaultNotificationId = -1
)

type result struct {
	AppId          string            `json:"appId"`
	Cluster        string            `json:"cluster"`
	NamespaceName  string            `json:"namespaceName"`
	Configurations map[string]string `json:"configurations"`
	ReleaseKey     string            `json:"releaseKey"`
}

type Client struct {
	host         string
	appId        string
	cluster      string
	cacheDir     string
	ip           string
	caches       *namespaceCache
	notification INotification
	rmx          *sync.RWMutex
	eventCh      chan *ChangeEvent
	done         uint32
	cancel       context.CancelFunc
	releaseRepo  *sync.Map
	client       *http.Client
}

func New(host, appId, cluster string) *Client {
	var netTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   90 * time.Second, //连接超时时间
			KeepAlive: 90 * time.Second, //连接保持超时时间
		}).DialContext,
		MaxIdleConns:        20,                //client对与所有host最大空闲连接数总和
		IdleConnTimeout:     100 * time.Second, //空闲连接在连接池中的超时时间
		TLSHandshakeTimeout: 100 * time.Second, //TLS安全连接握手超时时间
	}
	return &Client{
		host:         host,
		appId:        appId,
		cluster:      cluster,
		cacheDir:     os.TempDir(),
		caches:       newNamespaceCache(),
		notification: newNotificationRepo(host, appId, cluster),
		rmx:          &sync.RWMutex{},
		eventCh:      make(chan *ChangeEvent, 100),
		releaseRepo:  &sync.Map{},
		client: &http.Client{
			Timeout:   time.Second * 30,
			Transport: netTransport,
		},
	}
}

func (c *Client) SetCacheDir(dir string) {
	c.cacheDir = dir
}

func (c *Client) SetClientIp(ip string) {
	c.ip = ip
}

func (c *Client) preload(namespace string) {
	err := c.caches.load(namespace, filepath.Join(c.cacheDir, c.appId, namespace))
	if err != nil {
		logger.Printf("解析备份文件失败 -> %s", err)
	}
}

func (c *Client) sync(namespace string) (*ChangeEvent, error) {
	configUrl := fmt.Sprintf("%s/configs/%s/%s/%s?releaseKey=%s&ip=%s",
		c.host,
		url.QueryEscape(c.appId),
		url.QueryEscape(c.cluster),
		url.QueryEscape(namespace),
		c.GetReleaseKey(namespace),
		c.ip,
	)
	logger.Printf("正在获取最新配置 -> %s", configUrl)
	req, err := http.NewRequest("GET", configUrl, nil)
	if err != nil {
		logger.Printf("构建 Request 出错 -> %s", err)
		return nil, err
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, nil
	}
	if resp.StatusCode == http.StatusNotModified {
		return nil, nil
	}
	body, err := ioutil.ReadAll(resp.Body)

	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Printf("发起通知请求失败 -> %s - %d - %s ", configUrl, resp.StatusCode, string(body))
		return nil, err

	}
	logger.Printf("获取最新配置成功 -> %s - %s", configUrl, string(body))
	var result result
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Printf("解析服务端响应值失败 -> %s - %s - %s", configUrl, string(body), err)
		return nil, err
	}
	c.releaseRepo.Store(result.NamespaceName, result.ReleaseKey)

	return c.caches.store(result), nil
}

//AddNamespace 使用默认序列化器添加命名空间
func (c *Client) AddNamespace(name string) *Client {
	c.AddNamespaceWithSerializer(name, NewJsonSerializer())
	return c
}

//AddNamespaceAndSerializer 使用自定义序列化器添加命名空间
func (c *Client) AddNamespaceWithSerializer(namespace string, serializer Serializer) *Client {
	path := filepath.Join(c.cacheDir, c.appId, namespace)
	c.AddNamespaceWithSerializerWithPath(namespace, serializer, path)
	return c
}

func (c *Client) AddNamespaceWithSerializerWithPath(namespace string, serializer Serializer, filename string) *Client {
	c.notification.AddNamespace(namespace)
	c.caches.addSerializer(namespace, serializer)
	err := c.caches.load(namespace, filename)
	if err != nil {
		logger.Printf("解析备份文件失败 -> %s", err)
	}
	return c
}

func (c *Client) Run(ctx context.Context) error {
	if atomic.LoadUint32(&c.done) == 1 {
		return errors.New("apollo already running")
	}
	ctx1, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Printf("出现未处理异常 -> %s", err)
			}
			logger.Printf("通知变更监听已退出")
		}()

		for {
			select {
			case notify := <-c.notification.Watch():

				if event, err := c.sync(notify.NamespaceName); err != nil {
					logger.Printf("同步最新配置失败 -> %s - %s - %s - %s", c.appId, c.cluster, notify.NamespaceName, err)
					break
				} else if event != nil {
					logger.Printf("事件通知 -> %+v", event)
					select {
					case c.eventCh <- event:
					default:
					}
					_ = c.caches.dump(notify.NamespaceName)
				}

			case <-ctx1.Done():
				return
			}
		}
	}()
	return nil
}

func (c *Client) Close() error {
	c.rmx.Lock()
	defer c.rmx.Unlock()
	_ = c.caches.save()

	if c.cancel != nil {
		c.cancel()
	}
	c.caches = newNamespaceCache()
	_ = c.notification.Close()
	close(c.eventCh)

	return nil
}

func (c *Client) WatchUpdate() <-chan *ChangeEvent {
	return c.eventCh
}

// GetReleaseKey 获取指定命名空间的版本号.
func (c *Client) GetReleaseKey(namespace string) string {
	if releaseKey, ok := c.releaseRepo.Load(namespace); ok {
		return releaseKey.(string)
	}

	return ""
}

//GetValue 获取默认命名空间的指定键值.
func (c *Client) GetValue(key string) (val string, exist bool) {
	val, exist = c.caches.get(defaultNamespace, key)
	return
}

//GetValueWithNamespace 获取指定命名空间的指定键值.
func (c *Client) GetValueWithNamespace(namespace, key string) (val string, exist bool) {
	val, exist = c.caches.get(namespace, key)
	return
}

//GetContentWithNamespace 获取指定命名空间的内容.
func (c *Client) GetContentWithNamespace(namespace string) (val string, exist bool) {
	val, exist = c.caches.get(namespace, "content")
	return
}

func (c *Client) AllKeys(namespace string) []string {
	return c.caches.keys(namespace)
}
