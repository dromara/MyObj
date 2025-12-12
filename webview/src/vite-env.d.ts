/* eslint-disable */
/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface ImportMetaEnv {
  readonly MODE: string
  readonly VITE_APP_TITLE: string
  readonly VITE_APP_VERSION: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
