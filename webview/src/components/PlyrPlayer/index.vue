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
import { ref, watch, onMounted, onUnmounted, nextTick, getCurrentInstance } from 'vue'
// @ts-ignore - plyr 类型定义可能不完整
import Plyr from 'plyr'
import 'plyr/dist/plyr.css'

// Plyr 中文 i18n 配置
const plyrI18nZh = {
  restart: '重新开始',
  rewind: '后退 {seektime} 秒',
  play: '播放',
  pause: '暂停',
  fastForward: '快进 {seektime} 秒',
  seek: '跳转',
  seekLabel: '{currentTime} / {duration}',
  played: '已播放',
  buffered: '已缓冲',
  currentTime: '当前时间',
  duration: '总时长',
  volume: '音量',
  mute: '静音',
  unmute: '取消静音',
  enableCaptions: '启用字幕',
  disableCaptions: '禁用字幕',
  download: '下载',
  enterFullscreen: '进入全屏',
  exitFullscreen: '退出全屏',
  frameTitle: '{title} 播放器',
  captions: '字幕',
  settings: '设置',
  pip: '画中画',
  menuBack: '返回上一级菜单',
  speed: '播放速度',
  normal: '正常',
  quality: '画质',
  loop: '循环播放',
  start: '开始',
  end: '结束',
  all: '全部',
  reset: '重置',
  disabled: '已禁用',
  enabled: '已启用',
  advertisement: '广告',
  qualityBadge: {
    2160: '4K',
    1440: 'HD',
    1080: 'HD',
    720: 'HD',
    576: 'SD',
    480: 'SD',
  },
}

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
    i18n: plyrI18nZh, // 设置中文界面
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
    const errorMessage = detail?.message || '视频播放错误'
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
  background: #000;
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

