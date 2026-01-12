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
        <p class="subtitle">{{ t('login.subtitle') }}</p>
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
            <label>{{ t('login.username') }}</label>
            <el-input
              v-model="loginForm.username"
              :placeholder="t('login.usernamePlaceholder')"
              class="custom-input"
            />
          </div>
          <div class="input-group">
            <div class="label-row">
              <label>{{ t('login.password') }}</label>
            </div>
            <el-input
              v-model="loginForm.password"
              type="password"
              :placeholder="t('login.passwordPlaceholder')"
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
            <span v-else>{{ t('login.login') }}</span>
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
            <label>{{ t('login.newUsername') }}</label>
            <el-input
              v-model="registerForm.username"
              :placeholder="t('login.newUsernamePlaceholder')"
              class="custom-input"
            />
          </div>
          <div class="input-group">
            <label>{{ t('login.bindEmail') }}</label>
            <el-input
              v-model="registerForm.email"
              :placeholder="t('login.emailPlaceholder')"
              class="custom-input"
            />
          </div>
          <div class="input-group">
            <label>{{ t('login.setPassword') }}</label>
            <el-input
              v-model="registerForm.password"
              type="password"
              :placeholder="t('login.passwordPlaceholder2')"
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
            <span v-else>{{ t('login.register') }}</span>
          </button>
        </el-form>

        <div class="auth-switch">
          <span class="switch-text">
            {{ activeTab === 'login' ? t('login.noAccount') : t('login.hasAccount') }}
          </span>
          <span 
            v-if="activeTab === 'login' && (allowRegister || isFirstUse)"
            class="switch-link" 
            @click="toggleMode"
          >
            {{ t('login.freeRegister') }}
          </span>
          <span 
            v-else-if="activeTab === 'register'"
            class="switch-link" 
            @click="toggleMode"
          >
            {{ t('login.backToLogin') }}
          </span>
          <span 
            v-else
            class="switch-link disabled"
            @click="toggleMode"
          >
            {{ t('login.registerClosed') }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { login, register, getChallenge, getSysInfo } from '@/api/auth'
import { rsaEncrypt } from '@/utils/crypto'
import { useAuthStore } from '@/stores'
import { useI18n } from '@/composables/useI18n'

const { proxy } = getCurrentInstance() as ComponentInternalInstance
const activeTab = ref('login')
const loading = ref(false)
const isFirstUse = ref(false)
const hasUsers = ref(true)
const allowRegister = ref(true) // 是否允许注册

const loginFormRef = ref<FormInstance>()
const registerFormRef = ref<FormInstance>()

const loginForm = reactive({ username: '', password: '', challenge: '' })
const registerForm = reactive({ username: '', password: '', email: '', challenge: '' })

const { t } = useI18n()

const loginRules: FormRules = {
  username: [{ required: true, message: t('login.usernameRequired'), trigger: 'blur' }],
  password: [{ required: true, message: t('login.passwordRequired'), trigger: 'blur' }]
}

const registerRules: FormRules = {
  username: [{ required: true, message: t('login.usernameRequired'), trigger: 'blur' }],
  email: [{ required: true, message: t('login.emailRequired'), trigger: 'blur' }, { type: 'email', message: t('login.emailFormat'), trigger: 'blur' }],
  password: [{ required: true, message: t('login.passwordRequired'), trigger: 'blur' }, { min: 6, message: t('login.passwordMin'), trigger: 'blur' }]
}

const toggleMode = () => {
  if (isFirstUse.value) return 
  // 如果不允许注册，禁止切换到注册页面
  if (!allowRegister.value && activeTab.value === 'login') {
    proxy?.$modal.msgWarning(t('login.registerDisabled'))
    return
  }
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
          proxy?.$modal.msgError(t('login.connectFailed'))
          return
        }
        const encryptedPassword = rsaEncrypt(challengeRes.data.publicKey, loginForm.password)
        const res = await login({
          username: loginForm.username,
          password: encryptedPassword,
          challenge: challengeRes.data.id
        })
        if (res.data?.token) {
          // 使用 store 管理登录状态
          const authStore = useAuthStore()
          authStore.login(res.data.token, res.data.user_info)
          proxy?.$modal.msgSuccess(t('login.loginSuccess'))
          proxy?.$router.push('/files')
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('login.loginFailed'))
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
        proxy?.$modal.msgSuccess(t('login.registerSuccess'))
        // 切换到登录页面
        activeTab.value = 'login'
        // 自动填充用户名和密码
        loginForm.username = registerForm.username
        loginForm.password = registerForm.password
        // 如果是首次使用，尝试自动登录
        if (isFirstUse.value) {
          // 延迟一下，确保tab切换完成
          setTimeout(async () => {
            await handleLogin()
          }, 100)
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('login.registerFailed'))
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
      allowRegister.value = res.data.allow_register ?? true
      
      if (isFirstUse.value) {
        activeTab.value = 'register'
        // 首次使用，允许注册
        allowRegister.value = true
      } else {
        // 如果不允许注册且当前在注册页面，切换回登录页面
        if (!allowRegister.value && activeTab.value === 'register') {
          activeTab.value = 'login'
        }
      }
    }
  } catch (e) {
    // 如果获取失败，默认允许注册（向后兼容）
    allowRegister.value = true
  }
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
  background: var(--bg-color, #ffffff);
  position: relative;
  overflow: hidden;
}

