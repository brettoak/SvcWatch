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
      label: {
        emphasis: {
          show: false
        }
      },
      itemStyle: {
        normal: {
          areaColor: themeStore.isDark ? '#323c48' : '#e2e8f0',
          borderColor: themeStore.isDark ? '#111' : '#fff'
        },
        emphasis: {
          areaColor: themeStore.isDark ? '#2a333d' : '#cbd5e1'
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
          return Math.max((val[2] / maxCount) * 20, 5)
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
            color: '#3b82f6',
            shadowBlur: 10,
            shadowColor: '#3b82f6'
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
      <div v-if="loading" class="absolute inset-0 flex items-center justify-center bg-bg-secondary/50 backdrop-blur-sm z-10">
        <span class="text-text-secondary font-bold">Loading map data...</span>
      </div>
      <v-chart class="w-full h-full" :option="option" :autoresize="true" />
    </div>
  </div>
</template>
