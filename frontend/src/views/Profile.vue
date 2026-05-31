<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { passportApi, uploadAvatar } from '@/services/api'
import { useServerStore } from '@/stores/server'

interface UserProfile {
  id: number
  username: string
  email: string
  role: string
  status: string
  avatarUrl?: string
}

const profile = ref<UserProfile | null>(null)
const isLoading = ref(true)
const isUploadingAvatar = ref(false)
const error = ref('')
const avatarInput = ref<HTMLInputElement | null>(null)

// Server switching setup
const serverStore = useServerStore()
const newServerName = ref('')
const newBackendUrl = ref('')
const newPassportUrl = ref('')
const formError = ref('')

const fetchProfile = async () => {
  isLoading.value = true
  error.value = ''
  try {
    const response = await passportApi.get('/users/profile')
    if (response.data && response.data.code === 200) {
      profile.value = response.data.data
    } else {
      error.value = response.data.message || 'Failed to load profile'
    }
  } catch (err: any) {
    error.value = err.response?.data?.message || err.message || 'An error occurred while fetching profile'
    console.error('Profile fetch error:', err)
  } finally {
    isLoading.value = false
  }
}

const triggerAvatarUpload = () => {
  if (avatarInput.value) {
    avatarInput.value.click()
  }
}

const handleAvatarUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  
  if (!file) return

  if (!file.type.startsWith('image/')) {
    alert('Please select an image file')
    return
  }
  
  if (file.size > 2 * 1024 * 1024) {
    alert('File size must be less than 2MB')
    return
  }

  try {
    isUploadingAvatar.value = true
    const response = await uploadAvatar(file)
    if (response.data && response.data.code === 200) {
      if (profile.value) {
        await fetchProfile()
      }
    } else {
      alert(response.data.message || 'Failed to upload avatar')
    }
  } catch (error: any) {
    console.error('Avatar upload failed:', error)
    alert(error.response?.data?.message || 'Failed to upload avatar')
  } finally {
    isUploadingAvatar.value = false
    if (target) {
      target.value = ''
    }
  }
}

const handleSwitchServer = (serverName: string) => {
  if (serverName === serverStore.currentServer.name) {
    return
  }
  const confirmed = confirm(
    `Are you sure you want to switch to server [${serverName}]?\n\nWarning: Switching backends will clear your current session and you will need to sign in again. Please ensure the target server is online and has CORS enabled.`
  )
  if (confirmed) {
    serverStore.setActiveServer(serverName)
  }
}

const handleAddServer = () => {
  formError.value = ''
  
  const name = newServerName.value.trim()
  const backend = newBackendUrl.value.trim()
  const passport = newPassportUrl.value.trim()
  
  if (!name || !backend || !passport) {
    formError.value = 'All fields are required'
    return
  }
  
  // Basic format validation
  if (!backend.startsWith('http://') && !backend.startsWith('https://') && !backend.startsWith('/')) {
    formError.value = 'Core service URL must start with http:// or https:// (or use relative path /)'
    return
  }
  if (!passport.startsWith('http://') && !passport.startsWith('https://') && !passport.startsWith('/')) {
    formError.value = 'Passport URL must start with http:// or https:// (or use relative path /)'
    return
  }
  
  try {
    serverStore.addCustomServer({
      name,
      backendUrl: backend,
      passportUrl: passport
    })
    
    // Reset form
    newServerName.value = ''
    newBackendUrl.value = ''
    newPassportUrl.value = ''
    alert(`Custom server [${name}] added successfully!`)
  } catch (err: any) {
    formError.value = err.message || 'Failed to add custom server. The name might already exist.'
  }
}

const handleRemoveServer = (serverName: string) => {
  const confirmed = confirm(`Are you sure you want to delete custom server [${serverName}]?`)
  if (confirmed) {
    serverStore.removeCustomServer(serverName)
  }
}

onMounted(() => {
  fetchProfile()
})
</script>

