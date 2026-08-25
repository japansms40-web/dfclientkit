# dfclientkit

多个"账号队列驱动"型桌面客户端项目共用的 Go 基建。每个消费方独立打包、独立运行，
这里只提供可复用的底层能力。

> 前端的 Vue3 桌面壳组件（原 `ui-shell/`，npm 包 `@dongfang/df-ui-shell`）已拆分到
> 独立仓库 [df-ui-shell](https://github.com/japansms40-web/df-ui-shell)，本仓库只保留 Go 模块。

## 目录

- `go/` — Go 模块（`module github.com/japansms40-web/dfclientkit/go`），四个包：
  - `account` — 账号队列模型（CK/UA/IP + 状态统计）与磁盘持久化
  - `appconfig` — 通用的 JSON 配置读写 + 系统默认目录解析
  - `taskrunner` — 账号队列驱动的并发任务引擎（换号/多轮/暂停恢复/取消）
  - `runlog` — 把 taskrunner 的事件格式化成运行日志文案

## 如何被消费

私有模块，消费方需保证 `GOPRIVATE` 覆盖 `github.com/japansms40-web`，`go.mod` 里按版本引用：

```
require github.com/japansms40-web/dfclientkit/go v0.1.0
```

升级：`go get github.com/japansms40-web/dfclientkit/go@vX.Y.Z && go mod tidy`。

## 发布新版本

本模块位于仓库子目录 `go/`，按 Go 子目录模块规则，tag 必须带 `go/` 前缀：

```
git tag -a go/vX.Y.Z -m "..." && git push origin main go/vX.Y.Z
```
