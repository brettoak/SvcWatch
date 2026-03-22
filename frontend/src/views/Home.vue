<script setup lang="ts">
// Logic for the dashboard overview will go here
</script>

<template>
  <div class="home-content">
    <div class="welcome-card">
      <h1>Welcome to SvcWatch Dashboard</h1>
      <p>Your service monitoring overview will be displayed here.</p>
    </div>
  </div>
</template>

<style scoped>
.home-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.welcome-card {
  background-color: var(--bg-secondary);
  border-radius: 12px;
  padding: 2.5rem;
  box-shadow: var(--card-shadow);
  text-align: center;
}

.welcome-card h1 {
  color: var(--text-primary);
  font-size: 1.875rem;
  font-weight: 700;
  margin-bottom: 0.75rem;
}

.welcome-card p {
  color: var(--text-secondary);
  font-size: 1.125rem;
}
</style>

