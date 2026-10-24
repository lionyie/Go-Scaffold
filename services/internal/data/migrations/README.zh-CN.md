# 数据库迁移

此目录刻意留空。当前服务使用 GORM `AutoMigrate` 管理脚手架的 schema。

引入真正的迁移工具后（例如 `golang-migrate`），请把
`NNNN_name.up.sql` / `NNNN_name.down.sql` 成对放在这里，并在 `AutoMigrate`
之前执行。`internal/data/database/data.go` 的启动钩子已经会读取此目录，当目录
中存在文件但迁移器未启用时会打印警告。

使用 PostgreSQL 时，请使用有序的数字前缀命名，并且永远不要修改已执行过的
迁移文件。
