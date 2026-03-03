<script setup>
import { ref, onMounted, computed } from 'vue'
import { GetHideoutStations, SetHideoutStationLevel, GetItemRequirements } from '../../wailsjs/go/main/App'

const stations = ref([])
const items = ref([])
const loading = ref(false)
const error = ref('')
const showProgress = ref(false)

const hideoutItems = computed(() =>
  items.value.filter(i =>
    i.hideoutTotalFIR > 0 || i.hideoutTotalNorm > 0
  ).filter(i =>
    need(i.hideoutTotalFIR, i.hideoutUsedFIR, i.stashFIR) > 0 ||
    need(i.hideoutTotalNorm, i.hideoutUsedNorm, i.stashNorm) > 0
  )
)

function need(total, used, stash = 0) {
  return Math.max(0, total - used - stash)
}

function fmt(needed, total) {
  if (total === 0) return '-'
  if (needed === 0) return `✓ (${total})`
  return `${needed}/${total}`
}

function currentLevel(station) {
  const completed = station.levels.filter(l => l.completed)
  return completed.length > 0 ? Math.max(...completed.map(l => l.level)) : 0
}

onMounted(async () => {
  loading.value = true
  try {
    stations.value = await GetHideoutStations()
    items.value = await GetItemRequirements()
  } catch (e) {
    error.value = 'Failed to load: ' + e
  } finally {
    loading.value = false
  }
})

async function setLevel(station, level) {
  try {
    await SetHideoutStationLevel(station.id, level)
    stations.value = await GetHideoutStations()
    items.value = await GetItemRequirements()
  } catch (e) {
    error.value = 'Failed to update: ' + e
  }
}
</script>

<template>
  <div style="padding: 20px">
    <div style="display: flex; align-items: center; gap: 16px; margin-bottom: 20px">
      <h2 style="margin: 0">Hideout</h2>
      <button
        @click="showProgress = !showProgress"
        style="padding: 4px 12px; background: #2a2a4e; border: 1px solid #444; color: #aaa; cursor: pointer; border-radius: 4px; font-size: 0.85em"
      >
        {{ showProgress ? 'Hide Progress' : 'Manage Progress' }}
      </button>
    </div>

    <p v-if="loading">Loading...</p>
    <p v-if="error" style="color: red">{{ error }}</p>

    <!-- Progress grid -->
    <div v-if="showProgress" style="margin-bottom: 32px">
      <div style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px">
        <div
          v-for="station in stations"
          :key="station.id"
          style="background: #1a1a2e; border: 1px solid #2a2a4e; border-radius: 6px; padding: 12px"
        >
          <div style="font-weight: bold; margin-bottom: 8px">{{ station.name }}</div>
          <div style="display: flex; gap: 6px; flex-wrap: wrap">
            <button
              v-for="lvl in [0, ...station.levels.map(l => l.level)]"
              :key="lvl"
              @click="setLevel(station, lvl)"
              :style="{
                padding: '4px 10px',
                borderRadius: '4px',
                border: '1px solid #444',
                cursor: 'pointer',
                background: currentLevel(station) === lvl ? '#4a9' : '#0f0f1a',
                color: currentLevel(station) === lvl ? 'white' : '#888',
                fontWeight: currentLevel(station) === lvl ? 'bold' : 'normal'
              }"
            >
              {{ lvl === 0 ? 'None' : lvl }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Items needed -->
    <div v-if="!loading">
      <p v-if="hideoutItems.length === 0" style="color: #555">
        All hideout items collected!
      </p>
      <table v-else style="width: 100%; border-collapse: collapse; font-size: 0.9em">
        <thead>
          <tr style="text-align: left; border-bottom: 1px solid #333; color: #666">
            <th style="padding: 8px">Item</th>
            <th style="padding: 8px">FIR needed</th>
            <th style="padding: 8px">Needed</th>
            <th style="padding: 8px">In Stash</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in hideoutItems"
            :key="item.id"
            style="border-bottom: 1px solid #1a1a2e"
          >
            <td style="padding: 8px; display: flex; align-items: center; gap: 8px">
              <img v-if="item.iconPath" :src="item.iconPath" width="32" height="32" loading="lazy" style="image-rendering: pixelated" />
              <div v-else style="width: 32px; height: 32px; background: #1a1a2e; border-radius: 2px" />
              <span>{{ item.name }}</span>
            </td>
            <td style="padding: 8px" :style="need(item.hideoutTotalFIR, item.hideoutUsedFIR, item.stashFIR) > 0 ? 'color: #f90' : 'color: #FFF'">
              {{ fmt(need(item.hideoutTotalFIR, item.hideoutUsedFIR, item.stashFIR), item.hideoutTotalFIR) }}
            </td>
            <td style="padding: 8px" :style="need(item.hideoutTotalNorm, item.hideoutUsedNorm, item.stashNorm) > 0 ? 'color: #f90' : 'color: #FFF'">
              {{ fmt(need(item.hideoutTotalNorm, item.hideoutUsedNorm, item.stashNorm), item.hideoutTotalNorm) }}
            </td>
            <td style="padding: 8px; color: #aaa">
              <span v-if="item.stashFIR > 0 || item.stashNorm > 0">
                <span :style="item.stashFIR > 0 ? 'color: #f90' : 'color: #555'">{{ item.stashFIR }} FIR</span>
                /
                <span :style="item.stashNorm > 0 ? 'color: white' : 'color: #555'">{{ item.stashNorm }}</span>
              </span>
              <span v-else style="color: #333">—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

  </div>
</template>



//
AI PROMPT

"Analyze all attached images together as a single, global inventory dataset. You must return ONE unified JSON array. Follow these logic rules:

Deduplication & Global Sum: If an item with the same Template ID (tpl) and Found in Raid status (fir) appears in multiple images (or multiple times in one image), you must sum their quantities into a single entry.

Example: If Image 1 has 5 Bolts (FIR) and Image 2 has 5 Bolts (FIR), the output must be one entry: {"name": "Bolts", "tpl": "...", "quantity": 10, "fir": true}.

Official API Data: Use the Official Full Name and the 24-character hexadecimal tpl from the Tarkov API for every item.

UI Text Check: If an icon is visually ambiguous, prioritize the UI text label in the top-left of the item slot to determine the item type.

FIR Distinction: Only aggregate items if their fir status is identical. Keep FIR and non-FIR versions of the same item as separate entries.

Output Requirement: Return ONLY the raw JSON array. No conversational text, no Markdown code blocks.

JSON Schema:
[{"name": "string", "tpl": "string", "quantity": number, "fir": boolean}]"