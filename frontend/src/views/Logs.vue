<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { getLogs } from '@/services/api'
import type { Log, LogQueryParams } from '@/services/api'

const logs = ref<Log[]>([])
const total = ref(0)
const loading = ref(false)
const showAdvanced = ref(false)

const toggleAdvanced = () => {
  showAdvanced.value = !showAdvanced.value
}

const formatToDateTimeLocal = (date: Date) => {
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const getDefaultTimes = () => {
  const end = new Date()
  const start = new Date()
  start.setMonth(start.getMonth() - 1)
  return {
    start: formatToDateTimeLocal(start),
    end: formatToDateTimeLocal(end)
  }
}

const defaultTimes = getDefaultTimes()

const filters = ref<LogQueryParams>({
  page: 1,
  size: 50,
  start_time: defaultTimes.start,
  end_time: defaultTimes.end,
  source_id: 'access.log',
  ip: '',
  method: '',
  status: undefined,
  status_class: '',
  path_keyword: '',
  min_latency: undefined,
  max_latency: undefined,
  sort: 'time_desc'
})

const fetchLogs = async () => {
  loading.value = true
  try {
    const params = { ...filters.value }
    // Clean up empty params
    Object.keys(params).forEach(key => {
      const k = key as keyof LogQueryParams
      if (params[k] === '' || params[k] === undefined || params[k] === null) {
        delete params[k]
      }
    })
    
    const resp = await getLogs(params)
    if (resp.data && resp.data.code === 200) {
      logs.value = resp.data.data.items || []
      total.value = resp.data.data.total || 0
    }
  } catch (err) {
    console.error('Failed to fetch logs', err)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  filters.value.page = 1
  fetchLogs()
}

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

watch(timeFilter, (newVal) => {
  if (newVal !== 'custom') {
    const end = new Date()
    let start = new Date(end)
    switch (newVal) {
      case '5m': start.setMinutes(start.getMinutes() - 5); break
      case '30m': start.setMinutes(start.getMinutes() - 30); break
      case '1h': start.setHours(start.getHours() - 1); break
      case '6h': start.setHours(start.getHours() - 6); break
      case '24h': start.setHours(start.getHours() - 24); break
      case '7d': start.setDate(start.getDate() - 7); break
      case '30d': start.setDate(start.getDate() - 30); break
    }
    filters.value.start_time = formatToDateTimeLocal(start)
    filters.value.end_time = formatToDateTimeLocal(end)
    handleSearch()
  }
})

const handleReset = () => {
  const times = getDefaultTimes()
  filters.value = {
    page: 1,
    size: 50,
    start_time: times.start,
    end_time: times.end,
    source_id: 'access.log',
    ip: '',
    method: '',
    status: undefined,
    status_class: '',
    path_keyword: '',
    min_latency: undefined,
    max_latency: undefined,
    sort: 'time_desc'
  }
  if (timeFilter.value !== '30d') {
    timeFilter.value = '30d'
  } else {
    fetchLogs()
  }
}

const nextPage = () => {
  if (filters.value.page! * filters.value.size! < total.value) {
    filters.value.page!++
    fetchLogs()
  }
}

const prevPage = () => {
  if (filters.value.page! > 1) {
    filters.value.page!--
    fetchLogs()
  }
}

onMounted(() => {
  fetchLogs()
})

const getStatusColor = (status: number) => {
  if (status >= 200 && status < 300) return 'text-emerald-500 bg-emerald-500/10'
  if (status >= 300 && status < 400) return 'text-cyan-500 bg-cyan-500/10'
  if (status >= 400 && status < 500) return 'text-amber-500 bg-amber-500/10'
  if (status >= 500) return 'text-red-500 bg-red-500/10'
  return 'text-slate-500 bg-slate-500/10'
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const parseMethod = (request: string) => {
  if (!request) return '-'
  return request.split(' ')[0] || '-'
}

const parsePath = (request: string) => {
  if (!request) return '-'
  const parts = request.split(' ')
  return parts.length > 1 ? parts[1] : request
}
</script>

<template>
  <div class="flex flex-col gap-6 animate-fade-in">
    <div class="flex justify-between items-end">
      <div>
        <h1 class="text-3xl font-bold m-0 tracking-tight text-text-primary">Log Explorer</h1>
        <p class="text-text-secondary mt-2 font-medium">Query and analyze detailed access logs</p>
      </div>
      <div class="flex gap-3">
         <button 
           @click="handleReset" 
           class="px-4 py-2 rounded-xl text-sm font-bold border border-border-color bg-bg-secondary text-text-secondary hover:text-text-primary hover:border-text-secondary transition-all"
         >
           Reset
         </button>
         <button 
           @click="handleSearch" 
           class="px-6 py-2 rounded-xl text-sm font-bold bg-primary-blue text-white shadow-lg shadow-primary-blue/20 hover:brightness-110 active:scale-95 transition-all flex items-center gap-2"
         >
           <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
           Search
         </button>
      </div>
    </div>

    <!-- Filters Card -->
    <div class="bg-bg-secondary rounded-2xl p-6 shadow-card border border-border-color flex flex-col gap-6">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- Basic Filters -->
        <div class="flex flex-col gap-1.5">
          <label class="text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary ml-1">Source ID</label>
          <input v-model="filters.source_id" type="text" placeholder="e.g. access.log" class="bg-bg-primary border border-border-color rounded-xl px-4 py-2.5 text-sm outline-none focus:border-primary-blue transition-all" />
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary ml-1">IP Address</label>
          <input v-model="filters.ip" type="text" placeholder="Search IP..." class="bg-bg-primary border border-border-color rounded-xl px-4 py-2.5 text-sm outline-none focus:border-primary-blue transition-all" />
        </div>
        <div class="flex flex-col gap-1.5 lg:col-span-2">
          <label class="text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary ml-1">Time Range</label>
          <div class="flex flex-wrap items-center gap-2">
            <div class="flex bg-bg-primary rounded-xl p-1 border border-border-color w-fit">
              <button 
                type="button"
                v-for="opt in timeOptions" 
                :key="opt.value"
                class="px-3 py-1.5 rounded-lg text-xs font-bold cursor-pointer transition-all duration-200 border-none outline-none"
                :class="timeFilter === opt.value ? 'bg-primary-blue text-white shadow-md shadow-primary-blue/30' : 'bg-transparent text-text-secondary hover:text-text-primary'"
                @click="timeFilter = opt.value"
              >
                {{ opt.label }}
              </button>
            </div>
            
            <div v-if="timeFilter === 'custom'" class="flex items-center gap-1 animate-fade-in flex-1 min-w-[300px]">
              <input v-model="filters.start_time" type="datetime-local" class="bg-bg-primary border border-border-color rounded-xl px-3 py-1.5 text-sm outline-none focus:border-primary-blue transition-all w-full" />
              <span class="text-text-secondary text-xs font-bold px-1">to</span>
              <input v-model="filters.end_time" type="datetime-local" class="bg-bg-primary border border-border-color rounded-xl px-3 py-1.5 text-sm outline-none focus:border-primary-blue transition-all w-full" />
            </div>
          </div>
        </div>
      </div>

      <!-- Advanced Toggle -->
      <button 
        type="button"
        @click.stop.prevent="toggleAdvanced" 
        class="text-[0.75rem] font-bold text-primary-blue flex items-center gap-1 hover:underline w-fit bg-transparent border-none cursor-pointer p-1"
      >
        <span class="pointer-events-none">{{ showAdvanced ? 'Hide Advanced Filters' : 'Show Advanced Filters' }}</span>
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="showAdvanced ? 'rotate-180' : ''" class="transition-transform pointer-events-none"><polyline points="6 9 12 15 18 9"></polyline></svg>
      </button>

      <Transition name="slide">
        <div v-if="showAdvanced" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="flex flex-col gap-1.5">
          <label class="text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary ml-1">HTTP Method</label>
          <select v-model="filters.method" class="bg-bg-primary border border-border-color rounded-xl px-4 py-2.5 text-sm outline-none focus:border-primary-blue transition-all">
            <option value="">All Methods</option>
            <option value="GET">GET</option>
            <option value="POST">POST</option>
            <option value="PUT">PUT</option>
            <option value="DELETE">DELETE</option>
            <option value="HEAD">HEAD</option>
            <option value="OPTIONS">OPTIONS</option>
          </select>
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary ml-1">Status Code</label>
          <input v-model.number="filters.status" type="number" placeholder="e.g. 200" class="bg-bg-primary border border-border-color rounded-xl px-4 py-2.5 text-sm outline-none focus:border-primary-blue transition-all" />
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary ml-1">Status Class</label>
          <select v-model="filters.status_class" class="bg-bg-primary border border-border-color rounded-xl px-4 py-2.5 text-sm outline-none focus:border-primary-blue transition-all">
            <option value="">All</option>
            <option value="2xx">2xx Success</option>
            <option value="3xx">3xx Redirect</option>
            <option value="4xx">4xx Client Error</option>
            <option value="5xx">5xx Server Error</option>
          </select>
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary ml-1">Path Keyword</label>
          <input v-model="filters.path_keyword" type="text" placeholder="Search path..." class="bg-bg-primary border border-border-color rounded-xl px-4 py-2.5 text-sm outline-none focus:border-primary-blue transition-all" />
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary ml-1">Min Latency (ms)</label>
          <input v-model.number="filters.min_latency" type="number" placeholder="0" class="bg-bg-primary border border-border-color rounded-xl px-4 py-2.5 text-sm outline-none focus:border-primary-blue transition-all" />
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary ml-1">Max Latency (ms)</label>
          <input v-model.number="filters.max_latency" type="number" placeholder="5000" class="bg-bg-primary border border-border-color rounded-xl px-4 py-2.5 text-sm outline-none focus:border-primary-blue transition-all" />
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary ml-1">Sort By</label>
          <select v-model="filters.sort" class="bg-bg-primary border border-border-color rounded-xl px-4 py-2.5 text-sm outline-none focus:border-primary-blue transition-all">
            <option value="time_desc">Time (Newest First)</option>
            <option value="latency_desc">Latency (Highest First)</option>
          </select>
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary ml-1">Page Size</label>
          <select v-model.number="filters.size" class="bg-bg-primary border border-border-color rounded-xl px-4 py-2.5 text-sm outline-none focus:border-primary-blue transition-all">
            <option :value="20">20 per page</option>
            <option :value="50">50 per page</option>
            <option :value="100">100 per page</option>
            <option :value="200">200 per page</option>
          </select>
        </div>
      </div>
      </Transition>
    </div>

    <!-- Logs Table Card -->
    <div class="bg-bg-secondary rounded-2xl shadow-card border border-border-color overflow-hidden flex flex-col min-h-[400px]">
      <div class="overflow-x-auto flex-1">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-bg-primary/50 border-b border-border-color">
              <th class="px-6 py-4 text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary">Time</th>
              <th class="px-6 py-4 text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary">IP</th>
              <th class="px-6 py-4 text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary">Method</th>
              <th class="px-6 py-4 text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary">Path</th>
              <th class="px-6 py-4 text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary text-center">Status</th>
              <th class="px-6 py-4 text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary text-right">Latency</th>
              <th class="px-6 py-4 text-[0.7rem] font-bold uppercase tracking-widest text-text-secondary text-right">Size</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border-color">
            <tr v-if="loading" v-for="i in 5" :key="'skeleton-'+i" class="animate-pulse">
              <td colspan="7" class="px-6 py-4">
                <div class="h-4 bg-bg-primary rounded w-full"></div>
              </td>
            </tr>
            <tr v-else-if="logs.length === 0" class="text-center italic text-text-secondary py-10">
              <td colspan="7" class="px-6 py-20">No logs found matching criteria</td>
            </tr>
            <tr v-else v-for="log in logs" :key="log.entry.id" class="hover:bg-bg-primary/30 transition-colors group">
              <td class="px-6 py-4 text-xs font-medium text-text-secondary whitespace-nowrap">{{ new Date(log.entry.time_local).toLocaleString() }}</td>
              <td class="px-6 py-4 text-xs font-bold text-text-primary">{{ log.entry.remote_addr }}</td>
              <td class="px-6 py-4">
                <span class="px-2 py-0.5 rounded text-[0.65rem] font-black uppercase tracking-tighter" 
                  :class="{
                    'bg-emerald-500/10 text-emerald-500': parseMethod(log.entry.request) === 'GET',
                    'bg-blue-500/10 text-blue-500': parseMethod(log.entry.request) === 'POST',
                    'bg-amber-500/10 text-amber-500': parseMethod(log.entry.request) === 'PUT',
                    'bg-red-500/10 text-red-500': parseMethod(log.entry.request) === 'DELETE',
                    'bg-slate-500/10 text-slate-500': !['GET','POST','PUT','DELETE'].includes(parseMethod(log.entry.request))
                  }">
                  {{ parseMethod(log.entry.request) }}
                </span>
              </td>
              <td class="px-6 py-4">
                <div class="text-xs text-text-primary font-medium max-w-[300px] truncate" :title="parsePath(log.entry.request)">{{ parsePath(log.entry.request) }}</div>
              </td>
              <td class="px-6 py-4 text-center">
                <span class="px-2 py-1 rounded-md text-[0.75rem] font-black" :class="getStatusColor(log.entry.status)">
                  {{ log.entry.status }}
                </span>
              </td>
              <td class="px-6 py-4 text-right">
                <span class="text-xs font-bold" :class="log.entry.request_time > 0.5 ? 'text-red-500' : 'text-text-primary'">
                  {{ (log.entry.request_time * 1000).toFixed(0) }} ms
                </span>
              </td>
              <td class="px-6 py-4 text-right text-xs text-text-secondary font-medium">
                {{ formatBytes(log.entry.body_bytes_sent) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Footer -->
      <div class="px-6 py-4 bg-bg-primary/30 border-t border-border-color flex justify-between items-center">
        <div class="text-[0.75rem] font-bold text-text-secondary uppercase tracking-tight">
          Total: <span class="text-text-primary">{{ total }}</span> logs
        </div>
        <div class="flex items-center gap-4">
          <span class="text-xs font-bold text-text-secondary">
            Page <span class="text-text-primary">{{ filters.page }}</span> of {{ Math.ceil(total / (filters.size || 50)) || 1 }}
          </span>
          <div class="flex gap-2">
            <button 
              @click="prevPage" 
              :disabled="filters.page === 1 || loading"
              class="p-2 rounded-lg border border-border-color hover:bg-bg-secondary disabled:opacity-30 disabled:cursor-not-allowed transition-all"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"></polyline></svg>
            </button>
            <button 
              @click="nextPage" 
              :disabled="filters.page! * filters.size! >= total || loading"
              class="p-2 rounded-lg border border-border-color hover:bg-bg-secondary disabled:opacity-30 disabled:cursor-not-allowed transition-all"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"></polyline></svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Chrome, Safari, Edge, Opera */
input::-webkit-outer-spin-button,
input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

/* Firefox */
input[type=number] {
  -moz-appearance: textfield;
}

/* Transition for Advanced Filters */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease-out;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
