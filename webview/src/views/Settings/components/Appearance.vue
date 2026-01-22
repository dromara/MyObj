<template>
  <div class="appearance-settings">
    <el-form label-width="120px">
      <!-- 基础设置分组 -->
      <el-divider content-position="left">
        <span class="divider-title">{{ t('settings.groups.basic') }}</span>
      </el-divider>

      <!-- 主题设置 -->
      <el-form-item :label="t('settings.theme')">
        <el-radio-group v-model="currentTheme" @change="handleThemeChange">
          <el-radio-button value="light">
            <el-icon><Sunny /></el-icon>
            <span style="margin-left: 4px">{{ t('settings.light') }}</span>
          </el-radio-button>
          <el-radio-button value="dark">
            <el-icon><Moon /></el-icon>
            <span style="margin-left: 4px">{{ t('settings.dark') }}</span>
          </el-radio-button>
          <el-radio-button value="auto">
            <el-icon><Monitor /></el-icon>
            <span style="margin-left: 4px">{{ t('settings.auto') }}</span>
          </el-radio-button>
        </el-radio-group>
      </el-form-item>

      <!-- 语言设置 -->
      <el-form-item :label="t('settings.language')">
        <el-select v-model="currentLocale" @change="handleLocaleChange" style="width: 200px">
          <el-option :label="t('settings.chinese')" :value="LanguageEnum.zh_CN" />
          <el-option :label="t('settings.english')" :value="LanguageEnum.en_US" />
        </el-select>
      </el-form-item>

      <!-- 功能设置分组 -->
      <el-divider content-position="left">
        <span class="divider-title">{{ t('settings.groups.features') }}</span>
      </el-divider>

      <!-- 剪贴板监听 -->
      <el-form-item>
        <div class="clipboard-monitor-card" :class="{ 'is-enabled': clipboardMonitorEnabled }">
          <div class="clipboard-monitor-header">
            <div class="clipboard-monitor-title-group">
              <el-icon class="clipboard-monitor-icon" :size="24">
                <Link v-if="clipboardMonitorEnabled" />
                <Link v-else />
              </el-icon>
              <div class="clipboard-monitor-title-content">
                <h4 class="clipboard-monitor-title">{{ t('settings.clipboardMonitor.title') }}</h4>
                <el-tag
                  :type="clipboardMonitorEnabled ? 'success' : 'info'"
                  size="small"
                  class="clipboard-monitor-status"
                >
                  {{ clipboardMonitorEnabled ? t('settings.clipboardMonitor.enabled') : t('settings.clipboardMonitor.disabled') }}
                </el-tag>
              </div>
            </div>
            <el-switch
              :model-value="clipboardMonitorEnabled"
              @change="handleClipboardMonitorChange"
              size="large"
            />
          </div>
          <p class="clipboard-monitor-description">
            {{ t('settings.clipboardMonitor.description') }}
          </p>
          <p class="clipboard-monitor-note">
            {{ t('settings.clipboardMonitor.descriptionNote') }}
          </p>
          <div class="clipboard-monitor-features">
            <div class="feature-item">
              <el-icon class="feature-icon"><Link /></el-icon>
              <span class="feature-text">{{ t('settings.clipboardMonitor.linkType.http') }}</span>
            </div>
            <div class="feature-item">
              <el-icon class="feature-icon"><Link /></el-icon>
              <span class="feature-text">{{ t('settings.clipboardMonitor.linkType.magnet') }}</span>
            </div>
            <div class="feature-item">
              <el-icon class="feature-icon"><Document /></el-icon>
              <span class="feature-text">{{ t('settings.clipboardMonitor.linkType.torrent') }}</span>
            </div>
          </div>
        </div>
      </el-form-item>

      <!-- 主题与颜色分组 -->
      <el-divider content-position="left">
        <span class="divider-title">{{ t('settings.groups.theme') }}</span>
      </el-divider>

      <!-- 自定义主题色 -->
      <el-form-item :label="t('settings.themeColor')">
        <div class="color-picker-group">
          <div class="color-item">
            <label>{{ t('settings.primaryColor') }}</label>
            <el-color-picker v-model="displayColors.primary" @change="val => handleColorChange('primary', val)" />
            <el-button text size="small" @click="resetColor('primary')">{{ t('settings.reset') }}</el-button>
          </div>
          <div class="color-item">
            <label>{{ t('settings.successColor') }}</label>
            <el-color-picker v-model="displayColors.success" @change="val => handleColorChange('success', val)" />
            <el-button text size="small" @click="resetColor('success')">{{ t('settings.reset') }}</el-button>
          </div>
          <div class="color-item">
            <label>{{ t('settings.warningColor') }}</label>
            <el-color-picker v-model="displayColors.warning" @change="val => handleColorChange('warning', val)" />
            <el-button text size="small" @click="resetColor('warning')">{{ t('settings.reset') }}</el-button>
          </div>
          <div class="color-item">
            <label>{{ t('settings.dangerColor') }}</label>
            <el-color-picker v-model="displayColors.danger" @change="val => handleColorChange('danger', val)" />
            <el-button text size="small" @click="resetColor('danger')">{{ t('settings.reset') }}</el-button>
          </div>
        </div>
        <el-button type="primary" @click="resetAllColors" style="margin-top: 12px">
          {{ t('settings.resetAll') }}
        </el-button>
      </el-form-item>

      <!-- 辅助模式 -->
      <el-form-item :label="t('settings.auxiliaryModes')">
        <div class="auxiliary-modes">
          <el-switch
            v-model="currentGrayscale"
            :active-text="t('settings.grayscale')"
            @change="(val: string | number | boolean) => handleGrayscaleChange(val === true || val === 'true')"
          />
          <el-switch
            v-model="currentColourWeakness"
            :active-text="t('settings.colourWeakness')"
            @change="(val: string | number | boolean) => handleColourWeaknessChange(val === true || val === 'true')"
            style="margin-left: 24px"
          />
        </div>
      </el-form-item>

      <!-- 主题预设 -->
      <el-form-item :label="t('settings.themePreset')">
        <div class="preset-list">
          <div
            v-for="preset in themePresets"
            :key="preset.name"
            class="preset-card"
            :class="{ 'preset-active': selectedPreset === preset.name }"
            @click="handlePresetSelect(preset)"
          >
            <div class="preset-header">
              <div class="preset-title-group">
                <h5 class="preset-title">{{ getPresetName(preset.name) }}</h5>
              </div>
              <el-button
                :type="selectedPreset === preset.name ? 'primary' : 'default'"
                size="small"
                :class="{ 'preset-apply-btn': true, 'preset-apply-btn-active': selectedPreset === preset.name }"
                @click.stop="handlePresetChange(preset.name)"
              >
                {{ selectedPreset === preset.name ? t('settings.applied') : t('settings.apply') }}
              </el-button>
            </div>
            <p class="preset-desc">{{ getPresetDesc(preset.name) }}</p>
            <div class="preset-preview">
              <div class="preset-colors">
                <div
                  v-for="(color, key) in getDisplayColors(preset)"
                  :key="key"
                  class="preset-color-dot"
                  :style="{ backgroundColor: color }"
                  :class="{ 'is-primary': key === 'primary' }"
                  :title="key"
                />
              </div>
              <div class="preset-meta">
                <el-icon v-if="preset.theme === 'dark'"><Moon /></el-icon>
                <el-icon v-else-if="preset.theme === 'light'"><Sunny /></el-icon>
                <el-icon v-else><Monitor /></el-icon>
                <span v-if="preset.grayscale" class="preset-icon">🎨</span>
                <span v-if="preset.colourWeakness" class="preset-icon">👁️</span>
              </div>
            </div>
          </div>
        </div>
      </el-form-item>

      <!-- 背景图案 -->
      <el-form-item :label="t('settings.backgroundPattern')">
        <el-select v-model="backgroundPattern" @change="handleBackgroundPatternChange" style="width: 200px">
          <el-option :label="t('settings.none')" value="none" />
          <el-option :label="t('settings.grid')" value="grid" />
          <el-option :label="t('settings.dots')" value="dots" />
          <el-option :label="t('settings.gradient')" value="gradient" />
          <el-option :label="t('settings.waves')" value="waves" />
          <el-option :label="t('settings.particles')" value="particles" />
        </el-select>
      </el-form-item>

      <!-- 布局设置分组 -->
      <el-divider content-position="left">
        <span class="divider-title">{{ t('settings.groups.layout') }}</span>
      </el-divider>

      <!-- 布局模式 -->
      <el-form-item :label="t('layout.mode.title')">
        <div class="layout-mode-wrapper">
          <LayoutModeCard
            v-model="currentLayoutMode"
            :disabled="isMobile"
            @update:modelValue="handleLayoutModeChange"
          />
          <el-button type="primary" @click="handleResetLayoutMode" style="margin-top: 12px">
            <el-icon><RefreshLeft /></el-icon>
            <span style="margin-left: 4px">{{ t('settings.reset') }}</span>
          </el-button>
        </div>
        <div v-if="isMobile" class="layout-tip">
          <el-text type="info" size="small">{{ t('layout.mode.mobileTip') }}</el-text>
        </div>
      </el-form-item>

      <!-- 侧边栏设置 -->
      <el-form-item :label="t('layout.sidebar.title')">
        <div class="sidebar-settings">
          <div class="setting-item">
            <label>{{ t('layout.sidebar.width') }}</label>
            <el-input-number
              v-model="currentSidebarWidth"
              :min="200"
              :max="400"
              :step="10"
              @change="handleSidebarWidthChange"
            />
            <span class="unit">px</span>
            <el-button type="primary" @click="handleResetSidebarWidth" style="margin-left: 8px">
              <el-icon><RefreshLeft /></el-icon>
              <span style="margin-left: 4px">{{ t('settings.reset') }}</span>
            </el-button>
          </div>
          <div class="setting-item">
            <el-switch
              v-model="currentSidebarCollapsed"
              :active-text="t('layout.sidebar.collapsed')"
              @change="handleSidebarCollapsedChange"
            />
          </div>
        </div>
      </el-form-item>

      <!-- 标签页设置 -->
      <el-form-item :label="t('layout.tagsView.title')">
        <div class="setting-item">
          <el-switch
            v-model="currentTagsViewVisible"
            :active-text="t('layout.tagsView.visible')"
            @change="handleTagsViewVisibleChange"
          />
        </div>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
  import { Sunny, Moon, Monitor, RefreshLeft, Link, Document } from '@element-plus/icons-vue'
  import { useTheme, useI18n, useResponsive, useClipboardMonitor } from '@/composables'
  import { LanguageEnum } from '@/enums/LanguageEnum'
  import { themePresets } from '@/theme/presets'
  import { useLayoutStore } from '@/stores'
  import LayoutModeCard from '@/components/LayoutModeCard/index.vue'

  const {
    theme,
    customColors,
    grayscale,
    colourWeakness,
    setTheme,
    setCustomColors,
    resetCustomColors,
    setGrayscale,
    setColourWeakness,
    applyPreset
  } = useTheme()
  const { locale, setLocale } = useI18n()
  const { isMobile } = useResponsive()
  const layoutStore = useLayoutStore()
  const clipboardMonitor = useClipboardMonitor()
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  // 剪贴板监听状态
  const clipboardMonitorEnabled = computed(() => clipboardMonitor.isEnabled.value)
  
  // 处理剪贴板监听开关变化
  const handleClipboardMonitorChange = (value: string | number | boolean) => {
    const boolValue = value === true || value === 'true' || value === 1
    if (boolValue) {
      clipboardMonitor.enable()
      proxy?.$modal.msgSuccess(t('settings.clipboardMonitor.enabledSuccess'))
    } else {
      clipboardMonitor.disable()
      proxy?.$modal.msgSuccess(t('settings.clipboardMonitor.disabledSuccess'))
    }
  }

  const currentTheme = ref(theme.value)
  const currentLocale = ref(locale.value)
  const currentGrayscale = ref(grayscale.value)
  const currentColourWeakness = ref(colourWeakness.value)
  const backgroundPattern = ref<'none' | 'grid' | 'dots' | 'gradient' | 'waves' | 'particles'>('none')
  const selectedPreset = ref<string>('')

  // 布局相关状态
  const currentLayoutMode = ref(layoutStore.layoutMode)
  const currentSidebarWidth = ref(layoutStore.sidebarWidth)
  const currentSidebarCollapsed = ref(layoutStore.sidebarCollapsed)
  const currentTagsViewVisible = ref(layoutStore.tagsViewVisible)

  // 默认颜色值（用于初始化颜色选择器）
  const defaultColors = {
    primary: '#2563eb',
    success: '#10b981',
    warning: '#f59e0b',
    danger: '#ef4444'
  }

  // 确保颜色选择器有默认值（使用 ref 以支持 v-model）
  const displayColors = ref({
    primary: customColors.value.primary || defaultColors.primary,
    success: customColors.value.success || defaultColors.success,
    warning: customColors.value.warning || defaultColors.warning,
    danger: customColors.value.danger || defaultColors.danger
  })

  // 监听 customColors 变化，同步到 displayColors
  watch(
    customColors,
    newColors => {
      displayColors.value = {
        primary: newColors.primary || defaultColors.primary,
        success: newColors.success || defaultColors.success,
        warning: newColors.warning || defaultColors.warning,
        danger: newColors.danger || defaultColors.danger
      }
    },
    { deep: true, immediate: true }
  )

  // 从 localStorage 加载背景图案设置和布局设置
  onMounted(() => {
    const saved = localStorage.getItem('backgroundPattern')
    if (saved && ['none', 'grid', 'dots', 'gradient', 'waves', 'particles'].includes(saved)) {
      backgroundPattern.value = saved as any
    }

    // 初始化布局设置
    layoutStore.initLayout()
    currentLayoutMode.value = layoutStore.layoutMode
    currentSidebarWidth.value = layoutStore.sidebarWidth
    currentSidebarCollapsed.value = layoutStore.sidebarCollapsed
    currentTagsViewVisible.value = layoutStore.tagsViewVisible
  })

  // 监听主题变化
  watch(theme, newTheme => {
    currentTheme.value = newTheme
  })

  // 监听语言变化
  watch(locale, newLocale => {
    currentLocale.value = newLocale
  })

  // 监听自定义颜色变化，同步到显示颜色（displayColors 是 computed，会自动更新）

  // 监听灰度模式变化
  watch(grayscale, newValue => {
    currentGrayscale.value = newValue
  })

  // 监听色弱模式变化
  watch(colourWeakness, newValue => {
    currentColourWeakness.value = newValue
  })

  // 监听布局模式变化
  watch(
    () => layoutStore.layoutMode,
    newMode => {
      currentLayoutMode.value = newMode
    }
  )

  // 监听侧边栏宽度变化
  watch(
    () => layoutStore.sidebarWidth,
    newWidth => {
      currentSidebarWidth.value = newWidth
    }
  )

  // 监听侧边栏折叠状态变化
  watch(
    () => layoutStore.sidebarCollapsed,
    newCollapsed => {
      currentSidebarCollapsed.value = newCollapsed
    }
  )

  // 监听标签页显示状态变化
  watch(
    () => layoutStore.tagsViewVisible,
    newVisible => {
      currentTagsViewVisible.value = newVisible
    }
  )

  const handleThemeChange = (value: string | number | boolean | undefined) => {
    if (typeof value === 'string' && (value === 'light' || value === 'dark' || value === 'auto')) {
      setTheme(value)
      proxy?.$modal.msgSuccess(t('settings.themeChanged'))
    }
  }

  const handleLocaleChange = (value: LanguageEnum) => {
    setLocale(value)
    // 重新加载页面以应用 Element Plus 语言
    window.location.reload()
  }

  const handleColorChange = (colorKey: 'primary' | 'success' | 'warning' | 'danger', value: string | null) => {
    if (value) {
      setCustomColors({ [colorKey]: value })
      proxy?.$modal.msgSuccess(t('settings.colorUpdated'))
    }
  }

  const resetColor = (colorKey: 'primary' | 'success' | 'warning' | 'danger') => {
    const defaultColors: Record<string, string> = {
      primary: '#2563eb',
      success: '#10b981',
      warning: '#f59e0b',
      danger: '#ef4444'
    }

    setCustomColors({ [colorKey]: defaultColors[colorKey] })
    proxy?.$modal.msgSuccess(t('settings.colorReset'))
  }

  const resetAllColors = () => {
    resetCustomColors()
    proxy?.$modal.msgSuccess(t('settings.allColorsReset'))
  }

  const handleGrayscaleChange = (value: boolean) => {
    setGrayscale(value)
    proxy?.$modal.msgSuccess(value ? t('settings.grayscaleEnabled') : t('settings.grayscaleDisabled'))
  }

  const handleColourWeaknessChange = (value: boolean) => {
    setColourWeakness(value)
    proxy?.$modal.msgSuccess(value ? t('settings.colourWeaknessEnabled') : t('settings.colourWeaknessDisabled'))
  }

  // 获取预设的国际化名称
  const getPresetName = (presetKey: string) => {
    // 根据预设名称映射到国际化键
    if (presetKey.includes('默认') || presetKey.includes('Default')) {
      return t('settings.presets.default.name')
    } else if (presetKey.includes('亮色') || presetKey.includes('Light')) {
      return t('settings.presets.light.name')
    } else if (presetKey.includes('暗色') || presetKey.includes('Dark')) {
      return t('settings.presets.dark.name')
    } else if (presetKey.includes('灰度') || presetKey.includes('Grayscale')) {
      return t('settings.presets.grayscale.name')
    } else if (
      presetKey.includes('色弱') ||
      presetKey.includes('ColourWeakness') ||
      presetKey.includes('Color Weakness')
    ) {
      return t('settings.presets.colourWeakness.name')
    }
    return presetKey
  }

  // 获取预设的国际化描述
  const getPresetDesc = (presetKey: string) => {
    // 根据预设名称映射到国际化键
    if (presetKey.includes('默认') || presetKey.includes('Default')) {
      return t('settings.presets.default.desc')
    } else if (presetKey.includes('亮色') || presetKey.includes('Light')) {
      return t('settings.presets.light.desc')
    } else if (presetKey.includes('暗色') || presetKey.includes('Dark')) {
      return t('settings.presets.dark.desc')
    } else if (presetKey.includes('灰度') || presetKey.includes('Grayscale')) {
      return t('settings.presets.grayscale.desc')
    } else if (
      presetKey.includes('色弱') ||
      presetKey.includes('ColourWeakness') ||
      presetKey.includes('Color Weakness')
    ) {
      return t('settings.presets.colourWeakness.desc')
    }
    return presetKey
  }

  const handlePresetSelect = (preset: (typeof themePresets)[0]) => {
    selectedPreset.value = preset.name
    // 立即应用主题预设
    applyPreset(preset)
    proxy?.$modal.msgSuccess(t('settings.presetApplied', { name: getPresetName(preset.name) }))
  }

  const handlePresetChange = (presetName: string) => {
    const preset = themePresets.find(p => p.name === presetName)
    if (preset) {
      selectedPreset.value = presetName
      applyPreset(preset)
      proxy?.$modal.msgSuccess(t('settings.presetApplied', { name: getPresetName(preset.name) }))
    }
  }

  // 获取预设的显示颜色（根据预设类型调整）
  const getDisplayColors = (preset: (typeof themePresets)[0]) => {
    const colors = { ...preset.colors }

    // 如果是暗色预设，使用适合暗色主题的颜色
    if (preset.theme === 'dark') {
      // 暗色模式下，颜色应该更亮一些以便在暗色背景上显示
      return {
        primary: colors.primary || '#3b82f6',
        success: colors.success || '#10b981',
        warning: colors.warning || '#f59e0b',
        danger: colors.danger || '#ef4444',
        info: colors.info || '#06b6d4'
      }
    }

    // 如果是色弱预设，使用高对比度的颜色
    if (preset.colourWeakness) {
      return {
        primary: colors.primary || '#2563eb',
        success: colors.success || '#059669', // 更深的绿色，提高对比度
        warning: colors.warning || '#d97706', // 更深的橙色，提高对比度
        danger: colors.danger || '#dc2626', // 更深的红色，提高对比度
        info: colors.info || '#0284c7' // 更深的蓝色，提高对比度
      }
    }

    // 默认返回原始颜色
    return colors
  }

  // 初始化时设置当前预设
  onMounted(() => {
    // 根据当前主题设置匹配的预设
    const currentPreset = themePresets.find(p => {
      if (p.theme !== theme.value) return false
      if (p.grayscale !== grayscale.value) return false
      if (p.colourWeakness !== colourWeakness.value) return false
      return true
    })
    if (currentPreset) {
      selectedPreset.value = currentPreset.name
    }
  })

  const handleBackgroundPatternChange = (value: string) => {
    // 保存到 localStorage
    localStorage.setItem('backgroundPattern', value)

    // 立即更新本地状态（确保 UI 响应）
    backgroundPattern.value = value as any

    // 获取图案的显示名称
    const patternNames: Record<string, string> = {
      none: t('settings.none'),
      grid: t('settings.grid'),
      dots: t('settings.dots'),
      gradient: t('settings.gradient'),
      waves: t('settings.waves'),
      particles: t('settings.particles')
    }
    const patternName = patternNames[value] || value

    // 显示成功提示
    proxy?.$modal.msgSuccess(t('settings.backgroundPatternChanged', { pattern: patternName }))

    // 触发背景图案更新事件（使用 bubbles 和 cancelable 确保事件能正确传播）
    const event = new CustomEvent('background-pattern-changed', {
      detail: { pattern: value },
      bubbles: true,
      cancelable: true
    })
    window.dispatchEvent(event)
  }

  // 布局相关处理函数
  const handleLayoutModeChange = (mode: typeof layoutStore.layoutMode) => {
    layoutStore.setLayoutMode(mode)
    proxy?.$modal.msgSuccess(t('layout.mode.changed', { mode: t(`layout.mode.${mode}`) }))
  }

  const handleSidebarWidthChange = (width: number | undefined) => {
    if (width !== undefined && width !== null) {
      layoutStore.setSidebarWidth(width)
      proxy?.$modal.msgSuccess(t('layout.sidebar.widthChanged', { width }))
    }
  }

  const handleSidebarCollapsedChange = (val: string | number | boolean) => {
    const collapsed = val === true || val === 'true'
    layoutStore.setSidebarCollapsed(collapsed)
    proxy?.$modal.msgSuccess(collapsed ? t('layout.sidebar.collapsedEnabled') : t('layout.sidebar.collapsedDisabled'))
  }

  const handleTagsViewVisibleChange = (val: string | number | boolean) => {
    const visible = val === true || val === 'true'
    layoutStore.setTagsViewVisible(visible)
    proxy?.$modal.msgSuccess(visible ? t('layout.tagsView.visibleEnabled') : t('layout.tagsView.visibleDisabled'))
  }

  // 重置布局模式为默认值
  const handleResetLayoutMode = () => {
    layoutStore.setLayoutMode('vertical')
    currentLayoutMode.value = 'vertical'
    proxy?.$modal.msgSuccess(t('layout.mode.changed', { mode: t('layout.mode.vertical') }))
  }

  // 重置侧边栏宽度为默认值
  const handleResetSidebarWidth = () => {
    layoutStore.setSidebarWidth(240)
    currentSidebarWidth.value = 240
    proxy?.$modal.msgSuccess(t('layout.sidebar.widthChanged', { width: 240 }))
  }
