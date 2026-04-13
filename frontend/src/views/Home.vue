<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { getDashboardOverview, getStatusDistribution, getTimeSeriesStats } from '@/services/api'
import type { DashboardOverviewResponse, StatusDistributionResponse, TimeSeriesResponse } from '@/services/api'

type DashboardData = DashboardOverviewResponse['data']
type DistributionData = StatusDistributionResponse['data']
type TimeSeriesData = TimeSeriesResponse['data']

const timeFilter = ref('5m')
const timeOptions = [
  { label: '5m', value: '5m' },
  { label: '30m', value: '30m' },
  { label: '1h', value: '1h' },
  { label: '6h', value: '6h' },
  { label: '24h', value: '24h' },
  { label: '7d', value: '7d' },
  { label: 'Custom', value: 'custom' },
]

const customStartTime = ref('')
const customEndTime = ref('')

const dashboardData = ref<DashboardData | null>(null)
const distributionData = ref<DistributionData | null>(null)
const timeSeriesData = ref<TimeSeriesData | null>(null)

const selectedMetric = ref('qps')
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
    const [overviewResp, distResp] = await Promise.all([
      getDashboardOverview(range.startStr, range.endStr),
      getStatusDistribution(range.startStr, range.endStr)
    ])

    if (overviewResp.data && overviewResp.data.code === 200) {
      dashboardData.value = overviewResp.data.data
    }
    
    if (distResp.data && distResp.data.code === 200) {
      distributionData.value = distResp.data.data
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
const getTsBars = () => {
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
      // Format time based on interval roughly
      ts: new Date(p.ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }
  })
}

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
</script>

