<template>
  <div class="user-info-form">
    <el-form
      ref="formRef"
      :model="userForm"
      :rules="rules"
      label-width="120px"
      label-position="left"
    >
      <el-form-item label="用户名" prop="username">
        <el-input
          v-model="userForm.username"
          placeholder="请输入用户名"
          maxlength="50"
          show-word-limit
          clearable
        />
      </el-form-item>
      
      <el-form-item label="昵称" prop="nickname">
        <el-input
          v-model="userForm.nickname"
          placeholder="请输入昵称"
          maxlength="50"
          show-word-limit
          clearable
        />
      </el-form-item>
      
      <el-form-item label="邮箱" prop="email">
        <el-input
          v-model="userForm.email"
          type="email"
          placeholder="请输入邮箱地址"
          maxlength="100"
          clearable
        />
      </el-form-item>
      
      <el-form-item label="手机号" prop="phone">
        <el-input
          v-model="userForm.phone"
          placeholder="请输入手机号"
          maxlength="20"
          clearable
        />
      </el-form-item>
      
      <el-form-item>
        <el-button type="primary" :loading="saving" @click="handleSave">
          保存修改
        </el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, getCurrentInstance, ComponentInternalInstance } from 'vue'
import { updateUser } from '@/api/user'
import type { FormInstance, FormRules } from 'element-plus'
import type { UserInfo } from '@/types'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

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
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在 3 到 50 个字符', trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ]
}

// 加载用户信息
const loadUserInfo = () => {
  try {
    const user = proxy?.$cache.local.getJSON('userInfo') as UserInfo | null
    if (user) {
      userForm.username = user.user_name || ''
      userForm.nickname = user.name || ''
      userForm.email = user.email || ''
      userForm.phone = user.phone || ''
    }
  } catch (error) {
    proxy?.$log.error('加载用户信息失败', error)
  }
}

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
        proxy?.$modal.msgSuccess('用户信息更新成功')
        // 更新本地缓存的用户信息
        try {
          const user = proxy?.$cache.local.getJSON('userInfo') as UserInfo | null
          if (user) {
            user.user_name = userForm.username
            user.name = userForm.nickname
            user.email = userForm.email
            user.phone = userForm.phone
            proxy?.$cache.local.setJSON('userInfo', user)
          }
        } catch (error) {
          proxy?.$log.warn('更新本地用户信息失败', error)
        }
      } else {
        proxy?.$modal.msgError(result.message || '更新失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || '更新失败')
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
</style>

