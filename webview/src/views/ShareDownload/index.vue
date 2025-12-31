<template>
  <div class="share-download-page">
    <div class="share-container glass-panel">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-container">
        <el-icon class="is-loading" :size="48"><Loading /></el-icon>
        <p>加载中...</p>
      </div>

      <!-- 错误状态（致命错误，如分享不存在、已过期等） -->
      <div v-else-if="error && !needsPassword" class="error-container">
        <el-icon :size="64" color="#f56c6c"><WarningFilled /></el-icon>
        <h2>分享不存在或已失效</h2>
        <p>{{ error }}</p>
        <el-button type="primary" @click="goHome">返回首页</el-button>
      </div>

      <!-- 密码输入界面 -->
      <div v-else-if="needsPassword" class="password-container">
        <div class="share-header">
          <el-icon :size="48" color="var(--primary-color)"><Lock /></el-icon>
          <h1>分享需要密码</h1>
        </div>
        
        <div class="password-form-wrapper">
          <el-form @submit.prevent="handlePasswordSubmit" :model="{ password }">
            <el-form-item 
              label="分享密码"
              :error="passwordError"
              :validate-status="passwordError ? 'error' : ''"
            >
              <el-input
                v-model="password"
                type="password"
                placeholder="请输入分享密码"
                show-password
                size="large"
                @keyup.enter="handlePasswordSubmit"
                @input="passwordError = ''"
                autofocus
              />
            </el-form-item>
            <el-button
              type="primary"
              size="large"
              :loading="verifying"
              @click="handlePasswordSubmit"
              style="width: 100%"
            >
              {{ verifying ? '验证中...' : '验证密码' }}
            </el-button>
          </el-form>
        </div>
      </div>

      <!-- 分享文件信息 -->
      <div v-else-if="shareInfo && shareInfo.file_name" class="share-info-container">
        <div class="share-header">
          <el-icon :size="48" color="var(--primary-color)"><Share /></el-icon>
          <h1>文件分享</h1>
        </div>

        <div class="file-info-card">
          <div class="file-icon-wrapper">
            <file-icon
              :mime-type="shareInfo.mime_type"
              :file-name="shareInfo.file_name"
              :icon-size="64"
            />
          </div>
          
          <div class="file-details">
            <h2 class="file-name">{{ shareInfo.file_name }}</h2>
            <div class="file-meta">
              <div class="meta-item">
                <el-icon><Document /></el-icon>
                <span>文件大小：{{ formatSize(shareInfo.file_size) }}</span>
              </div>
              <div class="meta-item">
                <el-icon><Clock /></el-icon>
                <span>过期时间：{{ shareInfo.expires_at || '永久有效' }}</span>
              </div>
              <div class="meta-item" v-if="shareInfo.download_count > 0">
                <el-icon><Download /></el-icon>
                <span>已下载：{{ shareInfo.download_count }} 次</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 过期提示 -->
        <el-alert
          v-if="shareInfo.is_expired"
          title="分享已过期"
          type="warning"
          :closable="false"
          class="expire-alert"
        />

        <!-- 下载按钮 -->
        <div class="action-section">
          <el-button
            type="primary"
            size="large"
            :loading="downloading"
            @click="handleDownload"
            style="width: 100%"
            :icon="Download"
          >
            {{ downloading ? '准备下载中...' : '下载文件' }}
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getCurrentInstance, type ComponentInternalInstance } from 'vue'
import { getShareInfo, getShareDownloadUrl, type ShareInfoResponse } from '@/api/share'
import { formatSize } from '@/utils'
import FileIcon from '@/components/FileIcon/index.vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const { proxy } = getCurrentInstance() as ComponentInternalInstance

const loading = ref(true)
const error = ref('')
const passwordError = ref('') // 密码错误提示（不影响界面切换）
const shareInfo = ref<ShareInfoResponse | null>(null)
const password = ref('')
const downloading = ref(false)
const verifying = ref(false)
const needsPassword = ref(false)

// 从路由参数获取token
const token = computed(() => {
  return route.params.token as string
})

