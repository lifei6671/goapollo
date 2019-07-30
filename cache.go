package goapollo

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type namespaceCache struct {
	mux    *sync.RWMutex
	caches map[string]*sync.Map
}

func newNamespaceCache() *namespaceCache {

	return &namespaceCache{
		caches: map[string]*sync.Map{},
		mux:    &sync.RWMutex{},
	}
}

func (c *namespaceCache) load(name string) error {
	body, err := ioutil.ReadFile(name)
	if err != nil {
		logger.Printf("读取缓存文件失败 ->%s - %s", name, err)
		return err
	}
	var config Configuration
	err = json.Unmarshal(body, &config)
	if err != nil {
		return err
	}

	m := &sync.Map{}

	for k, v := range config.Configurations {
		m.Store(k, v)
	}
	c.mux.Lock()
	c.caches[config.NamespaceName] = m
	c.mux.Unlock()

	return nil
}

func (c *namespaceCache) save(dir string) error {
	if len(c.caches) <= 0 {
		return nil
	}
	c.mux.RLock()
	for name, m := range c.caches {
		config := Configuration{NamespaceName: name, Configurations: make(map[string]string)}
		path := filepath.Join(dir, name)

		m.Range(func(key, value interface{}) bool {
			config.Configurations[key.(string)] = value.(string)
			return true
		})
		if err := ioutil.WriteFile(path, []byte(config.String()), 0755); err != nil {
			return err
		}
	}
	c.mux.RUnlock()

	return nil
}

func (c *namespaceCache) dump(namespace, dir string) error {
	if len(c.caches) <= 0 {
		return nil
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	c.mux.RLock()
	defer c.mux.RUnlock()

	if m, ok := c.caches[namespace]; ok {
		config := Configuration{NamespaceName: namespace, Configurations: make(map[string]string)}
		path := filepath.Join(dir, namespace)

		m.Range(func(key, value interface{}) bool {
			config.Configurations[key.(string)] = value.(string)
			return true
		})
		if err := ioutil.WriteFile(path, []byte(config.String()), 0755); err != nil {
			return err
		}
		logger.Printf("备份文件已保存 -> %s - %s", namespace, path)
	}
	return nil
}

func (c *namespaceCache) store(result result) *ChangeEvent {
	event := ChangeEvent{Namespace: result.NamespaceName, Changes: make(map[string]*Change)}
	c.mux.Lock()
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
	c.mux.Unlock()

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
	if m, ok := c.caches[namespace]; ok {
		if val, ok := m.Load(key); ok {
			return val.(string), ok
		}
	}
	c.mux.RUnlock()
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
