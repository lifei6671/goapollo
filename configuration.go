package goapollo

import (
	"encoding/json"
	"sync"
)

type Configurations struct {
	m         sync.Map
	Namespace string
}

func NewConfiguration() *Configurations {
	return &Configurations{m: sync.Map{}}
}

func (c *Configurations) Load(k string) (string, bool) {
	if v, ok := c.m.Load(k); ok {
		s, ok := v.(string)
		return s, ok
	}
	return "", false
}

func (c *Configurations) Store(k, v string) {
	c.m.Store(k, v)
}

func (c *Configurations) Delete(k string) {
	c.m.Delete(k)
}

func (c *Configurations) Clear() {
	c.m = sync.Map{}
}

func (c *Configurations) String() string {
	b, err := c.Encode()
	if err != nil {
		return ""
	}
	return string(b)
}
func (c *Configurations) Range(f func(k, v string) bool) {
	c.m.Range(func(key, value interface{}) bool {
		k, _ := key.(string)
		v, _ := value.(string)
		return f(k, v)
	})
}

func (c *Configurations) Encode() ([]byte, error) {
	m := make(map[string]string)
	c.Range(func(key, value string) bool {
		m[key] = value
		return true
	})
	return json.Marshal(&m)
}

func (c *Configurations) Decode(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for k, v := range m {
		c.Store(k, v)
	}
	return nil
}
