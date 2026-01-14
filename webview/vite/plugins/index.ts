import vue from '@vitejs/plugin-vue'
import { codeInspectorPlugin } from 'code-inspector-plugin'
import createAutoImport from './auto-import'
import createComponents from './components'
import createCompression from './compression'
import path from 'path'

export default (env: any, isBuild = false): any[] => {
  const vitePlugins: any[] = []
  
  // Vue 插件配置（优化 HMR）
  vitePlugins.push(
    vue({
      // 模板编译选项
      template: {
        compilerOptions: {
          // 生产环境移除注释
          comments: isBuild ? false : true
        }
      },
      // 脚本设置选项
      script: {
        // 定义模型名称
        defineModel: true,
        // 启用 props 解构
        propsDestructure: true
      }
    })
  )
  
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

