package goapollo

import (
	"encoding/gob"
	"encoding/json"
)

type Configuration struct {
	NamespaceName  string            `json:"namespace_name"`
	Configurations map[string]string `json:"configurations"`
	ReleaseKey     string            `json:"release_key"`
}

func (c *Configuration) String() string {
	if c == nil {
		return ""
	}
	body, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(body)
}

type Config struct {
	Host       string   `json:"host"`
	AppId      string   `json:"app_id"`
	Cluster    string   `json:"cluster"`
	Namespaces []string `json:"namespaces"`
	IP         string   `json:"ip,omitempty"`
}

func init() {
	gob.Register(new(Config))
	gob.Register(new(Configuration))
}
