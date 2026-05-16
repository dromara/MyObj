<template>
  <div class="skeleton-container" :class="{ 'grid-view': viewMode === 'grid', 'list-view': viewMode === 'list' }">
    <!-- 网格视图骨架屏 -->
    <template v-if="viewMode === 'grid'">
      <div v-for="i in count" :key="i" class="skeleton-card" :style="{ animationDelay: `${(i - 1) * 0.05}s` }">
        <div class="skeleton-icon"></div>
        <div class="skeleton-text"></div>
        <div class="skeleton-text short"></div>
      </div>
    </template>

    <!-- 列表视图骨架屏 -->
    <template v-else>
      <div v-for="i in count" :key="i" class="skeleton-list-item" :style="{ animationDelay: `${(i - 1) * 0.03}s` }">
        <div class="skeleton-list-icon"></div>
        <div class="skeleton-list-content">
          <div class="skeleton-text"></div>
          <div class="skeleton-text short"></div>
        </div>
        <div class="skeleton-list-actions">
          <div class="skeleton-action"></div>
          <div class="skeleton-action"></div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
  interface Props {
    count?: number
    viewMode?: 'grid' | 'list'
  }

  withDefaults(defineProps<Props>(), {
    count: 12,
    viewMode: 'grid'
  })
</script>

<style scoped>
  .skeleton-container {
    width: 100%;
  }

  /* 网格视图骨架屏 */
  .skeleton-card {
    background: var(--card-bg, white);
    border-radius: 16px;
    padding: 12px;
    box-shadow:
      0 1px 3px rgba(0, 0, 0, 0.08),
      0 4px 12px rgba(0, 0, 0, 0.04);
    animation: skeletonFadeIn 0.4s ease-out backwards;
  }

  html.dark .skeleton-card {
    box-shadow:
      0 1px 3px rgba(0, 0, 0, 0.3),
      0 4px 12px rgba(0, 0, 0, 0.2);
  }

  .skeleton-icon {
    width: 100%;
    height: 80px;
    border-radius: 12px;
    background: linear-gradient(
      90deg,
      var(--el-fill-color-lighter) 25%,
      var(--el-fill-color-light) 50%,
      var(--el-fill-color-lighter) 75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
    margin-bottom: 8px;
  }

  .skeleton-text {
    height: 12px;
    border-radius: 6px;
    margin-top: 8px;
    background: linear-gradient(
      90deg,
      var(--el-fill-color-lighter) 25%,
      var(--el-fill-color-light) 50%,
      var(--el-fill-color-lighter) 75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
  }

  .skeleton-text.short {
    width: 60%;
    height: 10px;
    margin-top: 6px;
  }

  /* 列表视图骨架屏 */
  .skeleton-list-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: var(--card-bg, white);
    border-radius: 8px;
    margin-bottom: 8px;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    animation: skeletonFadeIn 0.4s ease-out backwards;
  }

  html.dark .skeleton-list-item {
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  }

  .skeleton-list-icon {
    width: 48px;
    height: 48px;
    border-radius: 8px;
    background: linear-gradient(
      90deg,
      var(--el-fill-color-lighter) 25%,
      var(--el-fill-color-light) 50%,
      var(--el-fill-color-lighter) 75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
    flex-shrink: 0;
  }

  .skeleton-list-content {
    flex: 1;
    min-width: 0;
  }

  .skeleton-list-content .skeleton-text {
    width: 100%;
    margin-top: 0;
  }

  .skeleton-list-content .skeleton-text.short {
    width: 40%;
    margin-top: 8px;
  }

  .skeleton-list-actions {
    display: flex;
    gap: 8px;
    flex-shrink: 0;
  }

  .skeleton-action {
    width: 32px;
    height: 32px;
    border-radius: 6px;
    background: linear-gradient(
      90deg,
      var(--el-fill-color-lighter) 25%,
      var(--el-fill-color-light) 50%,
      var(--el-fill-color-lighter) 75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
  }

  /* 网格布局 */
  .grid-view {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 20px;
    padding: 4px;
  }

  /* 动画 */
  @keyframes shimmer {
    0% {
      background-position: -200% 0;
    }
    100% {
      background-position: 200% 0;
    }
  }

  @keyframes skeletonFadeIn {
    from {
      opacity: 0;
      transform: translateY(10px) translateZ(0);
    }
    to {
      opacity: 1;
      transform: translateY(0) translateZ(0);
    }
  }

  /* 优化 shimmer 动画，使用更平滑的缓动函数 */
  .skeleton-icon,
  .skeleton-text,
  .skeleton-list-icon,
  .skeleton-action {
    animation: shimmer 1.5s ease-in-out infinite;
  }

  /* 优化动画性能 */
  .skeleton-card,
  .skeleton-list-item,
  .skeleton-icon,
  .skeleton-text,
  .skeleton-list-icon,
  .skeleton-action {
    will-change: background-position;
    transform: translateZ(0);
  }

  /* 响应式 */
  @media (max-width: 991px) {
    .grid-view {
      grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
      gap: 12px;
    }

    .skeleton-icon {
      height: 60px;
    }
  }

  @media (max-width: 480px) {
    .grid-view {
      grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
      gap: 8px;
    }

    .skeleton-icon {
      height: 50px;
    }

    .skeleton-list-item {
      padding: 8px 12px;
    }

    .skeleton-list-icon {
      width: 40px;
      height: 40px;
    }
  }
</style>
