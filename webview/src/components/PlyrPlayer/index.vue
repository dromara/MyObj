<template>
  <div class="plyr-player-wrapper">
    <video
      v-if="src"
      ref="videoElement"
      class="plyr-player"
      :playsinline="true"
      crossorigin="anonymous"
    ></video>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick, getCurrentInstance, computed } from 'vue'
// @ts-ignore - plyr 类型定义可能不完整
import Plyr from 'plyr'
import 'plyr/dist/plyr.css'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

// 根据当前语言获取 Plyr i18n 配置
const plyrI18n = computed(() => {
  const plyrKeys = [
    'restart', 'rewind', 'play', 'pause', 'fastForward', 'seek', 'seekLabel',
    'played', 'buffered', 'currentTime', 'duration', 'volume', 'mute', 'unmute',
    'enableCaptions', 'disableCaptions', 'download', 'enterFullscreen', 'exitFullscreen',
    'frameTitle', 'captions', 'settings', 'pip', 'menuBack', 'speed', 'normal',
    'quality', 'loop', 'start', 'end', 'all', 'reset', 'disabled', 'enabled', 'advertisement'
  ]
  
  const i18n: Record<string, any> = {}
  plyrKeys.forEach(key => {
    i18n[key] = t(`plyr.${key}`)
  })
  
  // 处理 qualityBadge
  i18n.qualityBadge = {
    '2160': t('plyr.qualityBadge.2160'),
    '1440': t('plyr.qualityBadge.1440'),
    '1080': t('plyr.qualityBadge.1080'),
    '720': t('plyr.qualityBadge.720'),
    '576': t('plyr.qualityBadge.576'),
    '480': t('plyr.qualityBadge.480')
  }
  
  return i18n
})

interface Props {
  src: string
  autoplay?: boolean
  loop?: boolean
  options?: Partial<Plyr.Options>
}

interface Emits {
  (e: 'ready'): void
  (e: 'error', error: string): void
  (e: 'play'): void
  (e: 'pause'): void
  (e: 'ended'): void
}

const props = withDefaults(defineProps<Props>(), {
  autoplay: false,
  loop: false,
  options: () => ({})
})

const emit = defineEmits<Emits>()

const videoElement = ref<HTMLVideoElement | null>(null)
let plyrInstance: Plyr | null = null

const proxy = getCurrentInstance()?.proxy

// 初始化 Plyr 播放器
const initPlyr = async () => {
  if (!videoElement.value || !props.src) return

  // 清理旧的实例
  if (plyrInstance) {
    plyrInstance.destroy()
    plyrInstance = null
  }

  await nextTick()

  // 初始化 Plyr（传入 video 元素）
  plyrInstance = new Plyr(videoElement.value, {
    autoplay: props.autoplay,
    loop: { active: props.loop },
    controls: ['play-large', 'play', 'progress', 'current-time', 'mute', 'volume', 'settings', 'pip', 'fullscreen'],
    settings: ['captions', 'quality', 'speed'],
    keyboard: { focused: true, global: false },
    tooltips: { controls: true, seek: true },
    clickToPlay: true,
    hideControls: true,
    resetOnEnd: false,
    i18n: plyrI18n.value, // 根据当前语言设置界面
    ...props.options // 允许外部传入自定义选项
  })

  // 通过 player.source 设置视频源（这是 Plyr 推荐的方式）
  plyrInstance.source = {
    type: 'video',
    sources: [
      {
        src: props.src,
        type: 'video/mp4' // 根据实际类型调整
      }
    ]
  }

  // 错误处理
  plyrInstance.on('error', (event: any) => {
    const detail = event.detail
    const errorMessage = detail?.message || t('plyr.playError')
    emit('error', errorMessage)
    proxy?.$log?.error('Plyr 播放错误', detail)
  })

  // Ready 事件
  plyrInstance.on('ready', () => {
    emit('ready')
  })

  // 播放事件
  plyrInstance.on('play', () => {
    emit('play')
  })

  // 暂停事件
  plyrInstance.on('pause', () => {
    emit('pause')
  })

  // 结束事件
  plyrInstance.on('ended', () => {
    emit('ended')
  })
}

// 清理 Plyr 实例
const destroyPlyr = () => {
  if (plyrInstance) {
    plyrInstance.destroy()
    plyrInstance = null
  }
}

// 监听 src 变化
watch(() => props.src, (newSrc) => {
  if (newSrc && plyrInstance && videoElement.value) {
    // 如果 Plyr 已初始化，更新 source
    plyrInstance.source = {
      type: 'video',
      sources: [
        {
          src: newSrc,
          type: 'video/mp4'
        }
      ]
    }
  } else if (newSrc) {
    // 如果还没有初始化，初始化 Plyr
    initPlyr()
  }
}, { immediate: false })

// 暴露方法给父组件
defineExpose({
  play: () => plyrInstance?.play(),
  pause: () => plyrInstance?.pause(),
  togglePlay: () => plyrInstance?.togglePlay(),
  stop: () => plyrInstance?.stop(),
  restart: () => plyrInstance?.restart(),
  getInstance: () => plyrInstance
})

onMounted(() => {
  if (props.src) {
    initPlyr()
  }
})

onUnmounted(() => {
  destroyPlyr()
})
</script>

<style scoped>
.plyr-player-wrapper {
  width: 100%;
  height: 100%;
  position: relative;
  background: var(--el-bg-color-page, #000);
  border-radius: 8px;
  overflow: hidden;
}

.plyr-player {
  width: 100%;
  height: 100%;
}

:deep(.plyr) {
  width: 100%;
  height: 100%;
}

:deep(.plyr__video-wrapper) {
  width: 100%;
  height: 100%;
}

:deep(video) {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
</style>

