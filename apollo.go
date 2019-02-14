package goapollo

import (
	"context"
	"sync"
)

type ChangeHandler func()

type ApolloClient struct {
	//Apollo配置服务的地址.
	serverUrl string
	config    ApolloConfig
	namespace *sync.Map
	ctx       context.Context
	cancel    func()
}

// NewApolloClient
func NewApolloClient(ctx context.Context, serverUrl string, config ApolloConfig, namespace string) *ApolloClient {

	c := &ApolloClient{
		serverUrl: serverUrl,
		config:    config,
		namespace: &sync.Map{},
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	return c
}

func (c *ApolloClient) Watch() {

	c.namespace.Range(func(key, value interface{}) bool {
		nsItem, _ := value.(*namespaceItem)
		go func(item *namespaceItem) {
			ch := make(chan *ChangeEvent)
			nsItem.Watch(ch)
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

func (c *ApolloClient) Stop() {
	c.cancel()
}
