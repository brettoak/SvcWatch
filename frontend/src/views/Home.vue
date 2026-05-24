<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { getDashboardOverview, getStatusDistribution, getTimeSeriesStats, getTopPaths } from '@/services/api'
import type { DashboardOverviewResponse, StatusDistributionResponse, TimeSeriesResponse, TopPathItem } from '@/services/api'
import { useAuthStore } from '@/stores/auth'

import TimeFilter from '@/components/TimeFilter.vue'

type DashboardData = DashboardOverviewResponse['data']
type DistributionData = StatusDistributionResponse['data']
type TimeSeriesData = TimeSeriesResponse['data']

const timeFilter = ref('7d')
const currentRange = ref<{ startStr: string; endStr: string } | null>(null)

const onTimeRangeChange = (range: { startStr: string; endStr: string }) => {
  currentRange.value = range
  fetchData()
}

const dashboardData = ref<DashboardData | null>(null)
const distributionData = ref<DistributionData | null>(null)
const realTimePoints = ref<any[]>([])
const isMockSimulate = ref(true)
const isTransitioning = ref(false)
const translateOffset = ref(0)
const topPathsData = ref<TopPathItem[]>([])
let statsWs: WebSocket | null = null
const statsWsStatus = ref<'connecting' | 'connected' | 'error' | 'closed'>('connecting')

watch(isMockSimulate, () => {
  if (statsWs) {
    statsWs.close()
  }
  connectStatsWebSocket()
})

const logsStream = ref<any[]>([])
const totalLogsReceived = ref(0)
const logContainerRef = ref<HTMLElement | null>(null)
let ws: WebSocket | null = null
const wsStatus = ref<'connecting' | 'connected' | 'error' | 'closed'>('connecting')

const connectWebSocket = () => {
  const authStore = useAuthStore()
  if (!authStore.token) return
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/api/sev/logs/ws?token=${authStore.token}`
  
  wsStatus.value = 'connecting'
  ws = new WebSocket(wsUrl)
  
  ws.onopen = () => {
    wsStatus.value = 'connected'
  }

  ws.onerror = (err) => {
    console.error('WebSocket Error:', err)
    wsStatus.value = 'error'
  }

  ws.onclose = () => {
    wsStatus.value = 'closed'
    // Optional: could implement reconnect logic here
  }
  
  let logIdCounter = 0
  ws.onmessage = (event) => {
    // Ignore empty heartbeat/ping messages
    if (!event.data || event.data.trim() === '') return

    let logData;
    try {
      logData = JSON.parse(event.data)
    } catch (e) {
      logData = { raw: event.data, _ts: Date.now() }
    }
    // Use a combination of timestamp, random and a counter for absolute uniqueness and stability
    logData._id = `${Date.now()}-${logIdCounter++}-${Math.random().toString(36).substring(2, 7)}`;
    totalLogsReceived.value++
    logsStream.value.push(logData)
    if (logsStream.value.length > 50) {
      logsStream.value.shift()
    }

    nextTick(() => {
      if (logContainerRef.value) {
        logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
      }
    })
  }
}

const connectStatsWebSocket = () => {
  const authStore = useAuthStore()
  if (!authStore.token) return
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/api/sev/stats/ws?token=${authStore.token}&simulate=${isMockSimulate.value}&interval=3s`
  
  statsWsStatus.value = 'connecting'
  statsWs = new WebSocket(wsUrl)
  
  statsWs.onopen = () => {
    statsWsStatus.value = 'connected'
  }

  statsWs.onerror = (err) => {
    console.error('Stats WebSocket Error:', err)
    statsWsStatus.value = 'error'
  }

  statsWs.onclose = () => {
    statsWsStatus.value = 'closed'
  }
  
  statsWs.onmessage = (event) => {
    if (!event.data || event.data.trim() === '') return
    
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'init') {
        realTimePoints.value = msg.data || []
        isTransitioning.value = false
        translateOffset.value = 0
      } else if (msg.type === 'update') {
        if (msg.data) {
          handleNewRealTimePoint(msg.data)
        }
      }
    } catch (e) {
      console.error('Error parsing stats websocket message:', e)
    }
  }
}

const chartGroupStyle = computed(() => {
  if (isTransitioning.value) {
    return {
      transform: `translate3d(${translateOffset.value}px, 0, 0)`,
      transition: 'transform 0.5s cubic-bezier(0.4, 0, 0.2, 1)'
    }
  }
  return {
    transform: 'translate3d(0, 0, 0)',
    transition: 'none'
  }
})

const handleNewRealTimePoint = (newPoint: any) => {
  // 1. If we are already transitioning, instantly finish it first to prevent overlapping ticks
  if (isTransitioning.value) {
    realTimePoints.value.shift()
    isTransitioning.value = false
    translateOffset.value = 0
  }

  // 2. Add the new point (this makes it 31 points)
  realTimePoints.value.push(newPoint)

  // 3. Trigger transition to slide left
  nextTick(() => {
    isTransitioning.value = true
    translateOffset.value = -20

    // 4. Set a timer to clean up the shift exactly when the transition ends (e.g. 500ms)
    setTimeout(() => {
      if (isTransitioning.value && realTimePoints.value.length > 30) {
        isTransitioning.value = false
        translateOffset.value = 0
        realTimePoints.value.shift()
      }
    }, 500)
  })
}


