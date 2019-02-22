package goapollo

import (
	"encoding/gob"
	"encoding/json"
	"os"
	"path/filepath"
)

type Serialization struct {
	ServerUrl      string                      `json:"server_url"`
	AppId          string                      `json:"appId,omitempty"`
	Cluster        string                      `json:"cluster,omitempty"`
	Namespace      string                      `json:"namespace"`
	ReleaseKey     string                      `json:"release_key"`
	Configurations []byte                      `json:"configurations"`
	Notification   []*ApolloConfigNotification `json:"notification"`
}

// NewSerializationWithFile 从文件序列化配置信息
func NewSerializationWithFile(filename string) (s *Serialization, err error) {

	if f, err := os.Open(filename); err != nil {
		return nil, err
	} else {
		var obj Serialization
		err = gob.NewDecoder(f).Decode(&obj)
		return &obj, err
	}
}
func (s *Serialization) Save(filename string) error {
	_ = os.MkdirAll(filepath.Dir(filename), 0755)

	if f, err := os.Create(filename); err != nil {
		return err
	} else {
		return gob.NewEncoder(f).Encode(s)
	}
}

func (s *Serialization) String() string {
	b, _ := json.Marshal(s)

	return string(b)
}

type SerializationObject struct {
	ServerUrl     string                    `json:"server_url"`
	AppId         string                    `json:"appId,omitempty"`
	Cluster       string                    `json:"cluster,omitempty"`
	Serialization map[string]*Serialization `json:"serialization"`
}

func NewSerializationObjectWithFile(filename string) (s *SerializationObject, err error) {

	if f, err := os.Open(filename); err != nil {
		return nil, err
	} else {
		var obj SerializationObject
		err = gob.NewDecoder(f).Decode(&obj)
		return &obj, err
	}
}

// Save 将对象序列化到文件中
func (s *SerializationObject) Save(filename string) error {

	_ = os.MkdirAll(filepath.Dir(filename), 0755)

	if f, err := os.Create(filename); err != nil {
		return err
	} else {
		return gob.NewEncoder(f).Encode(s)
	}
}

func init() {
	gob.Register(Serialization{})
	gob.Register(SerializationObject{})
}
