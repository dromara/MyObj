<template>
  <el-dialog
    v-model="visible"
    :title="t('uploadEncrypt.title')"
    width="500px"
    :close-on-click-modal="false"
    :close-on-press-escape="true"
    class="upload-encrypt-dialog"
    @close="handleClose"
  >
    <el-form :model="encryptForm" :rules="rules" ref="formRef" label-width="100px">
      <el-form-item :label="t('uploadEncrypt.isEncrypt')">
        <el-switch v-model="encryptForm.is_enc" />
      </el-form-item>

      <el-form-item v-if="encryptForm.is_enc" :label="t('uploadEncrypt.encryptPassword')" prop="file_password">
        <el-input
          v-model="encryptForm.file_password"
          type="password"
          :placeholder="t('uploadEncrypt.encryptPasswordPlaceholder')"
          show-password
          clearable
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleConfirm">{{ t('common.confirm') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { useI18n } from '@/composables'

  interface Props {
    modelValue: boolean
  }

  const props = defineProps<Props>()

  const emit = defineEmits<{
    'update:modelValue': [value: boolean]
    confirm: [config: { is_enc: boolean; file_password: string }]
  }>()

  const { t } = useI18n()

  const visible = computed({
    get: () => props.modelValue,
    set: val => emit('update:modelValue', val)
  })

  const formRef = ref<FormInstance>()

  const encryptForm = reactive({
    is_enc: false,
    file_password: ''
  })

  const rules = reactive<FormRules>({
    file_password: [
      {
        validator: (rule, value, callback) => {
          if (encryptForm.is_enc && !value) {
            callback(new Error(t('uploadEncrypt.passwordRequired')))
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ]
  })

  const handleConfirm = async () => {
    if (!formRef.value) return

    await formRef.value.validate(valid => {
      if (valid) {
        emit('confirm', {
          is_enc: encryptForm.is_enc,
          file_password: encryptForm.is_enc ? encryptForm.file_password : ''
        })
        handleClose()
      }
    })
  }

  const handleClose = () => {
    visible.value = false
    // 重置表单
    encryptForm.is_enc = false
    encryptForm.file_password = ''
    formRef.value?.clearValidate()
  }
</script>

<style scoped>
  .upload-encrypt-dialog :deep(.el-dialog__body) {
    padding: 24px;
  }
</style>
