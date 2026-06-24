<template>
  <div class="cloud-accounts">
    <h3>云盘账号管理</h3>
    <p class="desc">绑定云盘账号后，可直接解析分享链接并下载文件到本地</p>

    <div class="account-grid">
      <div
        v-for="item in providers"
        :key="item.key"
        class="account-card"
        :class="{ connected: getStatus(item.key).connected }"
      >
        <div class="card-header">
          <span class="icon">{{ item.icon }}</span>
          <span class="name">{{ item.name }}</span>
          <el-tag v-if="getStatus(item.key).connected" type="success" size="small">已连接</el-tag>
          <el-tag v-else type="info" size="small">未连接</el-tag>
        </div>

        <div class="card-body">
          <template v-if="getStatus(item.key).connected">
            <div class="account-info">
              <span>{{ getStatus(item.key).account_name || '已连接' }}</span>
            </div>
            <el-button type="danger" size="small" @click="handleDisconnect(item.key)">断开</el-button>
          </template>
          <template v-else>
            <!-- OAuth 登录 -->
            <el-button v-if="item.authType === 'oauth'" type="primary" size="small" @click="handleOAuth(item.key)">
              登录
            </el-button>
            <!-- 用户名密码登录 -->
            <el-button v-else-if="item.authType === 'login'" type="primary" size="small" @click="showLoginDialog(item.key)">
              登录
            </el-button>
            <!-- 直接可用 -->
            <el-tag v-else type="success" size="small">直接可用</el-tag>
          </template>
        </div>
      </div>
    </div>

    <!-- 登录对话框 -->
    <el-dialog v-model="loginVisible" :title="`登录${currentProvider?.name}`" width="400px">
      <el-form :model="loginForm" label-width="80px">
        <el-form-item label="账号">
          <el-input v-model="loginForm.username" placeholder="请输入账号" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="loginForm.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="loginVisible = false">取消</el-button>
        <el-button type="primary" :loading="logging" @click="handleLogin">登录</el-button>
      </template>
    </el-dialog>

  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  getCloudAccountStatus,
  deleteCloudAccount,
  getAliyunAuthUrl,
  getBaiduAuthUrl,
  getXunleiAuthUrl,
  getPikPakAuthUrl,
  xunleiLogin,
  tianyiLogin,
} from '@myobj/api/cloud-account'

const providers = [
  { key: 'aliyun', name: '阿里云盘', icon: '🅰️', authType: 'oauth' },
  { key: 'baidu', name: '百度网盘', icon: '🅱️', authType: 'oauth' },
  { key: 'xunlei', name: '迅雷网盘', icon: '⚡', authType: 'oauth' },
  { key: 'pikpak', name: 'PikPak', icon: '📦', authType: 'oauth' },
  { key: 'tianyi', name: '天翼云盘', icon: '☁️', authType: 'login' },
]

const accountStatus = ref<Record<string, any>>({})
const loginVisible = ref(false)
const logging = ref(false)
const currentProvider = ref<any>(null)
const loginForm = reactive({ username: '', password: '' })

const getStatus = (key: string) => accountStatus.value[key] || { connected: false, status: 'disconnected' }

const loadStatus = async () => {
  try {
    const res = await getCloudAccountStatus()
    accountStatus.value = res.data || res
  } catch (e) {
    console.error(e)
  }
}

const handleOAuth = async (provider: string) => {
  try {
    let res
    switch (provider) {
      case 'aliyun': res = await getAliyunAuthUrl(); break
      case 'baidu': res = await getBaiduAuthUrl(); break
      case 'xunlei': res = await getXunleiAuthUrl(); break
      case 'pikpak': res = await getPikPakAuthUrl(); break
      default: return
    }
    const data = res.data || res
    if (data.auth_url) {
      window.open(data.auth_url, `${provider}_auth`, 'width=600,height=700')
    }
  } catch (e: any) {
    // 捕获401错误，不跳转登录页
    if (e.message?.includes('登录已过期') || e.message?.includes('未授权')) {
      ElMessage.error('请先登录 MyObj 系统')
    } else {
      ElMessage.error(e.message || '获取授权链接失败')
    }
  }
}

const showLoginDialog = (provider: string) => {
  currentProvider.value = providers.find(p => p.key === provider)
  loginForm.username = ''
  loginForm.password = ''
  loginVisible.value = true
}

const handleLogin = async () => {
  logging.value = true
  try {
    const key = currentProvider.value?.key
    if (key === 'xunlei') {
      await xunleiLogin(loginForm.username, loginForm.password)
    } else if (key === 'tianyi') {
      await tianyiLogin(loginForm.username, loginForm.password)
    }
    ElMessage.success('登录成功')
    loginVisible.value = false
    loadStatus()
  } catch (e: any) {
    ElMessage.error(e.message || '登录失败')
  } finally {
    logging.value = false
  }
}

const handleDisconnect = async (provider: string) => {
  try {
    await deleteCloudAccount(provider)
    ElMessage.success('已断开')
    loadStatus()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

onMounted(() => {
  loadStatus()
  // 监听 OAuth 弹窗回调
  window.addEventListener('message', (event: MessageEvent) => {
    if (event.data?.type?.endsWith('_auth_success')) {
      loadStatus()
    }
  })
})
</script>

<style scoped>
.cloud-accounts { padding: 20px; }
.desc { color: var(--el-text-color-secondary); margin-bottom: 20px; }
.account-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; }
.account-card {
  border: 1px solid var(--el-border-color);
  border-radius: 12px;
  padding: 16px;
  transition: all 0.2s;
}
.account-card.connected { border-color: var(--el-color-success); background: var(--el-color-success-light-9); }
.card-header { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
.icon { font-size: 24px; }
.name { font-weight: 600; flex: 1; }
.card-body { display: flex; align-items: center; justify-content: space-between; }
.account-info { flex: 1; font-size: 13px; color: var(--el-text-color-secondary); }
</style>
