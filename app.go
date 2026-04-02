package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"lootmanager/backend/api"
	"lootmanager/backend/db"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

type SyncInfoDTO struct {
	ItemCount    int  `json:"itemCount"`
	NewItems     int  `json:"newItems"`
	NeedsIcons   bool `json:"needsIcons"`
	MissingIcons int  `json:"missingIcons"`
}

type ItemRequirementDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	IconPath         string `json:"iconPath"`
	HideoutTotalFIR  int    `json:"hideoutTotalFIR"`
	HideoutUsedFIR   int    `json:"hideoutUsedFIR"`
	HideoutTotalNorm int    `json:"hideoutTotalNorm"`
	HideoutUsedNorm  int    `json:"hideoutUsedNorm"`
	QuestTotalFIR    int    `json:"questTotalFIR"`
	QuestTotalNorm   int    `json:"questTotalNorm"`
	StashFIR         int    `json:"stashFIR"`
	StashNorm        int    `json:"stashNorm"`
}

type HideoutLevelDTO struct {
	ID        string `json:"id"`
	Level     int    `json:"level"`
	Completed bool   `json:"completed"`
}

type HideoutStationDTO struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Levels []HideoutLevelDTO `json:"levels"`
}

type QuestItemDTO struct {
	Name        string `json:"name"`
	Quantity    int    `json:"quantity"`
	FoundInRaid bool   `json:"foundInRaid"`
}

type QuestDTO struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Trader string         `json:"trader"`
	Items  []QuestItemDTO `json:"items"`
}

type StashItemDTO struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
	FIR      bool   `json:"fir"`
}

type StashResultDTO struct {
	Items      []StashItemDTO `json:"items"`
	Unmatched  []StashItemDTO `json:"unmatched"`
	ImageCount int            `json:"imageCount"`
}

type Config struct {
	GeminiAPIKey string `json:"gemini_api_key"`
}

const geminiUserPrompt = `Analyze all attached stash screenshots and return the unified JSON inventory array.`

func buildSystemInstruction(itemList string) string {
	return `You are a Tarkov stash analyzer. Analyze all provided images together as a single global inventory dataset.

The following is the complete list of valid item names:
` + itemList + `
Rules:
- Match every item you see in the screenshots to a name from the list above
- Use the exact name from the list in your response
- Deduplication & Global Sum: If the same item and fir status appears multiple times, sum quantities into one entry
- FIR Distinction: Keep FIR and non-FIR versions as separate entries
- Only include items that exist in the list above
- Output: Return ONLY a valid JSON array matching the schema exactly

JSON Schema: [{"name": "string", "quantity": number, "fir": boolean}]`
}

func buildManualPrompt(itemList string) string {
	return `Analyze all attached images together as a single, global inventory dataset. Return ONE unified JSON array.

The following is the complete list of valid item names:
` + itemList + `
Rules:
- Match every item you see to a name from the list above
- Use the exact name from the list in your response
- Deduplication & Global Sum: If the same item and fir status appears multiple times, sum quantities
- FIR Distinction: Keep FIR and non-FIR versions as separate entries
- Only include items that exist in the list above
- Output: Return ONLY the raw JSON array. No conversational text, no Markdown.

JSON Schema: [{"name": "string", "quantity": number, "fir": boolean}]`
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return &Config{}, nil
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid config.json: %w", err)
	}
	return &config, nil
}

func saveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("config.json", data, 0600)
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := db.Init("lootmanager.db"); err != nil {
		fmt.Println("DB init error:", err)
	}
	os.MkdirAll("screenshots", 0755)
}

func (a *App) GetSyncInfo() (*SyncInfoDTO, error) {
	beforeCount, err := db.CountItems()
	if err != nil {
		return nil, err
	}

	items, err := api.FetchItems()
	if err != nil {
		return nil, err
	}

	dbItems := make([]struct {
		ID       string
		Name     string
		IconLink string
	}, len(items))
	for i, item := range items {
		dbItems[i] = struct {
			ID       string
			Name     string
			IconLink string
		}{item.ID, item.Name, item.IconLink}
	}

	if err := db.UpsertItems(dbItems); err != nil {
		return nil, err
	}

	missing, err := db.GetItemsWithIconLinks()
	if err != nil {
		return nil, err
	}

	return &SyncInfoDTO{
		ItemCount:    len(items),
		NewItems:     len(items) - beforeCount,
		NeedsIcons:   len(missing) > 0,
		MissingIcons: len(missing),
	}, nil
}

func (a *App) DownloadIcons() error {
	items, err := db.GetItemsWithIconLinks()
	if err != nil {
		return err
	}

	downloadItems := make([]struct {
		ID       string
		IconLink string
	}, len(items))
	for i, item := range items {
		downloadItems[i] = struct {
			ID       string
			IconLink string
		}{item.ID, item.IconLink}
	}

	return api.DownloadIcons(downloadItems, "icons", func(done, total int) {
		runtime.EventsEmit(a.ctx, "icon-progress", map[string]int{
			"done":  done,
			"total": total,
		})
		if done == total {
			for _, item := range items {
				db.UpdateIconPath(item.ID, fmt.Sprintf("icons/%s.webp", item.ID))
			}
		}
	})
}

