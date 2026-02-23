import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Settings } from '../types'
import { fetchSettings, updateSettings } from '../api'

export const useSettingsStore = defineStore('settings', () => {
  const current = ref<Settings | null>(null)
  const draft = ref<Partial<Settings>>({})
  const saving = ref(false)
  const saved = ref(false)

  const isDirty = computed(() => {
    if (!current.value) return false
    return Object.keys(draft.value).some(
      key => (draft.value as any)[key] !== (current.value as any)[key]
    )
  })

  async function load() {
    current.value = await fetchSettings()
    draft.value = { ...current.value }
  }

  async function save() {
    saving.value = true
    saved.value = false
    try {
      current.value = await updateSettings(draft.value)
      draft.value = { ...current.value }
      saved.value = true
      setTimeout(() => { saved.value = false }, 2000)
    } finally {
      saving.value = false
    }
  }

  return { current, draft, saving, saved, isDirty, load, save }
})
