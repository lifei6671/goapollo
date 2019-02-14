package goapollo

import "sync"

type releaseMap struct {
	m *sync.Map
}

func NewReleaseMap() *releaseMap {
	return &releaseMap{m: &sync.Map{}}
}

func (r *releaseMap) Load(key string) (string, bool) {
	if r.m == nil {
		r.m = &sync.Map{}
	}
	if v, ok := r.m.Load(key); ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

func (r *releaseMap) Store(k, v string) {
	r.m.Store(k, v)
}

func (r *releaseMap) Delete(k string) {
	r.m.Delete(k)
}

func (r *releaseMap) Range(f func(k, v string) bool) {
	r.m.Range(func(key, value interface{}) bool {
		k, _ := key.(string)
		v, _ := value.(string)

		return f(k, v)
	})
}
