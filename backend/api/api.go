package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type Item struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IconLink string `json:"iconLink"`
}

type HideoutStation struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Levels []struct {
		ID               string `json:"id"`
		Level            int    `json:"level"`
		ItemRequirements []struct {
			ID         string `json:"id"`
			Item       struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"item"`
			Quantity   int `json:"quantity"`
			Attributes []struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"attributes"`
		} `json:"itemRequirements"`
	} `json:"levels"`
}

type Quest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Trader struct {
		Name string `json:"name"`
	} `json:"trader"`
	Objectives []struct {
		Item *struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"item"`
		Count       int  `json:"count"`
		FoundInRaid bool `json:"foundInRaid"`
	} `json:"objectives"`
}


func FetchItems() ([]Item, error) {
	query := `{"query": "{ items(lang: en) { id name iconLink } }"}`

	resp, err := http.Post(
		"https://api.tarkov.dev/graphql",
		"application/json",
		strings.NewReader(query),
	)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Items []Item `json:"items"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return result.Data.Items, nil
}

func DownloadIcons(
	items []struct {
		ID       string
		IconLink string
	},
	cacheDir string,
	onProgress func(done, total int),
) error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}

	total := len(items)
	var done atomic.Int32
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for _, item := range items {
		if item.IconLink == "" {
			done.Add(1)
			continue
		}

		wg.Add(1)
		go func(id, iconLink string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			ext := filepath.Ext(iconLink)
			if ext == "" {
				ext = ".webp"
			}
			destPath := filepath.Join(cacheDir, id+ext)

			resp, err := http.Get(iconLink)
			if err == nil {
				defer resp.Body.Close()
				f, err := os.Create(destPath)
				if err == nil {
					io.Copy(f, resp.Body)
					f.Close()
				}
			}

			current := int(done.Add(1))
			onProgress(current, total)
		}(item.ID, item.IconLink)
	}

	wg.Wait()
	return nil
}
func FetchHideoutStations() ([]HideoutStation, error) {
	query := `{"query": "{ hideoutStations { id name levels { id level itemRequirements { id item { id name } quantity attributes { type value } } } } }"}`

	resp, err := http.Post(
		"https://api.tarkov.dev/graphql",
		"application/json",
		strings.NewReader(query),
	)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			HideoutStations []HideoutStation `json:"hideoutStations"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return result.Data.HideoutStations, nil
}

func FetchQuests() ([]Quest, error) {
	query := `{"query": "{ tasks { id name trader { name } objectives { ... on TaskObjectiveItem { item { id name } count foundInRaid } } } }"}`

	resp, err := http.Post(
		"https://api.tarkov.dev/graphql",
		"application/json",
		strings.NewReader(query),
	)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Tasks []Quest `json:"tasks"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return result.Data.Tasks, nil
}