func (a *App) SyncHideoutAndQuests() error {
	stations, err := api.FetchHideoutStations()
	if err != nil {
		return fmt.Errorf("failed to fetch hideout: %w", err)
	}

	stationData := make([]struct{ ID, Name string }, len(stations))
	for i, s := range stations {
		stationData[i] = struct{ ID, Name string }{s.ID, s.Name}
	}
	if err := db.UpsertHideoutStations(stationData); err != nil {
		return err
	}

	for _, station := range stations {
		for _, level := range station.Levels {
			if err := db.UpsertHideoutLevel(level.ID, station.ID, level.Level); err != nil {
				return err
			}
			for _, req := range level.ItemRequirements {
				fir := false
				for _, attr := range req.Attributes {
					if attr.Type == "foundInRaid" && attr.Value == "true" {
						fir = true
					}
				}
				if err := db.UpsertHideoutRequirement(req.ID, level.ID, req.Item.ID, req.Quantity, fir); err != nil {
					return err
				}
			}
		}
	}

	quests, err := api.FetchQuests()
	if err != nil {
		return fmt.Errorf("failed to fetch quests: %w", err)
	}

	if err := db.ClearQuestRequirements(); err != nil {
		return fmt.Errorf("failed to clear quest requirements: %w", err)
	}

	for _, quest := range quests {
		if err := db.UpsertQuest(quest.ID, quest.Name, quest.Trader.Name); err != nil {
			return err
		}
		type questReqKey struct {
			ItemID string
			FIR    bool
		}
		// Tarkov API can emit duplicate item objectives for the same quest.
		// Collapse them by item+FIR and keep the max required quantity.
		deduped := make(map[questReqKey]int)
		for _, obj := range quest.Objectives {
			if obj.Item == nil {
				continue
			}
			key := questReqKey{ItemID: obj.Item.ID, FIR: obj.FoundInRaid}
			if obj.Count > deduped[key] {
				deduped[key] = obj.Count
			}
		}
		for key, quantity := range deduped {
			firFlag := 0
			if key.FIR {
				firFlag = 1
			}
			reqID := fmt.Sprintf("%s-%s-%d", quest.ID, key.ItemID, firFlag)
			if err := db.UpsertQuestRequirement(reqID, quest.ID, key.ItemID, quantity, key.FIR); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *App) GetHideoutStations() ([]HideoutStationDTO, error) {
	rows, err := db.GetHideoutStations()
	if err != nil {
		return nil, err
	}

	stations := make([]HideoutStationDTO, len(rows))
	for i, row := range rows {
		levels := make([]HideoutLevelDTO, len(row.Levels))
		for j, l := range row.Levels {
			levels[j] = HideoutLevelDTO{
				ID:        l.ID,
				Level:     l.Level,
				Completed: l.Completed,
			}
		}
		stations[i] = HideoutStationDTO{
			ID:     row.ID,
			Name:   row.Name,
			Levels: levels,
		}
	}
	return stations, nil
}

func (a *App) SetHideoutStationLevel(stationID string, level int) error {
	return db.SetHideoutStationLevel(stationID, level)
}

func (a *App) GetQuests() ([]QuestDTO, error) {
	rows, err := db.GetQuestsWithRequirements()
	if err != nil {
		return nil, err
	}

	quests := make([]QuestDTO, len(rows))
	for i, row := range rows {
		items := make([]QuestItemDTO, len(row.Items))
		for j, item := range row.Items {
			items[j] = QuestItemDTO{
				Name:        item.Name,
				Quantity:    item.Quantity,
				FoundInRaid: item.FoundInRaid,
			}
		}
		quests[i] = QuestDTO{
			ID:     row.ID,
			Name:   row.Name,
			Trader: row.Trader,
			Items:  items,
		}
	}
	return quests, nil
}

func (a *App) GetItemRequirements() ([]ItemRequirementDTO, error) {
	rows, err := db.GetItemRequirements()
	if err != nil {
		return nil, err
	}
	items := make([]ItemRequirementDTO, len(rows))
	for i, row := range rows {
		items[i] = ItemRequirementDTO{
			ID:               row.ID,
			Name:             row.Name,
			IconPath:         row.IconPath,
			HideoutTotalFIR:  row.HideoutTotalFIR,
			HideoutUsedFIR:   row.HideoutUsedFIR,
			HideoutTotalNorm: row.HideoutTotalNorm,
			HideoutUsedNorm:  row.HideoutUsedNorm,
			QuestTotalFIR:    row.QuestTotalFIR,
			QuestTotalNorm:   row.QuestTotalNorm,
			StashFIR:         row.StashFIR,
			StashNorm:        row.StashNorm,
		}
	}
	return items, nil
}

func (a *App) CountItems() (int, error) {
	return db.CountItems()
}

// --- Settings ---

func (a *App) GetGeminiKey() (string, error) {
	config, err := loadConfig()
	if err != nil {
		return "", err
	}
	return config.GeminiAPIKey, nil
}

func (a *App) SaveGeminiKey(key string) error {
	config, err := loadConfig()
	if err != nil {
		config = &Config{}
	}
	config.GeminiAPIKey = key
	return saveConfig(config)
}

func (a *App) HasGeminiKey() (bool, error) {
	config, err := loadConfig()
	if err != nil {
		return false, err
	}
	return config.GeminiAPIKey != "", nil
}

// --- Stash ---

func (a *App) GetManualPrompt() (string, error) {
	itemList, err := db.GetItemListForPrompt()
	if err != nil {
		return "", err
	}
	return buildManualPrompt(itemList), nil
}

func (a *App) ScanStash() (*StashResultDTO, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}
	if config.GeminiAPIKey == "" {
		return nil, fmt.Errorf("no Gemini API key configured")
	}

	entries, err := os.ReadDir("screenshots")
	if err != nil {
		return nil, fmt.Errorf("screenshots folder not found")
	}

	type imagePart struct {
		InlineData struct {
			MimeType string `json:"mime_type"`
			Data     string `json:"data"`
		} `json:"inline_data"`
	}

	var parts []interface{}
	imageCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			continue
		}

		data, err := os.ReadFile(filepath.Join("screenshots", entry.Name()))
		if err != nil {
			continue
		}

		mimeType := "image/png"
		if ext == ".jpg" || ext == ".jpeg" {
			mimeType = "image/jpeg"
		}

		var img imagePart
		img.InlineData.MimeType = mimeType
		img.InlineData.Data = base64.StdEncoding.EncodeToString(data)
		parts = append(parts, img)
		imageCount++
	}

	if imageCount == 0 {
		return nil, fmt.Errorf("no screenshots found in screenshots/ folder")
	}

	parts = append(parts, map[string]string{"text": geminiUserPrompt})

	itemList, err := db.GetItemListForPrompt()
	if err != nil {
		return nil, fmt.Errorf("failed to get item list: %w", err)
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"system_instruction": map[string]interface{}{
			"parts": []map[string]string{{"text": buildSystemInstruction(itemList)}},
		},
		"contents": []map[string]interface{}{
			{"role": "user", "parts": parts},
		},
		"generationConfig": map[string]string{
			"response_mime_type": "application/json",
		},
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s",
		config.GeminiAPIKey,
	)

	resp, err := http.Post(url, "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	responseText := apiResp.Candidates[0].Content.Parts[0].Text
	responseText = strings.ReplaceAll(responseText, "\u201c", "\"")
	responseText = strings.ReplaceAll(responseText, "\u201d", "\"")

	var scannedItems []StashItemDTO
	if err := json.Unmarshal([]byte(responseText), &scannedItems); err != nil {
		return nil, fmt.Errorf("failed to parse item list: %w", err)
	}

	return a.processStashItems(scannedItems, imageCount)
}