<template>
  <div class="dashboard-wrapper">
    <div class="dashboard-header">
      <h1 class="page-title">Overview Dashboard</h1>
      
      <div class="controls-container">
        <div class="last-updated">
          Last updated: {{ lastUpdated || '-' }}
          <button class="icon-btn refresh-btn" @click="fetchData" :disabled="loading" title="Refresh">
            <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="23 4 23 10 17 10"></polyline>
              <polyline points="1 20 1 14 7 14"></polyline>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
            </svg>
          </button>
        </div>

        <div class="time-filter-group">
          <div class="preset-filters">
            <button 
              v-for="opt in timeOptions" 
              :key="opt.value"
              :class="['filter-btn', { active: timeFilter === opt.value }]"
              @click="timeFilter = opt.value"
            >
              {{ opt.label }}
            </button>
          </div>

          <div v-if="timeFilter === 'custom'" class="custom-time-picker fade-in">
            <input type="datetime-local" v-model="customStartTime" class="time-input" />
            <span class="separator">to</span>
            <input type="datetime-local" v-model="customEndTime" class="time-input" />
            <button class="primary-btn" @click="fetchData" :disabled="loading">Search</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="errorMsg" class="error-banner">
      {{ errorMsg }}
    </div>

    <!-- Main Metric Cards -->
    <div class="stats-grid" :class="{ 'is-loading': loading }">
      <!-- Total Requests -->
      <div class="stat-card total-req-card">
        <h3 class="stat-title">Total Requests<span class="stat-icon">📈</span></h3>
        <div class="stat-main">
          <span class="stat-value">{{ dashboardData?.total_requests?.value || 0 }}</span>
        </div>
        <div class="stat-footer">
          <span class="compare-label">{{ dashboardData?.compare_type || 'vs yesterday' }}</span>
          <span :class="['trend', getTrendClass(dashboardData?.total_requests?.compare_percent || 0)]">
            {{ formatPercent(dashboardData?.total_requests?.compare_percent || 0) }}
          </span>
        </div>
        <div class="card-glow"></div>
      </div>

      <!-- Success Rate -->
      <div class="stat-card success-rate-card">
        <h3 class="stat-title">Success Rate<span class="stat-icon">✨</span></h3>
        <div class="stat-main">
          <span class="stat-value">{{ (dashboardData?.success_rate?.value || 0).toFixed(2) }}<span class="unit">%</span></span>
        </div>
        <div class="stat-footer">
          <span class="compare-label">{{ dashboardData?.compare_type || 'vs yesterday' }}</span>
          <span :class="['trend', getTrendClass(dashboardData?.success_rate?.compare_percent || 0)]">
            {{ formatPercent(dashboardData?.success_rate?.compare_percent || 0) }}
          </span>
        </div>
        <div class="card-glow"></div>
      </div>

      <!-- Error Rate -->
      <div class="stat-card error-rate-card">
        <h3 class="stat-title">Error Rate<span class="stat-icon">⚠️</span></h3>
        <div class="stat-main">
          <span class="stat-value">{{ (dashboardData?.error_rate?.value || 0).toFixed(2) }}<span class="unit">%</span></span>
        </div>
        <div class="stat-footer">
          <span class="compare-label">{{ dashboardData?.compare_type || 'vs yesterday' }}</span>
          <span :class="['trend', getTrendClass(dashboardData?.error_rate?.compare_percent || 0, true)]">
            {{ formatPercent(dashboardData?.error_rate?.compare_percent || 0) }}
          </span>
        </div>
        <div class="card-glow"></div>
      </div>

      <!-- Avg Response Time -->
      <div class="stat-card latency-card">
        <h3 class="stat-title">Avg Latency<span class="stat-icon">⚡</span></h3>
        <div class="stat-main">
          <span class="stat-value">{{ (dashboardData?.avg_response_time?.value || 0).toFixed(2) }}<span class="unit">ms</span></span>
        </div>
        <div class="stat-footer">
          <span class="compare-label">{{ dashboardData?.compare_type || 'vs yesterday' }}</span>
          <span :class="['trend', getTrendClass(dashboardData?.avg_response_time?.compare_percent || 0, true)]">
            {{ formatPercent(dashboardData?.avg_response_time?.compare_percent || 0) }}
          </span>
        </div>
        <div class="card-glow"></div>
      </div>
    </div>

    <div class="charts-row">
      <!-- Timeseries Bar Chart Card -->
      <div class="timeseries-card stat-card" :class="{ 'is-loading': loading || tsLoading }">
        <div class="stat-title-row">
          <h3 class="stat-title">Requests Over Time<span class="stat-icon">📈</span></h3>
          <div class="header-actions">
            <div class="metric-tabs">
              <button 
                v-for="opt in metricOptions" 
                :key="opt.value"
                :class="['metric-tab', { active: selectedMetric === opt.value }]"
                @click="selectedMetric = opt.value"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>
        </div>
        <div class="timeseries-chart-container">
          <svg v-if="timeSeriesData?.points?.length" viewBox="0 0 600 170" class="bar-chart-svg" preserveAspectRatio="none">
            <!-- Grid lines -->
            <g class="chart-grid">
              <line x1="0" y1="42.5" x2="600" y2="42.5" stroke="rgba(255,255,255,0.03)" stroke-width="1"/>
              <line x1="0" y1="85" x2="600" y2="85" stroke="rgba(255,255,255,0.03)" stroke-width="1"/>
              <line x1="0" y1="127.5" x2="600" y2="127.5" stroke="rgba(255,255,255,0.03)" stroke-width="1"/>
              <line x1="0" y1="170" x2="600" y2="170" stroke="rgba(255,255,255,0.1)" stroke-width="1"/>
            </g>
            
            <!-- Bars -->
            <g class="chart-bars">
              <rect
                v-for="(bar, idx) in getTsBars()" 
                :key="idx"
                :x="bar.x"
                :y="bar.y"
                :width="bar.w"
                :height="bar.h"
                fill="url(#barGradient)"
                class="bar-rect"
                rx="2"
                ry="2"
              >
                <title>{{ bar.ts }} - {{ formatBarTooltip(bar.val) }}</title>
              </rect>
            </g>

            <!-- Definitions -->
            <defs>
              <linearGradient id="barGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#3b82f6" />
                <stop offset="100%" stop-color="rgba(59, 130, 246, 0.2)" />
              </linearGradient>
            </defs>
          </svg>
          <div v-else class="empty-state">
            No timeseries data available
          </div>
          
          <div v-if="timeSeriesData?.points?.length" class="chart-labels">
             <span class="x-axis-label">{{ getTsBars()[0]?.ts || '' }}</span>
             <span class="x-axis-label">{{ getTsBars()[Math.floor(getTsBars().length / 2)]?.ts || '' }}</span>
             <span class="x-axis-label">{{ getTsBars()[getTsBars().length - 1]?.ts || '' }}</span>
          </div>
        </div>
      </div>

      <!-- Status Code Distribution Card -->
      <div class="distribution-card stat-card" :class="{ 'is-loading': loading }">
         <h3 class="stat-title">Status Code Distribution<span class="stat-icon">📊</span></h3>
       <div class="distribution-vertical">
          <!-- Donut Chart Top -->
          <div class="chart-section">
            <svg :width="chartSize" :height="chartSize" viewBox="0 0 180 180" class="donut-svg">
              <circle 
                :cx="center" :cy="center" :r="radius" 
                fill="transparent" :stroke-width="strokeWidth" 
                class="donut-bg-circle"
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
                class="donut-segment"
                transform="rotate(-90 90 90)"
              />
              <text :x="center" :y="center + 5" text-anchor="middle" class="chart-main-value">{{ getSuccessRate() }}%</text>
              <text :x="center" :y="center + 25" text-anchor="middle" class="chart-sub-label">Success</text>
            </svg>
          </div>

          <!-- Detailed List Bottom -->
          <div class="distribution-list">
              <div v-for="item in distributionData?.distribution" :key="item.code_class" class="distribution-row">
                 <div class="dist-row-main">
                    <div class="dist-label-info">
                       <span class="dist-bullet" :style="{ backgroundColor: getStatusColor(item.code_class) }"></span>
                       <span class="dist-code">{{ item.code_class }}</span>
                       <span class="dist-desc">{{ getStatusLabel(item.code_class) }}</span>
                    </div>
                    <div class="dist-bar-wrapper">
                       <div class="dist-bar-bg">
                          <div class="dist-bar-fill" :style="{ width: item.percentage + '%', backgroundColor: getStatusColor(item.code_class) }"></div>
                       </div>
                    </div>
                    <div class="dist-percent-val">{{ (item.percentage || 0).toFixed(2) }}%</div>
                 </div>
              </div>
              <div v-if="!distributionData?.distribution?.length" class="empty-state">
                No data available for this range
              </div>
          </div>
       </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-wrapper {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  padding: 1rem 0;
  animation: fadeIn 0.4s ease-out;
}

