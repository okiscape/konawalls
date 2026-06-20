package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type PostResponse []struct {
	Id       int    `json:"id"`
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
		fmt.Println("konawalls! - quick new wallpaper changer\nkonawalls get - get new wallpapers\n\nSources: https://github.com/okiscape/konawalls")
		return
	} else if args[0] == "get" {
		client := http.Client{Timeout: 5 * time.Second}

		url := "https://konachan.com/post.json?tags=" + strings.Join(cfg.Tags, "+") + "&limit=" + strconv.Itoa(cfg.Limit)

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

		if len(posts) == 0 {
			fmt.Printf("No arts was fetched with configured tags!\nTags: %s", strings.Join(cfg.Tags, " "))
			os.Exit(1)
		}

		randIndex := rand.N(len(posts))
		selectedWall := posts[randIndex]

		fmt.Printf("Downloading post: https://konachan.com/post/show/%d\n", selectedWall.Id)
		fmt.Printf("Tags: %s\n", selectedWall.Tags)

		fileName := filepath.Base(selectedWall.Jpeg_url)
		derr := downloadFile(selectedWall.Jpeg_url, cfg.SavePath)
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
