/**
 * 颜色工具函数
 * 参考 plus-ui 和 ruoyi-plus-soybean 的实现
 */

// hex颜色转rgb颜色
export const hexToRgb = (str: string): number[] => {
  str = str.replace('#', '')
  const hexs = str.match(/../g)
  if (!hexs) return [0, 0, 0]

  const rgb: number[] = []
  for (let i = 0; i < 3; i++) {
    rgb[i] = parseInt(hexs[i], 16)
  }
  return rgb
}

// rgb颜色转Hex颜色
export const rgbToHex = (r: number, g: number, b: number): string => {
  const hexs = [r.toString(16), g.toString(16), b.toString(16)]
  for (let i = 0; i < 3; i++) {
    if (hexs[i].length === 1) {
      hexs[i] = `0${hexs[i]}`
    }
  }
  return `#${hexs.join('')}`
}

// 变浅颜色值
export const getLightColor = (color: string, level: number): string => {
  const rgb = hexToRgb(color)
  const newRgb: number[] = []
  for (let i = 0; i < 3; i++) {
    const s = (255 - rgb[i]) * level + rgb[i]
    newRgb[i] = Math.floor(s)
  }
  return rgbToHex(newRgb[0], newRgb[1], newRgb[2])
}

// 变深颜色值
export const getDarkColor = (color: string, level: number): string => {
  const rgb = hexToRgb(color)
  const newRgb: number[] = []
  for (let i = 0; i < 3; i++) {
    newRgb[i] = Math.floor(rgb[i] * (1 - level))
  }
  return rgbToHex(newRgb[0], newRgb[1], newRgb[2])
}

/**
 * 生成完整的颜色调色板
 * 生成从 50 到 950 的颜色变体（类似 Tailwind CSS）
 */
export interface ColorPalette {
  50: string
  100: string
  200: string
  300: string
  400: string
  500: string
  600: string
  700: string
  800: string
  900: string
  950: string
}

/**
 * 生成颜色调色板
 * @param baseColor 基础颜色（hex格式）
 * @returns 颜色调色板对象
 */
export const generateColorPalette = (baseColor: string): ColorPalette => {
  const palette: ColorPalette = {
    50: getLightColor(baseColor, 0.95), // 最浅
    100: getLightColor(baseColor, 0.9),
    200: getLightColor(baseColor, 0.75),
    300: getLightColor(baseColor, 0.5),
    400: getLightColor(baseColor, 0.25),
    500: baseColor, // 基础色
    600: getDarkColor(baseColor, 0.2),
    700: getDarkColor(baseColor, 0.4),
    800: getDarkColor(baseColor, 0.6),
    900: getDarkColor(baseColor, 0.8),
    950: getDarkColor(baseColor, 0.9) // 最深
  }
  return palette
}

/**
 * 应用颜色调色板到 CSS 变量
 * @param colorName 颜色名称（如 'primary', 'success'）
 * @param palette 颜色调色板
 */
export const applyColorPaletteToCSS = (colorName: string, palette: ColorPalette) => {
  const root = document.documentElement

  // 设置基础颜色
  root.style.setProperty(`--${colorName}-color`, palette[500])

  // 设置所有调色板级别
  Object.entries(palette).forEach(([level, color]) => {
    root.style.setProperty(`--${colorName}-color-${level}`, color)
  })

  // 设置常用的 hover 和 active 颜色
  root.style.setProperty(`--${colorName}-color-hover`, palette[600])
  root.style.setProperty(`--${colorName}-color-active`, palette[700])
  root.style.setProperty(`--${colorName}-color-light`, palette[100])
  root.style.setProperty(`--${colorName}-color-lighter`, palette[50])
  root.style.setProperty(`--${colorName}-color-dark`, palette[800])
  root.style.setProperty(`--${colorName}-color-darker`, palette[900])
}

/**
 * 应用多个颜色的调色板
 * @param colors 颜色对象，键为颜色名称，值为颜色值
 */
export const applyColorPalettes = (colors: Record<string, string>) => {
  Object.entries(colors).forEach(([name, color]) => {
    if (color) {
      const palette = generateColorPalette(color)
      applyColorPaletteToCSS(name, palette)
    }
  })
}
