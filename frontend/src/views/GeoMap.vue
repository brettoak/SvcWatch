<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
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

import TimeFilter from '@/components/TimeFilter.vue'

const themeStore = useThemeStore()
const loading = ref(true)
const mapLoaded = ref(false)
const geoData = ref<GeoDistributionItem[]>([])

const timeFilter = ref('30d')
const currentRange = ref<{ startStr: string; endStr: string } | null>(null)
const sourceId = ref('')

const displayedCities = computed(() => {
  return geoData.value
    .map(item => {
      const displayName = item.city || item.region || item.country || 'Unknown location'

      return {
        ...item,
        displayName,
        displayRegion: [item.region, item.country]
          .filter(part => part && part !== displayName)
          .join(', '),
      }
    })
    .sort((a, b) => b.count - a.count)
})

const displayedRequestCount = computed(() => {
  return displayedCities.value.reduce((total, item) => total + item.count, 0)
})

const onTimeRangeChange = (range: { startStr: string; endStr: string }) => {
  currentRange.value = range
  fetchGeoData()
}

const calculateTimeRange = () => {
  return currentRange.value
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

          <TimeFilter v-model="timeFilter" @change="onTimeRangeChange" />
      </div>
    </div>
    
    <div class="flex-1 min-h-0 grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_18rem] gap-4">
      <div class="min-h-[28rem] lg:min-h-0 bg-bg-secondary rounded-2xl shadow-card border border-border-color p-4 relative overflow-hidden">
        <div v-if="loading || !mapLoaded" class="absolute inset-0 flex items-center justify-center bg-bg-secondary/50 backdrop-blur-sm z-10">
          <div class="flex flex-col items-center gap-3">
            <div class="w-8 h-8 border-4 border-primary-blue border-t-transparent rounded-full animate-spin"></div>
            <span class="text-text-secondary font-bold tracking-widest uppercase text-xs">Initializing Map...</span>
          </div>
        </div>
        <v-chart v-if="mapLoaded" class="w-full h-full" :option="option" :autoresize="true" />
      </div>

      <aside class="min-h-0 bg-bg-secondary rounded-2xl shadow-card border border-border-color overflow-hidden flex flex-col">
        <div class="p-4 border-b border-border-color">
          <div class="flex items-center justify-between gap-3">
            <div>
              <p class="m-0 text-xs font-bold uppercase tracking-widest text-text-secondary">Displayed Cities</p>
              <p class="m-0 mt-1 text-2xl font-bold text-text-primary">{{ displayedCities.length }}</p>
            </div>
            <div class="text-right">
              <p class="m-0 text-xs text-text-secondary">Requests</p>
              <p class="m-0 mt-1 text-sm font-bold text-primary-blue">{{ displayedRequestCount.toLocaleString() }}</p>
            </div>
          </div>
        </div>

        <div v-if="displayedCities.length" class="flex-1 overflow-y-auto divide-y divide-border-color">
          <div
            v-for="city in displayedCities"
            :key="`${city.city}-${city.region}-${city.country}-${city.latitude}-${city.longitude}`"
            class="flex items-center justify-between gap-3 px-4 py-3 transition-colors hover:bg-bg-primary/40"
          >
            <div class="min-w-0">
              <p class="m-0 text-sm font-bold text-text-primary truncate">{{ city.displayName }}</p>
              <p v-if="city.displayRegion" class="m-0 mt-0.5 text-xs text-text-secondary truncate">
                {{ city.displayRegion }}
              </p>
            </div>
            <span class="shrink-0 rounded-full bg-primary-blue/10 px-2.5 py-1 text-xs font-bold text-primary-blue">
              {{ city.count.toLocaleString() }}
            </span>
          </div>
        </div>

        <div v-else class="flex-1 flex items-center justify-center p-6 text-center">
          <p class="m-0 text-sm text-text-secondary">
            {{ loading ? 'Loading cities...' : 'No city data is available for this range.' }}
          </p>
        </div>
      </aside>
    </div>
  </div>
</template>
