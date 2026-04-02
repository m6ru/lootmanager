<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { GetItemRequirements, CountItems, GetSyncInfo, DownloadIcons, SyncHideoutAndQuests, GetGeminiKey, SaveGeminiKey, HasGeminiKey, ScanStash, ParseManualJSON, UpdateStash, ClearScreenshots, GetManualPrompt } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const allItems = ref([])
const search = ref('')
const loading = ref(false)
const error = ref('')
const hasData = ref(null)
const syncing = ref(false)
const progress = ref({ done: 0, total: 0 })

// Settings
const geminiKey = ref('')
const keySaved = ref(false)
const hasKey = ref(false)
const showSettings = ref(false)

// Stash
const scanning = ref(false)
const scanError = ref('')
const scanResult = ref(null)
const manualJSON = ref('')
const showStashUpdate = ref(false)

const results = computed(() => {
  if (!search.value) return []
  return allItems.value.filter(i =>
    i.name.toLowerCase().includes(search.value.toLowerCase())
  )
})
const showSearchArea = computed(() => !showSettings.value && !showStashUpdate.value)

function need(total, used, stash = 0) {
  return Math.max(0, total - used - stash)
}

function fmt(needed, total) {
  if (total === 0) return '-'
  if (needed === 0) return `✓ (${total})`
  return `${needed}/${total}`
}

function fmt2(total) {
  if (total === 0) return '-'
  return total
}

onMounted(async () => {
  EventsOn('icon-progress', (data) => { progress.value = data })
  await load()
  hasKey.value = await HasGeminiKey()
  const key = await GetGeminiKey()
  geminiKey.value = key
})

onUnmounted(() => {
  EventsOff('icon-progress')
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const count = await CountItems()
    hasData.value = count > 0
    if (!hasData.value) return
    allItems.value = await GetItemRequirements()
  } catch (e) {
    error.value = 'Failed to load: ' + e
  } finally {
    loading.value = false
  }
}

async function firstTimeSync() {
  syncing.value = true
  error.value = ''
  try {
    const info = await GetSyncInfo()
    progress.value = { done: 0, total: info.itemCount }
    if (info.needsIcons) await DownloadIcons()
    await SyncHideoutAndQuests()
    await load()
  } catch (e) {
    error.value = 'Sync failed: ' + e
  } finally {
    syncing.value = false
    progress.value = { done: 0, total: 0 }
  }
}

async function saveKey() {
  try {
    await SaveGeminiKey(geminiKey.value)
    hasKey.value = geminiKey.value !== ''
    keySaved.value = true
    setTimeout(() => keySaved.value = false, 2000)
  } catch (e) {
    error.value = 'Failed to save key: ' + e
  }
}

async function scanStash() {
  scanning.value = true
  scanError.value = ''
  scanResult.value = null
  try {
    scanResult.value = await ScanStash()
  } catch (e) {
    scanError.value = 'Scan failed: ' + e
  } finally {
    scanning.value = false
  }
}

async function copyPrompt() {
  try {
    const prompt = await GetManualPrompt()
    await navigator.clipboard.writeText(prompt)
  } catch (e) {
    scanError.value = 'Failed to copy: ' + e
  }
}

async function parseManual() {
  scanError.value = ''
  scanResult.value = null
  try {
    scanResult.value = await ParseManualJSON(manualJSON.value)
  } catch (e) {
    scanError.value = 'Invalid JSON: ' + e
  }
}

async function confirmStash() {
  try {
    await UpdateStash(scanResult.value.items)
    await ClearScreenshots()
    scanResult.value = null
    manualJSON.value = ''
    await load()
  } catch (e) {
    scanError.value = 'Failed to update stash: ' + e
  }
}

function cancelScan() {
  scanResult.value = null
  manualJSON.value = ''
  scanError.value = ''
}

function toggleStashUpdate() {
  const next = !showStashUpdate.value
  showStashUpdate.value = next
  if (next) {
    showSettings.value = false
  }
}

function toggleSettings() {
  const next = !showSettings.value
  showSettings.value = next
  if (next) {
    showStashUpdate.value = false
  }
}
</script>

