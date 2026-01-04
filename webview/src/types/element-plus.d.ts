import type * as ep from 'element-plus'

declare module 'element-plus/dist/locale/zh-cn';
declare global {
  // Element Plus 类型声明
  declare type FormInstance = ep.FormInstance
  declare type FormRules = ep.FormRules
  declare type TableInstance = ep.TableInstance
  declare type UploadInstance = ep.UploadInstance
  declare type ScrollbarInstance = ep.ScrollbarInstance
  declare type InputInstance = ep.InputInstance
  declare type SelectInstance = InstanceType<typeof ep.ElSelect>
  declare type DialogInstance = InstanceType<typeof ep.ElDialog>
  declare type TreeInstance = InstanceType<typeof ep.ElTree>
  declare type TreeSelectInstance = InstanceType<typeof ep.ElTreeSelect>
}

export {}
