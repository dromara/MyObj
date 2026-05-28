<template>
  <div class="cloud-download-panel">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="云盘账号" name="account">
    <el-form label-width="148px" class="cloud-form">
      <el-form-item :label="t('cloud.provider')">
        <el-select v-model="selectedProvider" :placeholder="t('cloud.selectProvider')" style="width: 100%" @change="resetState">
          <el-option
            v-for="p in cookieProviders"
            :key="p.id"
            :label="p.name"
            :value="p.id"
          >
            <span>{{ p.name }}</span>
            <span class="provider-desc">{{ p.description }}</span>
          </el-option>
        </el-select>
      </el-form-item>

      <el-form-item v-if="selectedProvider && savedBindings.length" label="已保存账号">
        <el-select
          v-model="selectedBindingId"
          clearable
          placeholder="选择已保存的凭据（可选）"
          style="width: 100%"
          @change="onBindingChange"
        >
          <el-option
            v-for="b in savedBindings"
            :key="b.id"
            :label="b.account_name || b.provider"
            :value="b.id"
          />
        </el-select>
      </el-form-item>

      <template v-if="selectedProvider && !selectedBindingId">
        <el-form-item
          v-for="field in credentialFields"
          :key="field.key"
          :label="credentialFieldLabel(field)"
        >
          <el-input
            v-model="credentialFieldsMap[field.key]"
            :type="field.secret ? 'password' : 'text'"
            :show-password="field.secret"
            :placeholder="credentialFieldPlaceholder(field)"
          />
        </el-form-item>

        <el-form-item v-if="!credentialFields.length" :label="credentialLabel">
          <el-input
            v-model="credential"
            type="textarea"
            :rows="3"
            :placeholder="credentialPlaceholder"
          />
          <div class="input-tip">
            <el-icon><InfoFilled /></el-icon>
            <span>{{ t('cloud.credentialTip') }}</span>
          </div>
        </el-form-item>
      </template>

      <el-form-item v-if="selectedProvider && !selectedBindingId">
        <el-checkbox v-model="saveBinding">验证成功后保存凭据</el-checkbox>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="validating" @click="handleValidate">
          {{ t('cloud.validate') }}
        </el-button>
      </el-form-item>

      <el-alert v-if="accountValidateError" type="error" :closable="true" show-icon class="form-error-alert" @close="accountValidateError = ''">
        {{ accountValidateError }}
      </el-alert>

      <el-alert v-if="userInfo" type="success" :closable="false" show-icon class="user-info-alert">
        {{ t('cloud.validatedAs', { name: userInfo.nickname }) }}
      </el-alert>
    </el-form>
      </el-tab-pane>

      <el-tab-pane label="OAuth 网盘" name="oauth">
        <el-form label-width="148px" class="cloud-form">
          <el-form-item label="OAuth 网盘">
            <el-select v-model="oauthProvider" style="width: 100%" @change="resetOAuthState">
              <el-option v-for="p in oauthProviders" :key="p.id" :label="p.name" :value="p.id" :disabled="!p.enabled" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="oauthBindingsForProvider.length" label="已授权账号">
            <el-select v-model="selectedOAuthBindingId" style="width: 100%" clearable placeholder="选择已授权账号">
              <el-option v-for="b in oauthBindingsForProvider" :key="b.id" :label="b.account_name || b.provider" :value="b.id" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleOAuthAuthorize">授权登录</el-button>
            <el-button type="success" :loading="validating" @click="handleOAuthValidate">验证并浏览</el-button>
          </el-form-item>
          <el-alert v-if="oauthValidateError" type="error" :closable="true" show-icon class="form-error-alert" @close="oauthValidateError = ''">
            {{ oauthValidateError }}
          </el-alert>
        </el-form>
        <div v-if="validated && activeTab === 'oauth'" class="file-browser">
          <div class="breadcrumb">
            <el-button link @click="navigateTo('')">{{ t('cloud.root') }}</el-button>
            <template v-for="(item, idx) in pathStack" :key="item.fid">
              <span>/</span>
              <el-button link @click="navigateToStack(idx)">{{ item.name }}</el-button>
            </template>
          </div>
          <el-table v-loading="loadingFiles" :data="files" height="260" @row-dblclick="handleRowDblClick">
            <el-table-column :label="t('tasks.fileName')" min-width="200">
              <template #default="{ row }">
                <el-icon v-if="row.is_dir"><Folder /></el-icon>
                <el-icon v-else><Document /></el-icon>
                <span class="file-name">{{ row.file_name }}</span>
              </template>
            </el-table-column>
            <el-table-column :label="t('tasks.fileSize')" width="120">
              <template #default="{ row }">{{ row.is_dir ? '-' : formatSize(row.size) }}</template>
            </el-table-column>
            <el-table-column :label="t('tasks.operation')" width="100">
              <template #default="{ row }">
                <el-button v-if="!row.is_dir" link type="primary" @click="selectFile(row)">{{ t('cloud.download') }}</el-button>
                <el-button v-else link @click="enterFolder(row)">{{ t('cloud.open') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <el-tab-pane label="分享链接" name="share">
        <el-form label-width="148px" class="cloud-form">
          <el-form-item label="分享类型">
            <el-select v-model="shareProvider" style="width: 100%">
              <el-option v-for="p in shareProviders" :key="p.id" :label="p.name" :value="p.id" />
              <el-option label="蓝奏云" value="lanzou" />
            </el-select>
          </el-form-item>
          <el-form-item label="分享链接">
            <el-input v-model="shareUrl" placeholder="粘贴分享链接" />
          </el-form-item>
          <el-form-item label="提取码">
            <el-input v-model="sharePassword" placeholder="可选" />
          </el-form-item>
          <el-form-item v-if="shareProvider === 'baidu_share'" label="百度 Cookie">
            <el-input v-model="shareExtraCookie" type="password" show-password placeholder="请输入 BDUSS" />
          </el-form-item>
          <el-form-item v-if="shareProvider === 'aliyun_share'" label="刷新令牌">
            <el-input v-model="shareExtraRefresh" type="password" show-password placeholder="阿里云盘 refresh_token" />
          </el-form-item>
          <el-form-item v-if="shareProvider === '115_share'" label="115 Cookie">
            <el-input v-model="shareExtraCookie" type="textarea" :rows="2" placeholder="请输入115账号 Cookie" />
          </el-form-item>
          <el-form-item label="保存路径">
            <el-tree-select v-model="shareForm.virtual_path" :data="folderTree" check-strictly style="width: 100%" :props="{ label: 'label', children: 'children' }" node-key="value" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="shareParsing" @click="handleShareParse">解析</el-button>
            <el-button type="success" :loading="shareCreating" :disabled="!shareParsed" @click="handleShareCreate">下载</el-button>
          </el-form-item>
          <el-alert v-if="shareParseError" type="error" :closable="true" show-icon class="share-parse-error" @close="shareParseError = ''">
            {{ shareParseError }}
          </el-alert>
          <el-alert v-if="shareParsed" type="success" :closable="false" show-icon>
            {{ shareParsed.file_name }} ({{ formatSize(shareParsed.file_size) }})
          </el-alert>
        </el-form>
      </el-tab-pane>
    </el-tabs>

    <div v-if="validated && activeTab === 'account'" class="file-browser">
      <div class="breadcrumb">
        <el-button link @click="navigateTo('')">{{ t('cloud.root') }}</el-button>
        <template v-for="(item, idx) in pathStack" :key="item.fid">
          <span>/</span>
          <el-button link @click="navigateToStack(idx)">{{ item.name }}</el-button>
        </template>
      </div>

      <el-table
        v-loading="loadingFiles"
        :data="files"
        height="260"
        @row-dblclick="handleRowDblClick"
      >
        <el-table-column :label="t('tasks.fileName')" min-width="200">
          <template #default="{ row }">
            <el-icon v-if="row.is_dir"><Folder /></el-icon>
            <el-icon v-else><Document /></el-icon>
            <span class="file-name">{{ row.file_name }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('tasks.fileSize')" width="120">
          <template #default="{ row }">
            {{ row.is_dir ? '-' : formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('tasks.operation')" width="100">
          <template #default="{ row }">
            <el-button v-if="!row.is_dir" link type="primary" @click="selectFile(row)">
              {{ t('cloud.download') }}
            </el-button>
            <el-button v-else link @click="enterFolder(row)">{{ t('cloud.open') }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="total > pageSize"
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="prev, pager, next"
        class="file-pagination"
        @current-change="loadFiles"
      />
    </div>

    <el-form v-if="selectedFile && activeTab !== 'share'" :model="form" label-width="148px" class="download-form cloud-form">
      <el-divider />
      <el-form-item :label="t('offline.saveLocation')">
        <el-tree-select
          v-model="form.virtual_path"
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
        <el-switch v-model="form.enable_encryption" />
      </el-form-item>
      <el-form-item v-if="form.enable_encryption" :label="t('offline.encryptPassword')">
        <el-input v-model="form.file_password" type="password" show-password />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="creating" @click="handleCreateDownload">
          {{ t('cloud.downloadFile', { name: selectedFile.file_name }) }}
        </el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Document, Folder, InfoFilled } from '@element-plus/icons-vue'
import {
  createCloudDownload,
  createCloudShareDownload,
  createLanzouDownload,
  getCloudProviders,
  listCloudCredentialBindings,
  listCloudFiles,
  listCloudOAuthBindings,
  parseCloudShare,
  parseLanzouShare,
  startCloudOAuth,
  validateCloudCredential,
  type CloudCredentialBinding,
  type CloudFileInfo,
  type CloudProviderInfo,
  type CredentialField,
  type LanzouParseResult,
  type OAuthProviderInfo
} from '@myobj/api/cloud'
import { useI18n } from '@/composables'
import { formatSize } from '@/utils'

const props = defineProps<{
  folderTree: any[]
}>()

const emit = defineEmits<{
  success: []
}>()

const { t } = useI18n()
const { proxy } = getCurrentInstance() as ComponentInternalInstance

const providers = ref<CloudProviderInfo[]>([])
const oauthProviders = ref<OAuthProviderInfo[]>([])
const oauthBindings = ref<Array<{ id: string; provider: string; account_name: string }>>([])
const bindings = ref<CloudCredentialBinding[]>([])
const activeTab = ref('account')
const selectedProvider = ref('')
const oauthProvider = ref('')
const selectedOAuthBindingId = ref('')
const activeOAuthBindingId = ref('')
const shareProvider = ref('baidu_share')
const shareUrl = ref('')
const sharePassword = ref('')
const shareExtraCookie = ref('')
const shareExtraRefresh = ref('')
const shareParsed = ref<LanzouParseResult | null>(null)
const shareParseError = ref('')
const accountValidateError = ref('')
const oauthValidateError = ref('')
const shareParsing = ref(false)
const shareCreating = ref(false)
const shareForm = reactive({ virtual_path: '/云盘分享下载/' })
const credential = ref('')
const credentialFieldsMap = reactive<Record<string, string>>({})
const selectedBindingId = ref('')
const saveBinding = ref(false)
const activeBindingId = ref('')
const validating = ref(false)
const validated = ref(false)
const userInfo = ref<{ nickname: string } | null>(null)
const loadingFiles = ref(false)
const files = ref<CloudFileInfo[]>([])
const page = ref(1)
const pageSize = ref(50)
const total = ref(0)
const currentDir = ref('')
const pathStack = ref<{ fid: string; name: string }[]>([])
const selectedFile = ref<CloudFileInfo | null>(null)
const creating = ref(false)

const form = reactive({
  virtual_path: '/云盘下载/',
  enable_encryption: false,
  file_password: ''
})

const cookieProviders = computed(() =>
  providers.value.filter(p => p.auth_type === 'cookie' || p.auth_type === 'refresh_token')
)

const shareProviders = computed(() =>
  providers.value.filter(p => p.auth_type === 'share_link' && p.id !== 'lanzou')
)

const oauthBindingsForProvider = computed(() =>
  oauthBindings.value.filter(b => b.provider === oauthProvider.value)
)

const currentProvider = computed(() => providers.value.find(p => p.id === selectedProvider.value))

const savedBindings = computed(() =>
  bindings.value.filter(b => b.provider === selectedProvider.value)
)

const credentialFields = computed((): CredentialField[] => currentProvider.value?.credential_fields || [])

const credentialFieldLabel = (field: CredentialField) => {
  const key = `cloud.field.${field.key}`
  const text = t(key)
  return text === key ? field.label : text
}

const credentialFieldPlaceholder = (field: CredentialField) => {
  if (field.help) return field.help
  const key = `cloud.fieldHelp.${field.key}`
  const text = t(key)
  return text === key ? field.label : text
}

const buildCredentialPayload = () => {
  if (activeTab.value === 'oauth' && activeOAuthBindingId.value) {
    return { oauth_binding_id: activeOAuthBindingId.value }
  }
  if (selectedBindingId.value) {
    return { binding_id: selectedBindingId.value }
  }
  if (credentialFields.value.length) {
    const fields = { ...credentialFieldsMap }
    const authType = currentProvider.value?.auth_type
    if (selectedProvider.value === '123pan') {
      const cid = (fields.client_id || '').trim()
      const csec = (fields.client_secret || '').trim()
      if (!cid || !csec) return null
      return { cookie: `${cid}|${csec}` }
    }
    if (authType === 'refresh_token') {
      const rt = (fields.refresh_token || '').trim()
      const cid = (fields.client_id || '').trim()
      const csec = (fields.client_secret || '').trim()
      if (!rt) return null
      const cookie = cid && csec ? `${rt}|${cid}|${csec}` : rt
      return { cookie }
    }
    const cookie = (fields.cookie || Object.values(fields).find(v => v?.trim()) || '').trim()
    return cookie ? { cookie } : null
  }
  const trimmed = credential.value.trim()
  return trimmed ? { cookie: trimmed } : null
}

const credentialLabel = computed(() => {
  if (currentProvider.value?.auth_type === 'refresh_token') {
    return t('cloud.refreshToken')
  }
  return t('cloud.cookie')
})

const credentialPlaceholder = computed(() => {
  if (currentProvider.value?.auth_type === 'refresh_token') {
    return t('cloud.refreshTokenPlaceholder')
  }
  return t('cloud.cookiePlaceholder')
})

const resetState = () => {
  validated.value = false
  userInfo.value = null
  files.value = []
  selectedFile.value = null
  currentDir.value = ''
  pathStack.value = []
  page.value = 1
  selectedBindingId.value = ''
  activeBindingId.value = ''
  saveBinding.value = false
  credential.value = ''
  accountValidateError.value = ''
  Object.keys(credentialFieldsMap).forEach(k => delete credentialFieldsMap[k])
}

const onBindingChange = () => {
  validated.value = false
  userInfo.value = null
  accountValidateError.value = ''
}

watch(credentialFields, fields => {
  fields.forEach(f => {
    if (!(f.key in credentialFieldsMap)) {
      credentialFieldsMap[f.key] = ''
    }
  })
})

const loadBindings = async () => {
  try {
    const res = await listCloudCredentialBindings()
    if (res.code === 200 && res.data) {
      bindings.value = res.data
    }
  } catch {
    // ignore
  }
}

const loadProviders = async () => {
  try {
    const res = await getCloudProviders()
    if (res.code === 200 && res.data) {
      providers.value = res.data.providers || []
      oauthProviders.value = res.data.oauth_providers || []
    }
  } catch {
    proxy?.$modal?.msgError(t('cloud.loadProvidersFailed'))
  }
}

const loadOAuthBindings = async () => {
  try {
    const res = await listCloudOAuthBindings()
    if (res.code === 200 && res.data) {
      oauthBindings.value = res.data
    }
  } catch {
    // ignore
  }
}

const resetOAuthState = () => {
  validated.value = false
  userInfo.value = null
  files.value = []
  selectedFile.value = null
  selectedOAuthBindingId.value = ''
  activeOAuthBindingId.value = ''
  oauthValidateError.value = ''
}

const handleOAuthAuthorize = async () => {
  if (!oauthProvider.value) {
    oauthValidateError.value = t('cloud.oauthProviderRequired')
    proxy?.$modal?.msgError(oauthValidateError.value)
    return
  }
  oauthValidateError.value = ''
  try {
    const res = await startCloudOAuth(oauthProvider.value)
    if (res.code === 200 && res.data?.authorize_url) {
      window.open(res.data.authorize_url, '_blank')
      proxy?.$modal?.msgSuccess('请在弹出窗口完成授权，然后重新选择已授权账号')
      await loadOAuthBindings()
    }
  } catch {
    proxy?.$modal?.msgError('授权失败')
  }
}

const handleOAuthValidate = async () => {
  if (!oauthProvider.value) {
    oauthValidateError.value = t('cloud.oauthProviderRequired')
    proxy?.$modal?.msgError(oauthValidateError.value)
    return
  }
  if (!selectedOAuthBindingId.value) {
    oauthValidateError.value = t('cloud.oauthBindingRequired')
    proxy?.$modal?.msgError(oauthValidateError.value)
    return
  }
  oauthValidateError.value = ''
  validating.value = true
  try {
    const res = await validateCloudCredential({
      provider: oauthProvider.value,
      oauth_binding_id: selectedOAuthBindingId.value
    })
    if (res.code === 200) {
      validated.value = true
      activeOAuthBindingId.value = selectedOAuthBindingId.value
      selectedProvider.value = oauthProvider.value
      userInfo.value = { nickname: res.data?.nickname || oauthProvider.value }
      await loadFiles()
      proxy?.$modal?.msgSuccess(t('cloud.validateSuccess'))
    } else {
      oauthValidateError.value = resolveApiError(res, t('cloud.validateFailed'))
      proxy?.$modal?.msgError(oauthValidateError.value)
    }
  } catch (error) {
    oauthValidateError.value = (error as Error).message || t('cloud.validateFailed')
    proxy?.$modal?.msgError(oauthValidateError.value)
  } finally {
    validating.value = false
  }
}

const buildShareExtra = () => {
  const extra: Record<string, string> = {}
  if (shareExtraCookie.value.trim()) extra.cookie = shareExtraCookie.value.trim()
  if (shareExtraRefresh.value.trim()) extra.refresh_token = shareExtraRefresh.value.trim()
  return extra
}

const resolveApiError = (res: { message?: string; msg?: string; data?: unknown }, fallback: string) => {
  if (typeof res.data === 'string' && res.data.trim()) {
    return res.data.replace(/^解析失败:\s*/, '')
  }
  return res.message || res.msg || fallback
}

const validateAccountCredential = (): string | null => {
  if (!selectedProvider.value) {
    return t('cloud.providerRequired')
  }
  if (selectedBindingId.value) {
    return null
  }
  if (credentialFields.value.length) {
    const missing = credentialFields.value.find(
      field => field.required !== false && !(credentialFieldsMap[field.key] || '').trim()
    )
    if (missing) {
      return t('cloud.credentialRequired', { label: credentialFieldLabel(missing) })
    }
    return null
  }
  if (!credential.value.trim()) {
    return currentProvider.value?.auth_type === 'refresh_token'
      ? t('cloud.refreshTokenRequired')
      : t('cloud.cookieRequired')
  }
  return null
}

const validateShareExtra = (): string | null => {
  if (shareProvider.value === 'baidu_share' && !shareExtraCookie.value.trim()) {
    return '请先填写 BDUSS（登录 pan.baidu.com 后，在浏览器 Cookie 中复制 BDUSS 的值）'
  }
  if (shareProvider.value === 'aliyun_share' && !shareExtraRefresh.value.trim()) {
    return '请先填写阿里云盘 refresh_token'
  }
  if (shareProvider.value === '115_share' && !shareExtraCookie.value.trim()) {
    return '请先填写 115 账号 Cookie'
  }
  return null
}

const handleShareParse = async () => {
  if (!shareUrl.value.trim()) {
    shareParseError.value = '请先填写分享链接'
    proxy?.$modal?.msgError(shareParseError.value)
    return
  }
  const extraError = validateShareExtra()
  if (extraError) {
    shareParseError.value = extraError
    proxy?.$modal?.msgError(extraError)
    return
  }
  shareParsing.value = true
  shareParsed.value = null
  shareParseError.value = ''
  try {
    if (shareProvider.value === 'lanzou') {
      const res = await parseLanzouShare({ share_url: shareUrl.value, password: sharePassword.value })
      if (res.code === 200 && res.data) {
        shareParsed.value = res.data
        proxy?.$modal?.msgSuccess(t('offline.parseSuccess'))
      } else {
        shareParseError.value = resolveApiError(res, t('offline.parseFailed'))
        proxy?.$modal?.msgError(shareParseError.value)
      }
    } else {
      const res = await parseCloudShare({
        provider: shareProvider.value,
        share_url: shareUrl.value,
        password: sharePassword.value,
        extra: buildShareExtra()
      })
      if (res.code === 200 && res.data) {
        shareParsed.value = res.data
        proxy?.$modal?.msgSuccess(t('offline.parseSuccess'))
      } else {
        shareParseError.value = resolveApiError(res, t('offline.parseFailed'))
        proxy?.$modal?.msgError(shareParseError.value)
      }
    }
  } catch (error) {
    shareParseError.value = (error as Error).message || t('offline.parseFailed')
    proxy?.$modal?.msgError(shareParseError.value)
  } finally {
    shareParsing.value = false
  }
}

const handleShareCreate = async () => {
  if (!shareUrl.value.trim()) return
  const extraError = validateShareExtra()
  if (extraError) {
    shareParseError.value = extraError
    proxy?.$modal?.msgError(extraError)
    return
  }
  shareCreating.value = true
  try {
    if (shareProvider.value === 'lanzou') {
      const res = await createLanzouDownload({
        share_url: shareUrl.value,
        password: sharePassword.value,
        virtual_path: shareForm.virtual_path
      })
      if (res.code !== 200) {
        proxy?.$modal?.msgError(resolveApiError(res, t('offline.taskCreatedFailed')))
        return
      }
    } else {
      const res = await createCloudShareDownload({
        provider: shareProvider.value,
        share_url: shareUrl.value,
        password: sharePassword.value,
        extra: buildShareExtra(),
        virtual_path: shareForm.virtual_path
      })
      if (res.code !== 200) {
        proxy?.$modal?.msgError(resolveApiError(res, t('offline.taskCreatedFailed')))
        return
      }
    }
    proxy?.$modal?.msgSuccess(t('offline.taskCreatedSuccess'))
    emit('success')
  } catch (error) {
    proxy?.$modal?.msgError((error as Error).message || t('offline.taskCreatedFailed'))
  } finally {
    shareCreating.value = false
  }
}

const handleValidate = async () => {
  const validationError = validateAccountCredential()
  if (validationError) {
    accountValidateError.value = validationError
    proxy?.$modal?.msgError(validationError)
    return
  }
  const payload = buildCredentialPayload()
  if (!payload) {
    accountValidateError.value = t('cloud.validateFailed')
    proxy?.$modal?.msgError(accountValidateError.value)
    return
  }
  accountValidateError.value = ''
  validating.value = true
  try {
    const res = await validateCloudCredential({
      provider: selectedProvider.value,
      ...payload,
      save_binding: saveBinding.value && !selectedBindingId.value
    })
    if (res.code === 200) {
      validated.value = true
      userInfo.value = { nickname: res.data?.nickname || selectedProvider.value }
      activeBindingId.value = res.data?.binding_id || selectedBindingId.value || ''
      activeOAuthBindingId.value = res.data?.oauth_binding_id || ''
      if (saveBinding.value) {
        await loadBindings()
      }
      await loadFiles()
      proxy?.$modal?.msgSuccess(t('cloud.validateSuccess'))
    } else {
      accountValidateError.value = resolveApiError(res, t('cloud.validateFailed'))
      proxy?.$modal?.msgError(accountValidateError.value)
    }
  } catch (error) {
    accountValidateError.value = (error as Error).message || t('cloud.validateFailed')
    proxy?.$modal?.msgError(accountValidateError.value)
  } finally {
    validating.value = false
  }
}

const loadFiles = async () => {
  if (!validated.value) return
  loadingFiles.value = true
  try {
    const credPayload = buildCredentialPayload()
    if (!credPayload && !activeBindingId.value && !activeOAuthBindingId.value) return
    const res = await listCloudFiles({
      provider: selectedProvider.value,
      ...(activeOAuthBindingId.value
        ? { oauth_binding_id: activeOAuthBindingId.value }
        : activeBindingId.value
          ? { binding_id: activeBindingId.value }
          : credPayload!),
      pdir_fid: currentDir.value,
      page: page.value,
      page_size: pageSize.value
    })
    if (res.code === 200 && res.data) {
      files.value = res.data.files || []
      total.value = res.data.total || 0
    }
  } catch {
    proxy?.$modal?.msgError(t('cloud.loadFilesFailed'))
  } finally {
    loadingFiles.value = false
  }
}

const enterFolder = (row: CloudFileInfo) => {
  pathStack.value.push({ fid: row.fid, name: row.file_name })
  currentDir.value = row.fid
  page.value = 1
  selectedFile.value = null
  loadFiles()
}

const navigateTo = (fid: string) => {
  currentDir.value = fid
  pathStack.value = []
  page.value = 1
  selectedFile.value = null
  loadFiles()
}

const navigateToStack = (idx: number) => {
  const item = pathStack.value[idx]
  pathStack.value = pathStack.value.slice(0, idx + 1)
  currentDir.value = item.fid
  page.value = 1
  selectedFile.value = null
  loadFiles()
}

const handleRowDblClick = (row: CloudFileInfo) => {
  if (row.is_dir) {
    enterFolder(row)
  } else {
    selectFile(row)
  }
}

const selectFile = (row: CloudFileInfo) => {
  selectedFile.value = row
}

const handleCreateDownload = async () => {
  if (!selectedFile.value) return
  if (form.enable_encryption && !form.file_password) {
    proxy?.$modal?.msgError(t('offline.encryptPasswordRequired'))
    return
  }
  creating.value = true
  try {
    const credPayload = buildCredentialPayload()
    const res = await createCloudDownload({
      provider: selectedProvider.value,
      ...(activeOAuthBindingId.value
        ? { oauth_binding_id: activeOAuthBindingId.value }
        : activeBindingId.value
          ? { binding_id: activeBindingId.value }
          : credPayload!),
      file_id: selectedFile.value.fid,
      file_name: selectedFile.value.file_name,
      file_size: selectedFile.value.size,
      virtual_path: form.virtual_path || '/云盘下载/',
      enable_encryption: form.enable_encryption,
      file_password: form.file_password
    })
    if (res.code === 200) {
      proxy?.$modal?.msgSuccess(t('offline.taskCreatedSuccess'))
      emit('success')
    } else {
      proxy?.$modal?.msgError(resolveApiError(res, t('offline.taskCreatedFailed')))
    }
  } catch {
    proxy?.$modal?.msgError(t('offline.taskCreatedFailed'))
  } finally {
    creating.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadProviders(), loadBindings(), loadOAuthBindings()])
})
</script>

<style scoped>
.cloud-download-panel :deep(.cloud-form .el-form-item) {
  align-items: flex-start;
  margin-bottom: 18px;
}
.cloud-download-panel :deep(.cloud-form .el-form-item__label) {
  white-space: normal;
  word-break: break-word;
  line-height: 1.45;
  height: auto;
  padding-top: 8px;
}
.cloud-download-panel .input-tip {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.provider-desc {
  float: right;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.user-info-alert,
.form-error-alert {
  margin-bottom: 12px;
}
.breadcrumb {
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}
.file-name {
  margin-left: 6px;
}
.file-pagination {
  margin-top: 8px;
  justify-content: center;
}
.download-form {
  margin-top: 8px;
}
</style>
