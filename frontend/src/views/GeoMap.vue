<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useThemeStore } from '@/stores/theme'
import { getGeoDistribution } from '@/services/api'
import type { GeoDistributionItem } from '@/services/api'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { EffectScatterChart, ScatterChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GeoComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import * as echarts from 'echarts/core'

use([
  CanvasRenderer,
  EffectScatterChart,
  ScatterChart,
  TitleComponent,
  TooltipComponent,
  GeoComponent,
])

const themeStore = useThemeStore()
const loading = ref(true)
const mapLoaded = ref(false)
const geoData = ref<GeoDistributionItem[]>([])

const timeFilter = ref('30d')
const timeOptions = [
  { label: '5m', value: '5m' },
  { label: '30m', value: '30m' },
  { label: '1h', value: '1h' },
  { label: '6h', value: '6h' },
  { label: '24h', value: '24h' },
  { label: '7d', value: '7d' },
  { label: '30d', value: '30d' },
  { label: 'Custom', value: 'custom' },
]

const customStartTime = ref('')
const customEndTime = ref('')
const sourceId = ref('')

const calculateTimeRange = () => {
  const end = new Date()
  let start = new Date(end)

  if (timeFilter.value === 'custom') {
    if (!customStartTime.value || !customEndTime.value) return null
    return {
      startStr: new Date(customStartTime.value).toISOString(),
      endStr: new Date(customEndTime.value).toISOString()
    }
  }

  switch (timeFilter.value) {
    case '5m': start.setMinutes(start.getMinutes() - 5); break
    case '30m': start.setMinutes(start.getMinutes() - 30); break
    case '1h': start.setHours(start.getHours() - 1); break
    case '6h': start.setHours(start.getHours() - 6); break
    case '24h': start.setHours(start.getHours() - 24); break
    case '7d': start.setDate(start.getDate() - 7); break
    case '30d': start.setDate(start.getDate() - 30); break
  }

  return {
    startStr: start.toISOString(),
    endStr: end.toISOString()
  }
}

const fetchGeoData = async () => {
  const range = calculateTimeRange()
  if (!range) return

  loading.value = true
  try {
    const res = await getGeoDistribution(range.startStr, range.endStr, sourceId.value)
    if (res.data && res.data.code === 200) {
      geoData.value = res.data.data || []
    }
  } catch (err) {
    console.error('Failed to load geo data', err)
  } finally {
    loading.value = false
  }
}

watch(timeFilter, (newVal) => {
  if (newVal !== 'custom') {
    fetchGeoData()
  }
})

onMounted(async () => {
  try {
    // Load world map json
    const res = await fetch('/world.json')
    const worldJson = await res.json()
    echarts.registerMap('world', worldJson)
    mapLoaded.value = true
  } catch (err) {
    console.error('Failed to load world.json', err)
  }

  const end = new Date()
  const start = new Date(end)
  start.setDate(start.getDate() - 30)
  
  const toLocalISO = (d: Date) => {
    const pad = (n: number) => n.toString().padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
  
  customEndTime.value = toLocalISO(end)
  customStartTime.value = toLocalISO(start)

  await fetchGeoData()
})

const option = computed(() => {
  const data = geoData.value.map(item => ({
    name: item.city || item.region || item.country,
    value: [item.longitude, item.latitude, item.count],
    ...item
  }))

  const maxCount = Math.max(...data.map(item => item.count), 1)

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        const d = params.data
        return `${d.name}<br/>Requests: ${d.count}`
      }
    },
    geo: {
      map: 'world',
      roam: true,
      zoom: 1.2,
      label: {
        emphasis: {
          show: false
        }
      },
      itemStyle: {
        normal: {
          areaColor: themeStore.isDark ? '#1e293b' : '#e0f2fe',
          borderColor: themeStore.isDark ? '#0f172a' : '#bae6fd',
          borderWidth: 1,
        },
        emphasis: {
          areaColor: themeStore.isDark ? '#334155' : '#bae6fd'
        }
      }
    },
    series: [
      {
        name: 'Traffic',
        type: 'effectScatter',
        coordinateSystem: 'geo',
        data: data,
        symbolSize: (val: any) => {
          return Math.max((val[2] / maxCount) * 25, 8)
        },
        showEffectOn: 'render',
        rippleEffect: {
          brushType: 'stroke'
        },
        hoverAnimation: true,
        label: {
          normal: {
            formatter: '{b}',
            position: 'right',
            show: false
          }
        },
        itemStyle: {
          normal: {
            color: themeStore.isDark ? '#06b6d4' : '#3b82f6', // Cyan in dark, Blue in light
            shadowBlur: 15,
            shadowColor: themeStore.isDark ? '#22d3ee' : '#60a5fa'
          }
        },
        zlevel: 1
      }
    ]
  }
})
</script>

