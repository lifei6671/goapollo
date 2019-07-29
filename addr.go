package goapollo

import "fmt"

func getConfigfiles(host, appId, cluster, namespace, ip string) string {
	return fmt.Sprintf("%s/configfiles/json/%s/%s/%s?ip=%s",
		host,
		appId,
		cluster,
		namespace,
		ip,
	)
}

func getConfigfilesNoCache(host, appId, cluster, namespace, releaseKey, ip string) string {
	return fmt.Sprintf(" %s/configs/%s/%s/%s?releaseKey=%s&ip=%s",
		host,
		appId,
		cluster,
		namespace,
		releaseKey,
		ip,
	)
}

func getNotification(host, appId, cluster, namespace, notifications string) string {
	return fmt.Sprintf("{config_server_url}/notifications/v2?appId={appId}&cluster={clusterName}&notifications={notifications}")
}
