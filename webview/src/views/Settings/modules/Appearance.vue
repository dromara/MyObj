<template>
  <div class="appearance-settings">
    <el-form label-width="120px">
      <!-- 主题设置 -->
      <el-form-item :label="t('settings.theme')">
        <el-radio-group v-model="currentTheme" @change="handleThemeChange">
          <el-radio-button label="light">
            <el-icon><Sunny /></el-icon>
            <span style="margin-left: 4px">{{ t('settings.light') }}</span>
          </el-radio-button>
          <el-radio-button label="dark">
            <el-icon><Moon /></el-icon>
            <span style="margin-left: 4px">{{ t('settings.dark') }}</span>
          </el-radio-button>
          <el-radio-button label="auto">
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

      <!-- 自定义主题色 -->
      <el-form-item :label="t('settings.themeColor')">
        <div class="color-picker-group">
          <div class="color-item">
            <label>{{ t('settings.primaryColor') }}</label>
            <el-color-picker v-model="displayColors.primary" @change="(val) => handleColorChange('primary', val)" />
            <el-button text size="small" @click="resetColor('primary')">{{ t('settings.reset') }}</el-button>
          </div>
          <div class="color-item">
            <label>{{ t('settings.successColor') }}</label>
            <el-color-picker v-model="displayColors.success" @change="(val) => handleColorChange('success', val)" />
            <el-button text size="small" @click="resetColor('success')">{{ t('settings.reset') }}</el-button>
          </div>
          <div class="color-item">
            <label>{{ t('settings.warningColor') }}</label>
            <el-color-picker v-model="displayColors.warning" @change="(val) => handleColorChange('warning', val)" />
            <el-button text size="small" @click="resetColor('warning')">{{ t('settings.reset') }}</el-button>
          </div>
          <div class="color-item">
            <label>{{ t('settings.dangerColor') }}</label>
            <el-color-picker v-model="displayColors.danger" @change="(val) => handleColorChange('danger', val)" />
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
        <el-select v-model="selectedPreset" @change="handlePresetChange" style="width: 300px">
          <el-option
            v-for="preset in themePresets"
            :key="preset.name"
            :label="getPresetName(preset.name)"
            :value="preset.name"
          >
            <div>
              <div style="font-weight: 500">{{ getPresetName(preset.name) }}</div>
              <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px">
                {{ getPresetDesc(preset.name) }}
              </div>
            </div>
          </el-option>
        </el-select>
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
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { useTheme } from '@/composables/useTheme'
import { useI18n } from '@/composables/useI18n'
import { LanguageEnum } from '@/enums/LanguageEnum'
import { themePresets } from '@/theme/presets'

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
const { proxy } = getCurrentInstance() as ComponentInternalInstance
const { t } = useI18n()

const currentTheme = ref(theme.value)
const currentLocale = ref(locale.value)
const currentGrayscale = ref(grayscale.value)
const currentColourWeakness = ref(colourWeakness.value)
const backgroundPattern = ref<'none' | 'grid' | 'dots' | 'gradient' | 'waves' | 'particles'>('none')
const selectedPreset = ref<string>('')

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
watch(customColors, (newColors) => {
  displayColors.value = {
    primary: newColors.primary || defaultColors.primary,
    success: newColors.success || defaultColors.success,
    warning: newColors.warning || defaultColors.warning,
    danger: newColors.danger || defaultColors.danger
  }
}, { deep: true, immediate: true })

// 从 localStorage 加载背景图案设置
onMounted(() => {
  const saved = localStorage.getItem('backgroundPattern')
  if (saved && ['none', 'grid', 'dots', 'gradient', 'waves', 'particles'].includes(saved)) {
    backgroundPattern.value = saved as any
  }
})

// 监听主题变化
watch(theme, (newTheme) => {
  currentTheme.value = newTheme
})

// 监听语言变化
watch(locale, (newLocale) => {
  currentLocale.value = newLocale
})

// 监听自定义颜色变化，同步到显示颜色（displayColors 是 computed，会自动更新）

// 监听灰度模式变化
watch(grayscale, (newValue) => {
  currentGrayscale.value = newValue
})

// 监听色弱模式变化
watch(colourWeakness, (newValue) => {
  currentColourWeakness.value = newValue
})

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
  }
  return presetKey
}

const handlePresetChange = (presetName: string) => {
  const preset = themePresets.find(p => p.name === presetName)
  if (preset) {
    applyPreset(preset)
    proxy?.$modal.msgSuccess(t('settings.presetApplied', { name: getPresetName(preset.name) }))
  }
}

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
</style>
