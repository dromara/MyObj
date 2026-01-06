![Logo](src/assets/images/LOGO.png)
# MyObj 网盘系统 - 前端项目

基于Vue3 + Vite构建的现代化网盘系统前端界面。

## 🚀 功能特性

### 核心功能
- ✨ **文件管理** - 文件上传、下载、删除、重命名、移动、复制
- 📁 **文件夹操作** - 创建、删除、重命名文件夹
- 🔍 **搜索功能** - 快速搜索文件
- 📤 **分享功能** - 生成分享链接、限时分享
- 🚀 **离线下载** - 远程资源异步下载
- 📦 **种子下载** - 支持磁力链和.torrent文件
- 👥 **用户管理** - 登录、注册、个人设置
- 💾 **存储管理** - 存储空间查看和配额管理

### 界面特性
- 🎨 现代化设计风格
- 📱 响应式布局
- 🌈 双视图模式（网格/列表）
- ⚡ 流畅的交互动画
- 🔐 安全的身份验证

## 📦 快速开始

### 安装依赖
```bash
npm install
```

### 开发模式
```bash
npm run dev
```
开发服务器默认运行在 `http://localhost:5173`
API请求会发送到 `http://localhost:8080` (Go后端服务器)

### 生产构建
```bash
npm run build
# 或
npm run build:prod
```

构建产物将输出到 `dist/` 目录。

### 预览生产构建
```bash
npm run preview
```

## ⚙️ 环境配置

### 开发环境
默认配置在 `.env.development` 文件中。
- API地址: `http://localhost:8080`
- 自动配置，无需额外设置

### 生产环境
生产环境有两种配置方式：

#### 方式1：使用当前域名（推荐）
前端和后端部署在同一域名下，API请求会自动使用当前域名。

#### 方式2：使用meta标签
在 `index.html` 中设置API地址：
```html
<meta name="api-url" content="http://your-api-domain.com">
```

后端可以在渲染HTML时动态注入此标签。

## 📁 项目结构

```
webview/
├── src/
│   ├── api/              # API接口定义
│   │   └── auth.js       # 认证相关API
│   ├── components/       # Vue组件
│   │   ├── Navbar.vue    # 顶部导航栏
│   │   ├── Sidebar.vue   # 侧边栏
│   │   └── FileList.vue  # 文件列表
│   ├── views/            # 页面视图
│   │   └── Login.vue     # 登录页面
│   ├── config/           # 配置文件
│   │   └── api.js        # API配置
│   ├── utils/            # 工具函数
│   │   └── request.js    # HTTP请求封装
│   ├── App.vue           # 根组件
│   ├── main.js           # 入口文件
│   └── style.css         # 全局样式
├── public/               # 静态资源
├── .env.development      # 开发环境配置
├── .env.production       # 生产环境配置
├── index.html            # HTML模板
├── package.json          # 项目配置
└── vite.config.js        # Vite配置
```

## 🔌 API集成

### 请求配置
所有API请求通过 `src/utils/request.js` 统一处理：
- 自动添加 Authorization Token
- 统一错误处理
- 401状态自动跳转登录

### API端点
详见 `src/config/api.js`，包含：
- 认证接口 (登录、注册、登出)
- 用户接口 (信息、设置)
- 文件接口 (上传、下载、管理)
- 分享接口 (创建、访问、管理)
- 下载接口 (离线、种子)

## 🚢 部署方案

### Nginx部署
```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    # 前端静态文件
    location / {
        root /path/to/dist;
        try_files $uri $uri/ /index.html;
    }
    
    # API代理到Go后端
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Docker部署
```dockerfile
FROM node:18-alpine as builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## 📝 开发说明

### 技术栈
- **Vue 3** - 渐进式JavaScript框架
- **Vite** - 下一代前端构建工具
- **原生CSS** - CSS变量主题系统
- **Fetch API** - 原生HTTP请求

### 代码规范
- 使用 Composition API
- 组件采用 `<script setup>` 语法
- 样式使用 scoped 隔离
- 遵循Vue 3最佳实践

### 添加新功能
1. 在 `src/api/` 添加API接口
2. 在 `src/components/` 或 `src/views/` 创建组件
3. 在组件中调用API并处理数据
4. 更新路由和导航

## 🔧 配置说明

### Vite配置
`vite.config.js` 可配置：
- 开发服务器端口
- 代理配置
- 构建优化
- 插件扩展

### API配置
`src/config/api.js` 包含：
- API基础URL配置
- 环境判断逻辑
- 端点定义
- 超时设置

## 📞 联系支持

- 项目主页: [GitHub Repository]
- 问题反馈: [GitHub Issues]

---

**注意**: 本前端项目需要配合Go后端服务使用，确保后端服务运行在 `http://localhost:8080`