html.dark .login-page {
  background: var(--bg-color);
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
  background: rgba(37, 99, 235, 0.1);
}

.blob-2 {
  width: 600px;
  height: 600px;
  bottom: -200px;
  right: -150px;
  background: rgba(79, 70, 229, 0.1);
}

html.dark .blob-1 {
  background: rgba(59, 130, 246, 0.15);
}

html.dark .blob-2 {
  background: rgba(99, 102, 241, 0.15);
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
  color: var(--text-primary, #111827);
  letter-spacing: -1.5px;
}

.badge {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--secondary-color) 100%);
  color: var(--text-primary-inverse, #fff);
  font-size: 10px;
  font-weight: 800;
  padding: 4px 8px;
  border-radius: 6px;
  letter-spacing: 0.5px;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

html.dark .badge {
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.subtitle {
  color: var(--text-secondary, #6b7280);
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
  color: var(--text-regular, #374151);
  margin-bottom: 8px;
}

:deep(.custom-input .el-input__wrapper) {
  background: var(--el-fill-color-lighter, #f9fafb) !important;
  box-shadow: none !important;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 12px;
  padding: 8px 16px;
  height: 50px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

:deep(.custom-input .el-input__wrapper:hover) {
  background: var(--el-fill-color-light, #f3f4f6) !important;
  border-color: var(--border-color, #d1d5db);
}

:deep(.custom-input .el-input__wrapper.is-focus) {
  background: var(--card-bg, #fff) !important;
  border-color: var(--primary-color) !important;
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.1) !important;
}

html.dark :deep(.custom-input .el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.2) !important;
}

:deep(.custom-input .el-input__inner) {
  color: var(--text-primary, #111827);
  font-weight: 500;
}

/* Submit Button (Electric Blue Gradient) */
.submit-btn {
  width: 100%;
  height: 54px;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--secondary-color) 100%);
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

html.dark .submit-btn {
  box-shadow: 0 10px 20px -5px rgba(59, 130, 246, 0.5);
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
  color: var(--text-placeholder, #9ca3af);
}

.switch-link {
  color: var(--primary-color);
  font-weight: 600;
  cursor: pointer;
  margin-left: 4px;
  transition: color 0.2s;
}

.switch-link:hover {
  color: var(--primary-hover);
  text-decoration: underline;
}

.switch-link.disabled {
  color: var(--text-placeholder, #9ca3af);
  cursor: not-allowed;
  text-decoration: none;
}

.switch-link.disabled:hover {
  color: var(--text-placeholder, #9ca3af);
  text-decoration: none;
}

/* 移动端响应式 */
@media (max-width: 1024px) {
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
