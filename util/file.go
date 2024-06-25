package util

import (
	"os"
)

// Read the file name form the specified path.
func GetFileName(path string) (string, error) {
	fileStat, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	return fileStat.Name(), nil
}

func ReadFile(path string) (*os.File, error) {
	return os.Open(path)
}
