<template>
  <el-tooltip
    :content="fileName || ''"
    :placement="placement"
    :disabled="disabled || !shouldShowTooltip"
    :popper-class="popperClass"
  >
    <component :is="tag" :class="computedClass" :style="computedStyle">
      <slot>{{ fileName }}</slot>
    </component>
  </el-tooltip>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { isFileNameTruncated } from '@/utils/file/fileName'

  interface Props {
    /** 文件名 */
    fileName: string | null | undefined
    /** 视图模式 */
    viewMode?: 'grid' | 'list' | 'table'
    /** 最大显示长度（可选，默认根据视图模式判断） */
    maxLength?: number
    /** Tooltip 位置 */
    placement?: 'top' | 'bottom' | 'left' | 'right'
    /** 是否禁用 tooltip */
    disabled?: boolean
    /** 自定义样式类 */
    customClass?: string
    /** 自定义样式 */
    customStyle?: string | Record<string, any>
    /** 渲染标签 */
    tag?: string
    /** Tooltip 的 popper 类名 */
    popperClass?: string
  }

  const props = withDefaults(defineProps<Props>(), {
    fileName: '',
    viewMode: 'table',
    maxLength: undefined,
    placement: 'top',
    disabled: false,
    customClass: '',
    customStyle: undefined,
    tag: 'span',
    popperClass: 'file-name-tooltip'
  })

  // 判断是否应该显示 tooltip
  const shouldShowTooltip = computed(() => {
    return isFileNameTruncated(props.fileName, props.maxLength, props.viewMode)
  })

  // 计算样式类
  const computedClass = computed(() => {
    const baseClass = 'file-name-text'
    const viewClass = `file-name-text--${props.viewMode}`
    return [baseClass, viewClass, props.customClass].filter(Boolean).join(' ')
  })

  // 计算样式
  const computedStyle = computed(() => {
    return props.customStyle
  })
</script>