<template>
  <div class="flex flex-col h-[calc(100vh-100px)] gap-6 animate-fade-in text-text-primary">
    <div class="flex flex-col gap-6 md:flex-row md:justify-between md:items-end">
      <h1 class="text-3xl font-bold m-0 tracking-tight">IP Distribution Map</h1>

      <div class="flex flex-col items-end gap-3">
        <div class="flex items-center gap-3 w-full md:w-auto">
          <input 
            type="text" 
            v-model="sourceId" 
            placeholder="Source ID (e.g. access.log)" 
            class="flex-1 md:w-48 bg-bg-secondary border border-border-color text-text-primary px-3 py-1.5 rounded-lg text-sm outline-none transition-all focus:border-primary-blue shadow-sm"
            @keyup.enter="fetchGeoData"
          />
          <button 
            class="bg-transparent border-none text-text-secondary cursor-pointer p-1.5 rounded-md flex items-center justify-center transition-all duration-200 hover:bg-bg-secondary hover:text-primary-blue disabled:opacity-50 disabled:cursor-not-allowed group" 
            @click="fetchGeoData" 
            :disabled="loading" 
            title="Refresh"
          >
            <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" class="group-active:rotate-180 transition-transform duration-300">
              <polyline points="23 4 23 10 17 10"></polyline>
              <polyline points="1 20 1 14 7 14"></polyline>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
            </svg>
          </button>
        </div>

        <div class="flex flex-col gap-2 items-end">
          <div class="flex bg-bg-secondary rounded-xl p-1 shadow-sm border border-border-color flex-wrap">
            <button 
              v-for="opt in timeOptions" 
              :key="opt.value"
              class="px-3.5 py-1.5 rounded-lg text-xs font-bold cursor-pointer transition-all duration-200"
              :class="timeFilter === opt.value ? 'bg-primary-blue text-white shadow-md shadow-primary-blue/30' : 'bg-transparent text-text-secondary hover:text-text-primary'"
              @click="timeFilter = opt.value"
            >
              {{ opt.label }}
            </button>
          </div>

          <div v-if="timeFilter === 'custom'" class="flex items-center gap-2 bg-bg-secondary p-1.5 rounded-xl shadow-sm border border-border-color animate-fade-in mt-1 flex-wrap">
            <input type="datetime-local" v-model="customStartTime" class="bg-transparent border border-border-color text-text-primary px-2.5 py-1.5 rounded-md text-sm outline-none transition-all focus:border-primary-blue" />
            <span class="text-text-secondary text-sm">to</span>
            <input type="datetime-local" v-model="customEndTime" class="bg-transparent border border-border-color text-text-primary px-2.5 py-1.5 rounded-md text-sm outline-none transition-all focus:border-primary-blue" />
            <button class="bg-primary-blue text-white border-none py-1.5 px-4 rounded-md text-sm font-bold cursor-pointer transition-all hover:brightness-110 hover:-translate-y-px disabled:opacity-60 disabled:cursor-not-allowed" @click="fetchGeoData" :disabled="loading">Search</button>
          </div>
        </div>
      </div>
    </div>
    
    <div class="flex-1 bg-bg-secondary rounded-2xl shadow-card border border-border-color p-4 relative overflow-hidden">
      <div v-if="loading || !mapLoaded" class="absolute inset-0 flex items-center justify-center bg-bg-secondary/50 backdrop-blur-sm z-10">
        <div class="flex flex-col items-center gap-3">
          <div class="w-8 h-8 border-4 border-primary-blue border-t-transparent rounded-full animate-spin"></div>
          <span class="text-text-secondary font-bold tracking-widest uppercase text-xs">Initializing Map...</span>
        </div>
      </div>
      <v-chart v-if="mapLoaded" class="w-full h-full" :option="option" :autoresize="true" />
    </div>
  </div>
</template>
