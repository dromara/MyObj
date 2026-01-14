<template>
  <Transition :name="transitionName">
    <div v-if="loading" class="loading-overlay" :class="overlayClass">
      <div class="loading-content">
        <div class="loading-spinner">
          <div class="spinner-ring"></div>
          <div class="spinner-ring"></div>
          <div class="spinner-ring"></div>
          <div class="spinner-ring"></div>
        </div>
        <p v-if="text" class="loading-text">{{ text }}</p>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
  interface Props {
    loading?: boolean
    text?: string
    /** 是否全屏显示（fixed 定位） */
    fullscreen?: boolean
    /** 加载类型：default（默认）或 route（路由加载，固定全屏） */
    type?: 'default' | 'route'
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
    text: '',
    fullscreen: false,
    type: 'default'
  })

  const overlayClass = computed(() => {
    if (props.type === 'route' || props.fullscreen) {
      return 'is-fullscreen'
    }
    return ''
  })

  const transitionName = computed(() => {
    return props.type === 'route' ? 'fade' : 'loading-fade'
  })
</script>

<style scoped>
  .loading-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(255, 255, 255, 0.8);
    backdrop-filter: blur(2px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
    transition: opacity 0.3s ease;
  }

  .loading-overlay.is-fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 9999;
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(4px);
  }

  html.dark .loading-overlay.is-fullscreen {
    background: rgba(15, 23, 42, 0.9);
  }

  html.dark .loading-overlay {
    background: rgba(15, 23, 42, 0.8);
  }

  .loading-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
  }

  .loading-spinner {
    position: relative;
    width: 48px;
    height: 48px;
  }

  .spinner-ring {
    position: absolute;
    width: 100%;
    height: 100%;
    border: 3px solid transparent;
    border-top-color: var(--primary-color);
    border-radius: 50%;
    animation: spin 1.2s cubic-bezier(0.5, 0, 0.5, 1) infinite;
  }

  .spinner-ring:nth-child(1) {
    animation-delay: -0.45s;
  }

  .spinner-ring:nth-child(2) {
    animation-delay: -0.3s;
    border-top-color: var(--secondary-color);
  }

  .spinner-ring:nth-child(3) {
    animation-delay: -0.15s;
    border-top-color: var(--primary-color);
    opacity: 0.7;
  }

  .spinner-ring:nth-child(4) {
    animation-delay: 0s;
    border-top-color: var(--secondary-color);
    opacity: 0.5;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  .loading-text {
    color: var(--text-primary);
    font-size: 14px;
    font-weight: 500;
    margin: 0;
    text-align: center;
  }

  /* 淡入淡出动画 */
  .loading-fade-enter-active,
  .loading-fade-leave-active,
  .fade-enter-active,
  .fade-leave-active {
    transition: opacity 0.3s ease;
  }

  .loading-fade-enter-from,
  .loading-fade-leave-to,
  .fade-enter-from,
  .fade-leave-to {
    opacity: 0;
  }
</style>