<template>
  <div style="padding: 20px">

    <!-- First launch -->
    <div v-if="hasData === false && !syncing" style="margin-top: 60px; text-align: center">
      <p style="color: #888; margin-bottom: 16px">No data yet. Download to get started.</p>
      <button @click="firstTimeSync" style="padding: 10px 24px; background: #2a2a4e; border: 1px solid #555; color: white; cursor: pointer; border-radius: 4px">
        Download
      </button>
      <p v-if="error" style="color: red; margin-top: 12px">{{ error }}</p>
    </div>

    <!-- Syncing progress -->
    <div v-else-if="syncing" style="margin-top: 60px; text-align: center">
      <p style="color: #888">Downloading data...</p>
      <div v-if="progress.total > 0" style="margin-top: 12px">
        <p>{{ progress.done }} / {{ progress.total }} items</p>
        <progress :value="progress.done" :max="progress.total" style="width: 300px" />
      </div>
    </div>

    <!-- Loading -->
    <div v-else-if="hasData === null || loading" style="margin-top: 60px; text-align: center">
      <p style="color: #888">Loading...</p>
    </div>

    <!-- Main content -->
    <div v-else>

      <!-- Top bar -->
      <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 24px; flex-wrap: wrap">
        <h2 style="margin: 0">Search</h2>
        <input
          v-model="search"
          placeholder="Search for an item..."
          style="padding: 6px 10px; width: 280px; background: #1a1a2e; border: 1px solid #333; color: white; border-radius: 4px; font-size: 1em"
          autofocus
        />
        <div style="margin-left: auto; display: flex; gap: 8px">
          <button
            @click="toggleStashUpdate"
            style="padding: 6px 12px; background: #1a2e1a; border: 1px solid #4a4; color: #4a4; cursor: pointer; border-radius: 4px; font-size: 0.85em"
          >
            {{ showStashUpdate ? 'Hide Stash Update' : '📦 Update Stash' }}
          </button>
          <button
              @click="toggleSettings"
              style="padding: 6px 12px; background: #2a2a4e; border: 1px solid #444; color: #aaa; cursor: pointer; border-radius: 4px; font-size: 0.85em"
            >
            {{ showSettings ? 'Hide Stash Scanner' : '⚙ Stash Scanner (Beta)' }}
  </button>
        </div>
      </div>

      <!-- Settings panel -->

        <div v-if="showSettings" style="background: #1a1a2e; border: 1px solid #2a2a4e; border-radius: 6px; padding: 16px; margin-bottom: 20px">
          <h3 style="margin: 0 0 8px 0; font-size: 0.95em; color: #aaa">Stash Scanner</h3>
          <p style="color: #f90; font-size: 0.85em; margin: 0">
            ⚠ Work in progress — needs more testing before it's ready.
          </p>
        </div>
 <!--      <div v-if="showSettings" style="background: #1a1a2e; border: 1px solid #2a2a4e; border-radius: 6px; padding: 16px; margin-bottom: 20px">
        <h3 style="margin: 0 0 12px 0; font-size: 0.95em; color: #aaa">Gemini API Key</h3>
        <p style="color: #666; font-size: 0.8em; margin-bottom: 10px">
          Get a free API key at <span style="color: #88a">aistudio.google.com</span>. Leave empty to use manual mode.
        </p>
        <div style="display: flex; gap: 8px; align-items: center">
          <input
            v-model="geminiKey"
            type="password"
            placeholder="Paste your Gemini API key..."
            style="flex: 1; padding: 6px 10px; background: #0f0f1a; border: 1px solid #333; color: white; border-radius: 4px"
          />
          <button
            @click="saveKey"
            style="padding: 6px 14px; background: #2a2a4e; border: 1px solid #555; color: white; cursor: pointer; border-radius: 4px"
          >
            {{ keySaved ? '✓ Saved' : 'Save' }}
          </button>
        </div>
      </div> -->

      <!-- Stash update panel -->
      <div v-if="showStashUpdate" style="background: #1a1a2e; border: 1px solid #2a2a4e; border-radius: 6px; padding: 16px; margin-bottom: 20px">
        <h3 style="margin: 0 0 12px 0; font-size: 0.95em; color: #aaa">Update Stash</h3>

        <!-- Scan result confirmation -->
        <div v-if="scanResult">
          <p style="color: #aaa; margin-bottom: 8px">
            Found <strong style="color: white">{{ scanResult.items.length }}</strong> matched items
            <span v-if="scanResult.unmatched.length > 0" style="color: #f90">
              ({{ scanResult.unmatched.length }} unrecognized)
            </span>
          </p>
          <div v-if="scanResult.unmatched.length > 0" style="margin-bottom: 12px">
            <p style="color: #666; font-size: 0.8em; margin-bottom: 4px">Unrecognized items (not saved):</p>
            <div style="max-height: 80px; overflow-y: auto">
              <span
                v-for="item in scanResult.unmatched"
                :key="item.tpl"
                style="display: inline-block; margin: 2px; padding: 2px 6px; background: #2a1a1a; border: 1px solid #633; border-radius: 3px; font-size: 0.75em; color: #f66"
              >
                {{ item.name }}
              </span>
            </div>
          </div>
          <div style="display: flex; gap: 8px">
            <button
              @click="confirmStash"
              style="padding: 6px 16px; background: #1a3a1a; border: 1px solid #4a4; color: #4a4; cursor: pointer; border-radius: 4px"
            >
              ✓ Confirm & Save
            </button>
            <button
              @click="cancelScan"
              style="padding: 6px 16px; background: #2a1a1a; border: 1px solid #633; color: #f66; cursor: pointer; border-radius: 4px"
            >
              Cancel
            </button>
          </div>
        </div>

        <!-- Scan controls -->
        <div v-else>
          <div v-if="hasKey" style="margin-bottom: 16px">
            <p style="color: #666; font-size: 0.8em; margin-bottom: 8px">
              Place junkbox screenshots in the <strong style="color: #aaa">screenshots/</strong> folder next to the app, then click scan.
            </p>
            <button
              @click="scanStash"
              :disabled="scanning"
              style="padding: 6px 16px; background: #1a2e1a; border: 1px solid #4a4; color: #4a4; cursor: pointer; border-radius: 4px"
            >
              {{ scanning ? 'Scanning...' : '🔍 Scan Stash' }}
            </button>
          </div>

          <div style="border-top: 1px solid #2a2a4e; padding-top: 16px" :style="hasKey ? '' : 'border-top: none; padding-top: 0'">
            <p style="color: #666; font-size: 0.8em; margin-bottom: 8px">
              {{ hasKey ? 'Or paste JSON manually:' : 'Copy the prompt, paste it with your screenshots into Gemini, then paste the result here:' }}
            </p>
            <button
              v-if="!hasKey"
              @click="copyPrompt"
              style="margin-bottom: 10px; padding: 6px 14px; background: #2a2a4e; border: 1px solid #555; color: #aaa; cursor: pointer; border-radius: 4px; font-size: 0.85em"
            >
              📋 Copy Prompt
            </button>
            <textarea
              v-model="manualJSON"
              placeholder='Paste JSON here: [{"name":"...","tpl":"...","quantity":1,"fir":true}]'
              style="width: 100%; height: 80px; background: #0f0f1a; border: 1px solid #333; color: white; border-radius: 4px; padding: 8px; font-size: 0.8em; resize: vertical; box-sizing: border-box"
            />
            <button
              @click="parseManual"
              :disabled="!manualJSON"
              style="margin-top: 8px; padding: 6px 14px; background: #2a2a4e; border: 1px solid #555; color: #aaa; cursor: pointer; border-radius: 4px"
            >
              Parse JSON
            </button>
          </div>

          <p v-if="scanError" style="color: red; margin-top: 8px">{{ scanError }}</p>
        </div>
      </div>

      <p v-if="error" style="color: red">{{ error }}</p>

      <div v-if="showSearchArea">
        <!-- Empty search state -->
        <p v-if="!search" style="color: #555">
          Search for an item to see how many are needed across hideout and quests.
        </p>

        <!-- No results -->
        <p v-else-if="results.length === 0" style="color: #555">
          No items found matching "{{ search }}"
        </p>

        <!-- Results table -->
        <table v-else style="width: 100%; border-collapse: collapse; font-size: 0.9em">
          <thead>
            <tr style="text-align: left; border-bottom: 1px solid #333; color: #666">
              <th style="padding: 8px">Item</th>
              <th style="padding: 8px">Hideout FIR</th>
              <th style="padding: 8px">Hideout</th>
              <th style="padding: 8px">Quest FIR</th>
              <th style="padding: 8px">Quest</th>
              <th style="padding: 8px">In Stash</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in results"
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
                {{  fmt2(item.questTotalFIR)  }}
              </td>
              <td style="padding: 8px; color: #aaa">
                {{ fmt2(item.questTotalNorm)  }}
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

  </div>
</template>