// 加载分享信息（不传密码，只检查是否需要密码）
const loadShareInfo = async () => {
  if (!token.value) {
    error.value = '分享链接无效'
    loading.value = false
    return
  }

  try {
    loading.value = true
    error.value = ''
    const res = await getShareInfo(token.value)
    
    if (res.code === 200 && res.data) {
      // 如果只返回了 has_password，说明需要密码
      if (res.data.has_password && !res.data.file_name) {
        needsPassword.value = true
        shareInfo.value = res.data
      } else {
        // 没有密码或密码已验证，显示文件信息
        needsPassword.value = false
        shareInfo.value = res.data
        // 如果已过期，直接显示错误
        if (res.data.is_expired) {
          error.value = '分享已过期'
        }
      }
    } else {
      error.value = res.message || '获取分享信息失败'
    }
  } catch (err: any) {
    error.value = err.message || '获取分享信息失败'
  } finally {
    loading.value = false
  }
}

// 验证密码并获取文件信息
const handlePasswordSubmit = async () => {
  if (!password.value) {
    return
  }

  try {
    verifying.value = true
    passwordError.value = '' // 清空之前的错误提示
    error.value = ''
    
    const res = await getShareInfo(token.value, password.value)
    
    if (res.code === 200 && res.data) {
      // 密码正确，显示文件信息
      needsPassword.value = false
      shareInfo.value = res.data
      passwordError.value = '' // 清空错误提示
      if (res.data.is_expired) {
        error.value = '分享已过期'
        needsPassword.value = false // 已过期，不再需要密码
      }
    } else {
      // 密码错误，保持在密码输入界面，只显示错误提示
      passwordError.value = res.message || '密码错误'
      password.value = ''
    }
  } catch (err: any) {
    // 密码验证失败，保持在密码输入界面
    passwordError.value = err.message || '密码验证失败'
    password.value = ''
  } finally {
    verifying.value = false
  }
}


// 下载文件（直接使用GET请求触发浏览器下载）
const handleDownload = () => {
  if (!shareInfo.value || !shareInfo.value.file_name) return

  // 获取下载URL（GET请求，直接触发浏览器下载）
  const downloadUrl = getShareDownloadUrl(
    token.value,
    shareInfo.value.has_password ? password.value : undefined
  )
  
  // 直接创建链接触发下载，浏览器会自动处理
  const link = document.createElement('a')
  link.href = downloadUrl
  link.download = shareInfo.value.file_name
  link.style.display = 'none'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  // 显示成功提示
  proxy?.$modal.msgSuccess('开始下载')
  
  // 延迟刷新分享信息（更新下载次数），避免影响下载
  setTimeout(async () => {
    try {
      const currentPassword = shareInfo.value?.has_password ? password.value : undefined
      const res = await getShareInfo(token.value, currentPassword)
      if (res.code === 200 && res.data && res.data.file_name && shareInfo.value) {
        // 只更新下载次数，保持其他状态不变
        shareInfo.value.download_count = res.data.download_count
      }
    } catch (err) {
      // 忽略刷新信息的错误，不影响用户体验
      console.warn('刷新分享信息失败', err)
    }
  }, 1000)
}

// 返回首页
const goHome = () => {
  router.push('/')
}

// 初始化加载
onMounted(() => {
  loadShareInfo()
})
</script>

<style scoped>
.share-download-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.share-container {
  width: 100%;
  max-width: 600px;
  padding: 40px;
  border-radius: 20px;
}

.loading-container,
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 60px 20px;
  text-align: center;
}

.error-container h2 {
  margin: 0;
  color: var(--text-primary);
}

.error-container p {
  color: var(--text-secondary);
  margin: 8px 0 24px;
}

.share-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 32px;
  justify-content: center;
}

.share-header h1 {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  color: var(--text-primary);
}

.file-info-card {
  display: flex;
  gap: 24px;
  padding: 24px;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 16px;
  margin-bottom: 24px;
}

.file-icon-wrapper {
  flex-shrink: 0;
}

.file-details {
  flex: 1;
  min-width: 0;
}

.file-name {
  margin: 0 0 16px 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  word-break: break-all;
}

.file-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-secondary);
  font-size: 14px;
}

.expire-alert {
  margin-bottom: 24px;
}

.action-section {
  margin-top: 24px;
}

.password-container {
  text-align: center;
}

.password-form-wrapper {
  margin-top: 32px;
  max-width: 400px;
  margin-left: auto;
  margin-right: auto;
}

.password-form {
  margin-bottom: 16px;
}

@media (max-width: 768px) {
  .share-container {
    padding: 24px;
  }

  .file-info-card {
    flex-direction: column;
    text-align: center;
  }

  .file-meta {
    align-items: center;
  }
}
</style>

