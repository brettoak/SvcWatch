<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useThemeStore } from '@/stores/theme'
import { watchEffect } from 'vue'

const themeStore = useThemeStore()

// Dynamically apply/remove the 'dark' class on the html element
watchEffect(() => {
  if (themeStore.isDark) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
})
</script>

<template>
  <RouterView />
</template>