const highlightRawLog = (text: string) => {
  if (!text) return ''
  let highlighted = text.replace(/</g, '&lt;').replace(/>/g, '&gt;')
  
  // Highlight IP Address - sleek sky blue
  highlighted = highlighted.replace(/^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})/, '<span class="text-sky-400 font-bold">$1</span>')
  
  // Highlight Status Code
  highlighted = highlighted.replace(/&quot; (\d{3}) /, (match, p1) => {
    const code = parseInt(p1, 10)
    let color = 'text-emerald-400'
    if (code >= 400) color = 'text-rose-500'
    else if (code >= 300) color = 'text-amber-400'
    return `&quot; <span class="${color} font-bold">${p1}</span> `
  })

  // Highlight Method and URI
  highlighted = highlighted.replace(/&quot;(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD) (.*?) (HTTP\/[0-9.]+)&quot;/, (match, method, uri, httpVer) => {
    let methodColor = 'text-slate-300'
    if (['GET', 'POST', 'PUT'].includes(method)) methodColor = 'text-amber-400'
    else if (method === 'PATCH') methodColor = 'text-sky-400'
    else if (method === 'DELETE') methodColor = 'text-rose-400'
    return `&quot;<span class="${methodColor} font-bold">${method}</span> <span class="text-slate-100">${uri}</span> <span class="text-slate-500 text-[0.6rem]">${httpVer}</span>&quot;`
  })
  
  // Highlight bytes sent at the end
  highlighted = highlighted.replace(/ (\d+) &quot;/, ' <span class="text-slate-500">$1</span> &quot;')
  
  return highlighted
}

const highlightRequestLine = (req: string) => {
  if (!req) return ''
  const parts = req.split(' ')
  if (parts.length >= 2) {
    const method = parts[0] || ''
    const uri = parts[1] || ''
    const httpVer = parts.length > 2 ? (parts[2] || '') : ''
    let methodColor = 'text-slate-300'
    if (['GET', 'POST', 'PUT'].includes(method)) methodColor = 'text-amber-400'
    else if (method === 'PATCH') methodColor = 'text-sky-400'
    else if (method === 'DELETE') methodColor = 'text-rose-400'
    return `<span class="${methodColor} font-bold">${method}</span> <span class="text-slate-100">${uri}</span> <span class="text-slate-500 text-[0.55rem]">${httpVer}</span>`
  }
  return `<span class="text-slate-200">${req}</span>`
}

const selectedMetric = ref('throughput')
const hoveredBarIdx = ref<number | null>(null)
const mouseX = ref(0)
const mouseY = ref(0)

const handleMouseMove = (e: MouseEvent) => {
  mouseX.value = e.clientX
  mouseY.value = e.clientY
}

const tooltipStyle = computed(() => {
  if (mouseX.value === 0 && mouseY.value === 0) return {}
  const tooltipWidth = 180
  const tooltipHeight = 120
  
  // Determine if the hovered point is in the right area (e.g. right 40%) of the Real-time Requests chart
  const isRightSideOfChart = hoveredBarIdx.value !== null && 
    tsDataSeries.value.success.length > 0 && 
    hoveredBarIdx.value >= tsDataSeries.value.success.length * 0.6
  
  const spaceOnBottom = window.innerHeight - mouseY.value
  
  const left = isRightSideOfChart 
    ? mouseX.value - tooltipWidth - 15 
    : mouseX.value + 15
    
  const top = spaceOnBottom < tooltipHeight + 20 
    ? mouseY.value - tooltipHeight - 15 
    : mouseY.value + 15
    
  return {
    left: `${left}px`,
    top: `${top}px`
  }
})

const metricOptions = [
  { label: 'QPS / Throughput', value: 'throughput' },
  { label: 'Error Rate', value: 'error_rate' },
]

const loading = ref(false)
const tsLoading = ref(false)
const lastUpdated = ref('')
const errorMsg = ref('')

