<template>
  <el-dialog
    v-model="visible"
    title="分享文件"
    width="600px"
    :close-on-click-modal="false"
    :close-on-press-escape="true"
    class="share-dialog"
    @close="handleClose"
  >
    <!-- 文件信息卡片 -->
    <div class="file-info-card">
      <el-icon :size="48" color="#409EFF"><Document /></el-icon>
      <div class="file-info-content">
        <div class="file-name">{{ fileInfo.file_name || '未知文件' }}</div>
        <div class="file-size" v-if="fileInfo.file_size">
          {{ formatFileSize(fileInfo.file_size) }}
        </div>
      </div>
    </div>

    <!-- 分享设置 -->
    <el-form :model="shareForm" label-width="100px" class="share-form">
      <el-form-item label="有效期">
        <!-- 移动端使用下拉选择框 -->
        <el-select 
          v-model="shareForm.expire_days" 
          class="expire-select mobile-only"
          @change="handleExpireChange"
        >
          <el-option
            v-for="option in expireOptions"
            :key="option.value"
            :label="option.label"
            :value="option.value"
          />
        </el-select>
        
        <!-- 桌面端使用单选按钮组 -->
        <el-radio-group 
          v-model="shareForm.expire_days" 
          class="expire-options desktop-only"
          @change="handleExpireChange"
        >
          <el-radio-button 
            v-for="option in expireOptions" 
            :key="option.value" 
            :label="option.value"
          >
            {{ option.label }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>
      
      <el-form-item label="访问密码">
        <el-input 
          v-model="shareForm.password" 
          placeholder="请输入访问密码（可选，留空则无需密码）"
          maxlength="20"
          show-word-limit
          clearable
        >
          <template #append>
            <el-button @click="generateRandomPassword" icon="Refresh">随机生成</el-button>
          </template>
        </el-input>
        <div class="form-tip">设置密码后，访问者需要输入密码才能下载文件；不设置密码则任何人都可以通过链接下载</div>
      </el-form-item>
    </el-form>

    <!-- 分享结果（分享成功后显示） -->
    <div v-if="shareResult" class="share-result">
      <el-alert
        type="success"
        :closable="false"
        show-icon
        class="result-alert"
      >
        <template #title>
          <div class="result-title">分享创建成功！</div>
        </template>
      </el-alert>
      
      <div class="share-link-section">
        <div class="link-label">分享链接</div>
        <div class="link-content">
          <el-input
            :model-value="shareResult.shareUrl"
            readonly
            class="link-input"
          >
            <template #append>
              <el-button 
                :icon="shareResult.copied ? 'Check' : 'CopyDocument'" 
                @click="copyShareLink"
                :type="shareResult.copied ? 'success' : 'primary'"
              >
                {{ shareResult.copied ? '已复制' : '复制链接' }}
              </el-button>
            </template>
          </el-input>
        </div>
        
        <div v-if="shareForm.password" class="password-section">
          <div class="link-label">访问密码</div>
          <div class="link-content">
            <el-input
              :model-value="shareForm.password"
              readonly
              class="link-input"
            >
              <template #append>
                <el-button 
                  :icon="shareResult.passwordCopied ? 'Check' : 'CopyDocument'" 
                  @click="copyPassword"
                  :type="shareResult.passwordCopied ? 'success' : 'primary'"
                >
                  {{ shareResult.passwordCopied ? '已复制' : '复制密码' }}
                </el-button>
              </template>
            </el-input>
          </div>
        </div>
        
        <div class="expire-info">
          <el-icon><Clock /></el-icon>
          <span>有效期：{{ shareResult.expireText }}</span>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button 
          v-if="!shareResult"
          type="primary" 
          :loading="sharing" 
          @click="handleConfirmShare"
        >
          创建分享
        </el-button>
        <el-button 
          v-else
          type="primary" 
          @click="handleCreateAnother"
        >
          继续分享
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { createShare } from '@/api/share'
import type { CreateShareRequest } from '@/types'
import { formatSize, generateRandomPassword as generatePassword, copyToClipboard, getShareUrl } from '@/utils'

interface Props {
  modelValue: boolean
  fileInfo: {
    file_id: string
    file_name: string
    file_size?: number
  }
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  fileInfo: () => ({
    file_id: '',
    file_name: '',
    file_size: 0
  })
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'success': [shareUrl: string, password: string]
}>()

const { proxy } = getCurrentInstance() as ComponentInternalInstance

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const sharing = ref(false)
const shareForm = reactive({
  expire_days: 7,
  password: ''
})

const shareResult = ref<{
  shareUrl: string
  expireText: string
  copied: boolean
  passwordCopied: boolean
} | null>(null)

const expireOptions = [
  { label: '1天', value: 1 },
  { label: '7天', value: 7 },
  { label: '30天', value: 30 },
  { label: '永久', value: 0 }
]

const formatFileSize = (size: number) => {
  return formatSize(size)
}

const generateRandomPassword = () => {
  shareForm.password = generatePassword()
}

const handleExpireChange = (value: string | number | boolean | undefined) => {
  // 确保值正确更新
  if (typeof value === 'number') {
    shareForm.expire_days = value
  }
}

const handleConfirmShare = async () => {
  sharing.value = true
  try {
    // 计算过期时间
    const expireDate = new Date()
    if (shareForm.expire_days > 0) {
      expireDate.setDate(expireDate.getDate() + shareForm.expire_days)
    } else {
      // 永久有效，设置为很远的未来
      expireDate.setFullYear(expireDate.getFullYear() + 100)
    }
    const expireStr = expireDate.toISOString().slice(0, 19).replace('T', ' ')
    
    const res = await createShare({
      file_id: props.fileInfo.file_id,
      expire: expireStr,
      password: shareForm.password
    } as CreateShareRequest)
    
    if (res.code === 200) {
      // 后端返回的 token，构建分享链接
      const token = res.data.split('/').pop()
      const shareUrl = getShareUrl(token || '')
      
      const expireText = shareForm.expire_days === 0 
        ? '永久有效' 
        : `${shareForm.expire_days}天后过期`
      
      shareResult.value = {
        shareUrl,
        expireText,
        copied: false,
        passwordCopied: false
      }
      
      // 自动复制链接
      await copyShareLink()
      
      emit('success', shareUrl, shareForm.password)
    } else {
      proxy?.$modal.msgError(res.message || '分享失败')
    }
  } catch (error: any) {
    proxy?.$modal.msgError(error.message || '分享失败')
  } finally {
    sharing.value = false
  }
}

const copyShareLink = async () => {
  if (!shareResult.value) return
  
  const success = await copyToClipboard(shareResult.value.shareUrl)
  if (success) {
    shareResult.value.copied = true
    proxy?.$modal.msgSuccess('链接已复制到剪贴板')
    setTimeout(() => {
      if (shareResult.value) {
        shareResult.value.copied = false
      }
    }, 2000)
  } else {
    proxy?.$modal.msgError('复制失败')
  }
}

const copyPassword = async () => {
  if (!shareResult.value || !shareForm.password) return
  
  const success = await copyToClipboard(shareForm.password)
  if (success) {
    shareResult.value.passwordCopied = true
    proxy?.$modal.msgSuccess('密码已复制到剪贴板')
    setTimeout(() => {
      if (shareResult.value) {
        shareResult.value.passwordCopied = false
      }
    }, 2000)
  } else {
    proxy?.$modal.msgError('复制失败')
  }
}

const handleClose = () => {
  visible.value = false
  // 重置表单
  shareForm.expire_days = 7
  shareForm.password = ''
  shareResult.value = null
}

const handleCreateAnother = () => {
  shareResult.value = null
  shareForm.expire_days = 7
  shareForm.password = ''
}
</script>

<style scoped>
.share-dialog :deep(.el-dialog) {
  box-sizing: border-box;
  /* 确保弹窗不会超出屏幕 */
  max-width: 100vw;
}

.share-dialog :deep(.el-dialog__body) {
  padding: 24px;
  box-sizing: border-box;
  width: 100%;
  max-width: 100%;
  /* 防止内容溢出 */
  overflow-x: hidden;
}

/* 确保所有内部元素使用 border-box */
.share-dialog :deep(*) {
  box-sizing: border-box;
}

.share-dialog :deep(.el-dialog__body > *) {
  max-width: 100%;
  overflow-x: hidden;
  box-sizing: border-box;
}

/* 确保弹窗容器不会超出屏幕 */
.share-dialog {
  max-width: 100vw;
  overflow-x: hidden;
}

.file-info-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border-radius: 12px;
  margin-bottom: 24px;
  border: 1px solid #bae6fd;
}

