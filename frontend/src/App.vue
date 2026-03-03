<template>
  <div style="display: flex; height: 100vh">
    <nav style="width: 160px; background: #1a1a2e; padding: 20px; display: flex; flex-direction: column; gap: 12px; justify-content: space-between">
      <div style="display: flex; flex-direction: column; gap: 12px">
        <router-link to="/">Search</router-link>
        <router-link to="/hideout">Hideout</router-link>
        <router-link to="/quests">Quests</router-link>
      </div>
      <button @click="syncAll" :disabled="syncing" style="font-size: 0.75em; padding: 6px; background: #2a2a4e; border: 1px solid #444; color: #aaa; cursor: pointer; border-radius: 4px">
        {{ syncing ? 'Syncing...' : '↺ Sync Data' }}
      </button>
    </nav>
    <main style="flex: 1; overflow-y: auto">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { ref, provide } from 'vue'
import { SyncHideoutAndQuests } from '../wailsjs/go/main/App'

const syncing = ref(false)
const lastSynced = ref(0)
const lastUpdated = ref(0)

async function syncAll() {
  syncing.value = true
  try {
    await SyncHideoutAndQuests()
    lastSynced.value = Date.now()
  } catch (e) {
    console.error('Sync failed:', e)
  } finally {
    syncing.value = false
  }
}

function notifyUpdated() {
  lastUpdated.value = Date.now()
}

provide('lastSynced', lastSynced)
provide('lastUpdated', lastUpdated)
provide('notifyUpdated', notifyUpdated)
</script>

<style>
a { color: #a0a0b0; text-decoration: none; }
a:hover, a.router-link-active { color: white; }
body { margin: 0; background: #0f0f1a; color: white; font-family: sans-serif; }
</style>