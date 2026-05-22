<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

const props = defineProps({
  modelValue: {
    type: String,
    default: '1h'
  },
  // Allows customizing the background of the button bar
  bgClass: {
    type: String,
    default: 'bg-bg-secondary border border-border-color shadow-sm'
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const timeOptions = [
  { label: '5m', value: '5m' },
  { label: '30m', value: '30m' },
  { label: '1h', value: '1h' },
  { label: '6h', value: '6h' },
  { label: '24h', value: '24h' },
  { label: '7d', value: '7d' },
  { label: '30d', value: '30d' },
  { label: 'Custom', value: 'custom' },
]

const customStartTime = ref('')
const customEndTime = ref('')

const toLocalISO = (d: Date) => {
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// Format date back to a clean string format for datetime-local
const formatToDateTimeLocal = (date: Date) => {
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

onMounted(() => {
  const end = new Date()
  const start = new Date(end)
  // Default custom range is 1 hour
  start.setHours(start.getHours() - 1)
  customEndTime.value = toLocalISO(end)
  customStartTime.value = toLocalISO(start)

  // Trigger initial calculation
  calculateAndEmit()
})

const calculateAndEmit = () => {
  const end = new Date()
  const start = new Date(end)

  if (props.modelValue === 'custom') {
    if (!customStartTime.value || !customEndTime.value) return
    const startD = new Date(customStartTime.value)
    const endD = new Date(customEndTime.value)
    emit('change', {
      startStr: startD.toISOString(),
      endStr: endD.toISOString(),
      localStart: formatToDateTimeLocal(startD),
      localEnd: formatToDateTimeLocal(endD),
      filter: 'custom'
    })
    return
  }

  switch (props.modelValue) {
    case '5m': start.setMinutes(start.getMinutes() - 5); break
    case '30m': start.setMinutes(start.getMinutes() - 30); break
    case '1h': start.setHours(start.getHours() - 1); break
    case '6h': start.setHours(start.getHours() - 6); break
    case '24h': start.setHours(start.getHours() - 24); break
    case '7d': start.setDate(start.getDate() - 7); break
    case '30d': start.setDate(start.getDate() - 30); break
  }

  emit('change', {
    startStr: start.toISOString(),
    endStr: end.toISOString(),
    localStart: formatToDateTimeLocal(start),
    localEnd: formatToDateTimeLocal(end),
    filter: props.modelValue
  })
}

// Watch modelValue changes (e.g. from parent or clicks)
watch(() => props.modelValue, (newVal) => {
  if (newVal !== 'custom') {
    calculateAndEmit()
  }
})

const selectFilter = (val: string) => {
  emit('update:modelValue', val)
}

const handleCustomSearch = () => {
  calculateAndEmit()
}
</script>

<template>
  <div class="flex flex-col gap-2">
    <div class="flex flex-wrap items-center gap-3">
      <!-- Button Filter bar -->
      <div class="flex rounded-xl p-1 shadow-sm shrink-0" :class="bgClass">
        <button 
          v-for="opt in timeOptions" 
          :key="opt.value"
          type="button"
          class="px-3.5 py-1.5 rounded-lg text-xs font-bold cursor-pointer transition-all duration-200 border-none outline-none"
          :class="modelValue === opt.value ? 'bg-primary-blue text-white shadow-md shadow-primary-blue/30' : 'bg-transparent text-text-secondary hover:text-text-primary'"
          @click="selectFilter(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>

      <!-- Inline custom time picker -->
      <div v-if="modelValue === 'custom'" class="flex items-center gap-2 bg-bg-secondary p-1.5 rounded-xl shadow-sm border border-border-color animate-fade-in flex-wrap">
        <input 
          type="datetime-local" 
          v-model="customStartTime" 
          step="1"
          class="bg-transparent border border-border-color text-text-primary px-2.5 py-1.5 rounded-md text-xs outline-none transition-all focus:border-primary-blue" 
        />
        <span class="text-text-secondary text-xs font-bold">to</span>
        <input 
          type="datetime-local" 
          v-model="customEndTime" 
          step="1"
          class="bg-transparent border border-border-color text-text-primary px-2.5 py-1.5 rounded-md text-xs outline-none transition-all focus:border-primary-blue" 
        />
        <button 
          type="button"
          class="bg-primary-blue text-white border-none py-1.5 px-4 rounded-md text-xs font-bold cursor-pointer transition-all hover:brightness-110 hover:-translate-y-px active:scale-95 shadow-sm"
          @click="handleCustomSearch"
        >
          Search
        </button>
      </div>
    </div>
  </div>
</template>