.file-info-content {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 13px;
  color: var(--text-secondary);
}

.share-form {
  margin-bottom: 24px;
}

/* 移动端下拉选择框 */
.expire-select.mobile-only {
  width: 100%;
  display: none; /* 默认隐藏，移动端显示 */
}

/* 桌面端单选按钮组 */
.expire-options.desktop-only {
  width: 100%;
  display: flex;
  gap: 8px;
}

.expire-options :deep(.el-radio-button) {
  flex: 1;
  min-width: 0;
}

.expire-options :deep(.el-radio-button__inner) {
  width: 100%;
  cursor: pointer;
  user-select: none;
  -webkit-tap-highlight-color: transparent;
}

.expire-options :deep(.el-radio-button__original-radio) {
  position: absolute;
  opacity: 0;
  width: 100%;
  height: 100%;
  margin: 0;
  cursor: pointer;
}

/* 桌面端显示单选按钮组，隐藏下拉选择框 */
@media screen and (min-width: 769px) {
  .expire-select.mobile-only {
    display: none !important;
  }
  
  .expire-options.desktop-only {
    display: flex !important;
  }
}

.form-tip {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
  line-height: 1.5;
}

.share-result {
  margin-top: 24px;
}

.result-alert {
  margin-bottom: 20px;
}

