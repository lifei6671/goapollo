package goapollo

import "sync"

type NamespaceList struct {
	m *sync.Map
}

func NewNamespaceList() *NamespaceList {
	return &NamespaceList{m: &sync.Map{}}
}
func (l *NamespaceList) Load(namespace string) (*namespaceItem, bool) {
	if v, ok := l.m.Load(namespace); ok {
		vv, ok := v.(*namespaceItem)
		return vv, ok
	}
	return nil, false
}

func (l *NamespaceList) Store(namespace string, item *namespaceItem) {
	l.m.Store(namespace, item)
}

func (l *NamespaceList) Delete(namespace string) {
	//先停止已存在的命名空间，再删除
	if item, ok := l.Load(namespace); ok {
		item.Stop()
		l.m.Delete(namespace)
	}
}

func (l *NamespaceList) Range(f func(key string, value *namespaceItem) bool) {
	l.m.Range(func(key, value interface{}) bool {
		k, _ := key.(string)
		v, _ := value.(*namespaceItem)
		return f(k, v)
	})
}

// GetValue 获取指定命名空间的配置值
func (l *NamespaceList) GetValue(namespace string, key string) (string, bool) {
	if v, ok := l.m.Load(namespace); ok {
		vv, _ := v.(*namespaceItem)
		//如果不是键值格式，则返回所有配置
		if vv.GetConfigType() != C_TYPE_POROPERTIES {
			key = "content"
		}
		return vv.GetValue(key)
	}
	return "", false
}