.dashboard-header {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

@media (min-width: 768px) {
  .dashboard-header {
    flex-direction: row;
    justify-content: space-between;
    align-items: flex-end;
  }
}

.page-title {
  color: var(--text-primary);
  font-size: 2rem;
  font-weight: 700;
  margin: 0;
  letter-spacing: -0.02em;
  text-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.controls-container {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.75rem;
}

.last-updated {
  font-size: 0.875rem;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.icon-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.icon-btn:hover:not(:disabled) {
  background: var(--bg-hover, rgba(0,0,0,0.05));
  color: var(--primary-color, #3b82f6);
}

.icon-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.refresh-btn:active:not(:disabled) svg {
  transform: rotate(180deg);
  transition: transform 0.3s ease;
}

.time-filter-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  align-items: flex-end;
}

.preset-filters {
  display: flex;
  background-color: var(--bg-secondary);
  border-radius: 10px;
  padding: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  border: 1px solid var(--border-color, rgba(255,255,255,0.05));
}

.filter-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 6px 14px;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.filter-btn:hover {
  color: var(--text-primary);
}

.filter-btn.active {
  background-color: var(--primary-color, #3b82f6);
  color: white;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.4);
}

.custom-time-picker {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background-color: var(--bg-secondary);
  padding: 6px 12px;
  border-radius: 10px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  border: 1px solid var(--border-color, rgba(255,255,255,0.05));
}

.time-input {
  background: transparent;
  border: 1px solid var(--border-color, rgba(255,255,255,0.1));
  color: var(--text-primary);
  padding: 6px 10px;
  border-radius: 6px;
  font-size: 0.875rem;
  outline: none;
  transition: border-color 0.2s;
}

.time-input:focus {
  border-color: var(--primary-color, #3b82f6);
}

.separator {
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.primary-btn {
  background-color: var(--primary-color, #3b82f6);
  color: white;
  border: none;
  padding: 6px 16px;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.primary-btn:hover:not(:disabled) {
  filter: brightness(1.1);
  transform: translateY(-1px);
}

.primary-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: translateY(0);
}

.error-banner {
  background-color: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  padding: 1rem 1.5rem;
  border-radius: 10px;
  border-left: 4px solid #ef4444;
  font-weight: 500;
  animation: slideIn 0.3s ease-out;
  backdrop-filter: blur(8px);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  transition: opacity 0.3s;
}

.stats-grid.is-loading {
  opacity: 0.5;
  pointer-events: none;
}

.stat-card {
  position: relative;
  background: linear-gradient(145deg, var(--bg-secondary) 0%, rgba(255, 255, 255, 0.02) 100%);
  border-radius: 16px;
  padding: 1.75rem;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  border: 1px solid var(--border-color, rgba(255,255,255,0.05));
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
  backdrop-filter: blur(12px);
  overflow: hidden;
  z-index: 1;
}

.stat-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.2);
}

.card-glow {
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle at center, rgba(255,255,255,0.08) 0%, transparent 60%);
  opacity: 0;
  transition: opacity 0.4s;
  pointer-events: none;
  z-index: -1;
}

.stat-card:hover .card-glow {
  opacity: 1;
}

.total-req-card { border-bottom: 3px solid #3b82f6; }
.success-rate-card { border-bottom: 3px solid #10b981; }
.error-rate-card { border-bottom: 3px solid #ef4444; }
.latency-card { border-bottom: 3px solid #f59e0b; }

.stat-title {
  color: var(--text-secondary);
  font-size: 0.9375rem;
  font-weight: 600;
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-icon {
  font-size: 1.25rem;
  opacity: 0.8;
}

.stat-main {
  display: flex;
  align-items: baseline;
}

.stat-value {
  color: var(--text-primary);
  font-size: 2.75rem;
  font-weight: 800;
  line-height: 1;
  font-feature-settings: "tnum";
  text-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.unit {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-secondary);
  margin-left: 4px;
}

.stat-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
  padding-top: 1rem;
  border-top: 1px solid var(--border-color, rgba(255,255,255,0.05));
}

.compare-label {
  color: var(--text-secondary);
  font-size: 0.8125rem;
  font-weight: 500;
}

.trend {
  font-size: 0.875rem;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 12px;
}

.trend-up {
  background-color: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.trend-down {
  background-color: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.trend-neutral {
  background-color: rgba(107, 114, 128, 0.15);
  color: var(--text-secondary);
}

.charts-row {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  margin-top: 1rem;
}

.timeseries-card {
  width: 100%;
  padding: 2.5rem;
}

.distribution-card {
  width: 100%;
  padding: 2.5rem;
}

@media (min-width: 1024px) {
  .charts-row {
    display: grid;
    grid-template-columns: 2fr 1fr;
  }
}

.stat-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.metric-tabs {
  display: flex;
  background-color: var(--bg-secondary);
  border-radius: 8px;
  padding: 3px;
  border: 1px solid var(--border-color, rgba(255,255,255,0.05));
}

.metric-tab {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.metric-tab:hover {
  color: var(--text-primary);
}

.metric-tab.active {
  background-color: var(--primary-color, #3b82f6);
  color: white;
  box-shadow: 0 2px 6px rgba(59, 130, 246, 0.3);
}

.timeseries-chart-container {
  margin-top: 1.5rem;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.bar-chart-svg {
  width: 100%;
  height: 170px;
  overflow: visible;
  filter: drop-shadow(0 4px 6px rgba(0,0,0,0.1));
}

.bar-rect {
  transition: height 0.8s ease-out, y 0.8s ease-out;
}

.bar-rect:hover {
  fill: #60a5fa;
  cursor: pointer;
}

.chart-labels {
  display: flex;
  justify-content: space-between;
  width: 100%;
  padding: 0 5px;
}

.x-axis-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.distribution-vertical {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
  align-items: center;
  margin-top: 1rem;
}

.chart-section {
  display: flex;
  justify-content: center;
  align-items: center;
}

.donut-svg {
  filter: drop-shadow(0 0 8px rgba(0, 242, 254, 0.2));
}

.donut-bg-circle {
  stroke: var(--text-secondary);
  stroke-opacity: 0.15;
}

.donut-segment {
  transition: stroke-dashoffset 0.8s cubic-bezier(0.34, 1.56, 0.64, 1);
  filter: drop-shadow(0 0 4px rgba(255, 255, 255, 0.1));
}

.chart-main-value {
  fill: var(--text-primary);
  font-size: 1.75rem;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.chart-sub-label {
  fill: var(--text-secondary);
  font-size: 0.8125rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.distribution-list {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  max-width: 600px;
}

.dist-row-main {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.dist-label-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 130px;
}

.dist-bullet {
  width: 10px;
  height: 10px;
  border-radius: 3px;
  flex-shrink: 0;
}

.dist-code {
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--text-primary);
}

.dist-desc {
  font-size: 0.8125rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.dist-bar-wrapper {
  flex: 1;
}

.dist-bar-bg {
  height: 4px;
  background-color: rgba(255, 255, 255, 0.05);
  border-radius: 2px;
  overflow: hidden;
}

.dist-bar-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 1s ease-in-out;
}

.dist-percent-val {
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--text-primary);
  min-width: 50px;
  text-align: right;
  font-feature-settings: "tnum";
}

.empty-state {
  text-align: center;
  color: var(--text-secondary);
  padding: 2rem;
  font-style: italic;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes slideIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.fade-in {
  animation: fadeIn 0.3s ease-out;
}
</style>
.stat-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.interval-selector, .selector-label, .modern-select {
  display: none;
}


.modern-select:focus {
  border-color: var(--primary-color, #3b82f6);
}

.metric-tabs {
  display: flex;
  gap: 0.5rem;
}
