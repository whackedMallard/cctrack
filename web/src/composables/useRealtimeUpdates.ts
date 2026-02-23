import { ref, onMounted, onUnmounted } from 'vue'
import { useDashboardStore } from '../stores/dashboard'
import type { ConnectionStatus } from '../types'

export function useRealtimeUpdates() {
  const store = useDashboardStore()
  const status = ref<ConnectionStatus>('reconnecting')
  let ws: WebSocket | null = null
  let retryDelay = 1000
  let closed = false

  function connect() {
    if (closed) return

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = location.host
    ws = new WebSocket(`${protocol}//${host}/api/v1/ws`)

    ws.onopen = () => {
      status.value = 'connected'
      retryDelay = 1000
    }

    ws.onmessage = (msg) => {
      try {
        const event = JSON.parse(msg.data)
        store.applyEvent(event)
      } catch (e) {
        // ignore malformed messages
      }
    }

    ws.onclose = () => {
      if (closed) return
      status.value = retryDelay >= 16000 ? 'offline' : 'reconnecting'
      setTimeout(() => {
        retryDelay = Math.min(retryDelay * 2, 30000)
        connect()
      }, retryDelay)
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  onMounted(connect)

  onUnmounted(() => {
    closed = true
    ws?.close()
  })

  return { status }
}
