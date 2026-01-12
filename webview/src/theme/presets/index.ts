/**
 * 主题预设
 */
import defaultPreset from './default.json'
import darkPreset from './dark.json'
import lightPreset from './light.json'
import grayscalePreset from './grayscale.json'

export interface ThemePreset {
  name: string
  desc: string
  theme: 'light' | 'dark' | 'auto'
  grayscale: boolean
  colourWeakness: boolean
  colors: {
    primary?: string
    secondary?: string
    success?: string
    warning?: string
    danger?: string
    info?: string
  }
}

export const themePresets: ThemePreset[] = [
  defaultPreset as ThemePreset,
  lightPreset as ThemePreset,
  darkPreset as ThemePreset,
  grayscalePreset as ThemePreset
]

export default themePresets