.result-title {
  font-size: 15px;
  font-weight: 600;
}

.share-link-section {
  background: var(--bg-color);
  border-radius: 8px;
  padding: 20px;
  border: 1px solid var(--border-color);
}

.link-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.link-content {
  margin-bottom: 16px;
}

.link-content:last-child {
  margin-bottom: 0;
}

.link-input {
  width: 100%;
}

.password-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}

.expire-info {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
  font-size: 13px;
  color: var(--text-secondary);
}

.expire-info .el-icon {
  font-size: 14px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 移动端响应式 */
@media (max-width: 1024px) {
  /* 确保弹窗容器不超出屏幕 */
  .share-dialog {
    width: 100vw !important;
    max-width: 100vw !important;
    overflow-x: hidden !important;
    position: fixed !important;
    left: 0 !important;
    right: 0 !important;
  }
  
  .share-dialog :deep(.el-dialog) {
    width: 100% !important;
    max-width: 100vw !important;
    margin: 0 !important;
    max-height: 100vh;
    box-sizing: border-box !important;
    /* 确保弹窗不会超出屏幕 */
    left: 0 !important;
    right: 0 !important;
    transform: none !important;
    border-radius: 0 !important;
    /* 防止横向滚动 */
    overflow-x: hidden !important;
    position: fixed !important;
  }
  
  /* 覆盖 Element Plus 的默认宽度 */
  .share-dialog :deep(.el-dialog__wrapper) {
    width: 100vw !important;
    max-width: 100vw !important;
    overflow-x: hidden !important;
  }
  
  .share-dialog :deep(.el-dialog__body) {
    padding: 16px;
    max-height: calc(100vh - 120px);
    overflow-y: auto;
    overflow-x: hidden !important;
    box-sizing: border-box;
    width: 100% !important;
    max-width: 100% !important;
  }
  
  .share-dialog :deep(.el-dialog__header) {
    padding: 16px;
    box-sizing: border-box;
    width: 100% !important;
    max-width: 100% !important;
    overflow-x: hidden;
  }
  
  .share-dialog :deep(.el-dialog__footer) {
    padding: 12px 16px;
    box-sizing: border-box;
    width: 100% !important;
    max-width: 100% !important;
    overflow-x: hidden;
  }
  
  /* 确保所有内部元素不超出 */
  .share-dialog :deep(.el-dialog__body > *) {
    max-width: 100% !important;
    box-sizing: border-box !important;
    overflow-x: hidden !important;
    width: 100% !important;
  }
  
  .file-info-card {
    padding: 16px;
    gap: 12px;
    flex-direction: row;
    align-items: center;
    width: 100% !important;
    max-width: 100% !important;
    box-sizing: border-box !important;
    overflow-x: hidden !important;
  }
  
  .file-info-card .el-icon {
    flex-shrink: 0;
  }
  
  .file-info-content {
    flex: 1;
    min-width: 0;
  }
  
  .file-name {
    font-size: 14px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  
  .file-size {
    font-size: 12px;
  }
  
  .share-form {
    margin-bottom: 20px;
  }
  
  .share-form :deep(.el-form-item) {
    margin-bottom: 24px;
    display: flex !important;
    flex-direction: column !important; /* 移动端垂直布局 */
    align-items: flex-start !important;
    /* 覆盖 Element Plus 默认的水平布局 */
    flex-wrap: nowrap;
  }
  
  /* 确保标签容器也使用垂直布局 */
  .share-form :deep(.el-form-item__label-wrap) {
    width: 100% !important;
    margin-right: 0 !important;
    padding-right: 0 !important;
  }
  
  .share-form :deep(.el-form-item__label) {
    width: 100% !important;
    text-align: left !important;
    margin-bottom: 10px;
    margin-right: 0 !important;
    padding: 0 !important;
    line-height: 1.5;
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary, #303133);
    /* 确保标签靠左对齐 */
    justify-content: flex-start;
    display: block;
  }
  
  .share-form :deep(.el-form-item__content) {
    margin-left: 0 !important;
    width: 100%;
    box-sizing: border-box;
  }
  
  /* 移动端显示下拉选择框，隐藏单选按钮组 */
  .expire-select.mobile-only {
    display: block !important;
    width: 100%;
  }
  
  .expire-options.desktop-only {
    display: none !important;
  }
  
  /* 输入框在移动端优化 */
  .share-form :deep(.el-input) {
    width: 100%;
    max-width: 100%;
    box-sizing: border-box;
  }
  
  .share-form :deep(.el-input__wrapper) {
    width: 100%;
    max-width: 100%;
    box-sizing: border-box;
    border-radius: 8px;
  }
  
  /* 下拉选择框样式优化 */
  .share-form :deep(.el-select) {
    width: 100%;
  }
  
  .share-form :deep(.el-select .el-input__wrapper) {
    border-radius: 8px;
  }
  
  .share-form :deep(.el-input__append) {
    padding: 0;
    flex-shrink: 0;
  }
  
  .share-form :deep(.el-input__append .el-button) {
    padding: 0 12px;
    font-size: 12px;
    white-space: nowrap;
    min-width: auto;
    height: 100%;
    border-radius: 0 8px 8px 0;
  }
  
  .form-tip {
    font-size: 12px;
    color: var(--text-secondary, #909399);
    margin-top: 8px;
    line-height: 1.5;
  }
  
  .share-result {
    margin-top: 20px;
  }
  
  .result-alert {
    margin-bottom: 16px;
  }
  
  .result-title {
    font-size: 14px;
  }
  
  .share-link-section {
    padding: 16px;
  }
  
  .link-label {
    font-size: 12px;
    margin-bottom: 8px;
    font-weight: 600;
  }
  
  .link-content {
    margin-bottom: 12px;
  }
  
  .link-input {
    width: 100%;
  }
  
  .link-input :deep(.el-input__wrapper) {
    width: 100%;
  }
  
  .link-input :deep(.el-input__append) {
    padding: 0;
  }
  
  .link-input :deep(.el-input__append .el-button) {
    padding: 0 12px;
    font-size: 12px;
    white-space: nowrap;
  }
  
  .link-input :deep(.el-input__append .el-button span) {
    display: inline; /* 平板端显示文字 */
  }
  
  .password-section {
    margin-top: 12px;
    padding-top: 12px;
  }
  
  .expire-info {
    margin-top: 12px;
    padding-top: 12px;
    font-size: 12px;
  }
  
  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    flex-wrap: wrap;
  }
  
  .dialog-footer .el-button {
    min-width: 80px;
  }
}

@media (max-width: 480px) {
  .share-dialog :deep(.el-dialog) {
    width: 100% !important;
    max-width: 100vw !important;
    margin: 0 !important;
    height: 100vh !important;
    border-radius: 0 !important;
    display: flex !important;
    flex-direction: column !important;
    max-height: 100vh !important;
    box-sizing: border-box;
    padding: 0;
  }
  
  .share-dialog :deep(.el-dialog__header) {
    flex-shrink: 0;
    padding: 12px 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
    box-sizing: border-box;
    width: 100%;
    max-width: 100%;
  }
  
  .share-dialog :deep(.el-dialog__body) {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 16px;
    -webkit-overflow-scrolling: touch;
    box-sizing: border-box;
    width: 100%;
    max-width: 100%;
  }
  
  .share-dialog :deep(.el-dialog__footer) {
    flex-shrink: 0;
    padding: 12px 16px;
    border-top: 1px solid var(--el-border-color-lighter);
    background: var(--el-bg-color);
    box-sizing: border-box;
    width: 100%;
    max-width: 100%;
  }
  
  .file-info-card {
    flex-direction: column;
    text-align: center;
    padding: 16px 12px;
    gap: 10px;
    margin-bottom: 20px;
  }
  
  .file-info-card .el-icon {
    margin: 0 auto;
  }
  
  .file-info-content {
    text-align: center;
    width: 100%;
  }
  
  .file-name {
    font-size: 14px;
    white-space: normal;
    word-break: break-all;
    overflow: visible;
    text-overflow: unset;
    line-height: 1.4;
  }
  
  .file-size {
    font-size: 12px;
    margin-top: 4px;
  }
  
  .share-form {
    margin-bottom: 20px;
  }
  
  .share-form :deep(.el-form-item) {
    margin-bottom: 20px;
    display: flex;
    flex-direction: column;
  }
  
  .share-form :deep(.el-form-item__label) {
    width: 100% !important;
    text-align: left;
    margin-bottom: 8px;
    padding: 0;
    line-height: 1.5;
    font-size: 13px;
    font-weight: 600;
  }
  
  .share-form :deep(.el-form-item__content) {
    margin-left: 0 !important;
    width: 100%;
  }
  
  /* 超小屏幕也使用下拉选择框 */
  .expire-select.mobile-only {
    display: block !important;
    width: 100%;
  }
  
  .expire-options.desktop-only {
    display: none !important;
  }
  
  /* 输入框在超小屏幕优化 */
  .share-form :deep(.el-input) {
    width: 100%;
  }
  
  .share-form :deep(.el-input__wrapper) {
    width: 100%;
  }
  
  .share-form :deep(.el-input__append) {
    padding: 0;
    flex-shrink: 0;
    min-width: auto;
  }
  
  .share-form :deep(.el-input__append .el-button) {
    padding: 0 8px;
    font-size: 12px;
    min-width: auto;
  }
  
  .share-form :deep(.el-input__append .el-button span) {
    display: none; /* 超小屏幕隐藏按钮文字 */
  }
  
  /* 确保输入框内部不会超出 */
  .share-form :deep(.el-input__inner) {
    max-width: 100%;
    box-sizing: border-box;
  }
  
  .form-tip {
    font-size: 11px;
    margin-top: 6px;
    line-height: 1.5;
  }
  
  .share-result {
    margin-top: 20px;
  }
  
  .result-alert {
    margin-bottom: 16px;
  }
  
  .result-title {
    font-size: 14px;
  }
  
  .share-link-section {
    padding: 16px 12px;
  }
  
  .link-label {
    font-size: 12px;
    margin-bottom: 8px;
    font-weight: 600;
  }
  
  .link-content {
    margin-bottom: 12px;
  }
  
  .link-input {
    width: 100%;
  }
  
  .link-input :deep(.el-input__wrapper) {
    width: 100%;
  }
  
  .link-input :deep(.el-input__inner) {
    font-size: 12px;
  }
  
  .link-input :deep(.el-input__append) {
    padding: 0;
  }
  
  .link-input :deep(.el-input__append .el-button) {
    padding: 0 10px;
    font-size: 12px;
  }
  
  .link-input :deep(.el-input__append .el-button span) {
    display: none; /* 超小屏幕隐藏按钮文字 */
  }
  
  .password-section {
    margin-top: 12px;
    padding-top: 12px;
  }
  
  .expire-info {
    margin-top: 12px;
    padding-top: 12px;
    font-size: 12px;
    gap: 6px;
  }
  
  .expire-info .el-icon {
    font-size: 14px;
  }
  
  .dialog-footer {
    display: flex;
    flex-direction: column;
    gap: 10px;
    width: 100%;
  }
  
  .dialog-footer .el-button {
    width: 100%;
    margin: 0;
    height: 44px;
    font-size: 15px;
  }
}
</style>

