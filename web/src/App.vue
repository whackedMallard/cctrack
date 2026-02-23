<template>
  <AppShell :connectionStatus="wsStatus">
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </AppShell>
</template>

<script setup lang="ts">
import AppShell from './components/layout/AppShell.vue'
import { useRealtimeUpdates } from './composables/useRealtimeUpdates'

const { status: wsStatus } = useRealtimeUpdates()
</script>
