package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Tags         []string `json:"tags"`
	SavePath     string   `json:"savePath"`
	ExecuteAfter *string  `json:"executeAfter"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "konawalls", "config.json"), nil
}

func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("Cannot receive home directory: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("Config was not found, using default")

		// Создаем дефолтный конфиг
		defaultConfig := Config{
			Tags:     []string{"red"},
			SavePath: filepath.Join(filepath.Dir(configPath), "downloads"),
		}

		err := os.MkdirAll(filepath.Dir(configPath), 0755)
		if err != nil {
			return nil, fmt.Errorf("Cannot create config directory: %w", err)
		}

		file, err := os.Create(configPath)
		if err != nil {
			return nil, fmt.Errorf("Cannot create config file: %w", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(&defaultConfig); err != nil {
			return nil, fmt.Errorf("Cannot create config file: %w", err)
		}

		return &defaultConfig, nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("Error occured while JSON parsing: %w", err)
	}

	return &cfg, nil
}

func downloadFile(url string, savePath string) error {
	var finalFilePath string

	ext := filepath.Ext(savePath)

	if ext != "" {
		finalFilePath = savePath

		dir := filepath.Dir(savePath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create directory: %w", err)
		}
	} else {
		if err := os.MkdirAll(savePath, 0755); err != nil {
			return fmt.Errorf("cannot create directory: %w", err)
		}
		fileName := filepath.Base(url)
		finalFilePath = filepath.Join(savePath, fileName)
	}

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server isn't ok: %s", resp.Status)
	}

	out, err := os.Create(finalFilePath)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error while writing: %w", err)
	}

	return nil
}
