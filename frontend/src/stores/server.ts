import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import api, { passportApi } from '@/services/api'

export interface ServerConfig {
  name: string
  backendUrl: string
  passportUrl: string
  isCustom?: boolean
}

export const useServerStore = defineStore('server', () => {
  const configServers = ref<ServerConfig[]>([])
  
  // Custom servers added by the user
  const customServers = ref<ServerConfig[]>(
    JSON.parse(localStorage.getItem('custom_servers') || '[]')
  )
  
  // The active server config
  const activeServerName = ref<string>(
    localStorage.getItem('active_server_name') || ''
  )

  const allServers = computed(() => {
    return [...configServers.value, ...customServers.value]
  })

  const currentServer = computed<ServerConfig>(() => {
    // Try to find the active server from all servers
    if (activeServerName.value) {
      const found = allServers.value.find(s => s.name === activeServerName.value)
      if (found) return found
    }
    
    // Fallback: use first server from config.json if available
    if (configServers.value.length > 0 && configServers.value[0]) {
      return configServers.value[0]
    }
    
    // absolute fallback
    return {
      name: 'Default Proxy Server',
      backendUrl: '/api/sev',
      passportUrl: '/api/passport'
    }
  })

  // Synchronously set baseline Axios base URLs using the stored or fallback values
  function initializeAxiosBaseURLs() {
    const srv = currentServer.value
    api.defaults.baseURL = srv.backendUrl
    passportApi.defaults.baseURL = srv.passportUrl
  }

  // Load public/config.json at startup
  async function loadConfig() {
    try {
      // Add timestamp to prevent caching issues
      const response = await fetch(`/config.json?t=${Date.now()}`)
      if (response.ok) {
        const data = await response.json()
        if (data && Array.isArray(data.servers)) {
          configServers.value = data.servers
        }
      }
    } catch (e) {
      console.warn('Failed to load public/config.json, using fallback proxy values.', e)
    } finally {
      // Make sure Axios is initialized with whatever we loaded
      updateAxiosBaseURLs()
    }
  }

  function updateAxiosBaseURLs() {
    const srv = currentServer.value
    api.defaults.baseURL = srv.backendUrl
    passportApi.defaults.baseURL = srv.passportUrl
    console.log(`[ServerStore] API baseURL set to: ${srv.backendUrl}`)
    console.log(`[ServerStore] Passport API baseURL set to: ${srv.passportUrl}`)
  }

  function setActiveServer(serverName: string) {
    if (!serverName) {
      activeServerName.value = ''
      localStorage.removeItem('active_server_name')
    } else {
      activeServerName.value = serverName
      localStorage.setItem('active_server_name', serverName)
    }
    
    updateAxiosBaseURLs()
    
    // Clear auth credentials as they won't be valid on the new backend
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  function addCustomServer(server: ServerConfig) {
    // Avoid duplicates
    if (allServers.value.some(s => s.name === server.name)) {
      throw new Error('A server with this name already exists')
    }
    
    const newServer = { ...server, isCustom: true }
    customServers.value.push(newServer)
    localStorage.setItem('custom_servers', JSON.stringify(customServers.value))
  }

  function removeCustomServer(serverName: string) {
    customServers.value = customServers.value.filter(s => s.name !== serverName)
    localStorage.setItem('custom_servers', JSON.stringify(customServers.value))
    
    // If the removed server was currently active, revert to default
    if (activeServerName.value === serverName) {
      setActiveServer('')
    }
  }

  // Pre-initialize axios right away at store definition import time
  initializeAxiosBaseURLs()

  return {
    configServers,
    customServers,
    activeServerName,
    allServers,
    currentServer,
    loadConfig,
    setActiveServer,
    addCustomServer,
    removeCustomServer
  }
})
