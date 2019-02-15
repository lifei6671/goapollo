package goapollo

import (
	"context"
	"log"
)

type ChangeHandler func()

type ApolloClient struct {
	//Apollo配置服务的地址.
	serverUrl   string
	config      ApolloConfig
	namespace   *NamespaceList
	ctx         context.Context
	cancel      func()
	changeEvent func(event *ChangeEvent)
}

// NewApolloClient
func NewApolloClient(ctx context.Context, serverUrl string, config ApolloConfig, namespace string) *ApolloClient {

	c := &ApolloClient{
		serverUrl: serverUrl,
		config:    config,
		namespace: NewNamespaceList(),
	}
	c.ctx, c.cancel = context.WithCancel(ctx)

	item, err := NewNamespaceItem(c.ctx, c.serverUrl, namespace, config)
	if err != nil {
		log.Fatalf("启动客户端失败 -> %s", err)
	}
	c.namespace.Store(namespace, item)

	return c
}

// Watch 启动监听所有命名空间的变更事件
func (c *ApolloClient) Watch() {

	c.namespace.Range(func(key string, nsItem *namespaceItem) bool {
		go func(item *namespaceItem) {
			log.Printf("【%s】 -> %s", item.namespace, item.GetConfigType())
			ch := make(chan *ChangeEvent)
			nsItem.Watch(ch)

			for {
				select {
				case item := <-ch:
					if c.changeEvent != nil {
						c.changeEvent(item)
					}
				case <-c.ctx.Done():
					log.Printf("[%s] 监听变更事件已退出", item.Namespace())
					return
				}
			}
		}(nsItem)
		return true
	})

}

// AddNamespace 增加命名空间
func (c *ApolloClient) AddNamespace(namespace string) error {
	ns, err := NewNamespaceItem(c.ctx, c.serverUrl, namespace, c.config)

	if err != nil {
		return err
	}
	c.namespace.Store(namespace, ns)
	return nil
}

func (c *ApolloClient) RemoveNamespace(namespace string) {
	c.namespace.Delete(namespace)
}

// OnChangeEvent 当配置变更后执行
func (c *ApolloClient) OnChangeEvent(f func(event *ChangeEvent)) {
	c.changeEvent = f
}

// Stop 停止所有配置同步
func (c *ApolloClient) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

// 定时将配置信息保存的文件中
func (c *ApolloClient) Save() {

}

// GetValue 获取默认命名空间指定键的配置值
// 如果配置不是 poroperties 类型，则返回所有配置信息
func (c *ApolloClient) GetValue(key string) (string, bool) {
	return c.namespace.GetValue("application", key)
}

// GetValueWithNamespace 获取指定命名空间指定键的配置值
// 如果配置不是 poroperties 类型，则返回所有配置信息
func (c *ApolloClient) GetValueWithNamespace(namespace string, key string) (string, bool) {
	return c.namespace.GetValue(namespace, key)
}
