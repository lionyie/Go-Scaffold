# 运行时插件

[English](README.md) | [简体中文](README.zh-CN.md)

> EXPERIMENTAL —— 不属于 MVP 验收项。

运行时插件由宿主在启动时加载，并通过 `/proxy/plugins/<id>/` 对外服务。每个
插件目录包含：

```text
<id>/
├── plugin.json     # 清单（id 必须等于目录名）
├── src/            # 插件源码
└── dist/           # 构建出的 ESM 入口与样式
```

脚手架中此目录刻意留空。新增功能请使用 `extensions/` 中的功能扩展机制，
那是受支持的路径。
