<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { passportApi, uploadAvatar } from '@/services/api'

interface UserProfile {
  id: number
  username: string
  email: string
  role: string
  status: string
  avatarUrl?: string
  // Add other fields as per the actual API response
}

const profile = ref<UserProfile | null>(null)
const isLoading = ref(true)
const isUploadingAvatar = ref(false)
const error = ref('')
const avatarInput = ref<HTMLInputElement | null>(null)

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

  // Basic validation
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
      // Assuming the backend returns the full user profile or the new avatar URL
      if (profile.value) {
        // We'll refetch the profile to get the updated data
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
    // Reset the input so the same file could be selected again if needed
    if (target) {
      target.value = ''
    }
  }
}

onMounted(() => {
  fetchProfile()
})
</script>

<template>
  <div class="flex flex-col gap-8 animate-fade-in">
    <div class="page-header">
      <h1 class="text-3xl font-bold text-text-primary mb-2 tracking-tight">User Profile</h1>
      <p class="text-text-secondary text-base">Manage your account settings and view your information.</p>
    </div>

    <div v-if="isLoading" class="flex flex-col items-center justify-center p-16 bg-bg-secondary rounded-xl shadow-card gap-4">
      <div class="w-10 h-10 border-3 border-border-color border-top-primary-blue rounded-full animate-spin"></div>
      <p class="text-text-secondary">Loading profile...</p>
    </div>

    <div v-else-if="error" class="flex flex-col items-center justify-center p-16 bg-bg-secondary rounded-xl shadow-card gap-4 text-red-600">
      <p>{{ error }}</p>
      <button @click="fetchProfile" class="px-4 py-2 bg-primary-blue text-white rounded-lg border-none cursor-pointer font-medium transition-all hover:bg-primary-blue-hover">Retry</button>
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
</template>

<style scoped>
/* Spinner border-top utility if needed, but Tailwind 4 might need a custom class if not standard */
.border-top-primary-blue {
  border-top-color: var(--color-primary-blue);
}
</style>
