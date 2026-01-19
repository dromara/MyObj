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

构建完成后，将 `docs/.vitepress/dist` 目录部署到静态文件服务器即可。
