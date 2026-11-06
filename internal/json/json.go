package json

import (
	"encoding/json"

	"github.com/MaximMarchenko/LocalFileSender/internal/manifest"
)

func Unmarshal(data []byte, m *manifest.Manifest) error {
	err := json.Unmarshal(data, m)
	if err != nil {
		return err
	}
	return nil
}

func Marshal(m *manifest.Manifest) ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}
