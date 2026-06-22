package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("konawalls! - quick new wallpaper changer\nkonawalls get - get new wallpapers\n\nSources: https://github.com/okiscape/konawalls")
		return
	}

	switch args[0] {
	case "get":
		runGet(args[1:])
	default:
		fmt.Println("konawalls! - quick new wallpaper changer\nkonawalls get - get new wallpapers\n\nSources: https://github.com/okiscape/konawalls")
	}
}

func runGet(argv []string) {
	getFlags := flag.NewFlagSet("get", flag.ExitOnError)
	providerFlag := getFlags.String("provider", "", "Provider name (konachan, safebooru, ...)")
	providerURL := getFlags.String("provider-url", "", "Provider base URL override")
	apiKey := getFlags.String("api-key", "", "API key for providers that require auth")
	apiUser := getFlags.String("api-user", "", "API username for providers that require auth")
	tagsFlag := getFlags.String("tags", "", "Comma-separated tags override")
	limitFlag := getFlags.Int("limit", 0, "Limit override")
	savePath := getFlags.String("save-path", "", "Save path override")
	executeAfter := getFlags.String("execute-after", "", "Post-download command override")
	getFlags.Parse(argv)

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Config error: %v\n", err)
		return
	}

	cliOverrides := map[string]string{}
	if *providerURL != "" {
		cliOverrides["baseUrl"] = *providerURL
	}
	if *apiKey != "" {
		cliOverrides["apiKey"] = *apiKey
	}
	if *apiUser != "" {
		cliOverrides["apiUser"] = *apiUser
	}

	pc, tags, limit, err := resolveActiveProvider(cfg, *providerFlag, cliOverrides)
	if err != nil {
		fmt.Printf("Provider error: %v\n", err)
		return
	}

	if *tagsFlag != "" {
		tags = strings.Split(*tagsFlag, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}
	if *limitFlag > 0 {
		limit = *limitFlag
	}

	savePathVal := cfg.SavePath
	if *savePath != "" {
		savePathVal = *savePath
	}

	executeAfterVal := cfg.ExecuteAfter
	seen := map[string]bool{}
	getFlags.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	if seen["execute-after"] {
		val := *executeAfter
		if val == "" {
			executeAfterVal = nil
		} else {
			executeAfterVal = &val
		}
	}

	posts, err := fetchPosts(pc, tags, limit)
	if err != nil {
		fmt.Printf("Fetch error: %v\n", err)
		return
	}

	if len(posts) == 0 {
		fmt.Printf("No arts fetched with configured tags!\nTags: %s\n", strings.Join(tags, " "))
		os.Exit(1)
	}

	randIndex := rand.N(len(posts))
	selected := posts[randIndex]

	fmt.Printf("Downloading post: %s\n", selected.PostURL)
	fmt.Printf("Tags: %s\n", selected.Tags)

	fileName := filepath.Base(selected.ImageURL)
	derr := downloadFile(selected.ImageURL, savePathVal)
	if derr != nil {
		fmt.Printf("Cannot download file %s: %v\n", fileName, derr)
	}

	fmt.Println("Saved gracefully!")
	if executeAfterVal != nil && *executeAfterVal != "" {
		cmd := exec.Command("bash", "-c", *executeAfterVal)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error while executing after-download: %v\n", err)
		}
	}
}

func resolveActiveProvider(cfg *Config, providerFlag string, cliOverrides map[string]string) (*ProviderConfig, []string, int, error) {
	if len(cfg.Providers) > 0 {
		var names []string

		if providerFlag != "" {
			if _, ok := cfg.Providers[providerFlag]; ok {
				names = []string{providerFlag}
			} else {
				return nil, nil, 0, fmt.Errorf("provider %q not found in providers map", providerFlag)
			}
		} else if cfg.Default != nil && len(cfg.Default) > 0 {
			names = resolveDefaultNames(cfg.Default, cfg.Providers)
		} else {
			for name := range cfg.Providers {
				names = append(names, name)
			}
		}

		if len(names) == 0 {
			return nil, nil, 0, fmt.Errorf("no providers selected by default config")
		}

		name := names[rand.N(len(names))]
		providerCfg := cfg.Providers[name]
		return resolveLabeledProvider(name, &providerCfg, cfg.Tags, cfg.Limit, cliOverrides)
	}

	providerRaw := cfg.Provider
	if providerFlag != "" {
		b, _ := json.Marshal(providerFlag)
		providerRaw = b
	}

	pc, err := resolveProvider(providerRaw, cliOverrides)
	if err != nil {
		return nil, nil, 0, err
	}

	return pc, cfg.Tags, cfg.Limit, nil
}
