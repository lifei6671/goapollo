package goapollo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Notification struct {
	NamespaceName  string `json:"namespaceName,omitempty"`
	NotificationId int    `json:"notificationId,omitempty"`
}

func (notify *Notification) String() string {
	if notify == nil {
		return ""
	}
	body, err := json.Marshal(notify)
	if err != nil {
		return ""
	}
	return string(body)
}

type INotification interface {
	io.Closer
	AddNamespace(namespace string)
	DeleteNamespace(namespace string)
	Watch() <-chan *Notification
}

type notificationRepo struct {
	notifications   *sync.Map
	notificationCh  chan *Notification
	client          *http.Client
	notificationUrl string
	cancel          context.CancelFunc
	once            *sync.Once
}

func newNotificationRepo(host, appId, cluster string) *notificationRepo {

	var netTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   90 * time.Second, //连接超时时间
			KeepAlive: 90 * time.Second, //连接保持超时时间
		}).DialContext,
		MaxIdleConns:        20,                //client对与所有host最大空闲连接数总和
		IdleConnTimeout:     100 * time.Second, //空闲连接在连接池中的超时时间
		TLSHandshakeTimeout: 100 * time.Second, //TLS安全连接握手超时时间
	}

	notificationUrl := fmt.Sprintf("%s/notifications/v2?appId=%s&cluster=%s&notifications=", host, appId, cluster)

	return &notificationRepo{
		notifications:   &sync.Map{},
		notificationUrl: notificationUrl,
		client: &http.Client{
			Timeout:   time.Second * 90,
			Transport: netTransport,
		},
		notificationCh: make(chan *Notification, 10),
		once:           &sync.Once{},
	}
}

func (n *notificationRepo) AddNamespace(namespace string) {
	n.notifications.Store(namespace, defaultNotificationId)
}

func (n *notificationRepo) DeleteNamespace(namespace string) {
	n.notifications.Delete(namespace)
}

func (n *notificationRepo) Watch() <-chan *Notification {
	n.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		n.cancel = cancel
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					{
						notificationUrl := n.notificationUrl + url.QueryEscape(n.String())
						logger.Printf("正在发起通知 -> %s\n", notificationUrl)
						req, err := http.NewRequest("GET", notificationUrl, nil)
						if err != nil {
							logger.Printf("构建 Request 出错 -> %s", err)
							break
						}
						req = req.WithContext(ctx)

						resp, err := n.client.Do(req)

						if err != nil {
							logger.Printf("发起通知请求失败 -> %s - %s", notificationUrl, err)
							break
						}
						if resp.StatusCode == http.StatusNotModified {
							logger.Printf("服务器端配置未改变 -> %d", resp.StatusCode)
							_ = resp.Body.Close()
							break
						}

						body, err := ioutil.ReadAll(resp.Body)

						_ = resp.Body.Close()

						if err != nil {
							logger.Printf("读取通知响应失败 -> %s - %s", notificationUrl, err)
							break
						}
						if resp.StatusCode != http.StatusOK {
							logger.Printf("服务器响应失败 -> %d - %s", resp.StatusCode, string(body))
							break
						}
						logger.Printf("正在解析通知 -> %s - %s", notificationUrl, string(body))
						var notifications []*Notification
						err = json.Unmarshal(body, &notifications)
						if err != nil {
							logger.Printf("解析通知响应失败 -> %s - %s - %s", notificationUrl, string(body), err)
							break
						}
						for i, item := range notifications {
							//这里预防将删除后的通知再次存入到缓存中
							if _, ok := n.notifications.Load(item.NamespaceName); ok {
								n.notifications.Store(item.NamespaceName, item.NotificationId)
								n.notificationCh <- notifications[i]
							}
						}
					}
				}
			}
		}()
	})

	return n.notificationCh
}

func (n *notificationRepo) String() string {
	var notifications []Notification

	n.notifications.Range(func(key, value interface{}) bool {
		item := Notification{NamespaceName: key.(string), NotificationId: value.(int)}
		notifications = append(notifications, item)
		return true
	})
	body, err := json.Marshal(notifications)
	if err != nil {
		return ""
	}
	return string(body)
}

func (n *notificationRepo) Close() error {
	if n.cancel != nil {
		n.cancel()
	}
	return nil
}
