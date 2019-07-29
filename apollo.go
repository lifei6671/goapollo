package goapollo

import (
	"os"
	"path/filepath"
	"time"
)

const (
	defaultConfigName = "app.properties"
	defaultNamespace  = "application"

	longPoolInterval      = time.Second * 2
	longPoolTimeout       = time.Second * 90
	queryTimeout          = time.Second * 2
	defaultNotificationId = -1
)

type Client struct {
	host, appId, cluster, cacheDir string
	caches                         *namespaceCache
	notification                   INotification
}

func New(host, appId, cluster string) *Client {
	return &Client{
		host:         host,
		appId:        appId,
		cluster:      cluster,
		cacheDir:     os.TempDir(),
		caches:       newNamespaceCache(),
		notification: newNotificationRepo(host, appId, cluster),
	}
}

func (c *Client) SetCacheDir(dir string) {
	c.cacheDir = dir
}

func (c *Client) preload(namespace string) {
	err := c.caches.load(filepath.Join(c.cacheDir, c.appId, namespace))
	if err != nil {

	}
}
