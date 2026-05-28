<template>
  <div class="lanzou-download-panel">
    <el-form :model="form" label-width="100px">
      <el-form-item :label="t('lanzou.shareUrl')">
        <el-input v-model="form.share_url" :placeholder="t('lanzou.shareUrlPlaceholder')" />
      </el-form-item>
      <el-form-item :label="t('lanzou.password')">
        <el-input v-model="form.password" :placeholder="t('lanzou.passwordPlaceholder')" show-password />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="parsing" @click="handleParse">
          {{ t('lanzou.parse') }}
        </el-button>
      </el-form-item>
    </el-form>

    <el-alert v-if="parseResult" type="success" :closable="false" show-icon class="parse-result">
      <template #title>
        {{ parseResult.file_name || t('lanzou.unknownFile') }}
        <span v-if="parseResult.file_size_text">（{{ parseResult.file_size_text }}）</span>
      </template>
    </el-alert>

    <el-form v-if="parseResult" :model="downloadForm" label-width="100px" class="download-form">
      <el-form-item :label="t('offline.saveLocation')">
        <el-tree-select
          v-model="downloadForm.virtual_path"
          :data="folderTree"
          :render-after-expand="false"
          :placeholder="t('offline.selectSaveDirectory')"
          style="width: 100%"
          check-strictly
          :props="{ label: 'label', children: 'children' }"
          node-key="value"
        />
      </el-form-item>
      <el-form-item :label="t('offline.encryptStorage')">
        <el-switch v-model="downloadForm.enable_encryption" />
      </el-form-item>
      <el-form-item v-if="downloadForm.enable_encryption" :label="t('offline.encryptPassword')">
        <el-input v-model="downloadForm.file_password" type="password" show-password />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="creating" @click="handleCreate">
          {{ t('offline.createDownload') }}
        </el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import {
  createLanzouDownload,
  parseLanzouShare,
  type LanzouParseResult
} from '@myobj/api/cloud'
import { useI18n } from '@/composables'

defineProps<{
  folderTree: any[]
}>()

const emit = defineEmits<{
  success: []
}>()

const { t } = useI18n()
const { proxy } = getCurrentInstance() as ComponentInternalInstance

const form = reactive({
  share_url: '',
  password: ''
})

const downloadForm = reactive({
  virtual_path: '/蓝奏云下载/',
  enable_encryption: false,
  file_password: ''
})

const parsing = ref(false)
const creating = ref(false)
const parseResult = ref<LanzouParseResult | null>(null)

const handleParse = async () => {
  if (!form.share_url.trim()) {
    proxy?.$modal?.msgError(t('lanzou.shareUrlRequired'))
    return
  }
  parsing.value = true
  parseResult.value = null
  try {
    const res = await parseLanzouShare({
      share_url: form.share_url.trim(),
      password: form.password || undefined
    })
    if (res.code === 200 && res.data) {
      parseResult.value = res.data
      proxy?.$modal?.msgSuccess(t('offline.parseSuccess'))
    } else {
      proxy?.$modal?.msgError(res.message || t('offline.parseFailed'))
    }
  } catch {
    proxy?.$modal?.msgError(t('offline.parseFailed'))
  } finally {
    parsing.value = false
  }
}

const handleCreate = async () => {
  if (!form.share_url.trim()) return
  if (downloadForm.enable_encryption && !downloadForm.file_password) {
    proxy?.$modal?.msgError(t('offline.encryptPasswordRequired'))
    return
  }
  creating.value = true
  try {
    const res = await createLanzouDownload({
      share_url: form.share_url.trim(),
      password: form.password || undefined,
      virtual_path: downloadForm.virtual_path || '/蓝奏云下载/',
      enable_encryption: downloadForm.enable_encryption,
      file_password: downloadForm.file_password
    })
    if (res.code === 200) {
      proxy?.$modal?.msgSuccess(t('offline.taskCreatedSuccess'))
      emit('success')
    } else {
      proxy?.$modal?.msgError(res.message || t('offline.taskCreatedFailed'))
    }
  } catch {
    proxy?.$modal?.msgError(t('offline.taskCreatedFailed'))
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.parse-result {
  margin-bottom: 12px;
}
.download-form {
  margin-top: 8px;
}
</style>
