package config

import (
	"encoding/json"
	"io"
	"os"
)

const configFileName string = ".gatorconfig.json"

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	filePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)

	config := Config{}
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	return config, err
}

func getConfigFilePath() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return dir + "/" + configFileName, err
}

func (cfg *Config) SetUser(user string) error {
	cfg.CurrentUserName = user
	write(*cfg)
	return nil
}

func write(cfg Config) error {
	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, os.ModePerm); err != nil {
		return err
	}

	return nil
}
