package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c Config) SetUser(username string) error {
	c.CurrentUserName = username
	return write(c)
}

func Read() *Config {
	var cfg Config

	file, err := getFile()
	if err != nil {
		fmt.Printf("unable to open config file: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		fmt.Printf("sta failed: %s\n", err)
		os.Exit(1)
	}

	buf := make([]byte, stat.Size())
	_, err = file.Read(buf)
	if err != nil {
		fmt.Printf("unable to read config file: %s\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(buf, &cfg)
	if err != nil {
		fmt.Printf("unable to parse config: %s\n", err)
		os.Exit(1)
	}

	return &cfg
}

func getFile() (*os.File, error) {
	// dir, err := os.Getwd()
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, configFileName)
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func write(cfg Config) error {
	buf, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	file, err := getFile()
	if err != nil {
		return err
	}
	defer file.Close()

	if err = file.Truncate(0); err != nil {
		return err
	}
	if _, err = file.Seek(0, 0); err != nil {
		return err
	}

	_, err = file.Write(buf)
	return err
}
