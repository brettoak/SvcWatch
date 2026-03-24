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
