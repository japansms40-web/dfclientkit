# dfclientkit

多个"账号队列驱动"型桌面客户端项目共用的基建。每个消费方独立打包、独立运行，
这里只提供可复用的底层能力。

## 目录

- `go/` — Go 模块（`module dfclientkit`），四个包：
  - `account` — 账号队列模型（CK/UA/IP + 状态统计）与磁盘持久化
  - `appconfig` — 通用的 JSON 配置读写 + 系统默认目录解析
  - `taskrunner` — 账号队列驱动的并发任务引擎（换号/多轮/暂停恢复/取消）
  - `runlog` — 把 taskrunner 的事件格式化成运行日志文案
- `ui-shell/` — npm 包 `@dongfang/df-ui-shell`：桌面壳的 Vue3 组件（标题栏/侧边导航/
  状态栏/日志面板/数字输入/结果弹窗）与基础主题，纯源码分发，不预编译。

## 本地开发中如何被消费

消费方项目在同级目录下（比如 `~/GolandProjects/<consumer>`）：

`go.mod`：
```
require dfclientkit v0.0.0
replace dfclientkit => ../dfclientkit/go
```

`frontend/package.json`：
```json
"@dongfang/df-ui-shell": "file:../../dfclientkit/ui-shell"
```