func (a *App) ParseManualJSON(jsonStr string) (*StashResultDTO, error) {
	jsonStr = strings.ReplaceAll(jsonStr, "\u201c", "\"")
	jsonStr = strings.ReplaceAll(jsonStr, "\u201d", "\"")
	var scannedItems []StashItemDTO
	if err := json.Unmarshal([]byte(jsonStr), &scannedItems); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return a.processStashItems(scannedItems, 0)
}

func (a *App) processStashItems(scannedItems []StashItemDTO, imageCount int) (*StashResultDTO, error) {
	nameMap, err := db.GetItemNameMap()
	if err != nil {
		return nil, err
	}

	matched := []StashItemDTO{}
	unmatched := []StashItemDTO{}

	for _, item := range scannedItems {
		id, ok := nameMap[strings.ToLower(item.Name)]
		if !ok {
			unmatched = append(unmatched, item)
			continue
		}
		item.ID = id
		matched = append(matched, item)
	}

	return &StashResultDTO{
		Items:      matched,
		Unmatched:  unmatched,
		ImageCount: imageCount,
	}, nil
}

func (a *App) UpdateStash(items []StashItemDTO) error {
	type stashEntry struct {
		quantity    int
		firQuantity int
	}
	grouped := make(map[string]*stashEntry)
	for _, item := range items {
		if _, ok := grouped[item.ID]; !ok {
			grouped[item.ID] = &stashEntry{}
		}
		if item.FIR {
			grouped[item.ID].firQuantity += item.Quantity
		} else {
			grouped[item.ID].quantity += item.Quantity
		}
	}

	dbItems := make([]struct {
		ItemID      string
		Quantity    int
		FIRQuantity int
	}, 0, len(grouped))

	for id, entry := range grouped {
		dbItems = append(dbItems, struct {
			ItemID      string
			Quantity    int
			FIRQuantity int
		}{id, entry.quantity, entry.firQuantity})
	}

	return db.UpdateStash(dbItems)
}

func (a *App) ClearScreenshots() error {
	entries, err := os.ReadDir("screenshots")
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			os.Remove(filepath.Join("screenshots", entry.Name()))
		}
	}
	return nil
}
