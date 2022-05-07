package fetch

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

func MakeDir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}

func UnmarshalFile(filePath string, data interface{}) error {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		return nil
	}
	if err := json.Unmarshal(bytes, data); err != nil {
		return err
	}
	return nil
}

func SaveFile(filePath, content string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.WriteString(out, content)
	if err != nil {
		return err
	}
	return nil
}