const formatDateStr = (date: Date) => {
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

const calculateTimeRange = () => {
  return currentRange.value
}

const tsDataSeries = computed(() => {
  if (realTimePoints.value.length === 0) return { success: [], errors: [], errorRate: [] }
  
  const maxRequestsVal = Math.max(...realTimePoints.value.map(p => Math.max(p.total - p.errors, p.errors, 1)))
  const maxErrorRateVal = Math.max(...realTimePoints.value.map(p => p.total > 0 ? (p.errors / p.total) * 100 : 0), 1)
  
  const height = 150
  const xSpan = 20
  
  const success = realTimePoints.value.map((p, i) => {
    const val = p.total - p.errors
    const h = (val / maxRequestsVal) * height
    return {
      x: i * xSpan + xSpan / 2,
      y: height - h,
      val: val,
      ts: formatDateStr(new Date(p.ts)),
      fullTs: new Date(p.ts).toLocaleString()
    }
  })

  const errors = realTimePoints.value.map((p, i) => {
    const val = p.errors
    const h = (val / maxRequestsVal) * height
    return {
      x: i * xSpan + xSpan / 2,
      y: height - h,
      val: val,
      ts: formatDateStr(new Date(p.ts)),
      fullTs: new Date(p.ts).toLocaleString()
    }
  })

  const errorRate = realTimePoints.value.map((p, i) => {
    const val = p.total > 0 ? (p.errors / p.total) * 100 : 0
    const h = (val / maxErrorRateVal) * height
    return {
      x: i * xSpan + xSpan / 2,
      y: height - h,
      val: val,
      ts: formatDateStr(new Date(p.ts)),
      fullTs: new Date(p.ts).toLocaleString()
    }
  })

  return { success, errors, errorRate }
})

const fetchData = async () => {
  const range = calculateTimeRange()
  if (!range) {
    errorMsg.value = 'Please select a complete custom time range'
    return
  }
  
  errorMsg.value = ''
  loading.value = true
  
  try {
    const [overviewResp, distResp, topPathsResp] = await Promise.all([
      getDashboardOverview(range.startStr, range.endStr),
      getStatusDistribution(range.startStr, range.endStr),
      getTopPaths(range.startStr, range.endStr, 10)
    ])

    if (overviewResp.data && overviewResp.data.code === 200) {
      dashboardData.value = overviewResp.data.data
    }
    
    if (distResp.data && distResp.data.code === 200) {
      distributionData.value = distResp.data.data
    }

    if (topPathsResp.data && topPathsResp.data.code === 200) {
      topPathsData.value = topPathsResp.data.data || []
    }

    lastUpdated.value = formatDateStr(new Date())
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || err.message || 'API request failed'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  connectWebSocket()
  connectStatsWebSocket()
})

onUnmounted(() => {
  if (ws) ws.close()
  if (statsWs) statsWs.close()
})

const getTrendClass = (val: number, isErrorRate = false) => {
  if (val === 0) return 'trend-neutral'
  if (isErrorRate) {
    return val > 0 ? 'trend-down' : 'trend-up'
  }
  return val > 0 ? 'trend-up' : 'trend-down'
}

const formatPercent = (val: number) => {
  const prefix = val > 0 ? '+' : ''
  return `${prefix}${val.toFixed(2)}%`
}

const getStatusLabel = (codeClass: string) => {
  switch (codeClass) {
    case '1xx': return 'Informational'
    case '2xx': return 'Success'
    case '3xx': return 'Redirect'
    case '4xx': return 'Error'
    case '5xx': return 'Server'
    default: return 'Unknown'
  }
}

const getStatusColor = (codeClass: string) => {
  switch (codeClass) {
    case '2xx': return '#10b981' // Vibrant Emerald
    case '3xx': return '#06b6d4' // Vibrant Cyan
    case '4xx': return '#f59e0b' // Amber
    case '5xx': return '#ef4444' // Red
    default: return '#94a3b8'
  }
}

// Donut Chart Helpers
const chartSize = 180
const radius = 75
const strokeWidth = 18
const center = chartSize / 2
const circumference = 2 * Math.PI * radius

const getDonutSegments = () => {
  if (!distributionData.value?.distribution) return []
  
  // Sort to ensure segments are rendered in order (2xx, 3xx, 4xx, 5xx)
  const sorted = [...distributionData.value.distribution].sort((a, b) => a.code_class.localeCompare(b.code_class))
  
  let currentOffset = 0
  return sorted
    .filter(item => item.percentage > 0)
    .map(item => {
      const percentage = item.percentage
      const segmentLength = (percentage / 100) * circumference
      // We add a tiny gap for the rounded caps to be visible if needed, 
      // but the design shows them touching. stroke-linecap: round will 
      // add length beyond the dash, so we should slightly reduce DashArray if we want exact 100%.
      // Actually, for simplicity and "touching" look, we just use the calculated length.
      const dashArray = `${segmentLength} ${circumference}`
      const dashOffset = -currentOffset
      currentOffset += segmentLength
      
      return {
        ...item,
        dashArray,
        dashOffset,
        color: getStatusColor(item.code_class)
      }
    })
}

const getSuccessRate = () => {
  if (!distributionData.value?.distribution) return '0.00'
  const s2xx = distributionData.value.distribution.find(i => i.code_class === '2xx')?.percentage || 0
  return s2xx.toFixed(2)
}

// Chart Helpers & Paths
const generateBezierPath = (points: { x: number; y: number }[]) => {
  if (points.length === 0 || !points[0]) return ''
  let d = `M -50 ${points[0].y} L ${points[0].x} ${points[0].y}`
  for (let i = 0; i < points.length - 1; i++) {
    const p0 = points[i]
    const p1 = points[i + 1]
    if (!p0 || !p1) continue
    const cp1x = p0.x + (p1.x - p0.x) / 2
    const cp1y = p0.y
    const cp2x = p1.x - (p1.x - p0.x) / 2
    const cp2y = p1.y
    d += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${p1.x} ${p1.y}`
  }
  const lastPt = points[points.length - 1]
  if (lastPt) {
    d += ` L 650 ${lastPt.y}`
  }
  return d
}

const generateAreaPath = (points: { x: number; y: number }[], height = 150) => {
  if (points.length === 0 || !points[0]) return ''
  let d = `M -50 ${height}`
  d += ` L -50 ${points[0].y}`
  d += ` L ${points[0].x} ${points[0].y}`
  for (let i = 0; i < points.length - 1; i++) {
    const p0 = points[i]
    const p1 = points[i + 1]
    if (!p0 || !p1) continue
    const cp1x = p0.x + (p1.x - p0.x) / 2
    const cp1y = p0.y
    const cp2x = p1.x - (p1.x - p0.x) / 2
    const cp2y = p1.y
    d += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${p1.x} ${p1.y}`
  }
  const lastPt = points[points.length - 1]
  if (lastPt) {
    d += ` L 650 ${lastPt.y}`
    d += ` L 650 ${height} Z`
  }
  return d
}

// Success Paths (Emerald Green)
const successLinePath = computed(() => generateBezierPath(tsDataSeries.value.success))
const successAreaPath = computed(() => generateAreaPath(tsDataSeries.value.success))

// Error Paths (Rose Red)
const errorLinePath = computed(() => generateBezierPath(tsDataSeries.value.errors))
const errorAreaPath = computed(() => generateAreaPath(tsDataSeries.value.errors))

// Error Rate Paths (Amber Orange)
const errorRateLinePath = computed(() => generateBezierPath(tsDataSeries.value.errorRate))
const errorRateAreaPath = computed(() => generateAreaPath(tsDataSeries.value.errorRate))

// Hover Tracker
const hoveredBar = computed(() => {
  if (hoveredBarIdx.value === null) return null
  const idx = hoveredBarIdx.value
  const sPt = tsDataSeries.value.success[idx]
  const ePt = tsDataSeries.value.errors[idx]
  const rPt = tsDataSeries.value.errorRate[idx]
  if (!sPt || !ePt || !rPt) return null
  return {
    ts: sPt.ts,
    fullTs: sPt.fullTs,
    successVal: sPt.val,
    errorVal: ePt.val,
    rateVal: rPt.val,
    successY: sPt.y,
    successX: sPt.x,
    errorY: ePt.y,
    errorX: ePt.x,
    rateY: rPt.y,
    rateX: rPt.x,
  }
})
</script>

<template>
  <div class="flex flex-col gap-8 py-4 animate-fade-in text-text-primary">
    <div class="flex flex-col gap-6 md:flex-row md:justify-between md:items-end">
      <h1 class="text-3xl font-bold m-0 tracking-tight">Overview Dashboard</h1>
      
      <div class="flex flex-col items-end gap-3">
        <div class="text-sm text-text-secondary flex items-center gap-2 font-medium">
          Last updated: {{ lastUpdated || '-' }}
          <button class="bg-transparent border-none text-text-secondary cursor-pointer p-1 rounded-md flex items-center justify-center transition-all duration-200 hover:bg-bg-secondary hover:text-primary-blue disabled:opacity-50 disabled:cursor-not-allowed group" @click="fetchData" :disabled="loading" title="Refresh">
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

    <div v-if="errorMsg" class="bg-red-500/10 text-red-500 px-6 py-4 rounded-xl border-l-4 border-red-500 font-semibold animate-slide-in backdrop-blur-md">
      {{ errorMsg }}
    </div>

    <!-- Main Metric Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 transition-opacity duration-300" :class="{ 'opacity-50 pointer-events-none': loading }">
      <!-- Total Requests -->
      <div class="relative bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 hover:-translate-y-1 hover:shadow-card-hover overflow-hidden z-20 border-b-4 border-b-blue-500 group animate-slide-in">
        <h3 class="text-text-secondary text-[0.75rem] font-bold uppercase tracking-widest flex justify-between items-center">Total Requests<span class="text-lg opacity-80">📈</span></h3>
        <div class="flex items-baseline">
          <span class="text-text-primary text-4xl font-extrabold tracking-tight">{{ dashboardData?.total_requests?.value || 0 }}</span>
        </div>
        <div class="flex justify-between items-center mt-auto pt-4 border-t border-border-color">
          <span class="text-text-secondary text-[0.7rem] font-semibold">{{ dashboardData?.compare_type || 'vs yesterday' }}</span>
          <span class="text-[0.85rem] font-bold px-2 py-1 rounded-full" :class="getTrendClass(dashboardData?.total_requests?.compare_percent || 0) === 'trend-up' ? 'bg-green-500/10 text-green-500' : 'bg-red-500/10 text-red-500'">
            {{ formatPercent(dashboardData?.total_requests?.compare_percent || 0) }}
          </span>
        </div>
      </div>

      <!-- Success Rate -->
      <div class="relative bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 hover:-translate-y-1 hover:shadow-card-hover overflow-hidden z-20 border-b-4 border-b-emerald-500 group animate-slide-in [animation-delay:0.1s]">
        <h3 class="text-text-secondary text-[0.75rem] font-bold uppercase tracking-widest flex justify-between items-center">Success Rate<span class="text-lg opacity-80">✨</span></h3>
        <div class="flex items-baseline">
          <span class="text-text-primary text-4xl font-extrabold tracking-tight">{{ (dashboardData?.success_rate?.value || 0).toFixed(2) }}<span class="text-base font-bold text-text-secondary ml-1">%</span></span>
        </div>
        <div class="flex justify-between items-center mt-auto pt-4 border-t border-border-color">
          <span class="text-text-secondary text-[0.7rem] font-semibold">{{ dashboardData?.compare_type || 'vs yesterday' }}</span>
          <span class="text-[0.85rem] font-bold px-2 py-1 rounded-full" :class="getTrendClass(dashboardData?.success_rate?.compare_percent || 0) === 'trend-up' ? 'bg-green-500/10 text-green-500' : 'bg-red-500/10 text-red-500'">
            {{ formatPercent(dashboardData?.success_rate?.compare_percent || 0) }}
          </span>
        </div>
      </div>

      <!-- Error Rate -->
      <div class="relative bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 hover:-translate-y-1 hover:shadow-card-hover overflow-hidden z-20 border-b-4 border-b-red-500 group animate-slide-in [animation-delay:0.2s]">
        <h3 class="text-text-secondary text-[0.75rem] font-bold uppercase tracking-widest flex justify-between items-center">Error Rate<span class="text-lg opacity-80">⚠️</span></h3>
        <div class="flex items-baseline">
          <span class="text-text-primary text-4xl font-extrabold tracking-tight">{{ (dashboardData?.error_rate?.value || 0).toFixed(2) }}<span class="text-base font-bold text-text-secondary ml-1">%</span></span>
        </div>
        <div class="flex justify-between items-center mt-auto pt-4 border-t border-border-color">
          <span class="text-text-secondary text-[0.7rem] font-semibold">{{ dashboardData?.compare_type || 'vs yesterday' }}</span>
          <span class="text-[0.85rem] font-bold px-2 py-1 rounded-full" :class="getTrendClass(dashboardData?.error_rate?.compare_percent || 0, true) === 'trend-up' ? 'bg-green-500/10 text-green-500' : 'bg-red-500/10 text-red-500'">
            {{ formatPercent(dashboardData?.error_rate?.compare_percent || 0) }}
          </span>
        </div>
      </div>

      <!-- Avg Response Time -->
      <div class="relative bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 hover:-translate-y-1 hover:shadow-card-hover overflow-hidden z-20 border-b-4 border-b-amber-500 group animate-slide-in [animation-delay:0.3s]">
        <h3 class="text-text-secondary text-[0.75rem] font-bold uppercase tracking-widest flex justify-between items-center">Avg Latency<span class="text-lg opacity-80">⚡</span></h3>
        <div class="flex items-baseline">
          <span class="text-text-primary text-4xl font-extrabold tracking-tight">{{ (dashboardData?.avg_response_time?.value || 0).toFixed(2) }}<span class="text-base font-bold text-text-secondary ml-1">ms</span></span>
        </div>
        <div class="flex justify-between items-center mt-auto pt-4 border-t border-border-color">
          <span class="text-text-secondary text-[0.7rem] font-semibold">{{ dashboardData?.compare_type || 'vs yesterday' }}</span>
          <span class="text-[0.85rem] font-bold px-2 py-1 rounded-full" :class="getTrendClass(dashboardData?.avg_response_time?.compare_percent || 0, true) === 'trend-up' ? 'bg-green-500/10 text-green-500' : 'bg-red-500/10 text-red-500'">
            {{ formatPercent(dashboardData?.avg_response_time?.compare_percent || 0) }}
          </span>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 items-stretch">
      <!-- Timeseries Bar Chart Card -->
      <div class="lg:col-span-2 relative bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 overflow-hidden z-10 h-full" :class="{ 'opacity-50 pointer-events-none': loading }">
        <div class="flex flex-col gap-4 sm:flex-row sm:justify-between sm:items-center">
          <div class="flex items-center gap-3 flex-wrap">
            <h3 class="text-text-secondary text-[0.75rem] font-bold uppercase tracking-widest m-0">Real-time Requests</h3>
            <div class="flex items-center gap-1.5 bg-bg-primary border border-border-color rounded-full px-2 py-0.5 shadow-sm">
              <span class="w-1.5 h-1.5 rounded-full" 
                    :class="{
                      'bg-emerald-500 animate-pulse shadow-[0_0_6px_rgba(16,185,129,0.8)]': statsWsStatus === 'connected',
                      'bg-amber-500 animate-pulse': statsWsStatus === 'connecting',
                      'bg-red-500': statsWsStatus === 'error' || statsWsStatus === 'closed'
                    }"></span>
              <span class="text-[0.55rem] font-black uppercase tracking-widest text-text-secondary">
                {{ statsWsStatus === 'connected' ? 'Live' : statsWsStatus }}
              </span>
            </div>
            <!-- Mock Simulation Toggle -->
            <div class="flex items-center gap-2 bg-bg-primary border border-border-color rounded-full px-2.5 py-0.5 shadow-sm hover:border-text-secondary/30 transition-colors duration-200">
              <span class="text-[0.55rem] font-black uppercase tracking-widest text-text-secondary">Mock Data</span>
              <button 
                @click="isMockSimulate = !isMockSimulate"
                class="relative inline-flex h-4 w-7 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                :class="isMockSimulate ? 'bg-primary-blue' : 'bg-slate-700'"
                :title="isMockSimulate ? 'Disable Mock Data' : 'Enable Mock Data'"
              >
                <span 
                  class="pointer-events-none inline-block h-3 w-3 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                  :class="isMockSimulate ? 'translate-x-3' : 'translate-x-0'"
                ></span>
              </button>
            </div>
          </div>
          <div class="flex bg-bg-primary rounded-lg p-1 border border-border-color shrink-0">
            <button 
              v-for="opt in metricOptions" 
              :key="opt.value"
              class="px-3 py-1.5 rounded-md text-[0.7rem] font-bold cursor-pointer transition-all duration-200"
              :class="selectedMetric === opt.value ? 'bg-primary-blue text-white shadow-sm' : 'bg-transparent text-text-secondary hover:text-text-primary'"
              @click="selectedMetric = opt.value"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>
        <div class="relative w-full h-[180px] mt-4 flex flex-col">
          <svg v-if="realTimePoints.length" viewBox="0 0 600 170" class="w-full h-full overflow-visible" preserveAspectRatio="none" @mousemove="handleMouseMove">
            <!-- Grid lines -->
            <g class="stroke-slate-200/50 dark:stroke-slate-700/50">
              <line x1="0" y1="42.5" x2="600" y2="42.5" stroke-width="1"/>
              <line x1="0" y1="85" x2="600" y2="85" stroke-width="1"/>
              <line x1="0" y1="127.5" x2="600" y2="127.5" stroke-width="1"/>
              <line x1="0" y1="170" x2="600" y2="170" class="stroke-slate-200 dark:stroke-slate-700" stroke-width="1"/>
            </g>
            
            <!-- ====== Metric Wave Paths ====== -->
            <g clip-path="url(#chartClip)">
              <g :style="chartGroupStyle">
                <!-- 1. QPS & Throughput Tab (Success + Errors) -->
              <template v-if="selectedMetric === 'throughput'">
                <!-- Success (Emerald Area & Line) -->
                <path
                  v-if="successAreaPath"
                  :d="successAreaPath"
                  fill="url(#successGradient)"
                  class="transition-all duration-300"
                />
                <path
                  v-if="successLinePath"
                  :d="successLinePath"
                  fill="none"
                  stroke="#10b981"
                  stroke-width="3"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  class="transition-all duration-300 drop-shadow-md"
                />

                <!-- Errors (Rose Area & Line) -->
                <path
                  v-if="errorAreaPath"
                  :d="errorAreaPath"
                  fill="url(#errorGradient)"
                  class="transition-all duration-300"
                />
                <path
                  v-if="errorLinePath"
                  :d="errorLinePath"
                  fill="none"
                  stroke="#f43f5e"
                  stroke-width="2.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  class="transition-all duration-300 drop-shadow-md"
                />
              </template>

              <!-- 2. Error Rate Tab -->
              <template v-else-if="selectedMetric === 'error_rate'">
                <path
                  v-if="errorRateAreaPath"
                  :d="errorRateAreaPath"
                  fill="url(#errorRateGradient)"
                  class="transition-all duration-300"
                />
                <path
                  v-if="errorRateLinePath"
                  :d="errorRateLinePath"
                  fill="none"
                  stroke="#f59e0b"
                  stroke-width="3"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  class="transition-all duration-300 drop-shadow-md"
                />
              </template>
              
              <!-- Hover detection regions (invisible columns) -->
              <g v-if="tsDataSeries.success.length">
                <rect
                  v-for="(pt, idx) in tsDataSeries.success" 
                  :key="idx"
                  :x="pt.x - 10"
                  :y="0"
                  :width="20"
                  :height="170"
                  fill="transparent"
                  class="cursor-pointer"
                  @mouseenter="hoveredBarIdx = idx"
                  @mouseleave="hoveredBarIdx = null"
                >
                </rect>
              </g>

              <!-- Highlight dots for hovered item -->
              <template v-if="hoveredBar">
                <!-- QPS / Throughput Tab: Two dots -->
                <template v-if="selectedMetric === 'throughput'">
                  <circle
                    :cx="hoveredBar.successX"
                    :cy="hoveredBar.successY"
                    r="5"
                    class="transition-all duration-200 stroke-white dark:stroke-slate-900"
                    fill="#10b981"
                    stroke-width="2"
                  />
                  <circle
                    :cx="hoveredBar.errorX"
                    :cy="hoveredBar.errorY"
                    r="5"
                    class="transition-all duration-200 stroke-white dark:stroke-slate-900"
                    fill="#f43f5e"
                    stroke-width="2"
                  />
                </template>
                <!-- Error Rate Tab: One dot -->
                <template v-else-if="selectedMetric === 'error_rate'">
                  <circle
                    :cx="hoveredBar.rateX"
                    :cy="hoveredBar.rateY"
                    r="5"
                    class="transition-all duration-200 stroke-white dark:stroke-slate-900"
                    fill="#f59e0b"
                    stroke-width="2"
                  />
                </template>
              </template>
              </g>
            </g>

            <!-- Definitions -->
            <defs>
              <clipPath id="chartClip">
                <rect x="0" y="-10" width="600" height="190" />
              </clipPath>
              <linearGradient id="successGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#10b981" stop-opacity="0.25" />
                <stop offset="100%" stop-color="#10b981" stop-opacity="0.0" />
              </linearGradient>
              <linearGradient id="errorGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#f43f5e" stop-opacity="0.2" />
                <stop offset="100%" stop-color="#f43f5e" stop-opacity="0.0" />
              </linearGradient>
              <linearGradient id="errorRateGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#f59e0b" stop-opacity="0.25" />
                <stop offset="100%" stop-color="#f59e0b" stop-opacity="0.0" />
              </linearGradient>
            </defs>
          </svg>
          

          
          <div v-if="!realTimePoints.length" class="flex-1 flex flex-col items-center justify-center text-text-secondary text-sm italic py-10">
            No real-time data available
          </div>
          
          <div v-if="realTimePoints.length && tsDataSeries.success.length" class="flex justify-between items-center text-[0.65rem] text-text-secondary font-bold uppercase tracking-tight mt-3 px-1">
             <span>{{ tsDataSeries.success[0]?.ts || '' }}</span>
             <span>{{ tsDataSeries.success[Math.floor(tsDataSeries.success.length / 2)]?.ts || '' }}</span>
             <span>{{ tsDataSeries.success[tsDataSeries.success.length - 1]?.ts || '' }}</span>
          </div>
        </div>
      </div>

      <!-- Status Code Distribution Card -->
      <div class="lg:col-span-1 relative bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 overflow-hidden z-10 h-full" :class="{ 'opacity-50 pointer-events-none': loading }">
         <h3 class="text-text-secondary text-[0.75rem] font-bold uppercase tracking-widest flex justify-between items-center">Status Distribution<span class="text-lg opacity-80 ml-2">📊</span></h3>
       <div class="flex flex-col xl:flex-row gap-6 items-center flex-1">
          <!-- Donut Chart Left (on large) / Top (on small) -->
          <div class="flex justify-center items-center py-2 shrink-0">
            <svg :width="chartSize" :height="chartSize" viewBox="0 0 180 180" class="filter drop-shadow-md">
              <circle 
                :cx="center" :cy="center" :r="radius" 
                fill="transparent" :stroke-width="strokeWidth" 
                class="stroke-slate-100 dark:stroke-slate-800"
              />
              <circle 
                v-for="seg in getDonutSegments()" :key="seg.code_class"
                :cx="center" :cy="center" :r="radius" 
                fill="transparent" 
                :stroke="seg.color" 
                :stroke-width="strokeWidth" 
                :stroke-dasharray="seg.dashArray" 
                :stroke-dashoffset="seg.dashOffset"
                stroke-linecap="round"
                class="transition-all duration-500 ease-out"
                transform="rotate(-90 90 90)"
              />
              <text :x="center" :y="center + 5" text-anchor="middle" class="fill-text-primary text-2xl font-extrabold">{{ getSuccessRate() }}%</text>
              <text :x="center" :y="center + 25" text-anchor="middle" class="fill-text-secondary text-[0.6rem] font-bold uppercase tracking-widest">Success</text>
            </svg>
          </div>

          <!-- Detailed List Right (on large) / Bottom (on small) -->
          <div class="flex flex-col gap-4 w-full flex-1">
              <div v-for="item in distributionData?.distribution" :key="item.code_class" class="flex flex-col gap-1.5">
                 <div class="flex items-center justify-between">
                    <div class="flex items-center gap-2">
                       <span class="w-2.5 h-2.5 rounded-full shadow-sm" :style="{ backgroundColor: getStatusColor(item.code_class) }"></span>
                       <span class="text-[0.75rem] font-bold text-text-primary">{{ item.code_class }}</span>
                    </div>
                    <span class="text-[0.75rem] font-extrabold text-text-primary">{{ (item.percentage || 0).toFixed(1) }}%</span>
                 </div>
                 <div class="h-1.5 w-full bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden shadow-inner">
                    <div class="h-full rounded-full transition-all duration-700 ease-out" :style="{ width: item.percentage + '%', backgroundColor: getStatusColor(item.code_class) }"></div>
                 </div>
              </div>
              <div v-if="!distributionData?.distribution?.length" class="flex flex-col items-center justify-center text-text-secondary text-sm italic py-4">
                No data available
              </div>
          </div>
       </div>
      </div>
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-2 gap-6 items-stretch animate-slide-in [animation-delay:0.4s]">
      <div class="relative bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 overflow-hidden h-[420px]">
        <h3 class="text-text-secondary text-[0.7rem] font-bold uppercase tracking-widest flex items-center justify-between">
          <div class="flex items-center gap-3">
            <span class="opacity-80">REAL-TIME LOG STREAM</span>
            <span class="text-emerald-500 font-black">{{ totalLogsReceived }} LINES</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="w-1.5 h-1.5 rounded-full" 
                  :class="{
                    'bg-emerald-500 animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.8)]': wsStatus === 'connected',
                    'bg-amber-500 animate-pulse': wsStatus === 'connecting',
                    'bg-red-500': wsStatus === 'error' || wsStatus === 'closed'
                  }"></span>
            <span class="text-[0.6rem] font-bold uppercase tracking-widest"
                  :class="{
                    'text-emerald-500': wsStatus === 'connected',
                    'text-amber-500': wsStatus === 'connecting',
                    'text-red-500': wsStatus === 'error' || wsStatus === 'closed'
                  }">
              {{ wsStatus }}
            </span>
          </div>
        </h3>
        <div ref="logContainerRef" class="overflow-y-auto flex-1 min-h-0 w-full flex flex-col gap-1.5 font-log text-[0.7rem] custom-scrollbar pr-2 py-1">
          <!-- Stable log container to prevent full re-render when switching from empty to populated -->
          <div class="relative min-h-full">
            <div v-if="!logsStream.length" class="absolute inset-0 flex flex-col items-center justify-center text-center italic text-text-secondary py-8 gap-2 z-10 pointer-events-none">
              <span v-if="wsStatus === 'connecting'">Connecting to log stream...</span>
              <span v-else-if="wsStatus === 'error'">Connection failed. Please check server.</span>
              <span v-else-if="wsStatus === 'closed'">Connection closed.</span>
              <span v-else>Waiting for logs...</span>
            </div>
            
            <TransitionGroup name="log-list" tag="div" class="flex flex-col gap-1.5">
              <div v-for="log in logsStream" :key="log._id" class="flex gap-4 bg-bg-primary/30 p-2 rounded-lg border border-border-color/30 hover:bg-bg-primary/60 transition-all duration-200 items-start relative z-20 group">
                <template v-if="log.raw">
                   <span class="text-slate-500/60 shrink-0 whitespace-nowrap text-[0.65rem] font-medium">{{ new Date(log._ts).toLocaleTimeString([], { hour12: false }) }}</span>
                   <div class="flex flex-col gap-1 w-full overflow-hidden">
                     <div class="text-slate-400 break-all leading-relaxed" v-html="highlightRawLog(log.raw)"></div>
                   </div>
                </template>
                <template v-else>
                   <span class="text-slate-500/60 shrink-0 whitespace-nowrap text-[0.65rem] font-medium">{{ new Date(log.time_local || log._ts || Date.now()).toLocaleTimeString([], { hour12: false }) }}</span>
                   <div class="flex flex-col gap-1 w-full overflow-hidden">
                     <div class="flex items-center gap-3">
                       <span class="text-sky-400 font-bold shrink-0">{{ log.remote_addr }}</span>
                       <span class="truncate" :title="log.request" v-html="highlightRequestLine(log.request)"></span>
                       <span class="font-bold shrink-0 ml-auto" 
                             :class="log.status >= 400 ? 'text-rose-500' : (log.status >= 300 ? 'text-amber-400' : 'text-emerald-400')">
                         {{ log.status || 200 }}
                       </span>
                       <span class="text-slate-500/50 shrink-0 min-w-[40px] text-right">{{ Math.round(log.body_bytes_sent / 1024) }}KB</span>
                     </div>
                   </div>
                </template>
              </div>
            </TransitionGroup>
          </div>
        </div>
      </div>

      <!-- Top Request Paths Card (Right) -->
      <div class="relative bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 h-[420px]" :class="{ 'opacity-50 pointer-events-none': loading }">
      <h3 class="text-text-secondary text-[0.75rem] font-bold uppercase tracking-widest flex items-center">Top Request Paths<span class="text-lg opacity-80 ml-2">🔥</span></h3>
      <div class="overflow-auto flex-1 min-h-0 w-full custom-scrollbar pr-2">
        <table class="w-full text-left border-collapse relative">
          <thead class="sticky top-0 bg-bg-secondary z-10 shadow-sm">
            <tr class="bg-bg-primary/90 backdrop-blur-sm border-b border-border-color">
              <th class="px-4 py-3 text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary">Path</th>
              <th class="px-4 py-3 text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary text-right">Hits</th>
              <th class="px-4 py-3 text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary text-right">Avg ms</th>
              <th class="px-4 py-3 text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary text-right">Err%</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border-color">
            <tr v-if="!topPathsData.length" class="text-center italic text-text-secondary py-4">
              <td colspan="4" class="px-4 py-8">No data available</td>
            </tr>
            <tr v-else v-for="(item, idx) in topPathsData" :key="idx" class="hover:bg-bg-primary/30 transition-colors">
              <td class="px-4 py-3 text-xs font-bold text-text-primary max-w-[400px] truncate" :title="item.uri">{{ item.uri }}</td>
              <td class="px-4 py-3 text-xs font-bold text-text-primary text-right">{{ item.request_count.toLocaleString() }}</td>
              <td class="px-4 py-3 text-right">
                <span class="text-xs font-bold" :class="item.avg_response_time > 0.5 ? 'text-red-500' : 'text-text-primary'">
                  {{ (item.avg_response_time * 1000).toFixed(0) }}
                </span>
              </td>
              <td class="px-4 py-3 text-right">
                <span class="text-xs font-bold" :class="item.error_rate > 5 ? 'text-red-500' : (item.error_rate > 0 ? 'text-amber-500' : 'text-emerald-500')">
                  {{ item.error_rate.toFixed(1) }}%
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    </div>

    <!-- Custom Tooltip (placed at root level to prevent z-index/clipping issues) -->
    <div 
      v-if="hoveredBar" 
      class="fixed pointer-events-none z-[100] bg-slate-900/90 text-white px-3.5 py-2.5 rounded-xl text-[0.7rem] shadow-xl backdrop-blur-md border border-white/10 flex flex-col gap-1.5 min-w-[155px] transition-opacity duration-200"
      :style="tooltipStyle"
    >
      <!-- Header: Time -->
      <div class="flex items-center gap-2 border-b border-white/10 pb-1.5 mb-0.5">
        <span class="font-bold text-white/90 text-[0.65rem] tracking-tight">{{ hoveredBar.fullTs }}</span>
      </div>

      <!-- Content for Throughput (Success + Errors) -->
      <template v-if="selectedMetric === 'throughput'">
        <div class="flex items-center justify-between gap-4">
          <div class="flex items-center gap-1.5">
            <span class="w-1.5 h-1.5 rounded-full bg-[#10b981]"></span>
            <span class="text-white/60 font-semibold uppercase text-[0.55rem] tracking-wider">Success</span>
          </div>
          <span class="text-[0.75rem] font-black text-white">{{ hoveredBar.successVal }} reqs</span>
        </div>
        <div class="flex items-center justify-between gap-4">
          <div class="flex items-center gap-1.5">
            <span class="w-1.5 h-1.5 rounded-full bg-[#f43f5e]"></span>
            <span class="text-white/60 font-semibold uppercase text-[0.55rem] tracking-wider">Failed</span>
          </div>
          <span class="text-[0.75rem] font-black text-[#f43f5e]">{{ hoveredBar.errorVal }} errs</span>
        </div>
        <div class="flex items-center justify-between gap-4 border-t border-white/5 pt-1.5 mt-0.5">
          <span class="text-white/40 font-bold uppercase text-[0.55rem] tracking-wider">Total</span>
          <span class="text-[0.75rem] font-black text-white/80">{{ hoveredBar.successVal + hoveredBar.errorVal }} reqs</span>
        </div>
      </template>

      <!-- Content for Error Rate -->
      <template v-else-if="selectedMetric === 'error_rate'">
        <div class="flex items-center justify-between gap-4">
          <div class="flex items-center gap-1.5">
            <span class="w-1.5 h-1.5 rounded-full bg-[#f59e0b]"></span>
            <span class="text-white/60 font-semibold uppercase text-[0.55rem] tracking-wider">Error Rate</span>
          </div>
          <span class="text-[0.75rem] font-black text-[#f59e0b]">{{ hoveredBar.rateVal.toFixed(2) }}%</span>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.log-list-enter-active,
.log-list-leave-active {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}
.log-list-enter-from {
  opacity: 0;
  transform: translateY(20px);
}
.log-list-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
.log-list-move {
  transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}
</style>
