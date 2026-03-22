<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { passportApi } from '@/services/api'

interface UserProfile {
  id: number
  username: string
  email: string
  role: string
  status: string
  // Add other fields as per the actual API response
}

const profile = ref<UserProfile | null>(null)
const isLoading = ref(true)
const error = ref('')

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

onMounted(() => {
  fetchProfile()
})
</script>

<template>
  <div class="profile-container">
    <div class="page-header">
      <h1>User Profile</h1>
      <p>Manage your account settings and view your information.</p>
    </div>

    <div v-if="isLoading" class="state-container">
      <div class="loading-spinner"></div>
      <p>Loading profile...</p>
    </div>

    <div v-else-if="error" class="state-container error">
      <p>{{ error }}</p>
      <button @click="fetchProfile" class="retry-btn">Retry</button>
    </div>

    <div v-else-if="profile" class="profile-card">
      <div class="card-header">
        <div class="avatar-placeholder">
          {{ profile.username.charAt(0).toUpperCase() }}
        </div>
        <div class="header-info">
          <h2>{{ profile.username }}</h2>
          <span class="role-badge">{{ profile.role || 'User' }}</span>
        </div>
      </div>

      <div class="card-body">
        <div class="info-group">
          <label>Email Address</label>
          <div class="info-value">{{ profile.email }}</div>
        </div>
        
        <div class="info-group">
          <label>User ID</label>
          <div class="info-value">#{{ profile.id }}</div>
        </div>

        <div class="info-group">
          <label>Account Status</label>
          <div class="info-value status">
            <span class="status-dot" :class="profile.status?.toLowerCase() || 'active'"></span>
            {{ profile.status || 'Active' }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-container {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.page-header h1 {
  font-size: 1.875rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.page-header p {
  color: var(--text-secondary);
  font-size: 1rem;
}

.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem;
  background: var(--bg-secondary);
  border-radius: 12px;
  box-shadow: var(--card-shadow);
  gap: 1rem;
}

.state-container.error {
  color: #dc2626;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-color);
  border-top: 3px solid var(--primary-blue);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.retry-btn {
  padding: 0.5rem 1rem;
  background-color: var(--primary-blue);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.profile-card {
  background: var(--bg-secondary);
  border-radius: 12px;
  box-shadow: var(--card-shadow);
  overflow: hidden;
}

.card-header {
  padding: 2.5rem;
  background: linear-gradient(135deg, var(--bg-primary) 0%, var(--bg-secondary) 100%);
  display: flex;
  align-items: center;
  gap: 1.5rem;
  border-bottom: 1px solid var(--border-color);
}

.avatar-placeholder {
  width: 80px;
  height: 80px;
  background-color: var(--primary-blue);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  font-weight: 700;
}

.header-info h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.25rem;
}

.role-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background-color: var(--bg-secondary);
  color: var(--primary-blue);
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: 9999px;
  border: 1px solid var(--border-color);
}

.card-body {
  padding: 2.5rem;
  display: grid;
  gap: 1.5rem;
}

.info-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.info-group label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
}

.info-value {
  font-size: 1rem;
  color: var(--text-primary);
  font-weight: 500;
}

.info-value.status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.active {
  background-color: #22c55e;
  box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.1);
}
</style>
