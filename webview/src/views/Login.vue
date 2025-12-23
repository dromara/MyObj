<template>
  <div class="login-page">
    <!-- Abstract Background -->
    <div class="background-blobs">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
    </div>

    <div class="login-content">
      <!-- Header -->
      <div class="brand-header">
        <div class="logo-row">
          <span class="logo-text">MyObj</span>
          <div class="badge">PRO</div>
        </div>
        <p class="subtitle">下一代智能云存储平台</p>
      </div>

      <!-- Auth Form -->
      <div class="auth-card">
        <!-- Login Form -->
        <el-form
          v-if="activeTab === 'login'"
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          @submit.prevent="handleLogin"
          class="auth-form"
          hide-required-asterisk
        >
          <div class="input-group">
            <label>账号</label>
            <el-input
              v-model="loginForm.username"
              placeholder="用户名 / 电子邮箱"
              class="custom-input"
            />
          </div>
          <div class="input-group">
            <div class="label-row">
              <label>密码</label>
            </div>
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="请输入密码"
              show-password
              class="custom-input"
            />
          </div>

          <button 
            type="submit"
            class="submit-btn" 
            :class="{ 'is-loading': loading }"
          >
            <el-icon v-if="loading" class="loading-icon"><Loading /></el-icon>
            <span v-else>立即登录</span>
          </button>
        </el-form>

        <!-- Register Form -->
        <el-form
          v-else
          ref="registerFormRef"
          :model="registerForm"
          :rules="registerRules"
          @submit.prevent="handleRegister"
          class="auth-form"
          hide-required-asterisk
        >
          <div class="input-group">
            <label>新用户名</label>
            <el-input
              v-model="registerForm.username"
              placeholder="设置用户名"
              class="custom-input"
            />
          </div>
          <div class="input-group">
            <label>绑定邮箱</label>
            <el-input
              v-model="registerForm.email"
              placeholder="name@example.com"
              class="custom-input"
            />
          </div>
          <div class="input-group">
            <label>设置密码</label>
            <el-input
              v-model="registerForm.password"
              type="password"
              placeholder="至少 6 位字符"
              show-password
              class="custom-input"
            />
          </div>

          <button 
            type="submit"
            class="submit-btn" 
            :class="{ 'is-loading': loading }"
          >
            <el-icon v-if="loading" class="loading-icon"><Loading /></el-icon>
            <span v-else>创建新账号</span>
          </button>
        </el-form>

        <div class="auth-switch">
          <span class="switch-text">
            {{ activeTab === 'login' ? '还没有账号？' : '已有账号？' }}
          </span>
          <span class="switch-link" @click="toggleMode">
            {{ activeTab === 'login' ? '免费注册' : '返回登录' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { login, register, getChallenge, getSysInfo } from '@/api/auth'
import { rsaEncrypt } from '@/utils/crypto'

const { proxy } = getCurrentInstance() as ComponentInternalInstance
const router = useRouter()
const activeTab = ref('login')
const loading = ref(false)
const isFirstUse = ref(false)
const hasUsers = ref(true)

const loginFormRef = ref<FormInstance>()
const registerFormRef = ref<FormInstance>()

const loginForm = reactive({ username: '', password: '', challenge: '' })
const registerForm = reactive({ username: '', password: '', email: '', challenge: '' })

const loginRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const registerRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [{ required: true, message: '请输入邮箱', trigger: 'blur' }, { type: 'email', message: '格式不正确', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '至少6位', trigger: 'blur' }]
}

const toggleMode = () => {
  if (isFirstUse.value) return 
  activeTab.value = activeTab.value === 'login' ? 'register' : 'login'
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  await loginFormRef.value.validate(async (valid: boolean) => {
    if (valid) {
      loading.value = true
      try {
        const challengeRes = await getChallenge()
        if (!challengeRes.data?.publicKey || !challengeRes.data.id) {
          proxy?.$modal.msgError('连接失败')
          return
        }
        const encryptedPassword = rsaEncrypt(challengeRes.data.publicKey, loginForm.password)
        const res = await login({
          username: loginForm.username,
          password: encryptedPassword,
          challenge: challengeRes.data.id
        })
        if (res.data?.token) {
          proxy?.$cache.local.set('token', res.data.token)
          proxy?.$cache.local.setJSON('userInfo', res.data.user_info)
          proxy?.$cache.local.set('username', loginForm.username)
          proxy?.$modal.msgSuccess('登录成功')
          router.push('/files')
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || '登录失败')
      } finally {
        loading.value = false
      }
    }
  })
}

const handleRegister = async () => {
  if (!registerFormRef.value) return
  await registerFormRef.value.validate(async (valid: boolean) => {
    if (valid) {
      loading.value = true
      try {
        const challengeRes = await getChallenge()
        if (!challengeRes.data?.publicKey || !challengeRes.data.id) return
        const encryptedPassword = rsaEncrypt(challengeRes.data.publicKey, registerForm.password)
        await register({
          username: registerForm.username,
          password: encryptedPassword,
          email: registerForm.email,
          challenge: challengeRes.data.id
        })
        proxy?.$modal.msgSuccess('注册成功')
        if (isFirstUse.value) {
          loginForm.username = registerForm.username
          loginForm.password = registerForm.password
          await handleLogin()
        } else {
          activeTab.value = 'login'
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || '注册失败')
      } finally {
        loading.value = false
      }
    }
  })
}

const checkSysInfo = async () => {
  try {
    const res = await getSysInfo()
    if (res.code === 200 && res.data) {
      isFirstUse.value = res.data.is_first_use
      hasUsers.value = !res.data.is_first_use
      if (isFirstUse.value) activeTab.value = 'register'
    }
  } catch (e) {}
}

onMounted(() => checkSysInfo())
</script>

<style scoped>
.login-page {
  width: 100%;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ffffff;
  position: relative;
  overflow: hidden;
}

/* --- Blue/Indigo Background Blobs --- */
.background-blobs {
  position: absolute;
  width: 100%;
  height: 100%;
  z-index: 0;
  overflow: hidden;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(90px);
  opacity: 0.4;
}

.blob-1 {
  width: 700px;
  height: 700px;
  top: -250px;
  left: -200px;
  background: #dbeafe; /* Blue-100 */
}

.blob-2 {
  width: 600px;
  height: 600px;
  bottom: -200px;
  right: -150px;
  background: #ede9fe; /* Violet-100 */
}

/* --- Content --- */
.login-content {
  position: relative;
  z-index: 10;
  width: 100%;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.brand-header {
  text-align: center;
  margin-bottom: 40px;
}

.logo-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-bottom: 8px;
}

.logo-text {
  font-size: 36px;
  font-weight: 800;
  color: #111827; /* Gray-900 */
  letter-spacing: -1.5px;
}

.badge {
  background: linear-gradient(135deg, #3b82f6 0%, #4f46e5 100%);
  color: #fff;
  font-size: 10px;
  font-weight: 800;
  padding: 4px 8px;
  border-radius: 6px;
  letter-spacing: 0.5px;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.subtitle {
  color: #6b7280;
  font-size: 14px;
  font-weight: 500;
}

/* --- Auth Card --- */
.auth-card {
  width: 100%;
  padding: 0 24px;
}

.input-group {
  margin-bottom: 20px;
}

.input-group label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
}

:deep(.custom-input .el-input__wrapper) {
  background: #f9fafb !important;
  box-shadow: none !important;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 8px 16px;
  height: 50px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

:deep(.custom-input .el-input__wrapper:hover) {
  background: #f3f4f6 !important;
  border-color: #d1d5db;
}

:deep(.custom-input .el-input__wrapper.is-focus) {
  background: #fff !important;
  border-color: #3b82f6 !important; /* Blue-500 */
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.1) !important;
}

:deep(.custom-input .el-input__inner) {
  color: #111827;
  font-weight: 500;
}

/* Submit Button (Electric Blue Gradient) */
.submit-btn {
  width: 100%;
  height: 54px;
  background: linear-gradient(135deg, #2563eb 0%, #4f46e5 100%); /* Blue to Indigo */
  color: white;
  border: none;
  border-radius: 14px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 12px;
  box-shadow: 0 10px 20px -5px rgba(37, 99, 235, 0.4);
}

.submit-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 15px 25px -5px rgba(37, 99, 235, 0.5);
  filter: brightness(1.05);
}

.submit-btn:active {
  transform: translateY(0);
}

.loading-icon {
  animation: rotate 1s linear infinite;
  margin-right: 8px;
}

@keyframes rotate { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

.auth-switch {
  margin-top: 32px;
  text-align: center;
  font-size: 14px;
}

.switch-text {
  color: #9ca3af;
}

.switch-link {
  color: #2563eb;
  font-weight: 600;
  cursor: pointer;
  margin-left: 4px;
  transition: color 0.2s;
}

.switch-link:hover {
  color: #1d4ed8;
  text-decoration: underline;
}

/* 移动端响应式 */
@media (max-width: 768px) {
  .login-content {
    max-width: 100%;
    padding: 0 20px;
  }
  
  .brand-header {
    margin-bottom: 32px;
  }
  
  .logo-text {
    font-size: 28px;
  }
  
  .badge {
    font-size: 9px;
    padding: 3px 6px;
  }
  
  .subtitle {
    font-size: 13px;
  }
  
  .auth-card {
    padding: 0 16px;
  }
  
  .input-group {
    margin-bottom: 16px;
  }
  
  .input-group label {
    font-size: 12px;
    margin-bottom: 6px;
  }
  
  :deep(.custom-input .el-input__wrapper) {
    height: 46px;
    padding: 6px 14px;
  }
  
  .submit-btn {
    height: 50px;
    font-size: 15px;
    margin-top: 10px;
  }
  
  .auth-switch {
    margin-top: 24px;
    font-size: 13px;
  }
  
  .blob-1 {
    width: 500px;
    height: 500px;
  }
  
  .blob-2 {
    width: 400px;
    height: 400px;
  }
}

@media (max-width: 480px) {
  .login-content {
    padding: 0 16px;
  }
  
  .brand-header {
    margin-bottom: 24px;
  }
  
  .logo-text {
    font-size: 24px;
  }
  
  .subtitle {
    font-size: 12px;
  }
  
  .auth-card {
    padding: 0 12px;
  }
  
  .input-group label {
    font-size: 11px;
  }
  
  :deep(.custom-input .el-input__wrapper) {
    height: 44px;
    padding: 6px 12px;
  }
  
  .submit-btn {
    height: 48px;
    font-size: 14px;
  }
  
  .auth-switch {
    font-size: 12px;
  }
}
</style>
