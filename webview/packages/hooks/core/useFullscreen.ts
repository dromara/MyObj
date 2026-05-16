import { ref, onUnmounted } from 'vue'

export function useFullscreen() {
  const isFullscreen = ref(false)

  const onChange = () => {
    isFullscreen.value = !!document.fullscreenElement
  }

  document.addEventListener('fullscreenchange', onChange)
  onUnmounted(() => document.removeEventListener('fullscreenchange', onChange))

  const toggle = async () => {
    if (!document.fullscreenElement) {
      await document.documentElement.requestFullscreen()
    } else {
      await document.exitFullscreen()
    }
  }

  return { isFullscreen, toggle }
}
