package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Installed        bool   `json:"installed"`
	Maintenance      bool   `json:"maintenance"`
	Port             int    `json:"port"`
	ReadOnly         bool   `json:"readOnly"`
	RootDataDir      string `json:"rootDataDir"`
	RootTempDir      string `json:"rootTempDir"`
	TokenSecret      string `json:"tokenSecret"`
	TokenCipherSuite string `json:"tokenCipherSuite"`
	ServeWeb         string `json:"serveWeb"`
	WebDir           string `json:"webDir"`
	WebURL           string `json:"webUR"`
}

type ConfigProvider struct {
	configFile string
}

func NewConfigProvider(configFile string) (*ConfigProvider, error) {
	return &ConfigProvider{configFile}, nil
}

func (cp *ConfigProvider) Parse() (*Config, error) {
	fd, err := os.Open(cp.configFile)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
func (cp *ConfigProvider) CreateNewConfig(cfg *Config) error {
	fd, err := os.Create(cp.configFile)
	if err != nil {
		return err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = fd.Write(data)
	if err != nil {
		return err
	}
	return nil
}
