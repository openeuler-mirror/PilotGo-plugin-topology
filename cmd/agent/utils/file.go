package utils

import (
	"io"
	"os"
	"github.com/pkg/errors"
)

func FileReadString(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", errors.New(err.Error())
	}

	return string(content), nil
}

func FileReadBytes(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer f.Close()

	var content []byte
	readbuff := make([]byte, 1024*4)
	for {
		n, err := f.Read(readbuff)
		if err != nil {
			if err == io.EOF {
				if n != 0 {
					content = append(content, readbuff[:n]...)
				}
				break
			}
			return nil, errors.New(err.Error())
		}
		content = append(content, readbuff[:n]...)
	}

	return content, nil
}
