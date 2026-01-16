<script setup lang="ts">
  // App只作为路由容器，逻辑由router守卫和各页面处理
  import { useTheme, useKeyboardShortcuts, useOnboarding } from '@/composables'
  import { useAppStore } from '@/stores'

  // 初始化主题系统
  useTheme()

  // 初始化快捷键系统
  useKeyboardShortcuts()

  // 初始化新手引导系统
  const { checkAndStartOnboarding, checkOnboardingStatus } = useOnboarding()

  // 获取 Element Plus 语言包
  const appStore = useAppStore()

  // 监听路由变化，检查是否需要启动引导
  const router = useRouter()
  router.afterEach(() => {
    // 延迟检查，确保页面已渲染
    setTimeout(() => {
      checkOnboardingStatus()
      checkAndStartOnboarding()
    }, 300)
  })

  // 初始化时也检查一次（处理首次加载或清除 localStorage 后的情况）
  onMounted(() => {
    checkOnboardingStatus()
    // 延迟检查，确保页面已渲染
    setTimeout(() => {
      checkAndStartOnboarding()
    }, 500)
  })
</script>

<template>
  <ElConfigProvider :locale="appStore.elementPlusLocale">
    <router-view />

    <!-- 快捷键帮助对话框 -->
    <ShortcutHelp />

    <!-- 新手引导欢迎对话框 -->
    <OnboardingWelcome />
  </ElConfigProvider>
</template>

<style scoped>
  /* 全局样式在style.css中定义 */

  /* 页面过渡动画 - 优化为更快速的淡入淡出 */
  .fade-enter-active,
  .fade-leave-active {
    transition: opacity 0.15s ease;
  }

  .fade-enter-from,
  .fade-leave-to {
    opacity: 0;
  }

  /* 滑动过渡 */
  .slide-enter-active,
  .slide-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .slide-enter-from {
    opacity: 0;
    transform: translateX(30px);
  }

  .slide-leave-to {
    opacity: 0;
    transform: translateX(-30px);
  }

  /* 缩放过渡 */
  .scale-enter-active,
  .scale-leave-active {
    transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  .scale-enter-from {
    opacity: 0;
    transform: scale(0.95);
  }

  .scale-leave-to {
    opacity: 0;
    transform: scale(1.05);
  }
</style>
