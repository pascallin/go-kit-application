package Config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Consul string `yaml:"consul"`
}

func ReadConfig() Config {
	filename, _ := filepath.Abs("config/default.yml")
	yamlFile, err := ioutil.ReadFile(filename)
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}
	return config
}
