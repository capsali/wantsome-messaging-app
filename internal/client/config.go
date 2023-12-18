package client

import (
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"
)

var DefaultConfig = Config{
	Client: ClientConfig{
		Server:     "127.0.0.1",
		Port:       8080,
		WsEndpoint: "/ws",
	},
	Storage: StorageConfig{
		Backend: "sqlite3",
		Path:    "server.sql",
	},
}

type Config struct {
	Client  ClientConfig
	Storage StorageConfig
}

type ClientConfig struct {
	Server     string
	Port       int
	WsEndpoint string
}

type StorageConfig struct {
	Backend string
	Path    string
}

func LoadConfig(configFile string) (*Config, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	configDir := filepath.Join(homeDir, ".go")
	configPath := filepath.Join(configDir, configFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file error: %v", err)
	} else if err != nil {
		return nil, err
	}
	conf := DefaultConfig
	if _, err := toml.DecodeFile(configPath, &conf); err != nil {
		log.Fatalf("Deconding config file failed: %v", err)
	}
	return &conf, nil
}
