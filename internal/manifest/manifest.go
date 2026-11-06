package manifest

import (
	"os"
	"path/filepath"
)

type Manifest struct {
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	IsDir   bool        `json:"isdir,omitempty"`
	Entries []*Manifest `json:"entries,omitempty"`
}

func ToAbsDir(dir string) (string, error) {
	if !filepath.IsAbs(dir) {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	return dir, nil
}

func Build(path string) (*Manifest, error) {
	path, err := ToAbsDir(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	var manifest []*Manifest
	var name, size, isDir = "", int64(0), false
	if info.IsDir() {
		var entries []os.DirEntry
		name, isDir = info.Name(), info.IsDir()
		entries, err = os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, en := range entries {
			m, err := Build(filepath.Join(en.Name()))
			if err != nil {
				return nil, err
			}
			manifest = append(manifest, m)
		}
	} else {
		name, size = info.Name(), info.Size()
	}
	return &Manifest{
		Name:    name,
		Size:    size,
		IsDir:   isDir,
		Entries: manifest,
	}, nil
}
