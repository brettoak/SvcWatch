<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { getDashboardOverview, getStatusDistribution, getTimeSeriesStats, getTopPaths } from '@/services/api'
import type { DashboardOverviewResponse, StatusDistributionResponse, TimeSeriesResponse, TopPathItem } from '@/services/api'
import { useAuthStore } from '@/stores/auth'

type DashboardData = DashboardOverviewResponse['data']
type DistributionData = StatusDistributionResponse['data']
type TimeSeriesData = TimeSeriesResponse['data']

const timeFilter = ref('7d')
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

const dashboardData = ref<DashboardData | null>(null)
const distributionData = ref<DistributionData | null>(null)
const timeSeriesData = ref<TimeSeriesData | null>(null)
const topPathsData = ref<TopPathItem[]>([])

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

const selectedMetric = ref('bandwidth')
const hoveredBarIdx = ref<number | null>(null)
const mouseX = ref(0)
const mouseY = ref(0)

const handleMouseMove = (e: MouseEvent) => {
  mouseX.value = e.clientX
  mouseY.value = e.clientY
}

const metricOptions = [
  { label: 'QPS', value: 'qps' },
  { label: 'Error Rate', value: 'error_rate' },
  { label: 'Latency', value: 'latency_p99' },
  { label: 'Bandwidth', value: 'bandwidth' },
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

const fetchTimeSeries = async () => {
  const range = calculateTimeRange()
  if (!range) return
  tsLoading.value = true
  try {
    const tsResp = await getTimeSeriesStats(selectedMetric.value, range.startStr, range.endStr)
    if (tsResp.data && tsResp.data.code === 200) {
      timeSeriesData.value = tsResp.data.data
    }
  } catch (err: any) {
    console.error('Timeseries load failed', err)
  } finally {
    tsLoading.value = false
  }
}

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

    await fetchTimeSeries()

    lastUpdated.value = formatDateStr(new Date())
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || err.message || 'API request failed'
  } finally {
    loading.value = false
  }
}

watch(selectedMetric, () => {
  fetchTimeSeries()
})

watch(timeFilter, (newVal) => {
  if (newVal !== 'custom') {
    fetchData()
  }
})

