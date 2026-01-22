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
      // 热重载优化配置
      hmr: {
        overlay: true, // 显示错误覆盖层
        protocol: 'ws', // 使用 WebSocket 协议
        host: 'localhost', // HMR 服务器主机
        port: Number(env.VITE_APP_PORT) || 5173 // HMR 服务器端口
      },
      // 监听配置
      watch: {
        // 忽略某些文件的监听，提升性能
        ignored: [
          '**/node_modules/**',
          '**/dist/**',
          '**/.git/**',
          '**/logs/**',
          '**/temp/**'
        ],
        // 使用轮询模式（在某些文件系统上更可靠）
        usePolling: false,
        // 聚合延迟（等待文件变化后再触发更新）
        aggregateTimeout: 300
      },
      // 预加载优化
      preTransformRequests: true,
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
      // 启用 CSS 代码分割
      cssCodeSplit: true,
      // 压缩配置
      minify: 'esbuild', // 使用 esbuild 压缩，速度更快
      // 清理输出目录
      emptyOutDir: true,
      rollupOptions: {
        output: {
          // 文件命名规则
          chunkFileNames: 'assets/js/[name]-[hash].js',
          entryFileNames: 'assets/js/[name]-[hash].js',
          assetFileNames: (assetInfo) => {
            const info = assetInfo.name?.split('.') || []
            const ext = info[info.length - 1]
            if (/\.(mp4|webm|ogg|mp3|wav|flac|aac)(\?.*)?$/i.test(assetInfo.name || '')) {
              return 'assets/media/[name]-[hash].[ext]'
            }
            if (/\.(png|jpe?g|gif|svg|webp|avif)(\?.*)?$/i.test(assetInfo.name || '')) {
              return 'assets/images/[name]-[hash].[ext]'
            }
            if (/\.(woff2?|eot|ttf|otf)(\?.*)?$/i.test(assetInfo.name || '')) {
              return 'assets/fonts/[name]-[hash].[ext]'
            }
            if (ext === 'css') {
              return 'assets/css/[name]-[hash].[ext]'
            }
            return 'assets/[name]-[hash].[ext]'
          }
        }
      }
    }
  }
})
