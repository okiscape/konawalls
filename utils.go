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
	Provider     json.RawMessage           `json:"defaultProvider,omitempty"`
	Providers    map[string]ProviderConfig `json:"providers,omitempty"`
	Default      json.RawMessage           `json:"default,omitempty"`
	Tags         []string                  `json:"tags,omitempty"`
	SavePath     string                    `json:"savePath"`
	Limit        int                       `json:"limit"`
	ExecuteAfter *string                   `json:"executeAfter"`
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

		defaultConfig := Config{
			Providers: map[string]ProviderConfig{
				"konachan": {
					Tags: []string{"konata_izumi", "s"},
				},
				"safebooru": {
					Tags: []string{"konata_izumi"},
				},
			},
			Default:  json.RawMessage(`["konachan", "safebooru"]`),
			Tags:     []string{"konata_izumi"},
			Limit:    100,
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

	if len(cfg.Providers) == 0 && len(cfg.Provider) == 0 {
		cfg.Provider = json.RawMessage(`"konachan"`)
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
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "konawalls/1.0")
	resp, err := client.Do(req)
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
