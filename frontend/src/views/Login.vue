<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { passportApi } from '@/services/api'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const isLoading = ref(false)
const showPassword = ref(false)

const togglePassword = () => {
  showPassword.value = !showPassword.value
}

const handleLogin = async () => {
  if (!email.value || !password.value) {
    error.value = 'Please enter both email and password'
    return
  }

  error.value = ''
  isLoading.value = true

  try {
    const response = await passportApi.post('/auth/login', {
      email: email.value,
      password: password.value,
    })

    if (response.data && response.data.code === 200) {
      const { token, email: userEmail, username } = response.data.data
      authStore.setToken(token)
      authStore.setUser({ email: userEmail, username })
      router.push('/')
    } else {
      error.value = response.data.message || 'Login failed'
    }
  } catch (err: any) {
    if (err.response) {
      error.value = err.response.data?.message || `Server error: ${err.response.status}`
    } else if (err.request) {
      error.value = 'No response from server. Please check your connection and proxy settings.'
    } else {
      error.value = err.message || 'An error occurred during login'
    }
    console.error('Login error:', err)
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="flex justify-center items-center min-h-screen bg-bg-primary">
    <div class="w-full max-w-[400px] p-10 bg-bg-secondary rounded-xl shadow-card animate-fade-in">
      <div class="text-center mb-8">
        <h1 class="text-4xl font-bold text-primary-blue mb-2 tracking-tight">SvcWatch</h1>
        <p class="text-text-secondary text-sm">Sign in to your account</p>
      </div>

      <form @submit.prevent="handleLogin" class="flex flex-col gap-5">
        <div class="flex flex-col gap-2">
          <label for="email" class="text-sm font-medium text-text-secondary">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            placeholder="admin@example.com"
            :disabled="isLoading"
            class="w-full p-3 bg-bg-secondary border border-border-color text-text-primary rounded-lg text-base outline-none transition-all focus:border-primary-blue focus:ring-3 focus:ring-primary-blue/10 disabled:opacity-50"
          />
        </div>

        <div class="flex flex-col gap-2">
          <label for="password" class="text-sm font-medium text-text-secondary">Password</label>
          <div class="relative flex items-center">
            <input
              id="password"
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="••••••••"
              :disabled="isLoading"
              class="w-full p-3 pr-10 bg-bg-secondary border border-border-color text-text-primary rounded-lg text-base outline-none transition-all focus:border-primary-blue focus:ring-3 focus:ring-primary-blue/10 disabled:opacity-50"
            />
            <button
              type="button"
              class="absolute right-3 bg-transparent border-none p-0 cursor-pointer text-text-secondary flex items-center justify-center transition-colors hover:text-primary-blue disabled:opacity-50"
              @click="togglePassword"
              :disabled="isLoading"
              tabindex="-1"
            >
              <svg
                v-if="showPassword"
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M9.88 9.88l-3.29-3.29m7.53 7.53l3.29 3.29M3 3l18 18" />
                <path
                  d="M10.12 10.12a3 3 0 0 0 3.76 3.76M14.73 9.27A3 3 0 0 0 12.03 7.27"
                />
                <path
                  d="M19.62 12.58a10.63 10.63 0 0 0-4-6.35M16.27 16.27A10.75 10.75 0 0 1 12 18c-4.14 0-7.79-2.51-9.38-6.16"
                />
              </svg>
              <svg
                v-else
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                <circle cx="12" cy="12" r="3" />
              </svg>
            </button>
          </div>
        </div>

        <div v-if="error" class="text-sm text-red-600 bg-red-50 dark:bg-red-500/10 p-3 rounded-lg border border-red-100 dark:border-red-500/20">
          {{ error }}
        </div>

        <button type="submit" class="mt-2 p-3 bg-primary-blue text-white border-none rounded-lg text-base font-semibold cursor-pointer transition-all hover:bg-primary-blue-hover hover:-translate-y-0.5 active:translate-y-0 disabled:opacity-70 disabled:cursor-not-allowed disabled:transform-none" :disabled="isLoading">
          <span v-if="isLoading">Signing in...</span>
          <span v-else>Sign In</span>
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
</style>
