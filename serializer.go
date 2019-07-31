package goapollo

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

type Serializer interface {
	Serialize(v *Configuration) ([]byte, error)
	Deserialize(body []byte, target *Configuration) error
}

type JsonSerializer struct{}

func NewJsonSerializer() *JsonSerializer {
	return &JsonSerializer{}
}
func (j *JsonSerializer) Serialize(v *Configuration) ([]byte, error) {
	return json.Marshal(v)
}

func (j *JsonSerializer) Deserialize(body []byte, target *Configuration) error {
	return json.Unmarshal(body, target)
}

type GobSerializer struct{}

func NewGobSerializer() *GobSerializer {
	return &GobSerializer{}
}
func (j *GobSerializer) Serialize(v *Configuration) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(v)

	return buf.Bytes(), err
}

func (j *GobSerializer) Deserialize(body []byte, target *Configuration) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(body))
	err := decoder.Decode(target)
	return err
}