</script>

<style scoped>
  .appearance-settings {
    padding: 8px 0;
    width: 100%;
    max-width: 1000px;
  }

  .color-picker-group {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .color-item {
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 32px;
  }

  .color-item label {
    width: 100px;
    font-size: 14px;
    color: var(--text-regular);
    flex-shrink: 0;
    font-weight: 500;
  }

  .color-item :deep(.el-color-picker) {
    margin-right: auto;
  }

  .color-item :deep(.el-button) {
    margin-left: auto;
  }

  .auxiliary-modes {
    display: flex;
    align-items: center;
    gap: 24px;
    flex-wrap: wrap;
  }

  .auxiliary-modes .el-switch {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  /* 主题预设卡片样式 */
  .preset-list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 16px;
    margin-top: 8px;
  }

  /* 超大屏幕：3列布局，更宽松 */
  @media (min-width: 1600px) {
    .preset-list {
      grid-template-columns: repeat(3, 1fr);
      gap: 20px;
    }
  }

  /* 大屏幕：3列布局 */
  @media (min-width: 1200px) and (max-width: 1599px) {
    .preset-list {
      grid-template-columns: repeat(3, 1fr);
      gap: 18px;
    }
  }

  /* 中等屏幕：2-3列自适应 */
  @media (min-width: 768px) and (max-width: 1199px) {
    .preset-list {
      grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
      gap: 16px;
    }
  }

  .preset-card {
    padding: 18px;
    background: var(--card-bg);
    border: 2px solid var(--border-light);
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
    overflow: hidden;
    min-width: 240px;
  }

  .preset-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: linear-gradient(90deg, var(--primary-color), var(--secondary-color));
    transform: scaleX(0);
    transition: transform 0.3s ease;
  }

  .preset-card:hover {
    border-color: var(--primary-color);
    box-shadow: 0 4px 12px rgba(37, 99, 235, 0.15);
    transform: translateY(-2px);
  }

  .preset-card:hover::before {
    transform: scaleX(1);
  }

  .preset-card.preset-active {
    border-color: var(--primary-color);
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.05) 0%, rgba(79, 70, 229, 0.05) 100%);
    box-shadow: 0 4px 16px rgba(37, 99, 235, 0.2);
  }

  .preset-card.preset-active::before {
    transform: scaleX(1);
  }

  html.dark .preset-card {
    background: rgba(30, 41, 59, 0.6);
    border-color: rgba(255, 255, 255, 0.1);
  }

  html.dark .preset-card:hover {
    border-color: var(--primary-color);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.25);
    background: rgba(30, 41, 59, 0.8);
  }

  html.dark .preset-card.preset-active {
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.1) 0%, rgba(99, 102, 241, 0.1) 100%);
    box-shadow: 0 4px 16px rgba(59, 130, 246, 0.3);
  }

  .preset-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 8px;
    gap: 12px;
  }

  .preset-apply-btn {
    flex-shrink: 0;
    min-width: 60px;
    font-weight: 500;
    transition: all 0.2s ease;
  }

  /* 未选中状态：使用默认按钮样式，确保文字清晰可见 */
  .preset-apply-btn:not(.preset-apply-btn-active) {
    color: var(--text-regular) !important;
    border-color: var(--border-color) !important;
    background: var(--card-bg) !important;
  }

  .preset-apply-btn:not(.preset-apply-btn-active):hover {
    background: var(--primary-color) !important;
    color: white !important;
    border-color: var(--primary-color) !important;
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(37, 99, 235, 0.2);
  }

  html.dark .preset-apply-btn:not(.preset-apply-btn-active) {
    color: var(--el-text-color-regular) !important;
    border-color: var(--el-border-color) !important;
    background: var(--card-bg) !important;
  }

  html.dark .preset-apply-btn:not(.preset-apply-btn-active):hover {
    background: var(--primary-color) !important;
    color: white !important;
    border-color: var(--primary-color) !important;
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
  }

  .preset-title-group {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 1;
    min-width: 0;
  }

  .preset-title {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    color: var(--text-primary);
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .preset-badge {
    flex-shrink: 0;
    opacity: 0.8;
  }

  .preset-desc {
    margin: 0 0 12px 0;
    font-size: 13px;
    color: var(--text-secondary);
    line-height: 1.5;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .preset-preview {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 12px;
    border-top: 1px solid var(--border-light);
  }

  .preset-colors {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .preset-color-dot {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    border: 2px solid rgba(255, 255, 255, 0.3);
    cursor: pointer;
    transition: all 0.2s ease;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .preset-color-dot:hover {
    transform: scale(1.15);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
  }

  .preset-color-dot.is-primary {
    border-color: var(--primary-color);
    box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.2);
  }

  html.dark .preset-color-dot {
    border-color: rgba(255, 255, 255, 0.2);
  }

  html.dark .preset-color-dot.is-primary {
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.3);
  }

  .preset-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--text-secondary);
  }

  .preset-icon {
    font-size: 16px;
    line-height: 1;
  }

  /* 响应式 - 移动端：2列布局 */
  @media (max-width: 768px) {
    .preset-list {
      grid-template-columns: repeat(2, 1fr);
      gap: 12px;
    }

    .preset-card {
      padding: 12px;
    }
  }

  /* 超小屏幕：单列布局 */
  @media (max-width: 480px) {
    .preset-list {
      grid-template-columns: 1fr;
    }
  }

  /* 深色模式优化 */
  html.dark .color-item label {
    color: var(--text-regular);
  }

  html.dark .appearance-settings :deep(.el-form-item__label) {
    color: var(--text-primary);
  }

  html.dark .appearance-settings :deep(.el-radio-button__inner) {
    background-color: var(--card-bg);
    border-color: var(--border-color);
    color: var(--text-primary);
  }

  html.dark .appearance-settings :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
    background-color: var(--primary-color);
    border-color: var(--primary-color);
    color: var(--el-text-color-primary);
  }

  /* 响应式布局 */
  @media (max-width: 768px) {
    .color-picker-group {
      gap: 12px;
    }

    .color-item {
      flex-wrap: wrap;
      gap: 8px;
    }

    .color-item label {
      width: 100%;
      margin-bottom: 4px;
    }

    .auxiliary-modes {
      flex-direction: column;
      align-items: flex-start;
      gap: 16px;
    }

    .auxiliary-modes .el-switch {
      margin-left: 0 !important;
    }
  }

  /* 布局设置样式 */
  .layout-tip {
    margin-top: 12px;
    padding: 8px 12px;
    background: var(--el-fill-color-light);
    border-radius: 6px;
  }

  html.dark .layout-tip {
    background: var(--el-fill-color-light);
  }

  .sidebar-settings {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .setting-item {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .setting-item label {
    min-width: 80px;
    font-size: 14px;
    color: var(--text-regular);
    font-weight: 500;
  }

  .setting-item .unit {
    font-size: 14px;
    color: var(--text-secondary);
  }

  /* 剪贴板监听卡片样式 */
  .clipboard-monitor-card {
    padding: 20px;
    background: var(--card-bg);
    border: 2px solid var(--border-light);
    border-radius: 12px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
    overflow: hidden;
  }

  .clipboard-monitor-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: linear-gradient(90deg, var(--primary-color), var(--primary-hover));
    transform: scaleX(0);
    transition: transform 0.3s ease;
  }

  .clipboard-monitor-card.is-enabled {
    border-color: var(--primary-color);
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.03) 0%, rgba(79, 70, 229, 0.03) 100%);
    box-shadow: 0 2px 8px rgba(37, 99, 235, 0.1);
  }

  .clipboard-monitor-card.is-enabled::before {
    transform: scaleX(1);
  }

  .clipboard-monitor-card:hover {
    border-color: var(--primary-color);
    box-shadow: 0 4px 12px rgba(37, 99, 235, 0.15);
    transform: translateY(-2px);
  }

  .clipboard-monitor-card:hover::before {
    transform: scaleX(1);
  }

  html.dark .clipboard-monitor-card {
    background: rgba(30, 41, 59, 0.6);
    border-color: rgba(255, 255, 255, 0.1);
  }

  html.dark .clipboard-monitor-card.is-enabled {
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.08) 0%, rgba(99, 102, 241, 0.08) 100%);
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.2);
  }

  html.dark .clipboard-monitor-card:hover {
    border-color: var(--primary-color);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.25);
    background: rgba(30, 41, 59, 0.8);
  }

  .clipboard-monitor-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 12px;
    gap: 16px;
  }

  .clipboard-monitor-title-group {
    display: flex;
    align-items: center;
    gap: 12px;
    flex: 1;
    min-width: 0;
  }

  .clipboard-monitor-icon {
    color: var(--primary-color);
    flex-shrink: 0;
  }

  .clipboard-monitor-title-content {
    display: flex;
    flex-direction: column;
    gap: 6px;
    flex: 1;
    min-width: 0;
  }

  .clipboard-monitor-title {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    line-height: 1.4;
  }

  .clipboard-monitor-status {
    align-self: flex-start;
    font-weight: 500;
  }

  .clipboard-monitor-description {
    margin: 0 0 8px 0;
    font-size: 14px;
    color: var(--text-secondary);
    line-height: 1.6;
  }

  .clipboard-monitor-note {
    margin: 0 0 16px 0;
    font-size: 12px;
    color: var(--el-color-warning);
    line-height: 1.5;
  }

  .clipboard-monitor-features {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    padding-top: 16px;
    border-top: 1px solid var(--border-light);
  }

  .feature-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: var(--el-fill-color-lighter);
    border-radius: 8px;
    transition: all 0.2s ease;
  }

  .feature-item:hover {
    background: var(--el-fill-color-light);
    transform: translateY(-1px);
  }

  .feature-icon {
    color: var(--primary-color);
    font-size: 16px;
  }

  .feature-text {
    font-size: 13px;
    color: var(--text-regular);
    font-weight: 500;
  }

  html.dark .feature-item {
    background: rgba(255, 255, 255, 0.05);
  }

  html.dark .feature-item:hover {
    background: rgba(255, 255, 255, 0.08);
  }

  /* 响应式优化 */
  @media (max-width: 768px) {
    .clipboard-monitor-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;
    }

    .clipboard-monitor-features {
      flex-direction: column;
      gap: 12px;
    }

    .feature-item {
      width: 100%;
    }
  }

  /* 分组标题样式 */
  .divider-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    padding: 0 8px;
  }

  html.dark .divider-title {
    color: var(--el-text-color-primary);
  }

  /* 布局模式包装器 */
  .layout-mode-wrapper {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }
</style>
