/**
 * ESLint 配置文件
 * 用于代码质量检查和规范
 */
module.exports = {
  root: true,
  env: {
    browser: true,
    es2021: true,
    node: true
  },
  extends: [
    'eslint:recommended',
    'plugin:vue/vue3-essential',
    'plugin:vue/vue3-strongly-recommended',
    'plugin:vue/vue3-recommended'
  ],
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module',
    parser: '@typescript-eslint/parser'
  },
  plugins: ['vue', '@typescript-eslint'],
  rules: {
    // Vue 相关规则
    'vue/multi-word-component-names': 'off', // 允许单单词组件名
    'vue/no-v-html': 'warn', // 警告使用 v-html
    'vue/require-default-prop': 'off', // 不要求 props 有默认值
    'vue/require-explicit-emits': 'warn', // 警告未显式声明的 emits
    'vue/html-self-closing': [
      'error',
      {
        html: {
          void: 'always',
          normal: 'never',
          component: 'always'
        },
        svg: 'always',
        math: 'always'
      }
    ],
    'vue/max-attributes-per-line': [
      'error',
      {
        singleline: 3,
        multiline: 1
      }
    ],
    'vue/singleline-html-element-content-newline': 'off',
    'vue/multiline-html-element-content-newline': 'off',
    
    // TypeScript 相关规则
    '@typescript-eslint/no-explicit-any': 'warn', // 警告使用 any
    '@typescript-eslint/no-unused-vars': [
      'warn',
      {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_'
      }
    ],
    '@typescript-eslint/explicit-function-return-type': 'off',
    '@typescript-eslint/explicit-module-boundary-types': 'off',
    
    // 通用规则
    'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
    'no-debugger': process.env.NODE_ENV === 'production' ? 'error' : 'off',
    'no-unused-vars': 'off', // 使用 TypeScript 版本
    'prefer-const': 'error',
    'no-var': 'error',
    'object-shorthand': 'error',
    'quote-props': ['error', 'as-needed'],
    'arrow-body-style': ['error', 'as-needed'],
    'prefer-arrow-callback': 'error',
    
    // 代码风格
    'indent': ['error', 2, { SwitchCase: 1 }],
    'quotes': ['error', 'single', { avoidEscape: true }],
    'semi': ['error', 'never'],
    'comma-dangle': ['error', 'never'],
    'max-len': [
      'warn',
      {
        code: 120,
        ignoreUrls: true,
        ignoreStrings: true,
        ignoreTemplateLiterals: true,
        ignoreRegExpLiterals: true
      }
    ],
    
    // 最佳实践
    'eqeqeq': ['error', 'always', { null: 'ignore' }],
    'curly': ['error', 'all'],
    'no-eval': 'error',
    'no-implied-eval': 'error',
    'no-new-func': 'error',
    'no-return-assign': 'error',
    'no-sequences': 'error',
    'no-throw-literal': 'error',
    'no-unmodified-loop-condition': 'error',
    'no-unused-expressions': 'error',
    'no-useless-call': 'error',
    'no-useless-concat': 'error',
    'no-useless-return': 'error',
    'prefer-promise-reject-errors': 'error',
    'radix': 'error',
    'require-await': 'warn',
    'yoda': 'error'
  },
  globals: {
    // 从 .eslintrc-auto-import.json 继承的全局变量
    ...require('./.eslintrc-auto-import.json').globals
  },
  overrides: [
    {
      files: ['*.vue'],
      rules: {
        'indent': 'off' // Vue 文件使用 prettier 处理缩进
      }
    },
    {
      files: ['*.ts', '*.tsx'],
      rules: {
        'no-undef': 'off' // TypeScript 会处理未定义变量
      }
    }
  ]
}
