<script setup>
import { ref, onMounted, computed } from 'vue'
import { GetQuests } from '../../wailsjs/go/main/App'

const quests = ref([])
const loading = ref(false)
const error = ref('')
const activeTrader = ref(null)

const traders = computed(() => {
  const names = [...new Set(quests.value.map(q => q.trader))].sort()
  return names
})

const filtered = computed(() => {
  if (!activeTrader.value) return []
  return quests.value.filter(q => q.trader === activeTrader.value)
})

onMounted(async () => {
  loading.value = true
  try {
    quests.value = await GetQuests()
    if (traders.value.length > 0) activeTrader.value = traders.value[0]
  } catch (e) {
    error.value = 'Failed to load quests: ' + e
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div style="padding: 20px">
    <h2 style="margin: 0; padding-bottom: 20px">Quests</h2>
    <p v-if="loading">Loading...</p>
    <p v-if="error" style="color: red">{{ error }}</p>

    <div v-if="!loading && quests.length">

      <!-- Trader buttons -->
      <div style="display: flex; flex-wrap: wrap; gap: 5px; margin-bottom: 24px">
        <button
          v-for="trader in traders"
          :key="trader"
          @click="activeTrader = trader"
          :style="{
            padding: '6px 20px',
            borderRadius: '4px',
            border: '1px solid #444',
            cursor: 'pointer',
            background: activeTrader === trader ? '#2a2a6e' : '#1a1a2e',
            color: activeTrader === trader ? 'white' : '#888',
            fontWeight: activeTrader === trader ? 'bold' : 'normal'
          }"
        >
          {{ trader || 'Unknown' }}
        </button>
      </div>

      <!-- Quest list for selected trader -->
      <div v-for="quest in filtered" :key="quest.id" style="margin-bottom: 20px; background: #1a1a2e; border: 1px solid #2a2a4e; border-radius: 6px; padding: 12px">
        <div style="font-weight: bold; margin-bottom: 8px">{{ quest.name }}</div>
        <table style="width: 100%; border-collapse: collapse; font-size: 0.85em">
         <thead>
              <tr style="color: #666; text-align: left; border-bottom: 1px solid #2a2a4e">
              <th style="padding: 4px 8px">Item</th> 
              <th style="padding: 4px 8px; width: 120px">Quantity</th>
              <th style="padding: 4px 8px; width: 80px">FIR</th>
              </tr>
            </thead>
          <tbody>
            <tr v-for="item in quest.items" :key="item.name" style="border-bottom: 1px solid #111">
              <td style="padding: 4px 8px">{{ item.name }}</td>
              <td style="padding: 4px 8px">{{ item.quantity }}</td>
              <td style="padding: 4px 8px">
                <span :style="item.foundInRaid ? 'color: #f90' : 'color: #555'">
                  {{ item.foundInRaid ? 'Yes' : 'No' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

    </div>
  </div>
</template>