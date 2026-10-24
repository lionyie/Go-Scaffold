# Runtime plugins

[English](README.md) | [简体中文](README.zh-CN.md)

> EXPERIMENTAL — not part of MVP acceptance.

Runtime plugins are loaded by the host at startup and served from
`/proxy/plugins/<id>/`. Each plugin directory contains:

```text
<id>/
├── plugin.json     # manifest (id must equal the directory name)
├── src/            # plugin source
└── dist/           # built ESM entry and styles
```

This directory is intentionally empty in the scaffold. The feature-extension
mechanism in `extensions/` is the supported path for new functionality.
