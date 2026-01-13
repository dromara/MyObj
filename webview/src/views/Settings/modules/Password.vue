<template>
  <div class="password-form">
    <el-form
      ref="formRef"
      :model="passwordForm"
      :rules="rules"
      label-width="120px"
      label-position="left"
    >
      <el-form-item :label="t('settings.password.oldPassword')" prop="oldPassword">
        <el-input
          v-model="passwordForm.oldPassword"
          type="password"
          :placeholder="t('settings.password.oldPasswordPlaceholder')"
          show-password
          clearable
        />
      </el-form-item>
      
      <el-form-item :label="t('settings.password.newPassword')" prop="newPassword">
        <el-input
          v-model="passwordForm.newPassword"
          type="password"
          :placeholder="t('settings.password.newPasswordPlaceholder')"
          show-password
          clearable
        />
        <div class="form-tip">{{ t('settings.password.passwordMin') }}</div>
      </el-form-item>
      
      <el-form-item :label="t('settings.password.confirmPassword')" prop="confirmPassword">
        <el-input
          v-model="passwordForm.confirmPassword"
          type="password"
          :placeholder="t('settings.password.confirmPasswordPlaceholder')"
          show-password
          clearable
        />
      </el-form-item>
      
      <el-form-item>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ t('settings.password.change') }}
        </el-button>
        <el-button @click="handleReset">{{ t('common.reset') }}</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { updatePassword } from '@/api/user'
import { getChallenge } from '@/api/auth'
import { rsaEncrypt } from '@/utils/crypto'
import { useI18n } from '@/composables/useI18n'

const { proxy } = getCurrentInstance() as ComponentInternalInstance
const { t } = useI18n()

const formRef = ref<FormInstance>()
const saving = ref(false)

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (value !== passwordForm.newPassword) {
    callback(new Error(t('settings.password.passwordMismatch')))
  } else {
    callback()
  }
}

const rules: FormRules = {
  oldPassword: [
    { required: true, message: t('settings.password.oldPasswordPlaceholder'), trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: t('settings.password.newPasswordPlaceholder'), trigger: 'blur' },
    { min: 6, message: t('settings.password.passwordMin'), trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: t('settings.password.confirmPasswordPlaceholder'), trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

// 修改密码
const handleSave = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    saving.value = true
    try {
      // 获取挑战值
      const challengeRes = await getChallenge()
      if (!challengeRes.data?.publicKey || !challengeRes.data.id) {
        proxy?.$modal.msgError(t('settings.password.getChallengeFailed'))
        return
      }
      
      // 加密密码
      const encryptedOldPassword = rsaEncrypt(challengeRes.data.publicKey, passwordForm.oldPassword)
      const encryptedNewPassword = rsaEncrypt(challengeRes.data.publicKey, passwordForm.newPassword)
      
      const result = await updatePassword({
        old_passwd: encryptedOldPassword,
        new_passwd: encryptedNewPassword,
        challenge: challengeRes.data.id
      })
      
      if (result.code === 200) {
        proxy?.$modal.msgSuccess(t('settings.password.changeSuccess'))
        // 延迟跳转到登录页
        setTimeout(() => {
          proxy?.$cache.local.remove('token')
          proxy?.$cache.local.remove('userInfo')
          window.location.href = '/login'
        }, 1500)
      } else {
        proxy?.$modal.msgError(result.message || t('settings.password.changeFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('settings.password.changeFailed'))
    } finally {
      saving.value = false
    }
  })
}

// 重置表单
const handleReset = () => {
  formRef.value?.resetFields()
}
</script>

<style scoped>
.password-form {
  width: 100%;
  max-width: 800px;
}

.form-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

@media (max-width: 768px) {
  .password-form :deep(.el-form-item__label) {
    width: 100px !important;
  }
}

/* 深色模式样式 */
html.dark .password-form :deep(.el-form-item__label) {
  color: var(--el-text-color-primary);
}

html.dark .password-form :deep(.el-input__wrapper) {
  background-color: var(--el-bg-color);
  border-color: var(--el-border-color);
}

html.dark .password-form :deep(.el-input__inner) {
  color: var(--el-text-color-primary);
}

html.dark .form-tip {
  color: var(--el-text-color-secondary);
}
</style>

