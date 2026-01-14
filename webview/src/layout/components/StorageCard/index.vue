<template>
  <div class="storage-card glass-panel-sm">
    <div class="storage-header">
      <span class="storage-title">{{ t('settings.storage') }}</span>
      <span class="storage-text" v-if="storageInfo.isUnlimited">{{ t('settings.unlimited') }}</span>
      <span class="storage-text" v-else>{{ storageInfo.percentage }}%</span>
    </div>

    <!-- 可视化图表 -->
    <div class="storage-chart">
      <div class="chart-container">
        <svg class="storage-ring" viewBox="0 0 100 100">
          <!-- 背景圆环 -->
          <circle cx="50" cy="50" r="40" fill="none" stroke="var(--border-light)" stroke-width="8" />
          <!-- 进度圆环 -->
          <circle
            v-if="!storageInfo.isUnlimited"
            cx="50"
            cy="50"
            r="40"
            fill="none"
            :stroke="progressColor"
            stroke-width="8"
            stroke-linecap="round"
            :stroke-dasharray="circumference"
            :stroke-dashoffset="dashOffset"
            transform="rotate(-90 50 50)"
            class="progress-ring"
          />
          <!-- 无限容量圆环 -->
          <circle
            v-else
            cx="50"
            cy="50"
            r="40"
            fill="none"
            stroke="var(--primary-color)"
            stroke-width="8"
            stroke-linecap="round"
            :stroke-dasharray="circumference"
            :stroke-dashoffset="0"
            transform="rotate(-90 50 50)"
            class="infinite-ring"
          />
        </svg>
        <div class="chart-center">
          <div class="chart-percentage">{{ storageInfo.isUnlimited ? '∞' : `${storageInfo.percentage}%` }}</div>
          <div class="chart-label">{{ t('settings.used') }}</div>
        </div>
      </div>
    </div>

    <el-progress
      :percentage="storageInfo.isUnlimited ? 100 : storageInfo.percentage"
      :color="storageInfo.isUnlimited ? 'var(--primary-color)' : customColors"
      :show-text="false"
      :stroke-width="8"
      class="storage-progress"
    />
    <div class="storage-detail">
      {{ formatStorageSize(storageInfo.used) }} /
      {{ storageInfo.isUnlimited ? '∞' : formatStorageSize(storageInfo.total) }}
    </div>
    <el-button v-if="!storageInfo.isUnlimited" type="primary" link class="upgrade-btn">{{
      t('storage.upgrade')
    }}</el-button>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from '@/composables'
  import { useUserStore } from '@/stores'

  const userStore = useUserStore()
  const { t } = useI18n()

  const customColors = computed(() => [
    { color: 'var(--success-color)', percentage: 60 },
    { color: 'var(--warning-color)', percentage: 80 },
    { color: 'var(--danger-color)', percentage: 100 }
  ])

  const formatStorageSize = (bytes: number): string => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // 使用 store 中的 storageInfo
  const storageInfo = computed(() => userStore.storageInfo)

  // 计算圆环进度
  const circumference = computed(() => 2 * Math.PI * 40) // 半径40

  const dashOffset = computed(() => {
    if (storageInfo.value.isUnlimited) return 0
    const percentage = storageInfo.value.percentage / 100
    return circumference.value * (1 - percentage)
  })

  // 根据使用率计算颜色
  const progressColor = computed(() => {
    const percentage = storageInfo.value.percentage
    if (percentage < 60) return 'var(--success-color)' // 绿色
    if (percentage < 80) return 'var(--warning-color)' // 橙色
    return 'var(--danger-color)' // 红色
  })
</script>

<style scoped>
  .storage-card {
    margin: 12px;
    padding: 16px;
    border-radius: 12px;
    background: var(--card-bg, var(--el-bg-color));
    border: 1px solid var(--border-light);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    animation: fade-in 0.3s ease-out;
  }

  .storage-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
    transform: translateY(-2px);
  }

  html.dark .storage-card {
    background: var(--card-bg, rgba(30, 41, 59, 0.6));
    border-color: rgba(255, 255, 255, 0.1);
  }

  html.dark .storage-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    background: rgba(30, 41, 59, 0.8);
  }

  .storage-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    margin-bottom: 12px;
    font-size: 13px;
  }

  .storage-title {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
  }

  .storage-text {
    font-size: 11px;
    color: var(--text-secondary);
    font-weight: 400;
  }

  .storage-chart {
    display: flex;
    justify-content: center;
    margin: 12px 0;
  }

  .chart-container {
    position: relative;
    width: 100px;
    height: 100px;
  }

  .storage-ring {
    width: 100%;
    height: 100%;
    transform: rotate(-90deg);
  }

  .progress-ring {
    transition:
      stroke-dashoffset 0.6s ease,
      stroke 0.3s ease;
    filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
  }

  .infinite-ring {
    /* 无限容量时显示完整的静态圆环，不旋转 */
    stroke-dasharray: 251.2 251.2;
    stroke-dashoffset: 0;
    opacity: 0.8;
    filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
  }

  .chart-center {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    text-align: center;
  }

  .chart-percentage {
    font-size: 18px;
    font-weight: 600;
    color: var(--primary-color);
    line-height: 1.2;
  }

  .chart-label {
    font-size: 10px;
    color: var(--text-secondary);
    margin-top: 2px;
  }

  .storage-progress {
    margin-bottom: 10px;
  }

  .storage-progress :deep(.el-progress-bar__outer) {
    border-radius: 4px;
  }

  .storage-detail {
    font-size: 11px;
    color: var(--text-secondary);
    text-align: center;
    margin-bottom: 8px;
    font-family: var(--font-family-mono, monospace);
  }

  .upgrade-btn {
    width: 100%;
    margin-top: 8px;
    border-radius: 8px;
    height: 32px;
    font-size: 12px;
    font-weight: 500;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s ease;
  }

  .upgrade-btn:hover {
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(37, 99, 235, 0.2);
  }
</style>
