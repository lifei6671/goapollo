package goapollo

import (
	"errors"
	"fmt"
	"net"
	"net/url"
)

const (
	NamespaceFormatProperties = "properties"
	NamespaceFormatXml        = "xml"
	NamespaceFormatYml        = "yml"
	NamespaceFormatYaml       = "yaml"
	NamespaceFormatJson       = "json"
)

var (
	ErrConfigUnmodified = errors.New("Apollo configuration not changed. ")
)

const (
	defaultNotificationId int64 = -1
)

// Apollo 通知信息
type ApolloConfigNotification struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationId int64  `json:"notificationId"`
}

func NewApolloConfigNotification() *ApolloConfigNotification {
	return &ApolloConfigNotification{NotificationId: defaultNotificationId}
}

func (notify *ApolloConfigNotification) String() string {
	return fmt.Sprintf("ApolloConfigNotification{ namespaceName='%s', notificationId=%d }", notify.NamespaceName, notify.NotificationId)
}

type ApolloConfig struct {
	AppId   string `json:"appId,omitempty"`
	Cluster string `json:"cluster,omitempty"`
}

//关联的Namespace.
type ApolloResult struct {
	ReleaseKey string `json:"release_key"`
	//应用的appId
	AppId string `json:"appId"`
	// 集群名称，一般情况下传入 default 即可。
	// 如果希望配置按集群划分，可以参考集群独立配置说明做相关配置，然后在这里填入对应的集群名.
	Cluster string `json:"cluster"`
	// 命名空间，如果没有新建过Namespace的话，传入application即可。
	// 如果创建了Namespace，并且需要使用该Namespace的配置，则传入对应的Namespace名字。
	// 需要注意的是对于properties类型的namespace，只需要传入namespace的名字即可，如application。
	// 对于其它类型的namespace，需要传入namespace的名字加上后缀名，如datasources.json
	NamespaceName string `json:"namespaceName"`
	//配置信息.
	Configurations map[string]string `json:"configurations"`
}

//获取本地的ip地址.
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ip4 := toIP4(a); ip4 != nil {
			return ip4.String()
		}
	}
	return ""
}

//将网络地址转换为ip地址.
func toIP4(addr net.Addr) net.IP {
	if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
		return ipnet.IP.To4()
	}
	return nil
}

//通知地址.
func getApolloRemoteNotificationUrl(serverUrl, appId, cluster string, messages *ApolloNotificationMessages) string {
	notificationUrl := fmt.Sprintf("%snotifications/v2?appId=%s&cluster=%s&notifications=%s",
		serverUrl,
		url.QueryEscape(appId),
		url.QueryEscape(cluster),
		url.QueryEscape(messages.String()))

	return notificationUrl
}

// 获取URL
func getApolloRemoteConfigFromCacheUrl(serverUrl, appId, cluster, namespace string) string {
	if cluster == "" {
		cluster = "default"
	}
	getApolloRemoteConfigFromCacheUrl := fmt.Sprintf("%sconfigfiles/json/%s/%s/%s?ip=%s",
		serverUrl,
		url.QueryEscape(appId),
		url.QueryEscape(cluster),
		url.QueryEscape(namespace),
		getLocalIP())

	return getApolloRemoteConfigFromCacheUrl
}

// 生成通过不带缓存的Http接口从Apollo读取配置的链接
func getApolloRemoteConfigFromDbUrl(serverUrl, appId, cluster, namespace, releaseKey string) string {
	if cluster == "" {
		cluster = "default"
	}

	getApolloRemoteConfigFromDbUrl := fmt.Sprintf("%sconfigs/%s/%s/%s?releaseKey=%s&ip=%s",
		serverUrl,
		url.QueryEscape(appId),
		url.QueryEscape(cluster),
		url.QueryEscape(namespace),
		releaseKey,
		getLocalIP(),
	)

	return getApolloRemoteConfigFromDbUrl
}
