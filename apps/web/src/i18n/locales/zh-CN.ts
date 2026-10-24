export default {
  app: { name: 'Scaffold' },
  nav: { home: '首页', plugins: '插件' },
  home: {
    title: '欢迎使用 Scaffold',
    subtitle: '一个多端脚手架：内置管理端、RBAC 与按需扩展。',
    apps: '多端',
    appsHint: '在 apps/ 下添加 web、admin、mobile 或 desktop。',
    extensions: '扩展',
    extensionsHint: '用 node scripts/module.mjs 安装功能模块与可运行示例。',
    admin: '管理端',
    adminLink: '打开管理端',
  },
  plugin: {
    title: '插件',
    installed: '已安装的运行时插件（实验性）',
    entry: '入口',
  },
  common: {
    refresh: '刷新',
    empty: '暂无数据',
    loading: '加载中',
  },
  errors: {
    UNAUTHORIZED: '请先登录',
    FORBIDDEN: '没有权限执行该操作',
    INTERNAL_ERROR: '服务器内部错误',
    INVALID_ARGUMENT: '输入不合法',
    MUST_CHANGE_PASSWORD: '请先修改密码',
  },
}
