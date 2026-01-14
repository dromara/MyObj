/**
 * 构建分析脚本
 * 分析打包后的文件大小
 */
import { readdirSync, statSync, existsSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

const distPath = resolve(__dirname, '../dist')

function formatSize(bytes) {
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)}KB`
  return `${(bytes / 1024 / 1024).toFixed(2)}MB`
}

function analyzeDirectory(dir) {
  const files = []
  
  try {
    const entries = readdirSync(dir)
    
    for (const entry of entries) {
      const fullPath = resolve(dir, entry)
      const stat = statSync(fullPath)
      
      if (stat.isFile()) {
        const size = stat.size
        files.push({
          name: entry,
          size,
          sizeFormatted: formatSize(size)
        })
      } else if (stat.isDirectory()) {
        files.push(...analyzeDirectory(fullPath))
      }
    }
  } catch (error) {
    console.error(`无法读取目录 ${dir}:`, error)
  }
  
  return files
}

console.log('📊 开始分析构建产物...\n')

if (!existsSync(distPath)) {
  console.error('❌ dist 目录不存在，请先执行构建！')
  process.exit(1)
}

const files = analyzeDirectory(distPath)
const totalSize = files.reduce((sum, file) => sum + file.size, 0)

console.log('📦 文件大小统计：\n')
console.log('文件名'.padEnd(50), '大小'.padStart(10))
console.log('-'.repeat(60))

files
  .sort((a, b) => b.size - a.size)
  .forEach(file => {
    console.log(file.name.padEnd(50), file.sizeFormatted.padStart(10))
  })

console.log('-'.repeat(60))
console.log('总计'.padEnd(50), formatSize(totalSize).padStart(10))
console.log('\n✅ 分析完成！')
