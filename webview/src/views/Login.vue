<template>
  <div class="login-container">
    <div class="login-background">
      <div class="bg-shape shape-1"></div>
      <div class="bg-shape shape-2"></div>
      <div class="bg-shape shape-3"></div>
    </div>
    
    <el-card class="login-card" shadow="always">
      <div class="logo-section">
        <el-icon :size="56" color="#409EFF"><Folder /></el-icon>
        <h1>MyObj 网盘</h1>
        <p>安全、高效的私有云存储</p>
      </div>
      
      <el-tabs v-model="activeTab" class="login-tabs" :before-leave="handleTabChange">
        <el-tab-pane label="登录" name="login" :disabled="isFirstUse">
          <el-form
            ref="loginFormRef"
            :model="loginForm"
            :rules="loginRules"
            @submit.prevent="handleLogin"
            class="login-form"
          >
            <el-form-item prop="username">
              <el-input
                v-model="loginForm.username"
                placeholder="请输入用户名"
                :prefix-icon="User"
                size="large"
                clearable
              />
            </el-form-item>
            
            <el-form-item prop="password">
              <el-input
                v-model="loginForm.password"
                type="password"
                placeholder="请输入密码"
                :prefix-icon="Lock"
                size="large"
                show-password
                @keyup.enter="handleLogin"
              />
            </el-form-item>
            
            <el-form-item>
              <el-button
                type="primary"
                size="large"
                :loading="loading"
                @click="handleLogin"
                style="width: 100%"
              >
                登录
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        
        <el-tab-pane label="注册" name="register" :disabled="!isFirstUse && hasUsers">
          <el-alert
            v-if="isFirstUse"
            title="欢迎使用 MyObj 网盘！"
            type="success"
            description="请创建第一个管理员账号"
            :closable="false"
            style="margin-bottom: 20px"
          />
          <el-form
            ref="registerFormRef"
            :model="registerForm"
            :rules="registerRules"
            @submit.prevent="handleRegister"
            class="login-form"
          >
            <el-form-item prop="username">
              <el-input
                v-model="registerForm.username"
                placeholder="请输入用户名"
                :prefix-icon="User"
                size="large"
              />
            </el-form-item>
            
            <el-form-item prop="email">
              <el-input
                v-model="registerForm.email"
                placeholder="请输入邮箱"
                :prefix-icon="Message"
                size="large"
              />
            </el-form-item>
            
            <el-form-item prop="password">
              <el-input
                v-model="registerForm.password"
                type="password"
                placeholder="请输入密码"
                :prefix-icon="Lock"
                size="large"
                show-password
              />
            </el-form-item>
            
            <el-form-item>
              <el-button
                type="primary"
                size="large"
                :loading="loading"
                @click="handleRegister"
                style="width: 100%"
              >
                注册
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
      
      <el-divider />
      
      <div class="features">
        <div class="feature-grid">
          <div class="feature-item">
            <el-icon :size="20" color="#409EFF"><Upload /></el-icon>
            <span>大文件分片上传、断点续传、秒传</span>
          </div>
          <div class="feature-item">
            <el-icon :size="20" color="#67C23A"><Lock /></el-icon>
            <span>文件加密存储、权限管理</span>
          </div>
          <div class="feature-item">
            <el-icon :size="20" color="#E6A23C"><Share /></el-icon>
            <span>文件分享、限时链接</span>
          </div>
          <div class="feature-item">
            <el-icon :size="20" color="#F56C6C"><Download /></el-icon>
            <span>离线下载、种子下载</span>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { User, Lock, Message, Folder, Upload, Share, Download } from '@element-plus/icons-vue'
import { login, register, getChallenge, getSysInfo } from '@/api/auth'
import { rsaEncrypt } from '@/utils/crypto'
import type { LoginRequest, RegisterRequest } from '@/types'

const router = useRouter()
const activeTab = ref('login')
const loading = ref(false)
const isFirstUse = ref(false)
const hasUsers = ref(true)

const loginFormRef = ref<FormInstance>()
const registerFormRef = ref<FormInstance>()

const loginForm = reactive({
  username: '',
  password: '',
  challenge: ''
})

const registerForm = reactive({
  username: '',
  password: '',
  email: '',
  challenge: ''
})

const loginRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ]
}

const registerRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, message: '用户名长度不能少于3位', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        // 获取密码挑战秘钥
        const challengeRes = await getChallenge()
        if (!challengeRes.data || !challengeRes.data.publicKey || !challengeRes.data.id) {
          ElMessage.error('获取密码挑战失败')
          return
        }
        
        // 使用公钥加密密码
        const encryptedPassword = rsaEncrypt(challengeRes.data.publicKey, loginForm.password)
        
        // 发送登录请求
        const loginData: LoginRequest = {
          username: loginForm.username,
          password: encryptedPassword,
          challenge: challengeRes.data.id
        }
        
        const res = await login(loginData)
        if (res.data && res.data.token) {
          localStorage.setItem('token', res.data.token)
          localStorage.setItem('userInfo', JSON.stringify(res.data.user_info))
          localStorage.setItem('username', loginForm.username)
          ElMessage.success(res.message || '登录成功')
          router.push('/files')
        } else {
          ElMessage.error('登录失败：未获取到token')
        }
      } catch (error: any) {
        ElMessage.error(error.message || '登录失败')
      } finally {
        loading.value = false
      }
    }
  })
}

const handleRegister = async () => {
  if (!registerFormRef.value) return
  
  await registerFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        // 获取密码挑战秘钥
        const challengeRes = await getChallenge()
        if (!challengeRes.data || !challengeRes.data.publicKey || !challengeRes.data.id) {
          ElMessage.error('获取密码挑战失败')
          return
        }
        
        // 使用公钥加密密码
        const encryptedPassword = rsaEncrypt(challengeRes.data.publicKey, registerForm.password)
        
        // 发送注册请求
        const registerData: RegisterRequest = {
          username: registerForm.username,
          password: encryptedPassword,
          email: registerForm.email,
          challenge: challengeRes.data.id
        }
        
        await register(registerData)
        ElMessage.success('注册成功，请登录')
        
        // 如果是首次使用，注册后直接自动登录
        if (isFirstUse.value) {
          loginForm.username = registerForm.username
          loginForm.password = registerForm.password
          await handleLogin()
        } else {
          activeTab.value = 'login'
          loginForm.username = registerForm.username
          registerForm.username = ''
          registerForm.email = ''
          registerForm.password = ''
        }
      } catch (error: any) {
        ElMessage.error(error.message || '注册失败')
      } finally {
        loading.value = false
      }
    }
  })
}

// 检查系统是否首次使用
const checkSysInfo = async () => {
  try {
    const res = await getSysInfo()
    if (res.code === 200 && res.data) {
      isFirstUse.value = res.data.is_first_use
      hasUsers.value = !res.data.is_first_use
      
      // 如果是首次使用，直接切换到注册页面
      if (isFirstUse.value) {
        activeTab.value = 'register'
      }
    }
  } catch (error) {
    console.error('获取系统信息失败:', error)
    // 失败时默认显示登录页面
  }
}

// 标签切换前的验证
const handleTabChange = (activeName: string, oldActiveName: string) => {
  // 如果是首次使用，不允许切换到登录
  if (isFirstUse.value && activeName === 'login') {
    return false
  }
  return true
}

// 页面加载时检查系统状态
onMounted(() => {
  checkSysInfo()
})
</script>

<style scoped>
.login-container {
  width: 100%;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.login-background {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  z-index: 0;
}

.bg-shape {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  animation: float 20s infinite ease-in-out;
}

.shape-1 {
  width: 400px;
  height: 400px;
  top: -100px;
  left: -100px;
  animation-delay: 0s;
}

.shape-2 {
  width: 300px;
  height: 300px;
  bottom: -50px;
  right: -50px;
  animation-delay: 5s;
}

.shape-3 {
  width: 200px;
  height: 200px;
  top: 50%;
  right: 10%;
  animation-delay: 10s;
}

@keyframes float {
  0%, 100% {
    transform: translate(0, 0) scale(1);
  }
  33% {
    transform: translate(30px, -30px) scale(1.1);
  }
  66% {
    transform: translate(-20px, 20px) scale(0.9);
  }
}

.login-card {
  width: 480px;
  padding: 40px;
  position: relative;
  z-index: 1;
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
}

.logo-section {
  text-align: center;
  margin-bottom: 32px;
}

.logo-section h1 {
  font-size: 32px;
  margin: 20px 0 8px;
  color: var(--el-text-color-primary);
  font-weight: 700;
}

.logo-section p {
  font-size: 15px;
  color: var(--el-text-color-secondary);
}

.login-tabs {
  margin-bottom: 24px;
}

.login-form {
  margin-top: 8px;
}

.features {
  margin-top: 20px;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  transition: all 0.3s;
}

.feature-item:hover {
  background: var(--el-fill-color);
  transform: translateY(-2px);
}

.feature-item span {
  flex: 1;
  line-height: 1.4;
}
</style>
