<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getTimeSeriesStats } from '@/services/api'
import type { TimeSeriesResponse } from '@/services/api'
import TimeFilter from '@/components/TimeFilter.vue'

// Basic States
const timeFilter = ref('24h')
const currentRange = ref<{ startStr: string; endStr: string } | null>(null)
const loading = ref(false)
const errorMsg = ref('')
const lastUpdated = ref('')

// Metrics historical points
const qpsPoints = ref<any[]>([])
const errorRatePoints = ref<any[]>([])
const latencyPoints = ref<any[]>([])
const bandwidthPoints = ref<any[]>([])

// Formatting helpers
const formatDateStr = (date: Date) => {
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

// Hover states for each of the 4 charts independently
const hoveredIdxQps = ref<number | null>(null)
const hoveredIdxErr = ref<number | null>(null)
const hoveredIdxLat = ref<number | null>(null)
const hoveredIdxBw = ref<number | null>(null)

// Mouse tracking for tooltips
const mouseX = ref(0)
const mouseY = ref(0)

const handleMouseMove = (e: MouseEvent) => {
  mouseX.value = e.clientX
  mouseY.value = e.clientY
}

// Fetch all 4 metrics parallelly from backend
const fetchData = async () => {
  if (!currentRange.value) {
    errorMsg.value = 'Please select a valid time range'
    return
  }

  errorMsg.value = ''
  loading.value = true

  try {
    const start = currentRange.value.startStr
    const end = currentRange.value.endStr

    const [qpsResp, errorResp, latencyResp, bandwidthResp] = await Promise.all([
      getTimeSeriesStats('qps', start, end),
      getTimeSeriesStats('error_rate', start, end),
      getTimeSeriesStats('latency_p99', start, end),
      getTimeSeriesStats('bandwidth', start, end)
    ])

    if (qpsResp.data && qpsResp.data.code === 200) {
      qpsPoints.value = qpsResp.data.data.points || []
    }
    if (errorResp.data && errorResp.data.code === 200) {
      errorRatePoints.value = errorResp.data.data.points || []
    }
    if (latencyResp.data && latencyResp.data.code === 200) {
      latencyPoints.value = latencyResp.data.data.points || []
    }
    if (bandwidthResp.data && bandwidthResp.data.code === 200) {
      bandwidthPoints.value = bandwidthResp.data.data.points || []
    }

    lastUpdated.value = formatDateStr(new Date())
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || err.message || 'Failed to fetch historical trend data'
  } finally {
    loading.value = false
  }
}

const onTimeRangeChange = (range: { startStr: string; endStr: string }) => {
  currentRange.value = range
  fetchData()
}

// Dynamic calculations for each metric series (min, max, average, formatted SVG paths)
const generateMetricAnalytics = (points: any[], metricType: 'qps' | 'error_rate' | 'latency' | 'bandwidth') => {
  if (points.length === 0) return { min: 0, max: 0, avg: 0, chartPoints: [], linePath: '', areaPath: '' }

  const values = points.map(p => p.value)
  const total = values.reduce((sum, v) => sum + v, 0)
  const avg = total / points.length
  const max = Math.max(...values)
  const min = Math.min(...values)

  // Scale settings
  const width = 550
  const height = 140
  const padLeft = 20
  const padRight = 20
  const padTop = 15
  const padBottom = 15

  const chartW = width - padLeft - padRight
  const chartH = height - padTop - padBottom

  const valRange = max - min === 0 ? 1 : max - min

  const chartPoints = points.map((p, i) => {
    const x = padLeft + (i / (points.length - 1 || 1)) * chartW
    const y = height - padBottom - ((p.value - min) / valRange) * chartH
    return {
      x,
      y,
      val: p.value,
      ts: new Date(p.ts).toLocaleString(),
      timeOnly: new Date(p.ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }
  })

  // Bezier curve path builders
  let linePath = ''
  let areaPath = ''

  const firstPt = chartPoints[0]
  const lastPt = chartPoints[chartPoints.length - 1]

  if (firstPt && lastPt && chartPoints.length > 0) {
    // Start line path
    linePath = `M ${firstPt.x} ${firstPt.y}`
    for (let i = 0; i < chartPoints.length - 1; i++) {
      const p0 = chartPoints[i]
      const p1 = chartPoints[i + 1]
      if (!p0 || !p1) continue
      const cp1x = p0.x + (p1.x - p0.x) / 2
      const cp1y = p0.y
      const cp2x = p1.x - (p1.x - p0.x) / 2
      const cp2y = p1.y
      linePath += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${p1.x} ${p1.y}`
    }

    // Start area path
    areaPath = `M ${firstPt.x} ${height - padBottom}`
    areaPath += ` L ${firstPt.x} ${firstPt.y}`
    for (let i = 0; i < chartPoints.length - 1; i++) {
      const p0 = chartPoints[i]
      const p1 = chartPoints[i + 1]
      if (!p0 || !p1) continue
      const cp1x = p0.x + (p1.x - p0.x) / 2
      const cp1y = p0.y
      const cp2x = p1.x - (p1.x - p0.x) / 2
      const cp2y = p1.y
      areaPath += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${p1.x} ${p1.y}`
    }
    areaPath += ` L ${lastPt.x} ${height - padBottom} Z`
  }

  return { min, max, avg, chartPoints, linePath, areaPath }
}

const qpsAnalysis = computed(() => generateMetricAnalytics(qpsPoints.value, 'qps'))
const errorRateAnalysis = computed(() => generateMetricAnalytics(errorRatePoints.value, 'error_rate'))
const latencyAnalysis = computed(() => generateMetricAnalytics(latencyPoints.value, 'latency'))
const bandwidthAnalysis = computed(() => generateMetricAnalytics(bandwidthPoints.value, 'bandwidth'))

// Computed properties for currently hovered points to secure type safety in templates
const hoveredQpsPt = computed(() => {
  if (hoveredIdxQps.value === null) return null
  return qpsAnalysis.value.chartPoints[hoveredIdxQps.value] || null
})

const hoveredErrPt = computed(() => {
  if (hoveredIdxErr.value === null) return null
  return errorRateAnalysis.value.chartPoints[hoveredIdxErr.value] || null
})

const hoveredLatPt = computed(() => {
  if (hoveredIdxLat.value === null) return null
  return latencyAnalysis.value.chartPoints[hoveredIdxLat.value] || null
})

const hoveredBwPt = computed(() => {
  if (hoveredIdxBw.value === null) return null
  return bandwidthAnalysis.value.chartPoints[hoveredIdxBw.value] || null
})

// Safe scale labels computations to avoid strict undefined check in Vue templates
const qpsLabels = computed(() => {
  const pts = qpsAnalysis.value.chartPoints
  if (pts.length === 0) return { start: '', mid: '', end: '' }
  const start = pts[0]?.timeOnly || ''
  const mid = pts[Math.floor(pts.length / 2)]?.timeOnly || ''
  const end = pts[pts.length - 1]?.timeOnly || ''
  return { start, mid, end }
})

const errLabels = computed(() => {
  const pts = errorRateAnalysis.value.chartPoints
  if (pts.length === 0) return { start: '', mid: '', end: '' }
  const start = pts[0]?.timeOnly || ''
  const mid = pts[Math.floor(pts.length / 2)]?.timeOnly || ''
  const end = pts[pts.length - 1]?.timeOnly || ''
  return { start, mid, end }
})

const latLabels = computed(() => {
  const pts = latencyAnalysis.value.chartPoints
  if (pts.length === 0) return { start: '', mid: '', end: '' }
  const start = pts[0]?.timeOnly || ''
  const mid = pts[Math.floor(pts.length / 2)]?.timeOnly || ''
  const end = pts[pts.length - 1]?.timeOnly || ''
  return { start, mid, end }
})

const bwLabels = computed(() => {
  const pts = bandwidthAnalysis.value.chartPoints
  if (pts.length === 0) return { start: '', mid: '', end: '' }
  const start = pts[0]?.timeOnly || ''
  const mid = pts[Math.floor(pts.length / 2)]?.timeOnly || ''
  const end = pts[pts.length - 1]?.timeOnly || ''
  return { start, mid, end }
})

// Formatting human-readable unit values
const formatValue = (val: number, metricType: 'qps' | 'error_rate' | 'latency' | 'bandwidth') => {
  if (metricType === 'qps') {
    return `${val.toFixed(2)} req/s`
  }
  if (metricType === 'error_rate') {
    return `${val.toFixed(2)}%`
  }
  if (metricType === 'latency') {
    // Original value from db is in seconds, e.g. 0.052, convert to ms
    const ms = val * 1000
    if (ms >= 1000) {
      return `${(ms / 1000).toFixed(2)} s`
    }
    return `${ms.toFixed(0)} ms`
  }
  if (metricType === 'bandwidth') {
    // Original value in bytes/s
    if (val >= 1024 * 1024) {
      return `${(val / (1024 * 1024)).toFixed(2)} MB/s`
    }
    if (val >= 1024) {
      return `${(val / 1024).toFixed(2)} KB/s`
    }
    return `${val.toFixed(0)} B/s`
  }
  return val.toString()
}

// Bounding containers reference for window scroll calculations
const showScrollTop = ref(false)
let mainScrollContainer: HTMLElement | null = null

const handleScroll = () => {
  if (mainScrollContainer) {
    showScrollTop.value = mainScrollContainer.scrollTop > 300
  }
}

const scrollToTop = () => {
  if (mainScrollContainer) {
    mainScrollContainer.scrollTo({
      top: 0,
      behavior: 'smooth'
    })
  }
}

onMounted(() => {
  mainScrollContainer = document.querySelector('main')
  if (mainScrollContainer) {
    mainScrollContainer.addEventListener('scroll', handleScroll)
  }
})

onUnmounted(() => {
  if (mainScrollContainer) {
    mainScrollContainer.removeEventListener('scroll', handleScroll)
  }
})
</script>

<template>
  <div class="flex flex-col gap-8 py-4 animate-fade-in text-text-primary">
    <!-- Header Block -->
    <div class="flex flex-col gap-6 md:flex-row md:justify-between md:items-end">
      <div>
        <h1 class="text-3xl font-bold m-0 tracking-tight">System Performance & Trends</h1>
        <p class="text-sm text-text-secondary mt-1 max-w-xl">
          Analyze historical workloads, latency percentiles, request success states, and bandwidth speeds pushed over mapped Nginx log targets.
        </p>
      </div>

      <div class="flex flex-col items-end gap-3 shrink-0">
        <div class="text-sm text-text-secondary flex items-center gap-2 font-medium">
          Last updated: {{ lastUpdated || '-' }}
          <button 
            class="bg-transparent border-none text-text-secondary cursor-pointer p-1.5 rounded-md flex items-center justify-center transition-all duration-200 hover:bg-bg-secondary hover:text-primary-blue disabled:opacity-50 disabled:cursor-not-allowed group" 
            @click="fetchData" 
            :disabled="loading" 
            title="Refresh Data"
          >
            <svg 
              viewBox="0 0 24 24" 
              width="16" 
              height="16" 
              stroke="currentColor" 
              stroke-width="2.5" 
              fill="none" 
              stroke-linecap="round" 
              stroke-linejoin="round" 
              class="group-active:rotate-180 transition-transform duration-500"
              :class="{ 'animate-spin text-primary-blue': loading }"
            >
              <polyline points="23 4 23 10 17 10"></polyline>
              <polyline points="1 20 1 14 7 14"></polyline>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
            </svg>
          </button>
        </div>

        <TimeFilter v-model="timeFilter" @change="onTimeRangeChange" />
      </div>
    </div>

    <!-- Error Banner -->
    <div v-if="errorMsg" class="bg-red-500/10 text-red-500 px-6 py-4 rounded-xl border-l-4 border-red-500 font-semibold animate-slide-in backdrop-blur-md">
      {{ errorMsg }}
    </div>

    <!-- Primary Layout Grid -->
    <div class="grid grid-cols-1 xl:grid-cols-2 gap-8 items-stretch" @mousemove="handleMouseMove">
      
      <!-- 1. QPS / Throughput Card -->
      <div 
        class="bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 relative"
        :class="{ 'opacity-50 pointer-events-none': loading }"
      >
        <!-- Header -->
        <div class="flex justify-between items-start">
          <div>
            <h3 class="text-sm font-bold text-text-primary uppercase tracking-wider flex items-center gap-2">
              <span class="text-blue-500">📈</span> QPS & Workload Throughput
            </h3>
            <p class="text-xs text-text-secondary mt-0.5">Average request volume over time interval</p>
          </div>
          <!-- Stats Summary Box -->
          <div class="flex gap-4 bg-bg-primary/50 px-3 py-1.5 rounded-xl border border-border-color/30 text-[0.68rem] font-bold text-text-secondary uppercase">
            <div>Avg: <span class="text-blue-500 font-extrabold">{{ formatValue(qpsAnalysis.avg, 'qps') }}</span></div>
            <div class="w-px bg-border-color/30 self-stretch"></div>
            <div>Max: <span class="text-text-primary font-extrabold">{{ formatValue(qpsAnalysis.max, 'qps') }}</span></div>
          </div>
        </div>

        <!-- SVG Chart -->
        <div class="relative w-full h-[140px] mt-2 select-none">
          <svg v-if="qpsPoints.length" viewBox="0 0 550 140" class="w-full h-full overflow-visible" preserveAspectRatio="none">
            <!-- Grid Lines -->
            <g class="stroke-slate-200/50 dark:stroke-slate-800/40" stroke-width="1">
              <line x1="20" y1="15" x2="530" y2="15" />
              <line x1="20" y1="65" x2="530" y2="65" />
              <line x1="20" y1="115" x2="530" y2="115" />
              <line x1="20" y1="125" x2="530" y2="125" class="stroke-slate-200 dark:stroke-slate-700" stroke-width="1.5" />
            </g>

            <!-- Line Curve & Fill Area -->
            <path v-if="qpsAnalysis.areaPath" :d="qpsAnalysis.areaPath" fill="url(#qpsGrad)" />
            <path v-if="qpsAnalysis.linePath" :d="qpsAnalysis.linePath" fill="none" stroke="#3b82f6" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="drop-shadow-[0_2px_8px_rgba(59,130,246,0.3)]" />

            <!-- Invisible vertical panels for hover tracking -->
            <g v-if="qpsAnalysis.chartPoints.length">
              <rect
                v-for="(pt, idx) in qpsAnalysis.chartPoints"
                :key="idx"
                :x="pt.x - 9"
                :y="0"
                :width="18"
                :height="140"
                fill="transparent"
                class="cursor-pointer"
                @mouseenter="hoveredIdxQps = idx"
                @mouseleave="hoveredIdxQps = null"
              />
            </g>

            <!-- Dot & crosshair highlighted on hover -->
            <g v-if="hoveredQpsPt">
              <!-- Vertical tracker line -->
              <line 
                :x1="hoveredQpsPt.x" 
                :y1="10" 
                :x2="hoveredQpsPt.x" 
                :y2="125" 
                stroke="#3b82f6" 
                stroke-width="1.5" 
                stroke-dasharray="3 3"
                class="opacity-60" 
              />
              <circle
                :cx="hoveredQpsPt.x"
                :cy="hoveredQpsPt.y"
                r="5"
                fill="#3b82f6"
                stroke="#ffffff"
                stroke-width="2"
                class="shadow-sm dark:stroke-slate-900"
              />
            </g>

            <!-- Definitions -->
            <defs>
              <linearGradient id="qpsGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.25" />
                <stop offset="100%" stop-color="#3b82f6" stop-opacity="0.0" />
              </linearGradient>
            </defs>
          </svg>

          <!-- Floating custom tooltip -->
          <div 
            v-if="hoveredQpsPt" 
            class="fixed pointer-events-none z-[100] bg-slate-900/90 dark:bg-slate-950/90 text-white px-3.5 py-2.5 rounded-xl text-[0.7rem] shadow-xl backdrop-blur-md border border-white/10 flex flex-col gap-1 min-w-[160px] transition-opacity duration-150 animate-fade-in"
            :style="{ left: mouseX + 15 + 'px', top: mouseY + 15 + 'px' }"
          >
            <div class="font-bold text-white/50 text-[0.62rem] uppercase tracking-wider">{{ hoveredQpsPt.ts }}</div>
            <div class="flex items-center justify-between gap-4 mt-0.5">
              <span class="font-semibold text-white/90">Request Rate</span>
              <span class="font-black text-blue-400 text-xs">{{ formatValue(hoveredQpsPt.val, 'qps') }}</span>
            </div>
          </div>

          <div v-if="!qpsPoints.length" class="absolute inset-0 flex items-center justify-center text-text-secondary text-sm italic">
            No historical QPS data available
          </div>
        </div>

        <!-- Timeline X-Axis scale labels -->
        <div v-if="qpsPoints.length && qpsAnalysis.chartPoints.length" class="flex justify-between items-center text-[0.65rem] text-text-secondary font-bold uppercase mt-1 px-1">
          <span>{{ qpsLabels.start }}</span>
          <span>{{ qpsLabels.mid }}</span>
          <span>{{ qpsLabels.end }}</span>
        </div>
      </div>

      <!-- 2. Error Rate Card -->
      <div 
        class="bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 relative"
        :class="{ 'opacity-50 pointer-events-none': loading }"
      >
        <!-- Header -->
        <div class="flex justify-between items-start">
          <div>
            <h3 class="text-sm font-bold text-text-primary uppercase tracking-wider flex items-center gap-2">
              <span class="text-rose-500">⚠️</span> Error Rate Trend
            </h3>
            <p class="text-xs text-text-secondary mt-0.5">Percentage of failed requests (status &gt;= 400)</p>
          </div>
          <!-- Stats Summary Box -->
          <div class="flex gap-4 bg-bg-primary/50 px-3 py-1.5 rounded-xl border border-border-color/30 text-[0.68rem] font-bold text-text-secondary uppercase">
            <div>Avg: <span class="text-rose-500 font-extrabold">{{ formatValue(errorRateAnalysis.avg, 'error_rate') }}</span></div>
            <div class="w-px bg-border-color/30 self-stretch"></div>
            <div>Max: <span class="text-text-primary font-extrabold">{{ formatValue(errorRateAnalysis.max, 'error_rate') }}</span></div>
          </div>
        </div>

        <!-- SVG Chart -->
        <div class="relative w-full h-[140px] mt-2 select-none">
          <svg v-if="errorRatePoints.length" viewBox="0 0 550 140" class="w-full h-full overflow-visible" preserveAspectRatio="none">
            <!-- Grid Lines -->
            <g class="stroke-slate-200/50 dark:stroke-slate-800/40" stroke-width="1">
              <line x1="20" y1="15" x2="530" y2="15" />
              <line x1="20" y1="65" x2="530" y2="65" />
              <line x1="20" y1="115" x2="530" y2="115" />
              <line x1="20" y1="125" x2="530" y2="125" class="stroke-slate-200 dark:stroke-slate-700" stroke-width="1.5" />
            </g>

            <!-- Line Curve & Fill Area -->
            <path v-if="errorRateAnalysis.areaPath" :d="errorRateAnalysis.areaPath" fill="url(#errGrad)" />
            <path v-if="errorRateAnalysis.linePath" :d="errorRateAnalysis.linePath" fill="none" stroke="#f43f5e" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="drop-shadow-[0_2px_8px_rgba(244,63,94,0.3)]" />

            <!-- Invisible vertical panels for hover tracking -->
            <g v-if="errorRateAnalysis.chartPoints.length">
              <rect
                v-for="(pt, idx) in errorRateAnalysis.chartPoints"
                :key="idx"
                :x="pt.x - 9"
                :y="0"
                :width="18"
                :height="140"
                fill="transparent"
                class="cursor-pointer"
                @mouseenter="hoveredIdxErr = idx"
                @mouseleave="hoveredIdxErr = null"
              />
            </g>

            <!-- Dot & crosshair highlighted on hover -->
            <g v-if="hoveredErrPt">
              <!-- Vertical tracker line -->
              <line 
                :x1="hoveredErrPt.x" 
                :y1="10" 
                :x2="hoveredErrPt.x" 
                :y2="125" 
                stroke="#f43f5e" 
                stroke-width="1.5" 
                stroke-dasharray="3 3"
                class="opacity-60" 
              />
              <circle
                :cx="hoveredErrPt.x"
                :cy="hoveredErrPt.y"
                r="5"
                fill="#f43f5e"
                stroke="#ffffff"
                stroke-width="2"
                class="shadow-sm dark:stroke-slate-900"
              />
            </g>

            <!-- Definitions -->
            <defs>
              <linearGradient id="errGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#f43f5e" stop-opacity="0.22" />
                <stop offset="100%" stop-color="#f43f5e" stop-opacity="0.0" />
              </linearGradient>
            </defs>
          </svg>

          <!-- Floating custom tooltip -->
          <div 
            v-if="hoveredErrPt" 
            class="fixed pointer-events-none z-[100] bg-slate-900/90 dark:bg-slate-950/90 text-white px-3.5 py-2.5 rounded-xl text-[0.7rem] shadow-xl backdrop-blur-md border border-white/10 flex flex-col gap-1 min-w-[160px] transition-opacity duration-150 animate-fade-in"
            :style="{ left: mouseX + 15 + 'px', top: mouseY + 15 + 'px' }"
          >
            <div class="font-bold text-white/50 text-[0.62rem] uppercase tracking-wider">{{ hoveredErrPt.ts }}</div>
            <div class="flex items-center justify-between gap-4 mt-0.5">
              <span class="font-semibold text-white/90">Error Ratio</span>
              <span class="font-black text-rose-400 text-xs">{{ formatValue(hoveredErrPt.val, 'error_rate') }}</span>
            </div>
          </div>

          <div v-if="!errorRatePoints.length" class="absolute inset-0 flex items-center justify-center text-text-secondary text-sm italic">
            No historical Error Rate data available
          </div>
        </div>

        <!-- Timeline X-Axis scale labels -->
        <div v-if="errorRatePoints.length && errorRateAnalysis.chartPoints.length" class="flex justify-between items-center text-[0.65rem] text-text-secondary font-bold uppercase mt-1 px-1">
          <span>{{ errLabels.start }}</span>
          <span>{{ errLabels.mid }}</span>
          <span>{{ errLabels.end }}</span>
        </div>
      </div>

      <!-- 3. P99 Latency Card -->
      <div 
        class="bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 relative"
        :class="{ 'opacity-50 pointer-events-none': loading }"
      >
        <!-- Header -->
        <div class="flex justify-between items-start">
          <div>
            <h3 class="text-sm font-bold text-text-primary uppercase tracking-wider flex items-center gap-2">
              <span class="text-purple-500">⚡</span> P99 Latency (Response Time)
            </h3>
            <p class="text-xs text-text-secondary mt-0.5">99th percentile maximum query latency</p>
          </div>
          <!-- Stats Summary Box -->
          <div class="flex gap-4 bg-bg-primary/50 px-3 py-1.5 rounded-xl border border-border-color/30 text-[0.68rem] font-bold text-text-secondary uppercase">
            <div>Avg: <span class="text-purple-500 font-extrabold">{{ formatValue(latencyAnalysis.avg, 'latency') }}</span></div>
            <div class="w-px bg-border-color/30 self-stretch"></div>
            <div>Max: <span class="text-text-primary font-extrabold">{{ formatValue(latencyAnalysis.max, 'latency') }}</span></div>
          </div>
        </div>

        <!-- SVG Chart -->
        <div class="relative w-full h-[140px] mt-2 select-none">
          <svg v-if="latencyPoints.length" viewBox="0 0 550 140" class="w-full h-full overflow-visible" preserveAspectRatio="none">
            <!-- Grid Lines -->
            <g class="stroke-slate-200/50 dark:stroke-slate-800/40" stroke-width="1">
              <line x1="20" y1="15" x2="530" y2="15" />
              <line x1="20" y1="65" x2="530" y2="65" />
              <line x1="20" y1="115" x2="530" y2="115" />
              <line x1="20" y1="125" x2="530" y2="125" class="stroke-slate-200 dark:stroke-slate-700" stroke-width="1.5" />
            </g>

            <!-- Line Curve & Fill Area -->
            <path v-if="latencyAnalysis.areaPath" :d="latencyAnalysis.areaPath" fill="url(#latGrad)" />
            <path v-if="latencyAnalysis.linePath" :d="latencyAnalysis.linePath" fill="none" stroke="#a855f7" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="drop-shadow-[0_2px_8px_rgba(168,85,247,0.3)]" />

            <!-- Invisible vertical panels for hover tracking -->
            <g v-if="latencyAnalysis.chartPoints.length">
              <rect
                v-for="(pt, idx) in latencyAnalysis.chartPoints"
                :key="idx"
                :x="pt.x - 9"
                :y="0"
                :width="18"
                :height="140"
                fill="transparent"
                class="cursor-pointer"
                @mouseenter="hoveredIdxLat = idx"
                @mouseleave="hoveredIdxLat = null"
              />
            </g>

            <!-- Dot & crosshair highlighted on hover -->
            <g v-if="hoveredLatPt">
              <!-- Vertical tracker line -->
              <line 
                :x1="hoveredLatPt.x" 
                :y1="10" 
                :x2="hoveredLatPt.x" 
                :y2="125" 
                stroke="#a855f7" 
                stroke-width="1.5" 
                stroke-dasharray="3 3"
                class="opacity-60" 
              />
              <circle
                :cx="hoveredLatPt.x"
                :cy="hoveredLatPt.y"
                r="5"
                fill="#a855f7"
                stroke="#ffffff"
                stroke-width="2"
                class="shadow-sm dark:stroke-slate-900"
              />
            </g>

            <!-- Definitions -->
            <defs>
              <linearGradient id="latGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#a855f7" stop-opacity="0.25" />
                <stop offset="100%" stop-color="#a855f7" stop-opacity="0.0" />
              </linearGradient>
            </defs>
          </svg>

          <!-- Floating custom tooltip -->
          <div 
            v-if="hoveredLatPt" 
            class="fixed pointer-events-none z-[100] bg-slate-900/90 dark:bg-slate-950/90 text-white px-3.5 py-2.5 rounded-xl text-[0.7rem] shadow-xl backdrop-blur-md border border-white/10 flex flex-col gap-1 min-w-[160px] transition-opacity duration-150 animate-fade-in"
            :style="{ left: mouseX + 15 + 'px', top: mouseY + 15 + 'px' }"
          >
            <div class="font-bold text-white/50 text-[0.62rem] uppercase tracking-wider">{{ hoveredLatPt.ts }}</div>
            <div class="flex items-center justify-between gap-4 mt-0.5">
              <span class="font-semibold text-white/90">P99 Latency</span>
              <span class="font-black text-purple-400 text-xs">{{ formatValue(hoveredLatPt.val, 'latency') }}</span>
            </div>
          </div>

          <div v-if="!latencyPoints.length" class="absolute inset-0 flex items-center justify-center text-text-secondary text-sm italic">
            No historical P99 Latency data available
          </div>
        </div>

        <!-- Timeline X-Axis scale labels -->
        <div v-if="latencyPoints.length && latencyAnalysis.chartPoints.length" class="flex justify-between items-center text-[0.65rem] text-text-secondary font-bold uppercase mt-1 px-1">
          <span>{{ latLabels.start }}</span>
          <span>{{ latLabels.mid }}</span>
          <span>{{ latLabels.end }}</span>
        </div>
      </div>

      <!-- 4. Bandwidth Speed Card -->
      <div 
        class="bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 relative"
        :class="{ 'opacity-50 pointer-events-none': loading }"
      >
        <!-- Header -->
        <div class="flex justify-between items-start">
          <div>
            <h3 class="text-sm font-bold text-text-primary uppercase tracking-wider flex items-center gap-2">
              <span class="text-cyan-500">📡</span> Bandwidth Consumption
            </h3>
            <p class="text-xs text-text-secondary mt-0.5">Network payload egress speed (bytes per second)</p>
          </div>
          <!-- Stats Summary Box -->
          <div class="flex gap-4 bg-bg-primary/50 px-3 py-1.5 rounded-xl border border-border-color/30 text-[0.68rem] font-bold text-text-secondary uppercase">
            <div>Avg: <span class="text-cyan-500 font-extrabold">{{ formatValue(bandwidthAnalysis.avg, 'bandwidth') }}</span></div>
            <div class="w-px bg-border-color/30 self-stretch"></div>
            <div>Max: <span class="text-text-primary font-extrabold">{{ formatValue(bandwidthAnalysis.max, 'bandwidth') }}</span></div>
          </div>
        </div>

        <!-- SVG Chart -->
        <div class="relative w-full h-[140px] mt-2 select-none">
          <svg v-if="bandwidthPoints.length" viewBox="0 0 550 140" class="w-full h-full overflow-visible" preserveAspectRatio="none">
            <!-- Grid Lines -->
            <g class="stroke-slate-200/50 dark:stroke-slate-800/40" stroke-width="1">
              <line x1="20" y1="15" x2="530" y2="15" />
              <line x1="20" y1="65" x2="530" y2="65" />
              <line x1="20" y1="115" x2="530" y2="115" />
              <line x1="20" y1="125" x2="530" y2="125" class="stroke-slate-200 dark:stroke-slate-700" stroke-width="1.5" />
            </g>

            <!-- Line Curve & Fill Area -->
            <path v-if="bandwidthAnalysis.areaPath" :d="bandwidthAnalysis.areaPath" fill="url(#bwGrad)" />
            <path v-if="bandwidthAnalysis.linePath" :d="bandwidthAnalysis.linePath" fill="none" stroke="#06b6d4" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="drop-shadow-[0_2px_8px_rgba(6,182,212,0.3)]" />

            <!-- Invisible vertical panels for hover tracking -->
            <g v-if="bandwidthAnalysis.chartPoints.length">
              <rect
                v-for="(pt, idx) in bandwidthAnalysis.chartPoints"
                :key="idx"
                :x="pt.x - 9"
                :y="0"
                :width="18"
                :height="140"
                fill="transparent"
                class="cursor-pointer"
                @mouseenter="hoveredIdxBw = idx"
                @mouseleave="hoveredIdxBw = null"
              />
            </g>

            <!-- Dot & crosshair highlighted on hover -->
            <g v-if="hoveredBwPt">
              <!-- Vertical tracker line -->
              <line 
                :x1="hoveredBwPt.x" 
                :y1="10" 
                :x2="hoveredBwPt.x" 
                :y2="125" 
                stroke="#06b6d4" 
                stroke-width="1.5" 
                stroke-dasharray="3 3"
                class="opacity-60" 
              />
              <circle
                :cx="hoveredBwPt.x"
                :cy="hoveredBwPt.y"
                r="5"
                fill="#06b6d4"
                stroke="#ffffff"
                stroke-width="2"
                class="shadow-sm dark:stroke-slate-900"
              />
            </g>

            <!-- Definitions -->
            <defs>
              <linearGradient id="bwGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#06b6d4" stop-opacity="0.25" />
                <stop offset="100%" stop-color="#06b6d4" stop-opacity="0.0" />
              </linearGradient>
            </defs>
          </svg>

          <!-- Floating custom tooltip -->
          <div 
            v-if="hoveredBwPt" 
            class="fixed pointer-events-none z-[100] bg-slate-900/90 dark:bg-slate-950/90 text-white px-3.5 py-2.5 rounded-xl text-[0.7rem] shadow-xl backdrop-blur-md border border-white/10 flex flex-col gap-1 min-w-[160px] transition-opacity duration-150 animate-fade-in"
            :style="{ left: mouseX + 15 + 'px', top: mouseY + 15 + 'px' }"
          >
            <div class="font-bold text-white/50 text-[0.62rem] uppercase tracking-wider">{{ hoveredBwPt.ts }}</div>
            <div class="flex items-center justify-between gap-4 mt-0.5">
              <span class="font-semibold text-white/90">Egress Rate</span>
              <span class="font-black text-cyan-400 text-xs">{{ formatValue(hoveredBwPt.val, 'bandwidth') }}</span>
            </div>
          </div>

          <div v-if="!bandwidthPoints.length" class="absolute inset-0 flex items-center justify-center text-text-secondary text-sm italic">
            No historical Bandwidth data available
          </div>
        </div>

        <!-- Timeline X-Axis scale labels -->
        <div v-if="bandwidthPoints.length && bandwidthAnalysis.chartPoints.length" class="flex justify-between items-center text-[0.65rem] text-text-secondary font-bold uppercase mt-1 px-1">
          <span>{{ bwLabels.start }}</span>
          <span>{{ bwLabels.mid }}</span>
          <span>{{ bwLabels.end }}</span>
        </div>
      </div>

    </div>

    <!-- Back to Top Floating Button -->
    <button 
      @click="scrollToTop" 
      class="fixed bottom-8 right-8 z-50 p-3.5 rounded-2xl bg-bg-secondary/80 backdrop-blur-md border border-border-color text-text-primary shadow-lg hover:shadow-primary-blue/15 hover:text-primary-blue hover:border-primary-blue/30 active:scale-95 transition-all duration-300 cursor-pointer flex items-center justify-center group"
      :class="showScrollTop ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4 pointer-events-none'"
      title="Back to Top"
    >
      <svg 
        xmlns="http://www.w3.org/2000/svg" 
        width="20" 
        height="20" 
        viewBox="0 0 24 24" 
        fill="none" 
        stroke="currentColor" 
        stroke-width="2.5" 
        stroke-linecap="round" 
        stroke-linejoin="round"
        class="transform group-hover:-translate-y-0.5 transition-transform duration-200"
      >
        <line x1="12" y1="19" x2="12" y2="5"></line>
        <polyline points="5 12 12 5 19 12"></polyline>
      </svg>
    </button>
  </div>
</template>

<style scoped>
/* High performance transitions for custom charts hover items */
.animate-fade-in {
  animation: fadeIn 0.4s cubic-bezier(0.4, 0, 0.2, 1) forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
