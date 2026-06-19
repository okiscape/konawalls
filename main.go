package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type PostResponse []struct {
	Tags     string `json:"tags"`
	Source   string `json:"source"`
	Jpeg_url string `json:"jpeg_url"`
}

func main() {
	args := os.Args[1:]
	cfg, lderr := loadConfig()
	if lderr != nil {
		fmt.Printf("Config error: %v\n", lderr)
		return
	}
	if len(args) == 0 {
		fmt.Println("konawalls!\nkonawalls get [opts | -h] - get new wallpapers")
		return
	} else if args[0] == "get" {
		client := http.Client{Timeout: 5 * time.Second}

		url := "https://konachan.com/post.json?tags=" + strings.Join(cfg.Tags, "+")

		resp, fetcherr := client.Get(url)
		if fetcherr != nil {
			fmt.Printf("Error happened: %v\n", fetcherr)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Server isnt ok: %s\n", resp.Status)
			return
		}

		var posts PostResponse
		var err = json.NewDecoder(resp.Body).Decode(&posts)
		if err != nil {
			fmt.Printf("Error while decoding JSON: %v\n", err)
			return
		}

		randIndex := rand.N(len(posts))
		fileName := filepath.Base(posts[randIndex].Jpeg_url)
		derr := downloadFile(posts[randIndex].Jpeg_url, cfg.SavePath)
		if derr != nil {
			fmt.Printf("Cannot download file %s: %v\n", fileName, derr)
		}

		fmt.Println("Saved gracefully!")
		if cfg.ExecuteAfter != nil {
			commandStr := *cfg.ExecuteAfter
			if commandStr != "" {
				cmd := exec.Command("bash", "-c", commandStr)
				if err := cmd.Run(); err != nil {
					fmt.Printf("Error while executing after-dowload: %v\n", err)
				}
			}
		}
	}
}
