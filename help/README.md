# MyObj 帮助文档

这是 MyObj 项目的 VitePress 帮助文档站点。

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建文档
npm run build

# 预览构建结果
npm run preview
```

## 文档结构

```
docs/
├── index.md              # 首页
├── guide/                # 使用指南
│   ├── getting-started.md
│   ├── installation.md
│   ├── configuration.md
│   ├── features.md
│   └── ...
├── api/                  # API 文档
│   ├── overview.md
│   └── ...
└── faq.md               # 常见问题
```

## 部署

### 构建

```bash
# 构建文档
npm run build
```

构建完成后，输出目录为 `docs/.vitepress/dist`。

### 部署到腾讯 Pages

1. **构建文档**：
   ```bash
   cd help
   npm run build
   ```

2. **部署 dist 目录**：
   - 将 `docs/.vitepress/dist` 目录下的**所有内容**部署到腾讯 Pages
   - 确保部署的根目录包含 `index.html`、`assets/` 目录等所有文件
   - **重要**：不要只部署 `dist` 目录本身，而是部署 `dist` 目录内的所有内容

3. **配置部署路径**：
   - 如果部署在子路径（如 `/help/`），确保 VitePress 的 `base` 配置与部署路径一致
   - 当前配置为 `base: '/help/'`，适用于部署在 `/help/` 子路径

4. **检查静态资源**：
   - 部署后检查 `assets/` 目录是否包含所有 CSS、JS 文件
   - 检查 `favicon.svg`、`LOGO.png` 等静态资源是否在根目录

### 常见问题

- **资源 404 错误**：确保部署的是 `dist` 目录内的所有内容，而不是 `dist` 目录本身
- **样式丢失**：检查 `assets/` 目录是否完整部署
- **路径错误**：确保 `base` 配置与部署路径一致
