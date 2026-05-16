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

const themeStore = useThemeStore()
const loading = ref(true)
const mapLoaded = ref(false)
const geoData = ref<GeoDistributionItem[]>([])

const fetchGeoData = async () => {
  loading.value = true
  try {
    const end = new Date()
    const start = new Date(end)
    start.setDate(start.getDate() - 30) // Default 30 days
    const res = await getGeoDistribution(start.toISOString(), end.toISOString())
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
  <div class="flex flex-col h-[calc(100vh-100px)] gap-6 animate-fade-in">
    <div class="flex justify-between items-end">
      <h1 class="text-3xl font-bold m-0 tracking-tight text-text-primary">IP Distribution Map</h1>
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
