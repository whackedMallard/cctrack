import { ref, watch, type Ref } from 'vue'

export function useCountUp(target: Ref<number>, duration = 800) {
  const display = ref(0)
  let raf: number

  watch(target, (to) => {
    const from = display.value
    const start = performance.now()
    cancelAnimationFrame(raf)

    function tick(now: number) {
      const elapsed = now - start
      const progress = Math.min(elapsed / duration, 1)
      const ease = 1 - Math.pow(1 - progress, 4) // easeOutQuart
      display.value = from + (to - from) * ease
      if (progress < 1) raf = requestAnimationFrame(tick)
    }

    raf = requestAnimationFrame(tick)
  }, { immediate: true })

  return display
}
