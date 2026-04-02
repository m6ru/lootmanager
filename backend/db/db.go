package db

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"strings"
)

var DB *sql.DB

func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if _, err := DB.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return fmt.Errorf("failed to set WAL mode: %w", err)
	}

	if err := initSchema(); err != nil {
		return fmt.Errorf("schema init failed: %w", err)
	}

	return nil
}

func initSchema() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id        TEXT PRIMARY KEY,
			name      TEXT NOT NULL,
			icon_link TEXT,
			icon_path TEXT
		);
		CREATE TABLE IF NOT EXISTS hideout_stations (
			id   TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS hideout_levels (
			id         TEXT PRIMARY KEY,
			station_id TEXT NOT NULL,
			level      INTEGER NOT NULL,
			completed  INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS hideout_requirements (
			id            TEXT PRIMARY KEY,
			level_id      TEXT NOT NULL,
			item_id       TEXT NOT NULL,
			quantity      INTEGER NOT NULL,
			found_in_raid INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS quests (
			id     TEXT PRIMARY KEY,
			name   TEXT NOT NULL,
			trader TEXT NOT NULL DEFAULT ''
		);
		CREATE TABLE IF NOT EXISTS quest_requirements (
			id            TEXT PRIMARY KEY,
			quest_id      TEXT NOT NULL,
			item_id       TEXT NOT NULL,
			quantity      INTEGER NOT NULL,
			found_in_raid INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS stash (
   			item_id      TEXT PRIMARY KEY,
    		quantity     INTEGER NOT NULL DEFAULT 0,
   			fir_quantity INTEGER NOT NULL DEFAULT 0
);
	`)
	return err
}

// --- Items ---

func UpsertItems(items []struct {
	ID       string
	Name     string
	IconLink string
}) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO items (id, name, icon_link)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name,
			icon_link=excluded.icon_link
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.Exec(item.ID, item.Name, item.IconLink); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func UpdateIconPath(id, path string) error {
	_, err := DB.Exec(`UPDATE items SET icon_path=? WHERE id=?`, path, id)
	return err
}

func GetItems() ([]struct {
	ID       string
	Name     string
	IconPath string
}, error) {
	rows, err := DB.Query(`SELECT id, name, COALESCE(icon_path, '') FROM items ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []struct {
		ID       string
		Name     string
		IconPath string
	}
	for rows.Next() {
		var item struct {
			ID       string
			Name     string
			IconPath string
		}
		if err := rows.Scan(&item.ID, &item.Name, &item.IconPath); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func GetItemsWithIconLinks() ([]struct {
	ID       string
	IconLink string
}, error) {
	rows, err := DB.Query(`
		SELECT id, COALESCE(icon_link, '')
		FROM items
		WHERE icon_link != '' AND (icon_path IS NULL OR icon_path = '')
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []struct {
		ID       string
		IconLink string
	}
	for rows.Next() {
		var item struct {
			ID       string
			IconLink string
		}
		if err := rows.Scan(&item.ID, &item.IconLink); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func CountItems() (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&count)
	return count, err
}

func GetItemListForPrompt() (string, error) {
	rows, err := DB.Query(`SELECT name FROM items WHERE id IN (SELECT item_id FROM hideout_requirements UNION SELECT item_id FROM quest_requirements) ORDER BY name`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return "", err
		}
		name = strings.ReplaceAll(name, "\"", "'")
		sb.WriteString(name)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func GetItemNameMap() (map[string]string, error) {
	rows, err := DB.Query(`SELECT id, LOWER(name) FROM items`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		name = strings.ReplaceAll(name, "\"", "'")
		m[name] = id
	}
	return m, nil
}

// --- Hideout ---

func UpsertHideoutStations(stations []struct {
	ID   string
	Name string
}) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, s := range stations {
		if _, err := tx.Exec(`
			INSERT INTO hideout_stations (id, name)
			VALUES (?, ?)
			ON CONFLICT(id) DO UPDATE SET name=excluded.name
		`, s.ID, s.Name); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func UpsertHideoutLevel(id, stationID string, level int) error {
	_, err := DB.Exec(`
		INSERT INTO hideout_levels (id, station_id, level)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO NOTHING
	`, id, stationID, level)
	return err
}

func UpsertHideoutRequirement(id, levelID, itemID string, quantity int, foundInRaid bool) error {
	fir := 0
	if foundInRaid {
		fir = 1
	}
	_, err := DB.Exec(`
		INSERT INTO hideout_requirements (id, level_id, item_id, quantity, found_in_raid)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			quantity=excluded.quantity,
			found_in_raid=excluded.found_in_raid
	`, id, levelID, itemID, quantity, fir)
	return err
}

func SetHideoutStationLevel(stationID string, level int) error {
	_, err := DB.Exec(`
		UPDATE hideout_levels
		SET completed = CASE WHEN level <= ? THEN 1 ELSE 0 END
		WHERE station_id = ?
	`, level, stationID)
	return err
}

func GetHideoutStations() ([]struct {
	ID     string
	Name   string
	Levels []struct {
		ID        string
		Level     int
		Completed bool
	}
}, error) {
	rows, err := DB.Query(`SELECT id, name FROM hideout_stations ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stations []struct {
		ID     string
		Name   string
		Levels []struct {
			ID        string
			Level     int
			Completed bool
		}
	}

	for rows.Next() {
		var s struct {
			ID     string
			Name   string
			Levels []struct {
				ID        string
				Level     int
				Completed bool
			}
		}
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}

		levelRows, err := DB.Query(`
			SELECT id, level, completed
			FROM hideout_levels
			WHERE station_id=? ORDER BY level
		`, s.ID)
		if err != nil {
			return nil, err
		}

		for levelRows.Next() {
			var l struct {
				ID        string
				Level     int
				Completed bool
			}
			var completed int
			if err := levelRows.Scan(&l.ID, &l.Level, &completed); err != nil {
				levelRows.Close()
				return nil, err
			}
			l.Completed = completed == 1
			s.Levels = append(s.Levels, l)
		}
		levelRows.Close()
		stations = append(stations, s)
	}
	return stations, nil
}

// --- Quests ---

func UpsertQuest(id, name, trader string) error {
	_, err := DB.Exec(`
		INSERT INTO quests (id, name, trader)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name,
			trader=excluded.trader
	`, id, name, trader)
	return err
}

func ClearQuestRequirements() error {
	_, err := DB.Exec(`DELETE FROM quest_requirements`)
	return err
}

func UpsertQuestRequirement(id, questID, itemID string, quantity int, foundInRaid bool) error {
	fir := 0
	if foundInRaid {
		fir = 1
	}
	_, err := DB.Exec(`
		INSERT INTO quest_requirements (id, quest_id, item_id, quantity, found_in_raid)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			quantity=excluded.quantity,
			found_in_raid=excluded.found_in_raid
	`, id, questID, itemID, quantity, fir)
	return err
}

func GetQuestsWithRequirements() ([]struct {
	ID     string
	Name   string
	Trader string
	Items  []struct {
		Name        string
		Quantity    int
		FoundInRaid bool
	}
}, error) {
	questRows, err := DB.Query(`
		SELECT DISTINCT q.id, q.name, q.trader
		FROM quests q
		INNER JOIN quest_requirements qr ON qr.quest_id = q.id
		INNER JOIN items i ON i.id = qr.item_id
		ORDER BY q.trader, q.name
	`)
	if err != nil {
		return nil, err
	}
	defer questRows.Close()

	var quests []struct {
		ID     string
		Name   string
		Trader string
		Items  []struct {
			Name        string
			Quantity    int
			FoundInRaid bool
		}
	}

	for questRows.Next() {
		var q struct {
			ID     string
			Name   string
			Trader string
			Items  []struct {
				Name        string
				Quantity    int
				FoundInRaid bool
			}
		}
		if err := questRows.Scan(&q.ID, &q.Name, &q.Trader); err != nil {
			return nil, err
		}

		itemRows, err := DB.Query(`
			SELECT i.name, qr.quantity, qr.found_in_raid
			FROM quest_requirements qr
			INNER JOIN items i ON i.id = qr.item_id
			WHERE qr.quest_id = ?
			ORDER BY i.name
		`, q.ID)
		if err != nil {
			return nil, err
		}

		for itemRows.Next() {
			var item struct {
				Name        string
				Quantity    int
				FoundInRaid bool
			}
			var fir int
			if err := itemRows.Scan(&item.Name, &item.Quantity, &fir); err != nil {
				itemRows.Close()
				return nil, err
			}
			item.FoundInRaid = fir == 1
			q.Items = append(q.Items, item)
		}
		itemRows.Close()
		quests = append(quests, q)
	}
	return quests, nil
}

// --- Requirements ---

func GetItemRequirements() ([]struct {
	ID               string
	Name             string
	IconPath         string
	HideoutTotalFIR  int
	HideoutUsedFIR   int
	HideoutTotalNorm int
	HideoutUsedNorm  int
	QuestTotalFIR    int
	QuestTotalNorm   int
	StashFIR         int
	StashNorm        int
}, error) {
	rows, err := DB.Query(`
		WITH hideout_totals AS (
			SELECT
				hr.item_id AS item_id,
				COALESCE(SUM(CASE WHEN hr.found_in_raid = 1 THEN hr.quantity ELSE 0 END), 0) AS hideout_total_fir,
				COALESCE(SUM(CASE WHEN hr.found_in_raid = 1 AND hl.completed = 1 THEN hr.quantity ELSE 0 END), 0) AS hideout_used_fir,
				COALESCE(SUM(CASE WHEN hr.found_in_raid = 0 THEN hr.quantity ELSE 0 END), 0) AS hideout_total_norm,
				COALESCE(SUM(CASE WHEN hr.found_in_raid = 0 AND hl.completed = 1 THEN hr.quantity ELSE 0 END), 0) AS hideout_used_norm
			FROM hideout_requirements hr
			LEFT JOIN hideout_levels hl ON hl.id = hr.level_id
			GROUP BY hr.item_id
		),
		quest_totals AS (
			SELECT
				qr.item_id AS item_id,
				COALESCE(SUM(CASE WHEN qr.found_in_raid = 1 THEN qr.quantity ELSE 0 END), 0) AS quest_total_fir,
				COALESCE(SUM(CASE WHEN qr.found_in_raid = 0 THEN qr.quantity ELSE 0 END), 0) AS quest_total_norm
			FROM quest_requirements qr
			GROUP BY qr.item_id
		)
		SELECT
			i.id,
			i.name,
			COALESCE(i.icon_path, ''),
			COALESCE(ht.hideout_total_fir, 0),
			COALESCE(ht.hideout_used_fir, 0),
			COALESCE(ht.hideout_total_norm, 0),
			COALESCE(ht.hideout_used_norm, 0),
			COALESCE(qt.quest_total_fir, 0),
			COALESCE(qt.quest_total_norm, 0),
			COALESCE(s.fir_quantity, 0),
			COALESCE(s.quantity, 0)
		FROM items i
		LEFT JOIN hideout_totals ht ON ht.item_id = i.id
		LEFT JOIN quest_totals qt ON qt.item_id = i.id
		LEFT JOIN stash s ON s.item_id = i.id
		ORDER BY i.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []struct {
		ID               string
		Name             string
		IconPath         string
		HideoutTotalFIR  int
		HideoutUsedFIR   int
		HideoutTotalNorm int
		HideoutUsedNorm  int
		QuestTotalFIR    int
		QuestTotalNorm   int
		StashFIR         int
		StashNorm        int
	}
	for rows.Next() {
		var item struct {
			ID               string
			Name             string
			IconPath         string
			HideoutTotalFIR  int
			HideoutUsedFIR   int
			HideoutTotalNorm int
			HideoutUsedNorm  int
			QuestTotalFIR    int
			QuestTotalNorm   int
			StashFIR         int
			StashNorm        int
		}
		if err := rows.Scan(
			&item.ID, &item.Name, &item.IconPath,
			&item.HideoutTotalFIR, &item.HideoutUsedFIR,
			&item.HideoutTotalNorm, &item.HideoutUsedNorm,
			&item.QuestTotalFIR, &item.QuestTotalNorm,
			&item.StashFIR, &item.StashNorm,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// --- Stash ---

func UpdateStash(items []struct {
	ItemID      string
	Quantity    int
	FIRQuantity int
}) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing stash
	if _, err := tx.Exec(`DELETE FROM stash`); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO stash (item_id, quantity, fir_quantity)
		VALUES (?, ?, ?)
		ON CONFLICT(item_id) DO UPDATE SET
			quantity=excluded.quantity,
			fir_quantity=excluded.fir_quantity
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.Exec(item.ItemID, item.Quantity, item.FIRQuantity); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func GetStash() (map[string]struct {
	Quantity    int
	FIRQuantity int
}, error) {
	rows, err := DB.Query(`SELECT item_id, quantity, fir_quantity FROM stash`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stash := make(map[string]struct {
		Quantity    int
		FIRQuantity int
	})
	for rows.Next() {
		var itemID string
		var qty, firQty int
		if err := rows.Scan(&itemID, &qty, &firQty); err != nil {
			return nil, err
		}
		stash[itemID] = struct {
			Quantity    int
			FIRQuantity int
		}{qty, firQty}
	}
	return stash, nil
}
