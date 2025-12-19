import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default (path: any) => {
  return Components({
    resolvers: [
      // 自动导入 Element Plus 组件
      ElementPlusResolver(),
    ],
    // 自动导入 src/components 目录下的组件
    dirs: ['src/components'],
    // 组件名称包含的目录
    directoryAsNamespace: false,
    // 类型声明文件生成路径
    dts: path.resolve(path.resolve(__dirname, '../../src'), 'types', 'components.d.ts'),
  })
}

