/**
 * Prettier 配置文件
 * 用于代码格式化
 */
module.exports = {
  // 基本配置
  semi: false, // 不使用分号（与 ESLint 保持一致）
  singleQuote: true, // 使用单引号
  quoteProps: 'as-needed', // 仅在需要时添加引号
  trailingComma: 'none', // 不使用尾随逗号（与 ESLint 保持一致）
  tabWidth: 2, // 缩进宽度
  useTabs: false, // 使用空格而非制表符
  printWidth: 120, // 行宽（与 ESLint max-len 保持一致）
  
  // 对象和数组
  bracketSpacing: true, // 对象括号内空格 { foo: bar }
  bracketSameLine: false, // 标签的 > 单独一行
  arrowParens: 'avoid', // 箭头函数参数尽可能省略括号
  
  // HTML/Vue
  htmlWhitespaceSensitivity: 'css', // HTML 空白字符敏感度
  vueIndentScriptAndStyle: true, // Vue 文件中的 <script> 和 <style> 标签缩进
  
  // 其他
  endOfLine: 'lf', // 使用 LF 作为行尾符
  embeddedLanguageFormatting: 'auto', // 自动格式化嵌入的语言
  insertPragma: false, // 不插入 @prettier 标记
  requirePragma: false, // 不需要 @prettier 标记
  
  // 文件覆盖
  overrides: [
    {
      files: '*.json',
      options: {
        printWidth: 200 // JSON 文件允许更长的行
      }
    },
    {
      files: '*.md',
      options: {
        printWidth: 100, // Markdown 文件使用较短的行宽
        proseWrap: 'preserve' // 保留换行
      }
    },
    {
      files: '*.vue',
      options: {
        parser: 'vue'
      }
    }
  ]
}
