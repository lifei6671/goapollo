package goapollo

import (
	"context"
	"errors"
	"os"
	"strings"
)

var defaultClient *Client

func Run(ctx context.Context) error {
	host := os.Getenv("APOLLO_HOST")
	appId := os.Getenv("APOLLO_APP_ID")
	cluster := os.Getenv("APOLLO_CLUSTER")
	namespace := os.Getenv("APOLLO_NAMESPACE")

	if host == "" || appId == "" {
		return errors.New("配置不完整")
	}
	if cluster == "" {
		cluster = "default"
	}
	var namespaces []string
	if namespace == "" {
		namespaces = append(namespaces, defaultNamespace)
	} else {
		namespaces = strings.Split(namespace, ";")
	}
	defaultClient = New(host, appId, cluster)
	for _, s := range namespaces {
		defaultClient.AddNamespace(s)
	}

	if err := defaultClient.Run(ctx); err != nil {
		logger.Printf("启动 Apollo 客户端失败 ->%s", err)
		return err
	}
	return nil
}

func WatchUpdate() <-chan *ChangeEvent {
	if defaultClient != nil {
		return defaultClient.eventCh
	}
	return nil
}

func GetValueWithNamespace(namespace, key string) (val string, exist bool) {
	if defaultClient == nil {
		return "", false
	}
	return defaultClient.GetValueWithNamespace(namespace, key)
}

func GetContentWithNamespace(namespace string) (val string, exist bool) {
	if defaultClient == nil {
		return "", false
	}
	return defaultClient.GetContentWithNamespace(namespace)
}

func GetValue(key string) (val string, exist bool) {
	if defaultClient == nil {
		return "", false
	}
	return defaultClient.GetValue(key)
}
func AllKeys(namespace string) []string {
	if defaultClient == nil {
		return nil
	}
	return defaultClient.AllKeys(namespace)
}
func Close() {
	if defaultClient != nil {
		_ = defaultClient.Close()
	}
}
