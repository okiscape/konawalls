package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Post struct {
	ID       int
	Tags     string
	ImageURL string
	PostURL  string
}

type ProviderConfig struct {
	Name            string   `json:"name,omitempty"`
	BaseURL         string   `json:"baseUrl,omitempty"`
	APIPath         string   `json:"apiPath,omitempty"`
	ImageField      string   `json:"imageField,omitempty"`
	IDField         string   `json:"idField,omitempty"`
	TagsField       string   `json:"tagsField,omitempty"`
	Nesting         string   `json:"nesting,omitempty"`
	APIKey          string   `json:"apiKey,omitempty"`
	APIUser         string   `json:"apiUser,omitempty"`
	PostURLTemplate string   `json:"postUrlTemplate,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Limit           int      `json:"limit,omitempty"`
}

var providerPresets = map[string]ProviderConfig{
	"konachan": {
		Name:            "konachan",
		BaseURL:         "https://konachan.com",
		APIPath:         "/post.json",
		ImageField:      "jpeg_url",
		IDField:         "id",
		TagsField:       "tags",
		PostURLTemplate: "/post/show/{id}",
	},
	"yandere": {
		Name:            "yandere",
		BaseURL:         "https://yande.re",
		APIPath:         "/post.json",
		ImageField:      "jpeg_url",
		IDField:         "id",
		TagsField:       "tags",
		PostURLTemplate: "/post/show/{id}",
	},
	"danbooru": {
		Name:            "danbooru",
		BaseURL:         "https://danbooru.donmai.us",
		APIPath:         "/posts.json",
		ImageField:      "file_url",
		IDField:         "id",
		TagsField:       "tag_string",
		PostURLTemplate: "/posts/{id}",
	},
	"safebooru": {
		Name:            "safebooru",
		BaseURL:         "https://safebooru.org",
		APIPath:         "/index.php?page=dapi&s=post&q=index&json=1",
		ImageField:      "file_url",
		IDField:         "id",
		TagsField:       "tags",
		PostURLTemplate: "/index.php?page=post&s=view&id={id}",
	},
	"gelbooru": {
		Name:            "gelbooru",
		BaseURL:         "https://gelbooru.com",
		APIPath:         "/index.php?page=dapi&s=post&q=index&json=1",
		ImageField:      "file_url",
		IDField:         "id",
		TagsField:       "tags",
		PostURLTemplate: "/index.php?page=post&s=view&id={id}",
	},
}

func resolveProvider(raw json.RawMessage, overrides map[string]string) (*ProviderConfig, error) {
	var name string
	if err := json.Unmarshal(raw, &name); err == nil {
		preset, ok := providerPresets[name]
		if !ok {
			return nil, fmt.Errorf("unknown provider: %s", name)
		}
		pc := preset
		applyOverrides(&pc, overrides)
		if err := validateProvider(&pc); err != nil {
			return nil, err
		}
		return &pc, nil
	}

	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("invalid provider config: %w", err)
	}

	var pc ProviderConfig
	if n, _ := obj["name"].(string); n != "" {
		if preset, ok := providerPresets[n]; ok {
			pc = preset
		}
	}

	if v, _ := obj["baseUrl"].(string); v != "" {
		pc.BaseURL = v
	}
	if v, _ := obj["apiPath"].(string); v != "" {
		pc.APIPath = v
	}
	if v, _ := obj["imageField"].(string); v != "" {
		pc.ImageField = v
	}
	if v, _ := obj["idField"].(string); v != "" {
		pc.IDField = v
	}
	if v, _ := obj["tagsField"].(string); v != "" {
		pc.TagsField = v
	}
	if v, _ := obj["nesting"].(string); v != "" {
		pc.Nesting = v
	}
	if v, _ := obj["apiKey"].(string); v != "" {
		pc.APIKey = v
	}
	if v, _ := obj["apiUser"].(string); v != "" {
		pc.APIUser = v
	}
	if v, _ := obj["postUrlTemplate"].(string); v != "" {
		pc.PostURLTemplate = v
	}

	applyOverrides(&pc, overrides)
	if err := validateProvider(&pc); err != nil {
		return nil, err
	}
	return &pc, nil
}

func applyOverrides(pc *ProviderConfig, overrides map[string]string) {
	if overrides == nil {
		return
	}
	if v, ok := overrides["baseUrl"]; ok && v != "" {
		pc.BaseURL = v
	}
	if v, ok := overrides["apiPath"]; ok && v != "" {
		pc.APIPath = v
	}
	if v, ok := overrides["imageField"]; ok && v != "" {
		pc.ImageField = v
	}
	if v, ok := overrides["idField"]; ok && v != "" {
		pc.IDField = v
	}
	if v, ok := overrides["tagsField"]; ok && v != "" {
		pc.TagsField = v
	}
	if v, ok := overrides["nesting"]; ok && v != "" {
		pc.Nesting = v
	}
	if v, ok := overrides["apiKey"]; ok && v != "" {
		pc.APIKey = v
	}
	if v, ok := overrides["apiUser"]; ok && v != "" {
		pc.APIUser = v
	}
	if v, ok := overrides["postUrlTemplate"]; ok && v != "" {
		pc.PostURLTemplate = v
	}
}

func validateProvider(pc *ProviderConfig) error {
	if pc.BaseURL == "" {
		return fmt.Errorf("baseUrl is required")
	}
	if pc.APIPath == "" {
		return fmt.Errorf("apiPath is required")
	}
	if pc.ImageField == "" {
		return fmt.Errorf("imageField is required")
	}
	if pc.IDField == "" {
		pc.IDField = "id"
	}
	if pc.TagsField == "" {
		pc.TagsField = "tags"
	}
	return nil
}

func buildAPIURL(pc *ProviderConfig, tags []string, limit int) string {
	base, _ := url.Parse(pc.BaseURL)
	ref, _ := url.Parse(pc.APIPath)
	full := base.ResolveReference(ref)

	q := full.Query()
	q.Set("tags", strings.Join(tags, " "))
	q.Set("limit", strconv.Itoa(limit))

	if pc.APIKey != "" {
		q.Set("api_key", pc.APIKey)
		if pc.APIUser != "" {
			q.Set("login", pc.APIUser)
		}
	}

	full.RawQuery = q.Encode()
	return full.String()
}

func fetchPosts(pc *ProviderConfig, tags []string, limit int) ([]Post, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	url := buildAPIURL(pc, tags, limit)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "konawalls/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	var rawPosts []any

	if pc.Nesting != "" {
		var obj map[string]any
		if err := json.Unmarshal(body, &obj); err != nil {
			return nil, fmt.Errorf("json error: %w", err)
		}
		nested, ok := obj[pc.Nesting].([]any)
		if !ok {
			return nil, fmt.Errorf("field %q is not an array", pc.Nesting)
		}
		rawPosts = nested
	} else {
		if err := json.Unmarshal(body, &rawPosts); err != nil {
			return nil, fmt.Errorf("json error: %w", err)
		}
	}

	posts := make([]Post, 0, len(rawPosts))
	for _, rp := range rawPosts {
		p, ok := rp.(map[string]any)
		if !ok {
			continue
		}

		id := getIntField(p, pc.IDField)
		tags := getStringField(p, pc.TagsField)
		imageURL := getStringField(p, pc.ImageField)

		if imageURL == "" {
			continue
		}

		postURL := ""
		if pc.PostURLTemplate != "" {
			postURL = pc.BaseURL + strings.ReplaceAll(pc.PostURLTemplate, "{id}", strconv.Itoa(id))
		}

		posts = append(posts, Post{
			ID:       id,
			Tags:     tags,
			ImageURL: imageURL,
			PostURL:  postURL,
		})
	}

	return posts, nil
}

func resolveDefaultNames(raw json.RawMessage, providers map[string]ProviderConfig) []string {
	if raw == nil || len(raw) == 0 {
		names := make([]string, 0, len(providers))
		for name := range providers {
			names = append(names, name)
		}
		return names
	}

	var single string
	if json.Unmarshal(raw, &single) == nil {
		if _, ok := providers[single]; ok {
			return []string{single}
		}
		return nil
	}

	var list []string
	if json.Unmarshal(raw, &list) == nil {
		valid := make([]string, 0, len(list))
		for _, name := range list {
			if _, ok := providers[name]; ok {
				valid = append(valid, name)
			}
		}
		return valid
	}

	return nil
}

func resolveLabeledProvider(label string, pc *ProviderConfig, defaultTags []string, defaultLimit int, overrides map[string]string) (*ProviderConfig, []string, int, error) {
	presetName := pc.Name
	if presetName == "" {
		presetName = label
	}

	var resolved ProviderConfig
	if preset, ok := providerPresets[presetName]; ok {
		resolved = preset
	}

	if pc.BaseURL != "" {
		resolved.BaseURL = pc.BaseURL
	}
	if pc.APIPath != "" {
		resolved.APIPath = pc.APIPath
	}
	if pc.ImageField != "" {
		resolved.ImageField = pc.ImageField
	}
	if pc.IDField != "" {
		resolved.IDField = pc.IDField
	}
	if pc.TagsField != "" {
		resolved.TagsField = pc.TagsField
	}
	if pc.Nesting != "" {
		resolved.Nesting = pc.Nesting
	}
	if pc.APIKey != "" {
		resolved.APIKey = pc.APIKey
	}
	if pc.APIUser != "" {
		resolved.APIUser = pc.APIUser
	}
	if pc.PostURLTemplate != "" {
		resolved.PostURLTemplate = pc.PostURLTemplate
	}

	applyOverrides(&resolved, overrides)

	if err := validateProvider(&resolved); err != nil {
		return nil, nil, 0, err
	}

	tags := defaultTags
	if len(pc.Tags) > 0 {
		tags = pc.Tags
	}
	limit := defaultLimit
	if pc.Limit > 0 {
		limit = pc.Limit
	}

	return &resolved, tags, limit, nil
}

func getStringField(obj map[string]any, field string) string {
	v, ok := obj[field]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func getIntField(obj map[string]any, field string) int {
	v, ok := obj[field]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	default:
		return 0
	}
}
