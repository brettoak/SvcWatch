<script setup lang="ts">
import { ref, h, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { useRouter, useRoute } from 'vue-router'

const authStore = useAuthStore()
const themeStore = useThemeStore()
const router = useRouter()
const route = useRoute()

const isSidebarCollapsed = ref(false)

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

// Icons as render functions
const OverviewIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('rect', { x: '3', y: '3', width: '18', height: '18', rx: '2', ry: '2' }),
  h('line', { x1: '3', y1: '9', x2: '21', y2: '9' }),
  h('line', { x1: '9', y1: '21', x2: '9', y2: '9' }),
])

const ProfileIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('path', { d: 'M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2' }),
  h('circle', { cx: '12', cy: '7', r: '4' }),
])

const SunIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('circle', { cx: '12', cy: '12', r: '5' }),
  h('line', { x1: '12', y1: '1', x2: '12', y2: '3' }),
  h('line', { x1: '12', y1: '21', x2: '12', y2: '23' }),
  h('line', { x1: '4.22', y1: '4.22', x2: '5.64', y2: '5.64' }),
  h('line', { x1: '18.36', y1: '18.36', x2: '19.78', y2: '19.78' }),
  h('line', { x1: '1', y1: '12', x2: '3', y2: '12' }),
  h('line', { x1: '21', y1: '12', x2: '23', y2: '12' }),
  h('line', { x1: '4.22', y1: '19.78', x2: '5.64', y2: '18.36' }),
  h('line', { x1: '18.36', y1: '5.64', x2: '19.78', y2: '4.22' }),
])

const MoonIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('path', { d: 'M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z' }),
])

const navItems = [
  { name: 'Overview', path: '/', icon: OverviewIcon },
  { name: 'Profile', path: '/profile', icon: ProfileIcon },
]

const currentThemeIcon = computed(() => themeStore.isDark ? SunIcon : MoonIcon)
</script>

<template>
  <div class="layout-container">
    <!-- Top Navigation Bar -->
    <header class="top-nav">
      <div class="header-left">
        <button @click="toggleSidebar" class="toggle-btn" title="Toggle Sidebar">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="3" y1="12" x2="21" y2="12"></line>
            <line x1="3" y1="6" x2="21" y2="6"></line>
            <line x1="3" y1="18" x2="21" y2="18"></line>
          </svg>
        </button>
        <div class="logo">SvcWatch</div>
      </div>
      <div class="user-actions">
        <button @click="themeStore.toggleTheme" class="theme-toggle" :title="themeStore.isDark ? 'Switch to Light Mode' : 'Switch to Dark Mode'">
          <component :is="currentThemeIcon" />
        </button>
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
      <aside class="sidebar" :class="{ collapsed: isSidebarCollapsed }">
        <nav class="side-nav">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="nav-link"
            :class="{ active: route.path === item.path }"
            :title="isSidebarCollapsed ? item.name : ''"
          >
            <span class="nav-icon-wrapper">
              <component :is="item.icon" />
            </span>
            <span class="nav-text" v-show="!isSidebarCollapsed">{{ item.name }}</span>
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
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

/* Top Navigation Bar */
.top-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 64px;
  padding: 0 1.5rem;
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  z-index: 20;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.toggle-btn {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.toggle-btn:hover {
  background-color: var(--bg-primary);
  color: var(--primary-blue);
}

.logo {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--primary-blue);
  letter-spacing: -0.025em;
}

.user-actions {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.theme-toggle {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.theme-toggle:hover {
  background-color: var(--bg-primary);
  color: var(--primary-blue);
}

.user-greeting {
  font-size: 0.875rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.logout-btn {
  background: none;
  border: none;
  color: var(--text-secondary);
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

.dark .logout-btn:hover {
  background-color: rgba(239, 68, 68, 0.1);
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
  background-color: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  z-index: 10;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar.collapsed {
  width: 72px;
}

.sidebar.collapsed .side-nav {
  padding: 1rem 0.5rem;
}

.sidebar.collapsed .nav-link {
  padding: 0.75rem 0;
  justify-content: center;
}

.side-nav {
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.nav-link {
  display: flex;
  align-items: center;
  padding: 0.75rem;
  color: var(--text-secondary);
  text-decoration: none;
  border-radius: 8px;
  font-weight: 500;
  font-size: 0.95rem;
  white-space: nowrap;
  transition: all 0.2s;
}

.nav-icon-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
}

.nav-text {
  margin-left: 0.25rem;
  transition: opacity 0.2s;
}

.sidebar.collapsed .nav-text {
  opacity: 0;
  pointer-events: none;
}

.nav-link:hover {
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

.nav-link.active {
  background-color: #eff6ff; /* Light theme active bg */
  color: #2563eb;
}

.dark .nav-link.active {
  background-color: rgba(59, 130, 246, 0.1);
  color: var(--primary-blue);
}

/* Content Area */
.content-area {
  flex: 1;
  overflow-y: auto;
  background-color: var(--bg-primary);
}

.content-wrapper {
  padding: 2rem;
}
</style>

