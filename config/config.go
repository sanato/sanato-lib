package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type config struct {
	Maintenance bool   `json:"maintenance"`
	ReadOnly    bool   `json:"readOnly"`
	RootDataDir string `json:"rootDataDir"`
}

type ConfigProvider struct {
	configFile string
}

func NewConfigProvider(configFile string) (*ConfigProvider, error) {
	return &ConfigProvider{configFile}, nil
}

func (cp *ConfigProvider) Parse() (*config, error) {
	fd, err := os.Open(cp.configFile)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	var config config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
