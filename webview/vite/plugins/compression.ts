import compression from 'vite-plugin-compression'

export default (env: any) => {
  const { VITE_BUILD_COMPRESS } = env
  const plugin: any[] = []
  
  if (VITE_BUILD_COMPRESS) {
    const compressList = VITE_BUILD_COMPRESS.split(',')
    
    if (compressList.includes('gzip')) {
      // Gzip 压缩配置
      // 生成 .gz 文件，保留原始文件
      plugin.push(
        compression({
          ext: '.gz',
          deleteOriginFile: false
        })
      )
    }
    
    if (compressList.includes('brotli')) {
      // Brotli 压缩配置
      // 生成 .br 文件，压缩率通常比 gzip 更高
      plugin.push(
        compression({
          ext: '.br',
          algorithm: 'brotliCompress',
          deleteOriginFile: false
        })
      )
    }
  }
  
  return plugin
}

