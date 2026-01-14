<template>
  <div class="empty-state" :class="[`empty-state-${type}`, { 'is-compact': compact }]">
    <!-- 插画区域 -->
    <div class="empty-illustration" v-if="showIllustration">
      <slot name="illustration">
        <div class="default-illustration" :class="`illustration-${type}`">
          <el-icon :size="illustrationSize">
            <component :is="illustrationIcon" />
          </el-icon>
        </div>
      </slot>
    </div>

    <!-- 内容区域 -->
    <div class="empty-content">
      <!-- 标题 -->
      <h3 v-if="title || $slots.title" class="empty-title">
        <slot name="title">{{ title || t(`empty.${type}.title`) }}</slot>
      </h3>

      <!-- 描述 -->
      <p v-if="description || $slots.description" class="empty-description">
        <slot name="description">{{ description || t(`empty.${type}.description`) }}</slot>
      </p>

      <!-- 操作按钮 -->
      <div v-if="showActions && ($slots.actions || actions.length > 0)" class="empty-actions">
        <slot name="actions">
          <template v-for="(action, index) in actions" :key="index">
            <el-button
              :type="action.type || 'primary'"
              :icon="action.icon"
              :size="compact ? 'small' : 'default'"
              @click="action.handler"
            >
              {{ action.label }}
            </el-button>
          </template>
        </slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from '@/composables'
  import {
    FolderOpened,
    Search,
    Document,
    Delete,
    Share,
    Upload,
    Refresh,
    QuestionFilled
  } from '@element-plus/icons-vue'

  interface EmptyAction {
    label: string
    handler: () => void
    type?: 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'default'
    icon?: any
  }

  interface Props {
    /** 空状态类型 */
    type?: 'folder' | 'search' | 'task' | 'trash' | 'share' | 'download' | 'default'
    /** 自定义标题 */
    title?: string
    /** 自定义描述 */
    description?: string
    /** 操作按钮列表 */
    actions?: EmptyAction[]
    /** 是否显示插画 */
    showIllustration?: boolean
    /** 是否显示操作按钮 */
    showActions?: boolean
    /** 紧凑模式 */
    compact?: boolean
    /** 插画大小 */
    illustrationSize?: number
  }

  const props = withDefaults(defineProps<Props>(), {
    type: 'default',
    title: undefined,
    description: undefined,
    actions: () => [],
    showIllustration: true,
    showActions: true,
    compact: false,
    illustrationSize: 120
  })

  const { t } = useI18n()

  // 根据类型选择默认图标
  const illustrationIcon = computed(() => {
    const iconMap: Record<string, any> = {
      folder: FolderOpened,
      search: Search,
      task: Document,
      trash: Delete,
      share: Share,
      download: Document,
      default: QuestionFilled
    }
    return iconMap[props.type] || QuestionFilled
  })
</script>

<style scoped>
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 20px;
    min-height: 400px;
    text-align: center;
  }

  .empty-state.is-compact {
    padding: 40px 20px;
    min-height: 300px;
  }

  /* 插画区域 */
  .empty-illustration {
    margin-bottom: 24px;
    animation: fade-in-up 0.6s ease-out;
  }

  .default-illustration {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 160px;
    height: 160px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--primary-color-light, #e3f2fd) 0%, var(--primary-color-lighter, #f3e5f5) 100%);
    color: var(--primary-color, #409eff);
    transition: all 0.3s ease;
  }

  html.dark .default-illustration {
    background: linear-gradient(135deg, rgba(64, 158, 255, 0.1) 0%, rgba(64, 158, 255, 0.05) 100%);
    color: var(--primary-color, #409eff);
  }

  .default-illustration:hover {
    transform: scale(1.05);
    box-shadow: 0 8px 24px rgba(64, 158, 255, 0.2);
  }

  html.dark .default-illustration:hover {
    box-shadow: 0 8px 24px rgba(64, 158, 255, 0.3);
  }

  /* 不同类型插画样式 */
  .illustration-folder {
    background: linear-gradient(135deg, #e3f2fd 0%, #f3e5f5 100%);
    color: #2196f3;
  }

  html.dark .illustration-folder {
    background: linear-gradient(135deg, rgba(33, 150, 243, 0.1) 0%, rgba(33, 150, 243, 0.05) 100%);
    color: #2196f3;
  }

  .illustration-search {
    background: linear-gradient(135deg, #fff3e0 0%, #fce4ec 100%);
    color: #ff9800;
  }

  html.dark .illustration-search {
    background: linear-gradient(135deg, rgba(255, 152, 0, 0.1) 0%, rgba(255, 152, 0, 0.05) 100%);
    color: #ff9800;
  }

  .illustration-task {
    background: linear-gradient(135deg, #e8f5e9 0%, #f1f8e9 100%);
    color: #4caf50;
  }

  html.dark .illustration-task {
    background: linear-gradient(135deg, rgba(76, 175, 80, 0.1) 0%, rgba(76, 175, 80, 0.05) 100%);
    color: #4caf50;
  }

  .illustration-trash {
    background: linear-gradient(135deg, #fce4ec 0%, #f3e5f5 100%);
    color: #e91e63;
  }

  html.dark .illustration-trash {
    background: linear-gradient(135deg, rgba(233, 30, 99, 0.1) 0%, rgba(233, 30, 99, 0.05) 100%);
    color: #e91e63;
  }

  .illustration-share {
    background: linear-gradient(135deg, #e1f5fe 0%, #e0f2f1 100%);
    color: #00bcd4;
  }

  html.dark .illustration-share {
    background: linear-gradient(135deg, rgba(0, 188, 212, 0.1) 0%, rgba(0, 188, 212, 0.05) 100%);
    color: #00bcd4;
  }

  /* 内容区域 */
  .empty-content {
    max-width: 500px;
    animation: fade-in-up 0.6s ease-out 0.2s both;
  }

  .empty-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--text-primary, #303133);
    margin: 0 0 12px 0;
    line-height: 1.4;
  }

  html.dark .empty-title {
    color: var(--text-primary, #e5eaf3);
  }

  .empty-description {
    font-size: 14px;
    color: var(--text-regular, #606266);
    margin: 0 0 24px 0;
    line-height: 1.6;
  }

  html.dark .empty-description {
    color: var(--text-regular, #a8abb2);
  }

  /* 操作按钮 */
  .empty-actions {
    display: flex;
    gap: 12px;
    justify-content: center;
    flex-wrap: wrap;
  }

  /* 动画 */
  @keyframes fade-in-up {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* 响应式 */
  @media (max-width: 768px) {
    .empty-state {
      padding: 40px 16px;
      min-height: 300px;
    }

    .empty-state.is-compact {
      padding: 30px 16px;
      min-height: 250px;
    }

    .default-illustration {
      width: 120px;
      height: 120px;
    }

    .empty-title {
      font-size: 18px;
    }

    .empty-description {
      font-size: 13px;
    }

    .empty-actions {
      flex-direction: column;
      width: 100%;
    }

    .empty-actions .el-button {
      width: 100%;
    }
  }
</style>
