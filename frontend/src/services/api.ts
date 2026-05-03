import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

const api = axios.create({
  baseURL: '/api/sev', // Proxied via Vite
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
})

// Request interceptor to add Bearer token
api.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    if (authStore.token) {
      if (config.headers && typeof config.headers.set === 'function') {
        config.headers.set('Authorization', `Bearer ${authStore.token}`)
      } else {
        config.headers = config.headers || {}
        config.headers.Authorization = `Bearer ${authStore.token}`
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle 401 errors
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    if (error.response && error.response.status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

export default api

// Special client for passport service requests (auth, users, etc.)
export const passportApi = axios.create({
  baseURL: '/api/passport', // Proxied via Vite
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
})

// Add interceptor to passportApi as well for protected endpoints like /users/profile
passportApi.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    if (authStore.token) {
      if (config.headers && typeof config.headers.set === 'function') {
        config.headers.set('Authorization', `Bearer ${authStore.token}`)
      } else {
        config.headers = config.headers || {}
        config.headers.Authorization = `Bearer ${authStore.token}`
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

export interface DashboardOverviewResponse {
  code: number
  message: string
  data: {
    total_requests: {
      value: number
      compare_percent: number
    }
    success_rate: {
      value: number
      compare_percent: number
    }
    error_rate: {
      value: number
      compare_percent: number
    }
    avg_response_time: {
      value: number
      compare_percent: number
    }
    compare_type: string
  }
}

export interface StatusDistributionResponse {
  code: number
  message: string
  data: {
    total: number
    distribution: Array<{
      code_class: string
      count: number
      percentage: number
    }>
  }
}

export interface TimeSeriesResponse {
  code: number
  message: string
  data: {
    metric: string
    interval: string
    points: Array<{
      ts: string
      value: number
    }>
  }
}

export interface LogEntry {
  remote_addr: string
  remote_user: string
  time_local: string
  request: string
  status: number
  body_bytes_sent: number
  http_referer: string
  http_user_agent: string
  request_time: number
}

export interface Log {
  source_id: string
  entry: LogEntry
}

export interface LogsResponse {
  code: number
  message: string
  data: {
    total: number
    page: number
    size: number
    items: Log[]
  }
}

export interface LogQueryParams {
  page?: number
  size?: number
  start_time?: string
  end_time?: string
  source_id?: string
  ip?: string
  method?: string
  status?: number
  status_class?: string
  path_keyword?: string
  min_latency?: number
  max_latency?: number
  sort?: string
}

export const getDashboardOverview = (startTime: string, endTime: string) => {
  return api.get<DashboardOverviewResponse>('/overview', {
    params: {
      start_time: startTime,
      end_time: endTime,
    },
  })
}

export const getStatusDistribution = (startTime: string, endTime: string) => {
  return api.get<StatusDistributionResponse>('/distribution', {
    params: {
      start_time: startTime,
      end_time: endTime,
    },
  })
}

export const getTimeSeriesStats = (metric: string, startTime: string, endTime: string) => {
  return api.get<TimeSeriesResponse>('/stats/timeseries', {
    params: {
      metric,
      start_time: startTime,
      end_time: endTime,
    },
  })
}

export const getLogs = (params: LogQueryParams) => {
  return api.get<LogsResponse>('/logs', { params })
}

export const uploadAvatar = (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  return passportApi.post('/users/avatar', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export interface TopPathItem {
  uri: string
  request_count: number
  avg_response_time: number
  error_rate: number
}

export interface TopPathsResponse {
  code: number
  message: string
  data: TopPathItem[]
}

export const getTopPaths = (startTime: string, endTime: string, limit: number = 10) => {
  return api.get<TopPathsResponse>('/stats/top-paths', {
    params: {
      start_time: startTime,
      end_time: endTime,
      limit,
    },
  })
}
