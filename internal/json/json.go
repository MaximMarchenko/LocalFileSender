package json

import (
	"encoding/json"

	"github.com/MaximMarchenko/LocalFileSender/internal/manifest"
)

func Unmarshal(data []byte) (*manifest.Manifest, error) {
	m := &manifest.Manifest{}
	err := json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func Marshal(m *manifest.Manifest) ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}
