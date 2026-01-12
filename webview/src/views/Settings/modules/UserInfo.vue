<template>
  <div class="user-info-form">
    <el-form
      ref="formRef"
      :model="userForm"
      :rules="rules"
      label-width="120px"
      label-position="left"
    >
      <el-form-item :label="t('settings.userInfo.username')" prop="username">
        <el-input
          v-model="userForm.username"
          :placeholder="t('settings.userInfo.usernamePlaceholder')"
          maxlength="50"
          show-word-limit
          clearable
        />
      </el-form-item>
      
      <el-form-item :label="t('settings.userInfo.nickname')" prop="nickname">
        <el-input
          v-model="userForm.nickname"
          :placeholder="t('settings.userInfo.nicknamePlaceholder')"
          maxlength="50"
          show-word-limit
          clearable
        />
      </el-form-item>
      
      <el-form-item :label="t('settings.userInfo.email')" prop="email">
        <el-input
          v-model="userForm.email"
          type="email"
          :placeholder="t('settings.userInfo.emailPlaceholder')"
          maxlength="100"
          clearable
        />
      </el-form-item>
      
      <el-form-item :label="t('settings.userInfo.phone')" prop="phone">
        <el-input
          v-model="userForm.phone"
          :placeholder="t('settings.userInfo.phonePlaceholder')"
          maxlength="20"
          clearable
        />
      </el-form-item>
      
      <el-form-item>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ t('settings.userInfo.save') }}
        </el-button>
        <el-button @click="handleReset">{{ t('settings.userInfo.reset') }}</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/user'
import { updateUser } from '@/api/user'
import { useI18n } from '@/composables/useI18n'

const { proxy } = getCurrentInstance() as ComponentInternalInstance
const userStore = useUserStore()
const { t } = useI18n()

const formRef = ref<FormInstance>()
const saving = ref(false)

const userForm = reactive({
  username: '',
  nickname: '',
  email: '',
  phone: ''
})

const rules: FormRules = {
  username: [
    { required: true, message: t('settings.userInfo.usernamePlaceholder'), trigger: 'blur' },
    { min: 3, max: 50, message: t('settings.userInfo.usernameLength'), trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: t('settings.userInfo.emailFormat'), trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: t('settings.userInfo.phoneFormat'), trigger: 'blur' }
  ]
}

// 加载用户信息
const loadUserInfo = () => {
  if (userStore.userInfo) {
    userForm.username = userStore.username
    userForm.nickname = userStore.nickname
    userForm.email = userStore.email
    userForm.phone = userStore.phone
  }
}

// 监听 store 中的用户信息变化
watch(() => userStore.userInfo, (newUserInfo) => {
  if (newUserInfo) {
    loadUserInfo()
  }
}, { immediate: true })

// 保存修改
const handleSave = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    saving.value = true
    try {
      const result = await updateUser({
        username: userForm.username,
        nickname: userForm.nickname,
        email: userForm.email,
        phone: userForm.phone
      })
      
      if (result.code === 200) {
        proxy?.$modal.msgSuccess(t('settings.userInfo.saveSuccess'))
        // 更新 store 中的用户信息
        if (userStore.userInfo) {
          userStore.updateUserInfo({
            user_name: userForm.username,
            name: userForm.nickname,
            email: userForm.email,
            phone: userForm.phone
          })
        }
      } else {
        proxy?.$modal.msgError(result.message || t('settings.userInfo.saveFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('settings.userInfo.saveFailed'))
    } finally {
      saving.value = false
    }
  })
}

// 重置表单
const handleReset = () => {
  formRef.value?.resetFields()
  loadUserInfo()
}

onMounted(() => {
  loadUserInfo()
})
</script>

<style scoped>
.user-info-form {
  max-width: 600px;
}

@media (max-width: 768px) {
  .user-info-form :deep(.el-form-item__label) {
    width: 100px !important;
  }
}

/* 深色模式样式 */
html.dark .user-info-form :deep(.el-form-item__label) {
  color: var(--el-text-color-primary);
}

html.dark .user-info-form :deep(.el-input__wrapper) {
  background-color: var(--el-bg-color);
  border-color: var(--el-border-color);
}

html.dark .user-info-form :deep(.el-input__inner) {
  color: var(--el-text-color-primary);
}
</style>

