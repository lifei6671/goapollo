package goapollo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type namespaceCache struct {
	mux         *sync.RWMutex
	caches      map[string]*sync.Map
	serializer  *sync.Map
	saves       *sync.Map
	releaseRepo *sync.Map
}

func newNamespaceCache() *namespaceCache {

	return &namespaceCache{
		caches:      map[string]*sync.Map{},
		mux:         &sync.RWMutex{},
		serializer:  &sync.Map{},
		saves:       &sync.Map{},
		releaseRepo: &sync.Map{},
	}
}

func (c *namespaceCache) load(namespace, path string) error {
	c.saves.Store(namespace, path)

	body, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Printf("读取缓存文件失败 ->%s - %s", path, err)
		return err
	}
	var config Configuration
	var serializer Serializer

	serializer, ok := c.getSerializer(namespace)

	if !ok {
		return fmt.Errorf("未找到可用的序列化器->%s", namespace)
	}
	err = serializer.Deserialize(body, &config)
	if err != nil {
		logger.Printf("反序列化对象失败 -> [namespace=%s] - [error=%s]", namespace, err)
		return err
	}

	m := &sync.Map{}

	for k, v := range config.Configurations {
		m.Store(k, v)
	}
	c.mux.Lock()
	c.caches[namespace] = m
	c.mux.Unlock()

	return nil
}

func (c *namespaceCache) save() error {
	if len(c.caches) <= 0 {
		return nil
	}
	c.mux.RLock()
	for name, m := range c.caches {
		config := &Configuration{NamespaceName: name, Configurations: make(map[string]string)}

		path, ok := c.getSave(name)
		if !ok {
			continue
		}
		m.Range(func(key, value interface{}) bool {
			config.Configurations[key.(string)] = value.(string)
			return true
		})
		serializer, ok := c.getSerializer(name)

		if !ok {
			serializer = NewJsonSerializer()
		}
		body, err := serializer.Serialize(config)
		if err != nil {
			logger.Printf("序列化对象失败 -> [namespace=%s] - [error=%s]", name, err)
			return err
		}
		if err := ioutil.WriteFile(path, body, 0755); err != nil {
			return err
		}
	}
	c.mux.RUnlock()

	return nil
}

func (c *namespaceCache) dump(namespace string) error {
	if len(c.caches) <= 0 {
		return nil
	}
	dir, ok := c.getSave(namespace)
	if !ok {
		logger.Printf("备份目录不存在 -> %s", namespace)
		return nil
	}

	if _, err := os.Stat(filepath.Dir(dir)); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Printf("创建目录失败 -> %s - %s", dir, err)
			return err
		}
	}
	c.mux.RLock()
	defer c.mux.RUnlock()

	if m, ok := c.caches[namespace]; ok {
		config := &Configuration{NamespaceName: namespace, Configurations: make(map[string]string)}

		m.Range(func(key, value interface{}) bool {
			config.Configurations[key.(string)] = value.(string)
			return true
		})
		serializer, ok := c.getSerializer(namespace)

		if !ok {
			serializer = NewJsonSerializer()
		}
		body, err := serializer.Serialize(config)
		if err != nil {
			logger.Printf("序列化对象失败 -> [namespace=%s] - [error=%s]", namespace, err)
			return err
		}
		if err := ioutil.WriteFile(dir, body, 0755); err != nil {
			logger.Printf("保存文件失败->[namespace=%s] - [error=%s]", namespace, err)
			return err
		}
		logger.Printf("备份文件已保存 -> %s - %s - %+v", namespace, dir, serializer)
	}
	return nil
}

func (c *namespaceCache) store(result result) *ChangeEvent {
	event := ChangeEvent{Namespace: result.NamespaceName, Changes: make(map[string]*Change)}
	c.mux.Lock()
	defer c.mux.Unlock()
	m, ok := c.caches[result.NamespaceName]
	if !ok {
		m = &sync.Map{}
		c.caches[result.NamespaceName] = m
	} else {
		m.Range(func(key, value interface{}) bool {
			k, _ := key.(string)
			v, _ := value.(string)
			event.Changes[k] = &Change{OldValue: v, ChangeType: -1}
			return true
		})
	}

	for k, v := range result.Configurations {
		m.Store(k, v)
		if change, ok := event.Changes[k]; ok {
			if v == change.OldValue {
				delete(event.Changes, k)
			} else {
				change.NewValue = v
				change.ChangeType = EventModify
			}
		} else {
			event.Changes[k] = &Change{NewValue: v, ChangeType: EventAdd}
		}
	}

	for k := range event.Changes {
		if _, ok := result.Configurations[k]; !ok {
			change := event.Changes[k]
			change.ChangeType = EventDelete
		}
	}

	return &event
}

func (c *namespaceCache) get(namespace string, key string) (string, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	if m, ok := c.caches[namespace]; ok {
		if val, ok := m.Load(key); ok {
			return val.(string), ok
		}
	}
	return "", false
}

func (c *namespaceCache) keys(namespace string) []string {
	var keys []string
	c.mux.RLock()
	if m, ok := c.caches[namespace]; ok {
		m.Range(func(key, value interface{}) bool {
			keys = append(keys, key.(string))
			return true
		})
	}
	c.mux.RUnlock()

	return keys
}

func (c *namespaceCache) addSerializer(namespace string, serializer Serializer) {
	c.serializer.Store(namespace, serializer)
}

func (c *namespaceCache) getSerializer(namespace string) (Serializer, bool) {
	var serializer Serializer

	if v, ok := c.serializer.Load(namespace); ok {
		serializer = v.(Serializer)
	} else {
		return nil, false
	}
	return serializer, true
}

func (c *namespaceCache) getSave(namespace string) (string, bool) {
	var path string

	if v, ok := c.saves.Load(namespace); ok {
		path = v.(string)
	} else {
		return "", false
	}
	return path, true
}
