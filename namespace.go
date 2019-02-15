package goapollo

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type namespaceItem struct {
	serverUrl        string
	config           ApolloConfig
	namespace        string
	releaseKey       string
	requester        requester
	cacheRequester   requester
	dataRequester    requester
	cancel           func()
	ctx              context.Context
	configurations   *Configurations
	notifications    *ApolloNotificationMessages
	notificationChan chan *ApolloNotificationMessages
	once             sync.Once
	backFile         string
}

func NewNamespaceItem(ctx context.Context, serverUrl, namespace string, config ApolloConfig) (*namespaceItem, error) {
	c := &namespaceItem{
		serverUrl:        serverUrl,
		namespace:        strings.TrimSpace(namespace),
		cacheRequester:   newHTTPRequester(&http.Client{Timeout: time.Second * 30}),
		dataRequester:    newHTTPRequester(&http.Client{Timeout: time.Second * 30}),
		requester:        newHTTPRequester(&http.Client{}),
		config:           config,
		configurations:   NewConfiguration(),
		notifications:    NewApolloNotificationMessages(),
		notificationChan: make(chan *ApolloNotificationMessages, 1),
	}
	c.notifications.Store(namespace, defaultNotificationId)
	if err := c.preload(); err != nil {
		log.Printf("[%s] 预加载失败 -> %s", c.namespace, err)
		return nil, err
	}
	c.ctx, c.cancel = context.WithCancel(ctx)

	return c, nil
}

// preload 预加载配置信息,先尝试从远程服务器拉取，如果失败，则尝试从备份文件中读取，如果再失败则返回错误.
func (c *namespaceItem) preload() error {
	if _, err := c.fetchFromDatabase(); err != nil {
		log.Printf("[%s] 从数据库拉取配置失败 -> %s", c.namespace, err)
		if b, err := ioutil.ReadFile(c.backFile); err == nil && len(b) > 0 {
			if err := c.configurations.Decode(b); err == nil {
				return nil
			}
		}
		return err
	}
	return nil
}

func (c *namespaceItem) Stop() {
	c.cancel()
}

// WatchWithContext 监听服务器端配置变更通知.
func (c *namespaceItem) Watch(ch chan *ChangeEvent) {
	c.once.Do(func() {
		go func() {
			timer := time.NewTimer(time.Second * 1)
			defer timer.Stop()
			for {
				select {
				case <-timer.C:
					message, err := c.request()
					if err != nil {
						if err == ErrConfigUnmodified {
							log.Printf("[%s] 没有最新通知 -> %s", c.namespace, err.Error())
							break
						}
						log.Printf("[%s] 请求 Apollo 服务器出错 -> %s", c.namespace, err)
					} else if message != nil {
						log.Printf("[%s] 通知结果 -> %s", c.namespace, message)

						select {
						case <-time.After(time.Second * 1):
							break
						case c.notificationChan <- message:
							//下次再监听通知需要用上一次的通知ID.
							message.Range(func(namespace string, id int64) bool {
								c.notifications.Put(namespace, id)
								return true
							})
							break
						}
					}
				case <-c.ctx.Done():
					log.Println("监听已停止 ->", c.namespace, c.serverUrl)
					return
				}
				timer.Reset(time.Second * 1)
			}
		}()

		go func() {
			for {
				select {
				case <-c.notificationChan:

					if kv, err := c.fetchFromCache(); err != nil {
						log.Printf("[%s] 处理变更失败 ->%s", c.namespace, err)
					} else {

						event := NewChangeEvent(c.namespace)
						if c.GetConfigType() == C_TYPE_POROPERTIES {
							//先遍历修改或删除的键值
							c.configurations.Range(func(k, v string) bool {
								vv, _ := kv[k]
								event.Store(k, v, vv)
								return true
							})

							for k, v := range kv {
								// 再处理新增的键值
								if vv, ok := c.configurations.Load(k); !ok {
									event.Store(k, vv, v)
								}
								c.configurations.Store(k, v)
							}
						} else if _, ok := kv["content"]; ok {
							oldValue, _ := c.configurations.Load("content")
							event.Store("content", oldValue, kv["content"])
							c.configurations.Store("content", kv["content"])
						}
						ch <- event
						log.Printf("[%s] 配置变更已处理 -> %+v", c.namespace, c.configurations)
					}
				case <-c.ctx.Done():
					log.Printf("[%s] 通知处理已退出", c.namespace)
					return
				}
			}
		}()
	})
}

