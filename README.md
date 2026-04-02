
# LootManager

A small desktop tool for tracking Escape from Tarkov items you still need for hideout and quests.

Data comes from [tarkov.dev](https://tarkov.dev).

## Usage

- Search any item and see how many you still need
- Track hideout progress by setting station levels
- View quest item requirements grouped by trader
- Stash Scanner (beta) — paste screenshots of your junkboxes into Gemini chat using the provided prompt, paste the JSON output back into the app to update stash counts. API mode available with a Gemini key, but need still some testing, so disabled for now.

## Stack

Wails v2 · Go · Vue 3 · SQLite · tarkov.dev API

## Setup

1. Add `lootmanager.exe` in its own folder
2. Launch and click Sync Data — fetches item, hideout and quest data, downloads icons (~1-2 min)

## Notes

- Data is stored locally (`lootmanager.db`)
- Internet is only needed for syncing data (and optional scanner workflow)
- Re-sync after game patches to keep data up to date