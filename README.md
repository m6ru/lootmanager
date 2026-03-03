LootManager
Tarkov item requirement tracker. Tracks what you still need for hideout and quests, with stash scanning (manually and via Gemini Vision (beta)).
Data sourced from tarkov.dev.

Setup

Place lootmanager.exe in its own folder
Launch and click Sync Data — fetches item, hideout and quest data and downloads icons (~1-2 min)

All data is stored locally. Internet only needed for sync.

Usage
Search — type any item name to see remaining hideout and quest requirements against your stash counts.
Hideout — set your current level per station, items list updates accordingly.
Quests — item requirements per quest, grouped by trader. Reference only.
Stash Scanner (beta) — paste screenshots of your junkboxes into Gemini chat using the provided prompt, paste the JSON output back into the app to update stash counts. API mode available with a Gemini key, but need still some testing, so disabled for now.
Re-sync after patches to get updated quest and hideout data.

Stack
Wails v2 · Vue 3 · SQLite · tarkov.dev API · Gemini Vision