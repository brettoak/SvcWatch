<script setup lang="ts">
import { ref, h, computed, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { useRouter, useRoute } from 'vue-router'

const authStore = useAuthStore()
const themeStore = useThemeStore()
const router = useRouter()
const route = useRoute()

const isSidebarCollapsed = ref(false)
const isFullWidth = ref(localStorage.getItem('svcwatch-layout-width') === 'full')

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

const toggleContentWidth = () => {
  isFullWidth.value = !isFullWidth.value
}

watch(isFullWidth, (value) => {
  localStorage.setItem('svcwatch-layout-width', value ? 'full' : 'contained')
})

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

const ConnectionIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('rect', { x: '2', y: '2', width: '20', height: '8', rx: '2', ry: '2' }),
  h('rect', { x: '2', y: '14', width: '20', height: '8', rx: '2', ry: '2' }),
  h('line', { x1: '6', y1: '6', x2: '6.01', y2: '6' }),
  h('line', { x1: '6', y1: '18', x2: '6.01', y2: '18' }),
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

const MapIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('circle', { cx: '12', cy: '10', r: '3' }),
  h('path', { d: 'M12 21.7C17.3 17 20 13 20 10a8 8 0 1 0-16 0c0 3 2.7 7 8 11.7z' }),
])

const StatsIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('line', { x1: '18', y1: '20', x2: '18', y2: '10' }),
  h('line', { x1: '12', y1: '20', x2: '12', y2: '4' }),
  h('line', { x1: '6', y1: '20', x2: '6', y2: '14' }),
])

const TrendIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('polyline', { points: '23 6 13.5 15.5 8.5 10.5 1 18' }),
  h('polyline', { points: '17 6 23 6 23 12' })
])

const DatabaseIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('ellipse', { cx: '12', cy: '5', rx: '9', ry: '3' }),
  h('path', { d: 'M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5' }),
  h('path', { d: 'M3 12c0 1.66 4 3 9 3s9-1.34 9-3' }),
])

const MaximizeIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('polyline', { points: '15 3 21 3 21 9' }),
  h('polyline', { points: '9 21 3 21 3 15' }),
  h('line', { x1: '21', y1: '3', x2: '14', y2: '10' }),
  h('line', { x1: '3', y1: '21', x2: '10', y2: '14' }),
])

const MinimizeIcon = h('svg', { xmlns: 'http://www.w3.org/2000/svg', width: '20', height: '20', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [
  h('polyline', { points: '4 14 10 14 10 20' }),
  h('polyline', { points: '20 10 14 10 14 4' }),
  h('line', { x1: '14', y1: '10', x2: '21', y2: '3' }),
  h('line', { x1: '3', y1: '21', x2: '10', y2: '14' }),
])

const navGroups = [
  {
    group: 'Dashboards & Metrics',
    items: [
      { name: 'Overview', path: '/', icon: OverviewIcon },
      { name: 'Trends', path: '/trends', icon: TrendIcon },
    ]
  },
  {
    group: 'Analytics & Space',
    items: [
      { name: 'Top Stats', path: '/top-stats', icon: StatsIcon },
      { name: 'Map', path: '/map', icon: MapIcon },
    ]
  },
  {
    group: 'Diagnostics & Data',
    items: [
      { name: 'Logs', path: '/logs', icon: LogsIcon },
      { name: 'Sources', path: '/sources', icon: DatabaseIcon },
    ]
  },
  {
    group: 'Settings',
    items: [
      { name: 'Profile', path: '/profile', icon: ProfileIcon },
      { name: 'Connection', path: '/connection', icon: ConnectionIcon },
    ]
  }
]

const currentThemeIcon = computed(() => themeStore.isDark ? SunIcon : MoonIcon)
const currentWidthIcon = computed(() => isFullWidth.value ? MinimizeIcon : MaximizeIcon)
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
        <button @click="toggleContentWidth" class="bg-transparent border-none text-text-secondary cursor-pointer p-2 rounded-md flex items-center justify-center transition-all duration-200 hover:bg-bg-primary hover:text-primary-blue" :title="isFullWidth ? 'Use Contained Width' : 'Use Full Width'">
          <component :is="currentWidthIcon" />
        </button>
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
        <nav class="flex flex-col gap-6 p-4" :class="{ 'px-2': isSidebarCollapsed }">
          <div v-for="group in navGroups" :key="group.group" class="flex flex-col gap-1">
            <div 
              class="px-3 mb-1 text-xs font-semibold text-text-secondary opacity-70 uppercase tracking-wider transition-opacity duration-200" 
              v-show="!isSidebarCollapsed"
            >
              {{ group.group }}
            </div>
            <router-link
              v-for="item in group.items"
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
          </div>
        </nav>
      </aside>

      <!-- Main Content Area -->
      <main class="flex-1 overflow-y-auto bg-bg-primary">
        <div class="p-8 mx-auto w-full transition-[max-width] duration-300 ease-in-out" :class="isFullWidth ? 'max-w-none' : 'max-w-[1600px]'">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* Only keeping structural-only or non-tailwind logic if absolutely needed, but here we can remove all */
</style>
