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

const LogsIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('line', { x1: '8', y1: '6', x2: '21', y2: '6' }),
  h('line', { x1: '8', y1: '12', x2: '21', y2: '12' }),
  h('line', { x1: '8', y1: '18', x2: '21', y2: '18' }),
  h('line', { x1: '3', y1: '6', x2: '3.01', y2: '6' }),
  h('line', { x1: '3', y1: '12', x2: '3.01', y2: '12' }),
  h('line', { x1: '3', y1: '18', x2: '3.01', y2: '18' }),
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
  { name: 'Logs', path: '/logs', icon: LogsIcon },
  { name: 'Profile', path: '/profile', icon: ProfileIcon },
]

const currentThemeIcon = computed(() => themeStore.isDark ? SunIcon : MoonIcon)
</script>

<template>
  <div class="flex flex-col h-screen overflow-hidden bg-bg-primary text-text-primary">
    <!-- Top Navigation Bar -->
    <header class="flex justify-between items-center h-16 px-6 bg-bg-secondary border-b border-border-color shadow-sm z-20">
      <div class="flex items-center gap-4">
        <button @click="toggleSidebar" class="bg-transparent border-none text-text-secondary cursor-pointer p-2 rounded-md flex items-center justify-center transition-all duration-200 hover:bg-bg-primary hover:text-primary-blue" title="Toggle Sidebar">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="3" y1="12" x2="21" y2="12"></line>
            <line x1="3" y1="6" x2="21" y2="6"></line>
            <line x1="3" y1="18" x2="21" y2="18"></line>
          </svg>
        </button>
        <div class="text-2xl font-bold text-primary-blue tracking-tight">SvcWatch</div>
      </div>
      <div class="flex items-center gap-6">
        <button @click="themeStore.toggleTheme" class="bg-transparent border-none text-text-secondary cursor-pointer p-2 rounded-md flex items-center justify-center transition-all duration-200 hover:bg-bg-primary hover:text-primary-blue" :title="themeStore.isDark ? 'Switch to Light Mode' : 'Switch to Dark Mode'">
          <component :is="currentThemeIcon" />
        </button>
        <span class="text-sm text-text-secondary font-medium">Hello, {{ authStore.user?.username || authStore.user?.email || 'Admin' }}</span>
        <button @click="handleLogout" class="bg-transparent border-none text-text-secondary cursor-pointer p-2 rounded-md flex items-center justify-center transition-all duration-200 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-500/10" title="Logout">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
            <polyline points="16 17 21 12 16 7"></polyline>
            <line x1="21" y1="12" x2="9" y2="12"></line>
          </svg>
        </button>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden">
      <!-- Left Sidebar -->
      <aside class="flex flex-col border-r border-border-color bg-bg-secondary z-10 transition-[width] duration-300 ease-in-out" :class="isSidebarCollapsed ? 'w-[72px]' : 'w-60'">
        <nav class="flex flex-col gap-2 p-4" :class="{ 'px-2': isSidebarCollapsed }">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="flex items-center p-3 text-text-secondary no-underline rounded-lg font-medium text-[0.95rem] whitespace-nowrap transition-all duration-200 hover:bg-bg-primary hover:text-text-primary"
            :class="[
              route.path === item.path ? 'bg-blue-50 text-blue-600 dark:bg-blue-500/10 dark:text-primary-blue' : '',
              isSidebarCollapsed ? 'justify-center py-3' : ''
            ]"
            :title="isSidebarCollapsed ? item.name : ''"
          >
            <span class="flex items-center justify-center min-w-[40px]">
              <component :is="item.icon" />
            </span>
            <span class="ml-1 transition-opacity duration-200" v-show="!isSidebarCollapsed">{{ item.name }}</span>
          </router-link>
        </nav>
      </aside>

      <!-- Main Content Area -->
      <main class="flex-1 overflow-y-auto bg-bg-primary">
        <div class="p-8">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* Only keeping structural-only or non-tailwind logic if absolutely needed, but here we can remove all */
</style>

