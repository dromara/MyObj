import { defineConfig, loadEnv } from 'vite'
import { resolve } from 'path'
import createPlugins from './vite/plugins'
// @ts-ignore - autoprefixer 类型定义可能不完整
import autoprefixer from 'autoprefixer'

// https://vitejs.dev/config/
export default defineConfig(({ mode, command }) => {
  // 加载环境变量
  const env = loadEnv(mode, process.cwd())
  
  return {
    // 部署路径（如果需要部署到子路径，可通过环境变量配置）
    base: env.VITE_APP_BASE_PATH || '/',
    
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src')
      },
      // 文件扩展名解析顺序
      extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json', '.vue']
    },
    
    // 插件配置
    plugins: createPlugins(env, command === 'build'),
    
    // 开发服务器配置
    server: {
      host: '0.0.0.0', // 允许外部访问
      port: Number(env.VITE_APP_PORT) || 5173,
      open: true, // 自动打开浏览器
      proxy: {
        [env.VITE_APP_BASE_API || '/dev-api']: {
          target: env.VITE_APP_BASE_URL || 'http://localhost:8080',
          changeOrigin: true,
          ws: true, // 支持 WebSocket
          // 将代理路径（如 /dev-api）重写为后端实际路径（/api）
          rewrite: (path) => path.replace(new RegExp('^' + (env.VITE_APP_BASE_API || '/dev-api')), '/api')
        }
      }
    },
    
    // CSS 配置
    css: {
      postcss: {
        plugins: [
          // 自动添加浏览器兼容性前缀
          autoprefixer(),
          // 移除 @charset 规则（避免警告）
          {
            postcssPlugin: 'internal:charset-removal',
            AtRule: {
              charset: (atRule: any) => {
                atRule.remove()
              }
            }
          }
        ]
      }
    },
    
    // 依赖预编译优化
    optimizeDeps: {
      include: [
        'vue',
        'vue-router',
        'element-plus',
        'element-plus/es/components/**/css',
        '@element-plus/icons-vue',
        'axios',
        'jsencrypt',
        'spark-md5'
      ]
    },
    
    // 构建配置
    build: {
      target: 'es2015', // 构建目标
      outDir: 'dist', // 输出目录
      assetsDir: 'assets', // 静态资源目录
      sourcemap: false, // 生产环境不生成 sourcemap
      // 禁用压缩大小警告
      chunkSizeWarningLimit: 1500,
      rollupOptions: {
        output: {
          // 手动分包策略
          manualChunks: {
            // Vue 相关
            'vue-vendor': ['vue', 'vue-router'],
            // Element Plus 相关
            'element-plus-vendor': ['element-plus', '@element-plus/icons-vue'],
            // 工具库
            'utils-vendor': ['axios', 'jsencrypt', 'spark-md5']
          }
        }
      }
    }
  }
})