onMounted(() => {
  const end = new Date()
  const start = new Date(end)
  start.setHours(start.getHours() - 1)
  
  const toLocalISO = (d: Date) => {
    const pad = (n: number) => n.toString().padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
  
  customEndTime.value = toLocalISO(end)
  customStartTime.value = toLocalISO(start)

  fetchData()
  connectWebSocket()
})

onUnmounted(() => {
  if (ws) ws.close()
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

// Bar Chart Helpers
const tsBars = computed(() => {
  if (!timeSeriesData.value?.points || timeSeriesData.value.points.length === 0) return []
  const pts = timeSeriesData.value.points
  const maxVal = Math.max(...pts.map(p => p.value), 1)
  const height = 150
  const width = 600
  const barWidth = Math.max((width / pts.length) - 8, 4)
  const xSpan = width / pts.length
  
  return pts.map((p, i) => {
    const h = (p.value / maxVal) * height
    return {
      x: i * xSpan + (xSpan - barWidth) / 2,
      y: height - h,
      w: barWidth,
      h: Math.max(h, 2), // minimum height 2px to be visible
      val: p.value,
      // Format time based on interval roughly, now includes date
      ts: new Date(p.ts).toLocaleString([], { 
        month: '2-digit', 
        day: '2-digit', 
        hour: '2-digit', 
        minute: '2-digit',
        hour12: false 
      }),
      fullTs: new Date(p.ts).toLocaleString()
    }
  })
})

const hoveredBar = computed(() => {
  if (hoveredBarIdx.value === null) return null
  return tsBars.value[hoveredBarIdx.value] || null
})

const formatBarTooltip = (val: number) => {
  if (selectedMetric.value === 'bandwidth') return Math.round(val / 1024) + ' KB/s'
  if (selectedMetric.value === 'latency_p99') return val.toFixed(1) + ' ms'
  if (selectedMetric.value === 'error_rate') return val.toFixed(2) + '%'
  return val.toFixed(0) + ' req/s'
}

const getTsMaxVal = () => {
  if (!timeSeriesData.value?.points || timeSeriesData.value.points.length === 0) return 0
  return Math.max(...timeSeriesData.value.points.map(p => p.value))
}

const tsWavePath = computed(() => {
  if (tsBars.value.length === 0) return ''
  const pts = tsBars.value.map(bar => ({ x: bar.x + bar.w / 2, y: bar.y }))
  if (pts.length === 0 || !pts[0]) return ''
  
  let d = `M ${pts[0].x} ${pts[0].y}`
  for (let i = 0; i < pts.length - 1; i++) {
    const p0 = pts[i]
    const p1 = pts[i + 1]
    if (!p0 || !p1) continue
    const cp1x = p0.x + (p1.x - p0.x) / 2
    const cp1y = p0.y
    const cp2x = p1.x - (p1.x - p0.x) / 2
    const cp2y = p1.y
    d += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${p1.x} ${p1.y}`
  }
  return d
})

const tsAreaPath = computed(() => {
  if (tsBars.value.length === 0) return ''
  const pts = tsBars.value.map(bar => ({ x: bar.x + bar.w / 2, y: bar.y }))
  if (pts.length === 0 || !pts[0]) return ''
  const height = 150
  
  let d = `M ${pts[0].x} ${height}`
  d += ` L ${pts[0].x} ${pts[0].y}`
  for (let i = 0; i < pts.length - 1; i++) {
    const p0 = pts[i]
    const p1 = pts[i + 1]
    if (!p0 || !p1) continue
    const cp1x = p0.x + (p1.x - p0.x) / 2
    const cp1y = p0.y
    const cp2x = p1.x - (p1.x - p0.x) / 2
    const cp2y = p1.y
    d += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${p1.x} ${p1.y}`
  }
  const lastPt = pts[pts.length - 1]
  if (lastPt) {
    d += ` L ${lastPt.x} ${height} Z`
  }
  return d
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

        <div class="flex flex-col gap-2 items-end">
          <div class="flex bg-bg-secondary rounded-xl p-1 shadow-sm border border-border-color">
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

          <div v-if="timeFilter === 'custom'" class="flex items-center gap-2 bg-bg-secondary p-1.5 rounded-xl shadow-sm border border-border-color animate-fade-in mt-1">
            <input type="datetime-local" v-model="customStartTime" class="bg-transparent border border-border-color text-text-primary px-2.5 py-1.5 rounded-md text-sm outline-none transition-all focus:border-primary-blue" />
            <span class="text-text-secondary text-sm">to</span>
            <input type="datetime-local" v-model="customEndTime" class="bg-transparent border border-border-color text-text-primary px-2.5 py-1.5 rounded-md text-sm outline-none transition-all focus:border-primary-blue" />
            <button class="bg-primary-blue text-white border-none py-1.5 px-4 rounded-md text-sm font-bold cursor-pointer transition-all hover:brightness-110 hover:-translate-y-px disabled:opacity-60 disabled:cursor-not-allowed" @click="fetchData" :disabled="loading">Search</button>
          </div>
        </div>
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
      <div class="lg:col-span-2 relative bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-5 transition-all duration-300 overflow-hidden z-10 h-full" :class="{ 'opacity-50 pointer-events-none': loading || tsLoading }">
        <div class="flex flex-col gap-4 sm:flex-row sm:justify-between sm:items-center">
          <h3 class="text-text-secondary text-[0.75rem] font-bold uppercase tracking-widest flex justify-between items-center">Requests Over Time<span class="text-lg opacity-80 ml-2">📈</span></h3>
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
          <svg v-if="timeSeriesData?.points?.length" viewBox="0 0 600 170" class="w-full h-full overflow-visible" preserveAspectRatio="none" @mousemove="handleMouseMove">
            <!-- Grid lines -->
            <g class="stroke-slate-200/50 dark:stroke-slate-700/50">
              <line x1="0" y1="42.5" x2="600" y2="42.5" stroke-width="1"/>
              <line x1="0" y1="85" x2="600" y2="85" stroke-width="1"/>
              <line x1="0" y1="127.5" x2="600" y2="127.5" stroke-width="1"/>
              <line x1="0" y1="170" x2="600" y2="170" class="stroke-slate-200 dark:stroke-slate-700" stroke-width="1"/>
            </g>
            
            <!-- Area Fill -->
            <path
              v-if="tsAreaPath"
              :d="tsAreaPath"
              fill="url(#waveGradient)"
              class="transition-all duration-300"
            />
            
            <!-- Smooth Line -->
            <path
              v-if="tsWavePath"
              :d="tsWavePath"
              fill="none"
              stroke="var(--color-primary-blue)"
              stroke-width="3"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="transition-all duration-300 drop-shadow-md"
            />
            
            <!-- Hover detection regions (invisible columns) -->
            <g>
              <rect
                v-for="(bar, idx) in tsBars" 
                :key="idx"
                :x="bar.x + bar.w / 2 - (600 / tsBars.length) / 2"
                :y="0"
                :width="600 / tsBars.length"
                :height="170"
                fill="transparent"
                class="cursor-pointer"
                @mouseenter="hoveredBarIdx = idx"
                @mouseleave="hoveredBarIdx = null"
              >
              </rect>
            </g>

            <!-- Highlight dot for hovered item -->
            <circle
              v-if="hoveredBar"
              :cx="hoveredBar.x + hoveredBar.w / 2"
              :cy="hoveredBar.y"
              r="4"
              class="transition-all duration-200 stroke-white dark:stroke-slate-900"
              fill="var(--color-primary-blue)"
              stroke-width="2"
            />

            <!-- Definitions -->
            <defs>
              <linearGradient id="waveGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="var(--color-primary-blue)" stop-opacity="0.3" />
                <stop offset="100%" stop-color="var(--color-primary-blue)" stop-opacity="0.0" />
              </linearGradient>
            </defs>
          </svg>
          
          <!-- Custom Tooltip -->
          <div 
            v-if="hoveredBar" 
            class="fixed pointer-events-none z-[100] bg-slate-900/90 text-white px-3 py-2 rounded-lg text-[0.7rem] shadow-xl backdrop-blur-md border border-white/10 flex flex-col gap-0.5 min-w-[120px] transition-opacity duration-200"
            :style="{ left: mouseX + 15 + 'px', top: mouseY + 15 + 'px' }"
          >
            <div class="flex items-center gap-2 mb-1 border-b border-white/10 pb-1">
              <span class="w-2 h-2 rounded-full bg-primary-blue"></span>
              <span class="font-bold text-white/90">{{ hoveredBar.fullTs }}</span>
            </div>
            <div class="flex justify-between items-baseline">
              <span class="text-white/50 uppercase text-[0.6rem] font-bold tracking-wider">{{ selectedMetric.replace('_', ' ') }}</span>
              <span class="text-[0.9rem] font-black text-white">{{ formatBarTooltip(hoveredBar.val || 0) }}</span>
            </div>
          </div>
          <div v-else class="flex-1 flex flex-col items-center justify-center text-text-secondary text-sm italic py-10">
            No timeseries data available
          </div>
          
          <div v-if="timeSeriesData?.points?.length" class="flex justify-between items-center text-[0.65rem] text-text-secondary font-bold uppercase tracking-tight mt-3 px-1">
             <span>{{ tsBars[0]?.ts || '' }}</span>
             <span>{{ tsBars[Math.floor(tsBars.length / 2)]?.ts || '' }}</span>
             <span>{{ tsBars[tsBars.length - 1]?.ts || '' }}</span>
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
