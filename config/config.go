package config

import (
	"encoding/json"
	"fmt"
	"os"
	"verivista/pt/interfaces"
)

var Config interfaces.Config

// GetConfig 获取配置信息
func GetConfig() error {
	configPath := os.Getenv("VERIVISTA_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.json"
	}
	jsonData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("[Config File Error or Not Exist]: %v", err)
	}

	var config interfaces.Config

	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		return fmt.Errorf("[Config Analyze Error]: %v", err)
	}
	Config = config
	return nil
}
