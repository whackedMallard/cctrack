<template>
  <button
    class="toggle"
    :class="{ on: modelValue }"
    role="switch"
    :aria-checked="modelValue"
    @click="$emit('update:modelValue', !modelValue)"
  >
    <div class="toggle-thumb"></div>
  </button>
</template>

<script setup lang="ts">
defineProps<{ modelValue: boolean }>()
defineEmits<{ 'update:modelValue': [value: boolean] }>()
</script>

<style scoped>
.toggle {
  width: 36px;
  height: 20px;
  border-radius: 10px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  cursor: pointer;
  position: relative;
  transition: background 150ms, border-color 150ms;
  padding: 0;
}
.toggle.on {
  background: var(--amber-500);
  border-color: var(--amber-600);
}
.toggle-thumb {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--text-primary);
  position: absolute;
  top: 2px;
  left: 2px;
  transition: transform 150ms;
}
.toggle.on .toggle-thumb {
  transform: translateX(16px);
}
</style>
