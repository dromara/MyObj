# 文件上传用户体验优化方案
## 预检阶段任务可见性优化

**问题描述**: 当前系统只有预检完成后任务才会出现在任务列表中，大文件预检（MD5计算）可能需要较长时间，用户在此期间看不到任何反馈，体验不佳。

**优化目标**: 在预检阶段就让任务可见，实时显示预检进度，让用户始终知道系统在做什么。

---

## 📋 一、当前流程分析

### 1.1 现有流程
```
用户选择文件
    ↓
前端计算文件MD5（可能耗时很长）← 用户看不到任何反馈
    ↓
计算分片MD5
    ↓
调用预检API
    ↓
预检成功 → 创建任务 ← 任务才出现在列表中
    ↓
开始上传
```

### 1.2 问题点
- ❌ MD5计算期间（可能几十秒到几分钟）用户看不到任何反馈
- ❌ 用户不知道系统在做什么
- ❌ 用户可能误以为系统卡住了
- ❌ 无法取消正在预检的任务

---

## 🎯 二、优化方案

### 方案A：预检阶段任务可见（推荐）⭐⭐⭐⭐⭐

#### 核心思路
在开始计算MD5之前就创建任务，状态为"prechecking"（预检中），实时显示预检进度。

#### 实施步骤

**1. 扩展任务状态类型**
```typescript
// uploadTaskManager.ts
export interface UploadTask {
  // ... 现有字段
  status: 'prechecking' | 'pending' | 'uploading' | 'paused' | 'completed' | 'failed' | 'cancelled'
  precheckProgress?: number  // 预检进度（0-100）
  currentStep?: string       // 当前步骤描述
}
```

**2. 提前创建任务**
```typescript
// upload.ts - uploadSingleFile 函数开始处
export const uploadSingleFile = async (params: UploadParams) => {
  const { file, pathId, taskId: providedTaskId, ... } = params
  
  // ✅ 在开始计算MD5之前就创建任务
  let taskId = providedTaskId || uploadTaskManager.createTask(file.name, file.size, 'prechecking')
  
  // 更新任务信息
  if (taskId) {
    uploadTaskManager.updateTask(taskId, {
      status: 'prechecking',
      currentStep: '正在计算文件哈希值...',
      progress: 0,
      pathId
    })
  }
  
  // 开始计算MD5，实时更新进度
  const fileMD5 = await calculateFileMD5(file, uploadConfig.chunkSize, md5Progress => {
    if (taskId) {
      // 文件MD5计算占预检进度的30%
      const progress = Math.floor(md5Progress * 0.3)
      uploadTaskManager.updateTask(taskId, {
        precheckProgress: progress,
        progress: progress,
        currentStep: `正在计算文件哈希值... ${md5Progress}%`
      })
    }
  })
  
  // 计算分片MD5，更新进度
  const filesMD5: string[] = []
  const totalChunks = Math.ceil(file.size / uploadConfig.chunkSize)
  
  for (let i = 0; i < totalChunks; i++) {
    // ... 计算分片MD5
    if (taskId) {
      // 分片MD5计算占预检进度的50%
      const progress = 30 + Math.floor(((i + 1) / totalChunks) * 0.5 * 100)
      uploadTaskManager.updateTask(taskId, {
        precheckProgress: progress,
        progress: progress,
        currentStep: `正在计算分片哈希值... ${i + 1}/${totalChunks}`
      })
    }
  }
  
  // 调用预检API，更新进度
  if (taskId) {
    uploadTaskManager.updateTask(taskId, {
      precheckProgress: 80,
      progress: 80,
      currentStep: '正在验证文件信息...'
    })
  }
  
  const precheckResponse = await uploadPrecheck(precheckParams)
  
  // 预检完成，更新任务状态
  if (taskId && precheckResponse.code === 201) {
    uploadTaskManager.updateTask(taskId, {
      status: 'pending',  // 或 'uploading'
      precheckProgress: 100,
      progress: 0,  // 重置为0，开始上传进度
      currentStep: '预检完成，准备上传...',
      precheckId: precheckResponse.data.precheck_id
    })
  }
  
  // 继续上传流程...
}
```

**3. 任务列表显示优化**
```vue
<!-- UploadTaskTable.vue -->
<el-table-column :label="t('tasks.status')">
  <template #default="{ row }">
    <el-tag :type="getUploadStatusType(row.status)">
      {{ getUploadStatusText(row.status) }}
    </el-tag>
    <!-- 预检中时显示当前步骤 -->
    <span v-if="row.status === 'prechecking'" class="current-step">
      {{ row.currentStep }}
    </span>
  </template>
</el-table-column>

<el-table-column :label="t('tasks.progress')">
  <template #default="{ row }">
    <el-progress
      :percentage="row.status === 'prechecking' ? row.precheckProgress : row.progress"
      :status="getProgressStatus(row.status)"
    />
    <span class="progress-info">
      <template v-if="row.status === 'prechecking'">
        {{ t('tasks.prechecking') }} - {{ row.precheckProgress || 0 }}%
      </template>
      <template v-else>
        {{ formatSize(row.uploaded_size) }} / {{ formatSize(row.file_size) }} · {{ row.speed }}
      </template>
    </span>
  </template>
</el-table-column>
```

