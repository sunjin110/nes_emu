package file

import (
	"fmt"
	"os"
)

func LoadNESFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed read file. path: %s, err: %w", path, err)
	}
	return data, nil
}
