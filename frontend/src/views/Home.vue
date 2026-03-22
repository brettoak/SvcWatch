<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="home-container">
    <header class="header">
      <div class="logo">SvcWatch</div>
      <div class="user-info">
        <span>Logged in as: {{ authStore.user?.email || 'Admin' }}</span>
        <button @click="handleLogout" class="logout-button">Logout</button>
      </div>
    </header>
    <main class="content">
      <h1>Welcome to SvcWatch</h1>
      <p>Your service monitoring dashboard is ready.</p>
    </main>
  </div>
</template>

<style scoped>
.home-container {
  min-height: 100vh;
  background-color: #f8fafc;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  background: white;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
}

.logo {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1e40af;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.logout-button {
  padding: 0.5rem 1rem;
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.logout-button:hover {
  background-color: #dc2626;
}

.content {
  padding: 2rem;
  text-align: center;
}

.content h1 {
  color: #1e293b;
  margin-bottom: 1rem;
}

.content p {
  color: #64748b;
}
</style>