<template>
  <div class="flex flex-col gap-8 animate-fade-in pb-16">
    <div class="page-header">
      <h1 class="text-3xl font-bold text-text-primary mb-2 tracking-tight">User Profile & Settings</h1>
      <p class="text-text-secondary text-base">Manage your account credentials and system connection settings.</p>
    </div>

    <!-- PROFILE SECTION CARD -->
    <div>
      <div v-if="isLoading" class="flex flex-col items-center justify-center p-16 bg-bg-secondary rounded-xl shadow-card gap-4">
        <div class="w-10 h-10 border-3 border-border-color border-top-primary-blue rounded-full animate-spin"></div>
        <p class="text-text-secondary">Loading profile...</p>
      </div>

      <div v-else-if="error" class="flex flex-col items-center justify-center p-12 bg-bg-secondary rounded-xl shadow-card gap-4 text-red-500 border border-red-500/10">
        <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
        <p class="font-medium text-center">Unable to connect to the active auth server</p>
        <p class="text-sm text-text-secondary -mt-2 max-w-md text-center">{{ error }}</p>
        <button @click="fetchProfile" class="px-5 py-2 bg-primary-blue text-white rounded-lg border-none cursor-pointer font-medium transition-all hover:bg-primary-blue-hover shadow-sm">Retry Connection</button>
      </div>

      <div v-else-if="profile" class="bg-bg-secondary rounded-xl shadow-card overflow-hidden">
        <div class="p-10 bg-gradient-to-br from-bg-primary to-bg-secondary flex items-center gap-6 border-b border-border-color">
          <div class="relative w-20 h-20 rounded-full cursor-pointer overflow-hidden shadow-md group" @click="triggerAvatarUpload">
            <input 
              type="file" 
              ref="avatarInput" 
              class="hidden" 
              accept="image/*" 
              @change="handleAvatarUpload" 
            />
            <img 
              v-if="profile.avatarUrl" 
              :src="profile.avatarUrl" 
              alt="User Avatar" 
              class="w-full h-full object-cover block" 
            />
            <div v-else class="w-full h-full bg-primary-blue text-white flex items-center justify-center text-3xl font-bold">
              {{ profile.username.charAt(0).toUpperCase() }}
            </div>
            <div class="absolute inset-0 bg-black/50 text-white flex items-center justify-center text-xs font-semibold opacity-0 group-hover:opacity-100 transition-opacity" :class="{ 'opacity-100 bg-black/70': isUploadingAvatar }">
              <span v-if="isUploadingAvatar" class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
              <span v-else>Update</span>
            </div>
          </div>
          <div>
            <h2 class="text-2xl font-bold text-text-primary mb-1">{{ profile.username }}</h2>
            <span class="inline-block px-3 py-1 bg-bg-secondary text-primary-blue text-[0.7rem] font-bold rounded-full border border-border-color uppercase tracking-wider">{{ profile.role || 'User' }}</span>
          </div>
        </div>

        <div class="p-10 grid gap-8">
          <div class="flex flex-col gap-1">
            <label class="text-[0.7rem] font-bold uppercase tracking-wider text-text-secondary">Email Address</label>
            <div class="text-base text-text-primary font-medium">{{ profile.email }}</div>
          </div>
          
          <div class="flex flex-col gap-1">
            <label class="text-[0.7rem] font-bold uppercase tracking-wider text-text-secondary">User ID</label>
            <div class="text-base text-text-primary font-medium">#{{ profile.id }}</div>
          </div>

          <div class="flex flex-col gap-1">
            <label class="text-[0.7rem] font-bold uppercase tracking-wider text-text-secondary">Account Status</label>
            <div class="flex items-center gap-2 text-base text-text-primary font-medium">
              <span class="w-2 h-2 rounded-full" :class="profile.status?.toLowerCase() === 'active' || !profile.status ? 'bg-green-500 ring-4 ring-green-500/10' : 'bg-gray-400'"></span>
              {{ profile.status || 'Active' }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- BACKEND SERVICES CONFIGURATION CARD (Always visible for easy switching) -->
    <div class="bg-bg-secondary rounded-xl shadow-card overflow-hidden">
      <div class="p-8 border-b border-border-color bg-gradient-to-br from-bg-primary to-bg-secondary">
        <h2 class="text-xl font-bold text-text-primary flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-primary-blue">
            <rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect>
            <rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect>
            <line x1="6" y1="6" x2="6.01" y2="6"></line>
            <line x1="6" y1="18" x2="6.01" y2="18"></line>
          </svg>
          Backend Connection Settings
        </h2>
        <p class="text-text-secondary text-sm mt-1">Configure the backend API endpoints connected by the frontend. If the active service is unreachable or you need to connect to another cluster, select or add a server environment below.</p>
      </div>

      <div class="p-8 flex flex-col gap-8">
        <!-- Current active highlights -->
        <div class="p-5 rounded-lg bg-blue-50/50 dark:bg-blue-500/5 border border-blue-500/10 flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
          <div>
            <div class="text-[0.7rem] font-bold uppercase tracking-wider text-blue-600 dark:text-primary-blue mb-1">Current Active Server</div>
            <div class="text-lg font-bold text-text-primary flex items-center gap-2">
              {{ serverStore.currentServer.name }}
              <span class="w-2.5 h-2.5 rounded-full bg-green-500 ring-4 ring-green-500/10 animate-pulse" title="Active Connection"></span>
            </div>
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs font-mono w-full md:w-auto">
            <div class="p-3 bg-bg-primary rounded border border-border-color">
              <span class="text-text-secondary block font-sans font-semibold text-[0.65rem] uppercase mb-1">Backend Service URL (backendUrl)</span>
              <span class="text-text-primary font-medium">{{ serverStore.currentServer.backendUrl }}</span>
            </div>
            <div class="p-3 bg-bg-primary rounded border border-border-color">
              <span class="text-text-secondary block font-sans font-semibold text-[0.65rem] uppercase mb-1">Passport Auth URL (passportUrl)</span>
              <span class="text-text-primary font-medium">{{ serverStore.currentServer.passportUrl }}</span>
            </div>
          </div>
        </div>

        <!-- Servers list -->
        <div>
          <h3 class="text-sm font-bold uppercase tracking-wider text-text-secondary mb-4">Available Environments List</h3>
          <div class="border border-border-color rounded-lg overflow-hidden bg-bg-primary">
            <div v-for="srv in serverStore.allServers" :key="srv.name" class="p-5 border-b border-border-color last:border-0 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 transition-all hover:bg-bg-secondary/40">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-1">
                  <span class="font-bold text-text-primary text-base">{{ srv.name }}</span>
                  <span v-if="srv.name === serverStore.currentServer.name" class="px-2 py-0.5 text-[0.65rem] font-bold bg-green-500/10 text-green-600 dark:text-green-400 rounded-full border border-green-500/20">Active</span>
                  <span v-if="srv.isCustom" class="px-2 py-0.5 text-[0.65rem] font-bold bg-yellow-500/10 text-yellow-600 dark:text-yellow-400 rounded-full border border-yellow-500/20">Custom</span>
                  <span v-else class="px-2 py-0.5 text-[0.65rem] font-bold bg-blue-500/10 text-blue-600 dark:text-primary-blue rounded-full border border-blue-500/20">Predefined</span>
                </div>
                <div class="text-xs text-text-secondary font-mono flex flex-col gap-1 mt-2">
                  <div class="flex gap-2">
                    <span class="font-sans font-semibold inline-block w-24">Backend API:</span>
                    <span class="text-text-primary">{{ srv.backendUrl }}</span>
                  </div>
                  <div class="flex gap-2">
                    <span class="font-sans font-semibold inline-block w-24">Auth Center:</span>
                    <span class="text-text-primary">{{ srv.passportUrl }}</span>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-2 self-end sm:self-center">
                <button 
                  v-if="srv.name !== serverStore.currentServer.name"
                  @click="handleSwitchServer(srv.name)"
                  class="px-4 py-2 text-xs font-semibold bg-bg-secondary text-text-primary border border-border-color rounded-lg cursor-pointer hover:bg-bg-primary transition-all flex items-center gap-1 shadow-sm"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="23 4 23 10 17 10"></polyline>
                    <polyline points="1 20 1 14 7 14"></polyline>
                    <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
                  </svg>
                  Switch Server
                </button>
                <button 
                  v-else
                  disabled
                  class="px-4 py-2 text-xs font-semibold bg-green-500/5 text-green-600 border border-green-500/10 rounded-lg cursor-default flex items-center gap-1"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="20 6 9 17 4 12"></polyline>
                  </svg>
                  Active
                </button>
                <button 
                  v-if="srv.isCustom"
                  @click="handleRemoveServer(srv.name)"
                  class="p-2 text-red-500 bg-transparent border border-transparent rounded-lg cursor-pointer hover:bg-red-50 dark:hover:bg-red-500/10 hover:border-red-500/15 transition-all"
                  title="Delete Server"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="3 6 5 6 21 6"></polyline>
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                    <line x1="10" y1="11" x2="10" y2="17"></line>
                    <line x1="14" y1="11" x2="14" y2="17"></line>
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Add custom server form -->
        <div class="border border-border-color rounded-lg p-6 bg-bg-primary">
          <h3 class="text-sm font-bold uppercase tracking-wider text-text-primary mb-4 flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-primary-blue">
              <line x1="12" y1="5" x2="12" y2="19"></line>
              <line x1="5" y1="12" x2="19" y2="12"></line>
            </svg>
            Add Custom Backend Server
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div class="flex flex-col gap-1.5">
              <label class="text-xs font-semibold text-text-secondary">Environment Name</label>
              <input 
                type="text" 
                v-model="newServerName"
                placeholder="e.g. Production Cluster B" 
                class="px-4 py-2.5 rounded-lg border border-border-color bg-bg-secondary text-text-primary text-sm focus:outline-none focus:border-primary-blue"
              />
            </div>
            <div class="flex flex-col gap-1.5">
              <label class="text-xs font-semibold text-text-secondary">Core Service Base URL (backendUrl)</label>
              <input 
                type="text" 
                v-model="newBackendUrl"
                placeholder="e.g. http://192.168.1.100:8080/api/v1/sev" 
                class="px-4 py-2.5 rounded-lg border border-border-color bg-bg-secondary text-text-primary text-sm focus:outline-none focus:border-primary-blue font-mono"
              />
            </div>
            <div class="flex flex-col gap-1.5">
              <label class="text-xs font-semibold text-text-secondary">Passport Auth Base URL (passportUrl)</label>
              <input 
                type="text" 
                v-model="newPassportUrl"
                placeholder="e.g. http://192.168.1.100:8089/api/v1" 
                class="px-4 py-2.5 rounded-lg border border-border-color bg-bg-secondary text-text-primary text-sm focus:outline-none focus:border-primary-blue font-mono"
              />
            </div>
          </div>

          <div v-if="formError" class="mt-4 text-xs text-red-500 font-medium flex items-center gap-1.5">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
            {{ formError }}
          </div>

          <div class="mt-6 flex justify-end">
            <button 
              @click="handleAddServer"
              class="px-5 py-2.5 bg-primary-blue text-white rounded-lg border-none cursor-pointer font-semibold text-sm transition-all hover:bg-primary-blue-hover shadow-sm flex items-center gap-1.5"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path>
                <polyline points="17 21 17 13 7 13 7 21"></polyline>
                <polyline points="7 3 7 8 15 8"></polyline>
              </svg>
              Save Environment
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.border-top-primary-blue {
  border-top-color: var(--color-primary-blue);
}
</style>


