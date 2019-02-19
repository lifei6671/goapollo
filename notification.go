package goapollo

import (
	"encoding/json"
	"sync"
)

type ApolloNotificationMessages struct {
	*sync.Map
}

func NewApolloNotificationMessages() *ApolloNotificationMessages {
	return &ApolloNotificationMessages{&sync.Map{}}
}

func (m *ApolloNotificationMessages) String() string {
	details := make([]*ApolloConfigNotification, 0)
	m.Range(func(key string, value int64) bool {
		notify := &ApolloConfigNotification{
			NamespaceName:  key,
			NotificationId: value,
		}
		details = append(details, notify)
		return true
	})
	if len(details) == 0 {
		return ""
	}
	b, _ := json.Marshal(details)

	return string(b)
}

func (m *ApolloNotificationMessages) Notifications() []*ApolloConfigNotification {
	details := make([]*ApolloConfigNotification, 0)
	m.Range(func(key string, value int64) bool {
		notify := &ApolloConfigNotification{
			NamespaceName:  key,
			NotificationId: value,
		}
		details = append(details, notify)
		return true
	})
	return details
}

func (m *ApolloNotificationMessages) Put(key string, value int64) {
	m.Store(key, value)
}

func (m *ApolloNotificationMessages) Get(key string) int64 {
	if v, ok := m.Load(key); ok {
		return v.(int64)
	}
	return 0
}

func (m *ApolloNotificationMessages) Delete(key string) {
	m.Delete(key)
}

func (m *ApolloNotificationMessages) Clear() {
	*(m.Map) = sync.Map{}
}

func (m *ApolloNotificationMessages) Range(f func(namespace string, id int64) bool) {
	m.Map.Range(func(key, value interface{}) bool {
		k, _ := key.(string)
		v, _ := value.(int64)
		return f(k, v)
	})
}
