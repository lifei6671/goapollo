package goapollo

import (
	"context"
	"log"
)

type ChangeHandler func()

type ApolloClient struct {
	//Apollo配置服务的地址.
	serverUrl     string
	config        ApolloConfig
	namespaceList *NamespaceList
	ctx           context.Context
	cancel        func()
	changeEvent   func(event *ChangeEvent)
	backFile      string
}

// NewApolloClient
func NewApolloClient(ctx context.Context, serverUrl string, config ApolloConfig, namespace string) *ApolloClient {

	c := &ApolloClient{
		serverUrl:     serverUrl,
		config:        config,
		namespaceList: NewNamespaceList(),
	}
	c.ctx, c.cancel = context.WithCancel(ctx)

	item, err := NewNamespaceItem(c.ctx, c.serverUrl, namespace, config)
	if err != nil {
		log.Fatalf("启动客户端失败 -> %s", err)
	}
	c.namespaceList.Store(namespace, item)

	return c
}

// NewApolloClientWithFile 从文件中初始
func NewApolloClientWithFile(ctx context.Context, filename string) *ApolloClient {
	c := &ApolloClient{
		namespaceList: NewNamespaceList(),
		backFile:      filename,
	}
	c.ctx, c.cancel = context.WithCancel(ctx)

	if obj, err := NewSerializationObjectWithFile(filename); err != nil {
		log.Fatalf("从文件加载配置失败 -> %s", err)
	} else {
		c.serverUrl = obj.ServerUrl
		c.config = ApolloConfig{AppId: obj.AppId, Cluster: obj.Cluster}
		for _, serialization := range obj.Serialization {
			item, err := NewNamespaceWithSerialization(c.ctx, serialization)
			if err != nil {
				log.Fatalf("命名空间初始化失败 -> %s : %s", serialization, err)
			}
			c.namespaceList.Store(serialization.Namespace, item)
		}
	}

	return c
}

// Watch 启动监听所有命名空间的变更事件
func (c *ApolloClient) Watch() {

	c.namespaceList.Range(func(key string, nsItem *namespaceItem) bool {
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
					c.save()
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
	c.namespaceList.Store(namespace, ns)
	return nil
}

// RemoveNamespace 移除指定的命名空间
func (c *ApolloClient) RemoveNamespace(namespace string) {
	c.namespaceList.Delete(namespace)
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

func (c *ApolloClient) SetBackFile(filename string) {
	c.backFile = filename
}

// 定时将配置信息保存的文件中
func (c *ApolloClient) save() {
	if c.backFile != "" {
		obj := SerializationObject{
			ServerUrl:     c.serverUrl,
			AppId:         c.config.AppId,
			Cluster:       c.config.Cluster,
			Serialization: make(map[string]*Serialization),
		}
		c.namespaceList.Range(func(key string, value *namespaceItem) bool {
			item, err := value.Dump()
			if err != nil {
				log.Printf("保存命名空间失败 -> %s", key)
			} else {
				obj.Serialization[key] = item
			}
			return true
		})
		if err := obj.Save(c.backFile); err != nil {
			log.Printf("保存备份文件失败 -> %s", err)
		} else {
			log.Printf("保存备份文件成功 -> %s", c.backFile)
		}
	}
}

// GetValue 获取默认命名空间指定键的配置值
// 如果配置不是 poroperties 类型，则返回所有配置信息
func (c *ApolloClient) GetValue(key string) (string, bool) {
	return c.namespaceList.GetValue("application", key)
}

// GetValueWithNamespace 获取指定命名空间指定键的配置值
// 如果配置不是 poroperties 类型，则返回所有配置信息
func (c *ApolloClient) GetValueWithNamespace(namespace string, key string) (string, bool) {
	return c.namespaceList.GetValue(namespace, key)
}
