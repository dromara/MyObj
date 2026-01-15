/**
 * TypeScript 类型检查脚本
 */
import { execSync } from 'child_process'
import { resolve } from 'path'

const projectRoot = resolve(__dirname, '..')

console.log('🔍 开始类型检查...\n')

try {
  execSync('tsc --noEmit', {
    cwd: projectRoot,
    stdio: 'inherit'
  })
  console.log('\n✅ 类型检查通过！')
  process.exit(0)
} catch (error) {
  console.error('\n❌ 类型检查失败！')
  process.exit(1)
}
