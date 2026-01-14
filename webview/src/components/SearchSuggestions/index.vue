<template>
  <Transition name="fade-scale">
    <div v-if="visible && suggestions.length > 0" class="search-suggestions">
      <div class="suggestions-header">
        <span class="suggestions-title">{{ t('searchSuggestions.title') }}</span>
        <el-button v-if="showClear" text size="small" @click="handleClear">
          {{ t('searchSuggestions.clearHistory') }}
        </el-button>
      </div>
      <div class="suggestions-list">
        <div
          v-for="(item, index) in suggestions"
          :key="index"
          class="suggestion-item"
          @click="handleSelect(item)"
          @mouseenter="hoveredIndex = index"
          @mouseleave="hoveredIndex = -1"
          :class="{ 'is-hovered': hoveredIndex === index }"
        >
          <el-icon class="suggestion-icon"><Clock /></el-icon>
          <span class="suggestion-text">{{ item }}</span>
          <el-button v-if="showDelete" text size="small" class="suggestion-delete" @click.stop="handleDelete(item)">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
  import { useI18n } from '@/composables'

  interface Props {
    suggestions: string[]
    visible: boolean
    showClear?: boolean
    showDelete?: boolean
  }

  const props = withDefaults(defineProps<Props>(), {
    showClear: true,
    showDelete: true
  })

  const emit = defineEmits<{
    select: [keyword: string]
    clear: []
    delete: [keyword: string]
  }>()

  const { t } = useI18n()
  const hoveredIndex = ref(-1)

  const handleSelect = (keyword: string) => {
    emit('select', keyword)
  }

  const handleClear = () => {
    emit('clear')
  }

  const handleDelete = (keyword: string) => {
    emit('delete', keyword)
  }
</script>

<style scoped>
  .search-suggestions {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    margin-top: 4px;
    background: var(--card-bg, white);
    border: 1px solid var(--border-color, #e5e7eb);
    border-radius: 8px;
    box-shadow:
      0 4px 12px rgba(0, 0, 0, 0.1),
      0 2px 4px rgba(0, 0, 0, 0.06);
    z-index: 1000;
    max-height: 300px;
    overflow-y: auto;
  }

  /* 组件进入/退出动画 */
  .fade-scale-enter-active {
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .fade-scale-leave-active {
    transition: all 0.15s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .fade-scale-enter-from {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }

  .fade-scale-leave-to {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }

  .suggestions-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-light, #f3f4f6);
  }

  .suggestions-title {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary, #6b7280);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .suggestions-list {
    padding: 4px;
  }

  .suggestion-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;
  }

  .suggestion-item:hover,
  .suggestion-item.is-hovered {
    background: var(--card-hover-bg, rgba(37, 99, 235, 0.05));
  }

  .suggestion-icon {
    color: var(--text-secondary, #6b7280);
    font-size: 14px;
    flex-shrink: 0;
  }

  .suggestion-text {
    flex: 1;
    font-size: 14px;
    color: var(--text-regular, #374151);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .suggestion-delete {
    opacity: 0;
    transition: opacity 0.2s;
    flex-shrink: 0;
  }

  .suggestion-item:hover .suggestion-delete {
    opacity: 1;
  }

  html.dark .search-suggestions {
    background: rgba(30, 41, 59, 0.95);
    border-color: rgba(255, 255, 255, 0.1);
  }

  html.dark .suggestion-item:hover {
    background: rgba(51, 65, 85, 0.8);
  }
</style>
