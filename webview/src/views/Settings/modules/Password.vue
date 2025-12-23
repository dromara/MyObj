<template>
  <div class="password-form">
    <el-form
      ref="formRef"
      :model="passwordForm"
      :rules="rules"
      label-width="120px"
      label-position="left"
    >
      <el-form-item label="当前密码" prop="oldPassword">
        <el-input
          v-model="passwordForm.oldPassword"
          type="password"
          placeholder="请输入当前密码"
          show-password
          clearable
        />
      </el-form-item>
      
      <el-form-item label="新密码" prop="newPassword">
        <el-input
          v-model="passwordForm.newPassword"
          type="password"
          placeholder="请输入新密码"
          show-password
          clearable
        />
        <div class="form-tip">密码长度至少 6 位</div>
      </el-form-item>
      
      <el-form-item label="确认新密码" prop="confirmPassword">
        <el-input
          v-model="passwordForm.confirmPassword"
          type="password"
          placeholder="请再次输入新密码"
          show-password
          clearable
        />
      </el-form-item>
      
      <el-form-item>
        <el-button type="primary" :loading="saving" @click="handleSave">
          修改密码
        </el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, getCurrentInstance, ComponentInternalInstance } from 'vue'
import { updatePassword } from '@/api/user'
import { getChallenge } from '@/api/auth'
import { rsaEncrypt } from '@/utils/crypto'
import type { FormInstance, FormRules } from 'element-plus'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

const formRef = ref<FormInstance>()
const saving = ref(false)

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (value !== passwordForm.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  oldPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
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
        proxy?.$modal.msgError('获取挑战值失败')
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
        proxy?.$modal.msgSuccess('密码修改成功，请重新登录')
        // 延迟跳转到登录页
        setTimeout(() => {
          proxy?.$cache.local.remove('token')
          proxy?.$cache.local.remove('userInfo')
          window.location.href = '/login'
        }, 1500)
      } else {
        proxy?.$modal.msgError(result.message || '密码修改失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || '密码修改失败')
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
  max-width: 600px;
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
</style>

