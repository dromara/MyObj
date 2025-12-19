import vue from '@vitejs/plugin-vue'
import { codeInspectorPlugin } from 'code-inspector-plugin'
import createAutoImport from './auto-import'
import createComponents from './components'
import createCompression from './compression'
import path from 'path'

export default (env: any, isBuild = false): any[] => {
  const vitePlugins: any[] = []
  
  vitePlugins.push(vue())
  
  // 开发环境才启用代码检查器
  if (!isBuild) {
    vitePlugins.push(
      codeInspectorPlugin({
        bundler: 'vite',
      })
    )
  }
  
  vitePlugins.push(createAutoImport(path))
  vitePlugins.push(createComponents(path))
  
  // 构建时才启用压缩
  if (isBuild) {
    vitePlugins.push(...createCompression(env))
  }
  
  return vitePlugins
}