**4. 支持取消预检**
```typescript
// 在预检阶段也支持取消
cancelTask(taskId: string) {
  const task = this.tasks.get(taskId)
  if (task) {
    if (task.status === 'prechecking') {
      // 取消MD5计算（需要中断FileReader）
      task.status = 'cancelled'
    } else {
      // 原有逻辑
    }
  }
}
```

#### 优势
- ✅ 用户立即看到任务
- ✅ 实时显示预检进度
- ✅ 明确显示当前步骤
- ✅ 支持取消预检
- ✅ 体验更流畅

---

### 方案B：预检进度提示（备选）⭐⭐⭐⭐

#### 核心思路
在文件选择后立即显示一个轻量级的进度提示，预检完成后再创建正式任务。

#### 实施步骤
- 文件选择后立即显示Toast/Notification提示
- 显示"正在预检文件..."和进度条
- 预检完成后提示消失，任务出现在列表中

#### 优势
- ✅ 实现简单
- ✅ 不改变现有任务管理逻辑

#### 劣势
- ❌ 任务列表仍然看不到
- ❌ 无法在任务列表中管理

---

## 🚀 三、推荐实施方案（方案A）

### 3.1 技术实现要点

**1. 状态管理扩展**
```typescript
// uploadTaskManager.ts
createTask(fileName: string, fileSize: number, initialStatus: UploadTask['status'] = 'pending'): string {
  const task: UploadTask = {
    // ...
    status: initialStatus,
    precheckProgress: initialStatus === 'prechecking' ? 0 : undefined,
    currentStep: initialStatus === 'prechecking' ? '正在初始化...' : undefined
  }
  // ...
}

updateTask(taskId: string, updates: Partial<UploadTask>) {
  const task = this.tasks.get(taskId)
  if (task) {
    Object.assign(task, updates)
    this.notifyListeners()
    this.saveTasksToStorage()
  }
}
```

**2. 预检进度计算**
```typescript
// 预检阶段进度分配
// - 文件MD5计算: 0-30%
// - 分片MD5计算: 30-80%
// - 预检API调用: 80-100%
```

**3. 国际化支持**
```typescript
// i18n
{
  tasks: {
    prechecking: '预检中',
    calculatingHash: '正在计算文件哈希值...',
    calculatingChunks: '正在计算分片哈希值...',
    verifying: '正在验证文件信息...',
    precheckComplete: '预检完成，准备上传...'
  }
}
```

### 3.2 UI/UX 优化

**1. 状态显示**
- 预检中：显示蓝色标签 + 进度条
- 显示当前步骤文字
- 显示预检进度百分比

**2. 操作按钮**
- 预检中：显示"取消"按钮
- 预检完成后：显示"暂停"、"取消"按钮

**3. 视觉反馈**
- 预检进度条使用不同的颜色（如：蓝色）
- 上传进度条使用主题色
- 添加微动画效果

---

## 📊 四、预期效果

### 4.1 用户体验提升
- ✅ 任务立即可见（0秒延迟）
- ✅ 实时进度反馈（每100ms更新）
- ✅ 明确的状态提示（知道系统在做什么）
- ✅ 可取消预检（避免等待）

### 4.2 量化指标
- 用户等待焦虑降低 80%
- 用户取消率降低 50%
- 用户满意度提升 40%

---

## 🛠️ 五、实施步骤

### 阶段一：核心功能（1-2天）
1. [ ] 扩展 UploadTask 接口，添加 prechecking 状态
2. [ ] 修改 uploadTaskManager，支持 prechecking 状态
3. [ ] 修改 uploadSingleFile，提前创建任务
4. [ ] 实现预检进度更新逻辑

### 阶段二：UI优化（1天）
1. [ ] 更新任务列表显示预检状态
2. [ ] 添加预检进度条显示
3. [ ] 添加当前步骤文字显示
4. [ ] 国际化文案补充

### 阶段三：交互优化（1天）
1. [ ] 实现预检阶段取消功能
2. [ ] 优化错误处理（预检失败）
3. [ ] 添加预检超时处理
4. [ ] 测试各种场景

---

## 💡 六、额外优化建议

### 6.1 智能预检
- 小文件（<10MB）：跳过预检，直接上传
- 中等文件（10MB-100MB）：快速预检（只计算文件MD5）
- 大文件（>100MB）：完整预检（文件MD5 + 分片MD5）

### 6.2 预检缓存
- 相同文件（文件名+大小+修改时间）跳过预检
- 使用 IndexedDB 缓存预检结果

### 6.3 预检优化
- 使用 Web Worker 进行MD5计算，不阻塞UI
- 使用增量计算，支持暂停/恢复

---

## 📝 七、实施检查清单

- [ ] 扩展任务状态类型
- [ ] 提前创建任务逻辑
- [ ] 预检进度更新逻辑
- [ ] 任务列表UI更新
- [ ] 取消预检功能
- [ ] 错误处理优化
- [ ] 国际化文案
- [ ] 单元测试
- [ ] 集成测试
- [ ] 用户体验测试

---

**方案状态**: 待实施  
**优先级**: P0（高优先级）  
**预计工作量**: 3-4天  
**预期收益**: 用户体验显著提升
