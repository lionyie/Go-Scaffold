# Migrations

[English](README.md) | [简体中文](README.zh-CN.md)

This directory is intentionally empty. The service currently uses GORM
`AutoMigrate` for the scaffold schema.

When a real migration tool is introduced (for example `golang-migrate`), put
`NNNN_name.up.sql` / `NNNN_name.down.sql` pairs here and run them before
`AutoMigrate`. The startup hook in `internal/data/database/data.go` already
reads this directory and warns when files are present but the migrator is not
enabled.

For PostgreSQL, use ordered numeric prefixes and never edit an applied file.
