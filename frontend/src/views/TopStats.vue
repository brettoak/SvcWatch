<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getTopIPs, getTopUserAgents } from '@/services/api'
import type { TopIPItem, TopUserAgentItem } from '@/services/api'
import TimeFilter from '@/components/TimeFilter.vue'

const timeFilter = ref('7d')
const currentRange = ref<{ startStr: string; endStr: string } | null>(null)
const loading = ref(false)
const errorMsg = ref('')
const lastUpdated = ref('')

const topIpsData = ref<TopIPItem[]>([])
const topUasData = ref<TopUserAgentItem[]>([])

const ipSearchQuery = ref('')
const uaSearchQuery = ref('')
const expandedUaIdx = ref<number | null>(null)
const copySuccessIdx = ref<number | null>(null)

// Format helper
const formatDateStr = (date: Date) => {
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

// Fetch data function
const fetchData = async () => {
  if (!currentRange.value) {
    errorMsg.value = 'Please select a valid time range'
    return
  }

  errorMsg.value = ''
  loading.value = true

  try {
    const [ipsResp, uasResp] = await Promise.all([
      getTopIPs(currentRange.value.startStr, currentRange.value.endStr, 100),
      getTopUserAgents(currentRange.value.startStr, currentRange.value.endStr, 100)
    ])

    if (ipsResp.data && ipsResp.data.code === 200) {
      topIpsData.value = ipsResp.data.data || []
    }
    if (uasResp.data && uasResp.data.code === 200) {
      topUasData.value = uasResp.data.data || []
    }

    lastUpdated.value = formatDateStr(new Date())
  } catch (err: unknown) {
    if (err && typeof err === 'object' && 'message' in err) {
      const errorObj = err as { message: string; response?: { data?: { message?: string } } }
      errorMsg.value = errorObj.response?.data?.message || errorObj.message || 'Failed to fetch rankings statistics'
    } else {
      errorMsg.value = 'Failed to fetch rankings statistics'
    }
  } finally {
    loading.value = false
  }
}

const onTimeRangeChange = (range: { startStr: string; endStr: string }) => {
  currentRange.value = range
  fetchData()
}

// IP computations
const totalIpRequests = computed(() => {
  return topIpsData.value.reduce((acc, curr) => acc + curr.request_count, 0)
})

const maxIpCount = computed(() => {
  if (topIpsData.value.length === 0) return 1
  return Math.max(...topIpsData.value.map(i => i.request_count))
})

const filteredIps = computed(() => {
  const query = ipSearchQuery.value.trim().toLowerCase()
  let list = topIpsData.value
  if (query) {
    list = list.filter(item => 
      item.ip.toLowerCase().includes(query) ||
      (item.country && item.country.toLowerCase().includes(query)) ||
      (item.region && item.region.toLowerCase().includes(query)) ||
      (item.city && item.city.toLowerCase().includes(query))
    )
  }
  return list.map(item => {
    const share = totalIpRequests.value > 0 ? (item.request_count / totalIpRequests.value) * 100 : 0
    return { ...item, share }
  })
})

// UA computations
const totalUaRequests = computed(() => {
  return topUasData.value.reduce((acc, curr) => acc + curr.request_count, 0)
})

const parseUA = (uaString: string) => {
  if (!uaString) return { browser: 'Unknown', os: 'Unknown', isBot: false }
  const ua = uaString.toLowerCase()
  let browser = 'Unknown'
  let os = 'Unknown'
  let isBot = false

  // Detect bots
  if (ua.includes('bot') || ua.includes('spider') || ua.includes('crawler') || ua.includes('slurp') || ua.includes('googlebot') || ua.includes('bingbot')) {
    isBot = true
    if (ua.includes('googlebot')) browser = 'GoogleBot'
    else if (ua.includes('bingbot')) browser = 'BingBot'
    else browser = 'Bot/Spider'
  } else {
    // Detect browsers/clients
    if (ua.includes('chrome') || ua.includes('chromium')) {
      browser = 'Chrome'
      if (ua.includes('edg/')) browser = 'Edge'
    } else if (ua.includes('firefox')) {
      browser = 'Firefox'
    } else if (ua.includes('safari') && !ua.includes('chrome') && !ua.includes('android')) {
      browser = 'Safari'
    } else if (ua.includes('postman')) {
      browser = 'Postman'
    } else if (ua.includes('curl')) {
      browser = 'curl'
    } else if (ua.includes('python')) {
      browser = 'Python Client'
    } else if (ua.includes('go-http-client')) {
      browser = 'Go Client'
    } else if (ua.includes('wget')) {
      browser = 'wget'
    } else if (ua.includes('okhttp')) {
      browser = 'OkHttp'
    } else if (ua.includes('insomnia')) {
      browser = 'Insomnia'
    }
  }

  // Detect OS
  if (ua.includes('windows')) {
    os = 'Windows'
  } else if (ua.includes('macintosh') || ua.includes('mac os x')) {
    os = 'macOS'
  } else if (ua.includes('linux') && !ua.includes('android')) {
    os = 'Linux'
  } else if (ua.includes('iphone') || ua.includes('ipad') || ua.includes('ipod')) {
    os = 'iOS'
  } else if (ua.includes('android')) {
    os = 'Android'
  }

  return { browser, os, isBot }
}

const filteredUas = computed(() => {
  const query = uaSearchQuery.value.trim().toLowerCase()
  let list = topUasData.value
  if (query) {
    list = list.filter(item => item.user_agent.toLowerCase().includes(query))
  }
  return list.map(item => {
    const share = totalUaRequests.value > 0 ? (item.request_count / totalUaRequests.value) * 100 : 0
    const parsed = parseUA(item.user_agent)
    return { ...item, share, parsed }
  })
})

const getLatencyClass = (avgLatency: number) => {
  if (avgLatency > 1.0) return 'text-red-500 dark:text-red-400 font-extrabold'
  if (avgLatency > 0.3) return 'text-amber-500 dark:text-amber-400 font-bold'
  return 'text-text-primary font-medium'
}

const getErrorRateClass = (rate: number) => {
  if (rate > 5.0) return 'bg-red-500/10 text-red-500 border border-red-500/20'
  if (rate > 0.0) return 'bg-amber-500/10 text-amber-500 border border-amber-500/20'
  return 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20'
}

const toggleExpandUa = (idx: number) => {
  if (expandedUaIdx.value === idx) {
    expandedUaIdx.value = null
  } else {
    expandedUaIdx.value = idx
  }
}

const copyToClipboard = async (text: string, idx: number) => {
  try {
    await navigator.clipboard.writeText(text)
    copySuccessIdx.value = idx
    setTimeout(() => {
      if (copySuccessIdx.value === idx) {
        copySuccessIdx.value = null
      }
    }, 2000)
  } catch (err) {
    console.error('Failed to copy text: ', err)
  }
}

// Scroll to Top logic
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
        <h1 class="text-3xl font-bold m-0 tracking-tight">Traffic Rankings & Client Analytics</h1>
        <p class="text-sm text-text-secondary mt-1 max-w-xl">
          Analyze top requested client IP addresses and user agents to inspect device distribution, latency, and error distribution.
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
    <div class="grid grid-cols-1 xl:grid-cols-2 gap-8 items-stretch">
      
      <!-- LEFT PANEL: Top Client IPs -->
      <div 
        class="bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-6 transition-all duration-300 min-h-[500px]"
        :class="{ 'opacity-50 pointer-events-none': loading }"
      >
        <!-- Card Header & Metadata -->
        <div class="flex flex-col sm:flex-row sm:justify-between sm:items-start gap-4">
          <div>
            <h3 class="text-lg font-bold text-text-primary flex items-center gap-2">
              <span class="text-xl">🌐</span> Top Requested Client IPs
            </h3>
            <p class="text-xs text-text-secondary mt-0.5">Ranked by request hits from clients</p>
          </div>
          
          <div class="flex items-center gap-2 shrink-0 bg-bg-primary/50 px-3 py-1.5 rounded-xl border border-border-color/30 text-[0.7rem] font-bold text-text-secondary uppercase tracking-wider">
            <span>Hits: {{ totalIpRequests.toLocaleString() }}</span>
            <span class="w-1.5 h-1.5 bg-border-color rounded-full"></span>
            <span>Unique: {{ topIpsData.length }}</span>
          </div>
        </div>

        <!-- Local Search Input -->
        <div class="relative w-full">
          <span class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-text-secondary">
            <svg class="h-4.5 w-4.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </span>
          <input 
            type="text" 
            v-model="ipSearchQuery"
            placeholder="Search IP address..." 
            class="w-full bg-bg-primary/40 border border-border-color text-text-primary pl-10 pr-4 py-2.5 rounded-xl text-sm outline-none transition-all focus:bg-bg-primary/80 focus:border-primary-blue focus:ring-2 focus:ring-primary-blue/15"
          />
          <button 
            v-if="ipSearchQuery"
            @click="ipSearchQuery = ''"
            class="absolute inset-y-0 right-0 pr-3.5 flex items-center text-text-secondary hover:text-text-primary bg-transparent border-none cursor-pointer"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- IP Table Container -->
        <div class="overflow-x-auto flex-1 custom-scrollbar min-h-[300px]">
          <table class="w-full text-left border-collapse">
            <thead class="sticky top-0 bg-bg-secondary z-10">
              <tr class="bg-bg-primary/60 backdrop-blur-sm border-b border-border-color">
                <th class="px-4 py-3 text-[0.65rem] font-bold uppercase tracking-wider text-text-secondary w-14 text-center">Rank</th>
                <th class="px-4 py-3 text-[0.65rem] font-bold uppercase tracking-wider text-text-secondary">IP Address / Request Share</th>
                <th class="px-4 py-3 text-[0.65rem] font-bold uppercase tracking-wider text-text-secondary w-44">Location</th>
                <th class="px-4 py-3 text-[0.65rem] font-bold uppercase tracking-wider text-text-secondary text-right w-20">Hits</th>
                <th class="px-4 py-3 text-[0.65rem] font-bold uppercase tracking-wider text-text-secondary text-right w-24">Avg Latency</th>
                <th class="px-4 py-3 text-[0.65rem] font-bold uppercase tracking-wider text-text-secondary text-right w-20">Err%</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border-color/50">
              <tr v-if="filteredIps.length === 0" class="text-center italic text-text-secondary py-8">
                <td colspan="6" class="px-4 py-12 text-sm text-text-secondary">No matching IP analytics found</td>
              </tr>
              <tr v-else v-for="(item, idx) in filteredIps" :key="item.ip" class="hover:bg-bg-primary/20 transition-all duration-150 group">
                <!-- Rank Badge -->
                <td class="px-4 py-4 text-center font-bold text-xs text-text-secondary">
                  <span 
                    class="inline-flex items-center justify-center w-6 h-6 rounded-lg text-[0.7rem]"
                    :class="[
                      idx === 0 ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400 font-extrabold shadow-sm' : '',
                      idx === 1 ? 'bg-slate-400/15 text-slate-600 dark:text-slate-400 font-bold' : '',
                      idx === 2 ? 'bg-yellow-700/15 text-yellow-700 dark:text-yellow-400 font-bold' : '',
                      idx > 2 ? 'bg-bg-primary text-text-secondary' : ''
                    ]"
                  >
                    {{ idx + 1 }}
                  </span>
                </td>
                
                <!-- IP & Share Slider -->
                <td class="px-4 py-4">
                  <div class="flex flex-col gap-1.5">
                    <span class="text-sm font-semibold tracking-tight text-text-primary font-mono group-hover:text-primary-blue transition-colors duration-150">{{ item.ip }}</span>
                    <!-- Dynamic Progress Bar -->
                    <div class="flex items-center gap-2">
                      <div class="h-1.5 w-full max-w-[200px] bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden shadow-inner">
                        <div 
                          class="h-full rounded-full bg-gradient-to-r from-blue-500 to-indigo-500 transition-all duration-700 ease-out" 
                          :style="{ width: `${(item.request_count / maxIpCount) * 100}%` }"
                        ></div>
                      </div>
                      <span class="text-[0.65rem] font-bold text-text-secondary shrink-0">{{ item.share.toFixed(1) }}%</span>
                    </div>
                  </div>
                </td>

                <!-- Location -->
                <td class="px-4 py-4 whitespace-nowrap">
                  <div class="flex items-center gap-1.5">
                    <span v-if="item.country" class="inline-flex items-center gap-1 bg-primary-blue/5 border border-primary-blue/15 text-primary-blue text-[0.7rem] font-bold px-2 py-0.5 rounded-md">
                      🌍 {{ item.country }}
                    </span>
                    <span v-else class="inline-flex items-center gap-1 bg-slate-500/5 border border-slate-500/15 text-text-secondary text-[0.7rem] font-bold px-2 py-0.5 rounded-md">
                      ❓ Unknown
                    </span>
                    <span v-if="item.city || item.region" class="text-xs font-semibold text-text-primary max-w-[120px] truncate" :title="[item.city, item.region].filter(Boolean).join(', ')">
                      {{ item.city || item.region }}
                    </span>
                  </div>
                </td>

                <!-- Hits Count -->
                <td class="px-4 py-4 text-right text-xs font-bold text-text-primary whitespace-nowrap">
                  {{ item.request_count.toLocaleString() }}
                </td>

                <!-- Average Response Time -->
                <td class="px-4 py-4 text-right text-xs whitespace-nowrap">
                  <span :class="getLatencyClass(item.avg_response_time)">
                    {{ item.avg_response_time >= 1.0 ? item.avg_response_time.toFixed(2) : (item.avg_response_time * 1000).toFixed(0) }}
                  </span>
                  <span class="text-[0.65rem] text-text-secondary ml-0.5">{{ item.avg_response_time >= 1.0 ? 's' : 'ms' }}</span>
                </td>

                <!-- Error Rate Badge -->
                <td class="px-4 py-4 text-right whitespace-nowrap">
                  <span 
                    class="text-[0.7rem] font-bold px-2 py-0.5 rounded-full"
                    :class="getErrorRateClass(item.error_rate)"
                  >
                    {{ item.error_rate.toFixed(1) }}%
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- RIGHT PANEL: Top User Agents -->
      <div 
        class="bg-bg-secondary rounded-2xl p-7 shadow-card border border-border-color flex flex-col gap-6 transition-all duration-300 min-h-[500px]"
        :class="{ 'opacity-50 pointer-events-none': loading }"
      >
        <!-- Card Header & Metadata -->
        <div class="flex flex-col sm:flex-row sm:justify-between sm:items-start gap-4">
          <div>
            <h3 class="text-lg font-bold text-text-primary flex items-center gap-2">
              <span class="text-xl">🖥️</span> Top Requested User Agents
            </h3>
            <p class="text-xs text-text-secondary mt-0.5">Ranked by browser, client tools, and device request hits</p>
          </div>
          
          <div class="flex items-center gap-2 shrink-0 bg-bg-primary/50 px-3 py-1.5 rounded-xl border border-border-color/30 text-[0.7rem] font-bold text-text-secondary uppercase tracking-wider">
            <span>Hits: {{ totalUaRequests.toLocaleString() }}</span>
            <span class="w-1.5 h-1.5 bg-border-color rounded-full"></span>
            <span>Unique: {{ topUasData.length }}</span>
          </div>
        </div>

        <!-- Local Search Input -->
        <div class="relative w-full">
          <span class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-text-secondary">
            <svg class="h-4.5 w-4.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </span>
          <input 
            type="text" 
            v-model="uaSearchQuery"
            placeholder="Search User-Agent string..." 
            class="w-full bg-bg-primary/40 border border-border-color text-text-primary pl-10 pr-4 py-2.5 rounded-xl text-sm outline-none transition-all focus:bg-bg-primary/80 focus:border-primary-blue focus:ring-2 focus:ring-primary-blue/15"
          />
          <button 
            v-if="uaSearchQuery"
            @click="uaSearchQuery = ''"
            class="absolute inset-y-0 right-0 pr-3.5 flex items-center text-text-secondary hover:text-text-primary bg-transparent border-none cursor-pointer"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- UA Table Container -->
        <div class="overflow-x-auto flex-1 custom-scrollbar min-h-[300px]">
          <table class="w-full text-left border-collapse">
            <thead class="sticky top-0 bg-bg-secondary z-10">
              <tr class="bg-bg-primary/60 backdrop-blur-sm border-b border-border-color">
                <th class="px-4 py-3 text-[0.65rem] font-bold uppercase tracking-wider text-text-secondary w-14 text-center">Rank</th>
                <th class="px-4 py-3 text-[0.65rem] font-bold uppercase tracking-wider text-text-secondary">Client details & User-Agent</th>
                <th class="px-4 py-3 text-[0.65rem] font-bold uppercase tracking-wider text-text-secondary text-right w-20">Hits</th>
                <th class="px-4 py-3 text-[0.65rem] font-bold uppercase tracking-wider text-text-secondary text-right w-24">Avg Latency</th>
                <th class="px-4 py-3 text-[0.65rem] font-bold uppercase tracking-wider text-text-secondary text-right w-20">Err%</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border-color/50">
              <tr v-if="filteredUas.length === 0" class="text-center italic text-text-secondary py-8">
                <td colspan="5" class="px-4 py-12 text-sm text-text-secondary">No matching User Agent analytics found</td>
              </tr>
              <template v-else v-for="(item, idx) in filteredUas" :key="idx">
                <tr class="hover:bg-bg-primary/10 transition-all duration-150 group">
                  <!-- Rank Badge -->
                  <td class="px-4 py-4 text-center font-bold text-xs text-text-secondary">
                    <span 
                      class="inline-flex items-center justify-center w-6 h-6 rounded-lg text-[0.7rem]"
                      :class="[
                        idx === 0 ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400 font-extrabold shadow-sm animate-pulse' : '',
                        idx === 1 ? 'bg-slate-400/15 text-slate-600 dark:text-slate-400 font-bold' : '',
                        idx === 2 ? 'bg-yellow-700/15 text-yellow-700 dark:text-yellow-400 font-bold' : '',
                        idx > 2 ? 'bg-bg-primary text-text-secondary' : ''
                      ]"
                    >
                      {{ idx + 1 }}
                    </span>
                  </td>
                  
                  <!-- Browser/OS Badges & Raw agent snippet -->
                  <td class="px-4 py-4">
                    <div class="flex flex-col gap-2 max-w-[320px] sm:max-w-[420px]">
                      <!-- Parse badges -->
                      <div class="flex flex-wrap items-center gap-1.5">
                        <!-- Browser Badge -->
                        <span 
                          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[0.62rem] font-bold uppercase border"
                          :class="[
                            item.parsed.isBot ? 'bg-red-500/10 text-red-500 border-red-500/20' : '',
                            item.parsed.browser === 'Chrome' ? 'bg-blue-500/10 text-blue-500 border-blue-500/20' : '',
                            item.parsed.browser === 'Edge' ? 'bg-cyan-500/10 text-cyan-500 border-cyan-500/20' : '',
                            item.parsed.browser === 'Firefox' ? 'bg-orange-500/10 text-orange-500 border-orange-500/20' : '',
                            item.parsed.browser === 'Safari' ? 'bg-teal-500/10 text-teal-500 border-teal-500/20' : '',
                            ['curl', 'Postman', 'Python Client', 'Go Client', 'wget'].includes(item.parsed.browser) ? 'bg-slate-500/10 text-slate-600 dark:text-slate-300 border-slate-500/20' : '',
                            item.parsed.browser === 'Unknown' ? 'bg-text-secondary/10 text-text-secondary border-text-secondary/20' : ''
                          ]"
                        >
                          <span v-if="item.parsed.isBot">🤖</span>
                          <span v-else-if="['curl', 'Postman', 'Python Client', 'Go Client', 'wget'].includes(item.parsed.browser)">🔌</span>
                          <span v-else>🌐</span>
                          {{ item.parsed.browser }}
                        </span>

                        <!-- OS Badge -->
                        <span 
                          v-if="item.parsed.os !== 'Unknown'"
                          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[0.62rem] font-bold uppercase bg-bg-primary text-text-secondary border border-border-color"
                        >
                          🖥️ {{ item.parsed.os }}
                        </span>
                      </div>
                      
                      <!-- Raw Text Snippet & Expand Action -->
                      <div class="flex items-center gap-1">
                        <span class="text-xs text-text-secondary font-mono truncate max-w-[280px] sm:max-w-[360px] select-all">{{ item.user_agent }}</span>
                        <button 
                          @click="toggleExpandUa(idx)" 
                          class="bg-transparent border-none p-1 cursor-pointer rounded text-text-secondary hover:text-primary-blue hover:bg-bg-primary shrink-0 transition-colors"
                          :title="expandedUaIdx === idx ? 'Collapse details' : 'Expand details'"
                        >
                          <svg 
                            class="h-3.5 w-3.5 transform transition-transform duration-200" 
                            :class="{ 'rotate-180': expandedUaIdx === idx }"
                            fill="none" 
                            viewBox="0 0 24 24" 
                            stroke="currentColor" 
                            stroke-width="3"
                          >
                            <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </td>

                  <!-- Hits Count -->
                  <td class="px-4 py-4 text-right text-xs font-bold text-text-primary whitespace-nowrap">
                    {{ item.request_count.toLocaleString() }}
                  </td>

                  <!-- Latency -->
                  <td class="px-4 py-4 text-right text-xs whitespace-nowrap">
                    <span :class="getLatencyClass(item.avg_response_time)">
                      {{ item.avg_response_time >= 1.0 ? item.avg_response_time.toFixed(2) : (item.avg_response_time * 1000).toFixed(0) }}
                    </span>
                    <span class="text-[0.65rem] text-text-secondary ml-0.5">{{ item.avg_response_time >= 1.0 ? 's' : 'ms' }}</span>
                  </td>

                  <!-- Error Rate -->
                  <td class="px-4 py-4 text-right whitespace-nowrap">
                    <span 
                      class="text-[0.7rem] font-bold px-2 py-0.5 rounded-full"
                      :class="getErrorRateClass(item.error_rate)"
                    >
                      {{ item.error_rate.toFixed(1) }}%
                    </span>
                  </td>
                </tr>

                <!-- Expanded Raw UA Details Drawer -->
                <tr v-if="expandedUaIdx === idx">
                  <td colspan="5" class="bg-bg-primary/20 px-8 py-4 animate-slide-in">
                    <div class="flex flex-col gap-2.5 bg-bg-secondary p-4 rounded-xl border border-border-color/60">
                      <div class="flex items-center justify-between">
                        <span class="text-[0.65rem] font-black uppercase tracking-wider text-text-secondary">Full User-Agent String</span>
                        <button 
                          @click="copyToClipboard(item.user_agent, idx)"
                          class="inline-flex items-center gap-1.5 text-[0.68rem] font-extrabold bg-bg-primary border border-border-color hover:bg-bg-secondary text-text-primary py-1 px-3 rounded-lg cursor-pointer transition-all active:scale-95"
                        >
                          <svg v-if="copySuccessIdx !== idx" class="h-3.5 w-3.5 text-text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                          </svg>
                          <svg v-else class="h-3.5 w-3.5 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                          </svg>
                          <span>{{ copySuccessIdx === idx ? 'Copied!' : 'Copy to Clipboard' }}</span>
                        </button>
                      </div>
                      <p class="text-xs text-text-primary font-mono select-all bg-bg-primary/60 p-3 rounded-lg border border-border-color/30 leading-relaxed break-all m-0 shadow-inner">
                        {{ item.user_agent }}
                      </p>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
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
/* Scoped micro-transition styles if any are needed; Tailwind handles the rest! */
</style>
