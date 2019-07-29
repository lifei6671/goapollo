package goapollo

import (
	"encoding/json"
	"io/ioutil"
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

func (space *namespaceCache) load(name string) error {
	body, err := ioutil.ReadFile(name)
	if err != nil {
		logger.Errorf("读取缓存文件失败 ->%s - %s", name, err)
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
	space.mux.Lock()
	space.caches[config.NamespaceName] = m
	space.mux.Unlock()

	return nil
}

func (space *namespaceCache) save(dir string) error {
	if len(space.caches) <= 0 {
		return nil
	}
	space.mux.RLock()
	for name, m := range space.caches {
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
	space.mux.RUnlock()

	return nil
}
