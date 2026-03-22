<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter, useRoute } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

const navItems = [
  { name: 'Overview', path: '/' },
  // Future items:
  // { name: 'Stats', path: '/stats' },
  // { name: 'Logs', path: '/logs' },
]
</script>

<template>
  <div class="layout-container">
    <!-- Top Navigation Bar -->
    <header class="top-nav">
      <div class="logo">SvcWatch</div>
      <div class="user-actions">
        <span class="user-greeting">Hello, {{ authStore.user?.username || authStore.user?.email || 'Admin' }}</span>
        <button @click="handleLogout" class="logout-btn" title="Logout">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
            <polyline points="16 17 21 12 16 7"></polyline>
            <line x1="21" y1="12" x2="9" y2="12"></line>
          </svg>
        </button>
      </div>
    </header>

    <div class="main-body">
      <!-- Left Sidebar -->
      <aside class="sidebar">
        <nav class="side-nav">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="nav-link"
            :class="{ active: route.path === item.path }"
          >
            {{ item.name }}
          </router-link>
        </nav>
      </aside>

      <!-- Main Content Area -->
      <main class="content-area">
        <div class="content-wrapper">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.layout-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  background-color: #f8fafc;
}

/* Top Navigation Bar */
.top-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 64px;
  padding: 0 1.5rem;
  background-color: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  z-index: 20;
}

.logo {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1e40af; /* Primary Blue */
  letter-spacing: -0.025em;
}

.user-actions {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.user-greeting {
  font-size: 0.875rem;
  color: #475569;
  font-weight: 500;
}

.logout-btn {
  background: none;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.logout-btn:hover {
  color: #ef4444;
  background-color: #fef2f2;
}

/* Main Body Layout */
.main-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* Left Sidebar */
.sidebar {
  width: 240px;
  background-color: #ffffff;
  border-right: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  z-index: 10;
}

.side-nav {
  padding: 1.5rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.nav-link {
  display: block;
  padding: 0.75rem 1rem;
  color: #475569;
  text-decoration: none;
  border-radius: 8px;
  font-weight: 500;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.nav-link:hover {
  background-color: #f1f5f9;
  color: #1e293b;
}

.nav-link.active {
  background-color: #eff6ff; /* Light Blue BG */
  color: #2563eb; /* Primary Blue Text */
}

/* Content Area */
.content-area {
  flex: 1;
  overflow-y: auto;
  background-color: #f8fafc;
}

.content-wrapper {
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
}
</style>
