package util

import (
	"os"

	"gopkg.in/yaml.v2"
)

func LoadFromYaml(filePath string, v interface{}) error {
	yamlFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	if err := yaml.NewDecoder(yamlFile).Decode(v); err != nil {
		return err
	}
	return yamlFile.Close()
}

func SaveToYaml(filePath string, v interface{}) error {
	yamlFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	if err := yaml.NewEncoder(yamlFile).Encode(v); err != nil {
		return err
	}
	return yamlFile.Close()
}