// FetchAllTask 全量拉取配置
func (c *namespaceItem) FetchAllTask(t time.Duration) {
	go func() {
		timer := time.NewTimer(t)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				kv, err := c.fetchFromCache()
				if err != nil {
					if err == ErrConfigUnmodified {
						log.Printf("[%s] 配置未改变", c.namespace)
					} else {
						log.Printf("[%s] 全量拉取配置出错%s", c.namespace, err.Error())
					}
					continue
				}
				for k, v := range kv {
					c.configurations.Store(k, v)
				}
			case <-c.ctx.Done():
				log.Printf("[%s] 定时全量拉取已退出", c.namespace)
				return

			}
		}
	}()
}

// Namespace 获取命名空间名称
func (c *namespaceItem) Namespace() string {
	return c.namespace
}

func (c *namespaceItem) request() (*ApolloNotificationMessages, error) {

	serverUrl := getApolloRemoteNotificationUrl(c.serverUrl, c.config.AppId, c.config.Cluster, c.notifications)

	log.Println("正在发起通知获取请求 ->", serverUrl)

	bts, err := c.requester.request(serverUrl)

	if err != nil || len(bts) == 0 {
		return nil, err
	}
	ret := NewApolloNotificationMessages()

	var m []*ApolloConfigNotification

	if err := json.Unmarshal(bts, &m); err != nil {
		return nil, err
	}
	if m != nil {
		for _, item := range m {
			ret.Put(item.NamespaceName, item.NotificationId)
		}
	}
	return ret, nil
}

// fetchFromDatabase 从数据库读取配置信息
func (c *namespaceItem) fetchFromDatabase() (*ApolloResult, error) {

	serverUrl := getApolloRemoteConfigFromDbUrl(c.serverUrl, c.config.AppId, c.config.Cluster, c.namespace, c.releaseKey)

	log.Printf("[%s] 从远程配置系统拉取配置信息 ->%s", c.namespace, serverUrl)
	body, err := c.dataRequester.request(serverUrl)

	if err != nil {
		log.Fatalf("%+v", err)
		if err == ErrConfigUnmodified {
			return nil, ErrConfigUnmodified
		}
		return nil, err
	}
	if len(body) <= 0 {
		return nil, errors.New("没有配置信息")
	}

	config := ApolloResult{}

	//解码后，如果解码成功，则将新值复制给原对象
	err = json.Unmarshal(body, &config)
	if err == nil {
		c.releaseKey = config.ReleaseKey
		for k, v := range config.Configurations {
			c.configurations.Store(k, v)
		}
	}

	return &config, err
}

//fetchFromCache 从缓存中读取配置信息
func (c *namespaceItem) fetchFromCache() (map[string]string, error) {
	serverUrl := getApolloRemoteConfigFromCacheUrl(c.serverUrl, c.config.AppId, c.config.Cluster, c.namespace)

	body, err := c.cacheRequester.request(serverUrl)

	if err != nil {
		return nil, err
	}
	kv := make(map[string]string, 0)
	err = json.Unmarshal(body, &kv)
	if err != nil {
		return nil, err
	}
	//log.Printf("[%s] 从缓存中拉取配置 %s - %+v\n", c.namespace, serverUrl, kv)

	return kv, nil
}

// GetValue 获取配置
func (c *namespaceItem) GetValue(key string) (string, bool) {
	return c.configurations.Load(key)
}

// GetConfigType 获取配置类型
func (c *namespaceItem) GetConfigType() ConfigType {
	switch filepath.Ext(c.namespace) {
	case ".json":
		return C_TYPE_JSON
	case ".xml":
		return C_TYPE_XML
	case ".yaml":
		return C_TYPE_YAML
	case ".yml":
		return C_TYPE_YML
	}
	return C_TYPE_POROPERTIES
}
