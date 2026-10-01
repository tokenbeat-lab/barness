---
status: proposed
date: 2026-10-01
---

# barness-ai 以 npm 发布件重建冻结 pi，并以分类账判定差分

规范要求同一批 E2E 场景与冻结 pi-ai `0.87.1`（commit `898ab804`）比较请求和结果，且 CI 不依赖开发者本机研究目录。冻结 commit 缺少 `src/providers/data` 生成数据，`generate-models` 抓取线上目录，结果随时间漂移。barness 在 `ai/internal/testkit/pioracle/node` 中按 lockfile 固定 npm `@earendil-works/pi-ai@0.87.1`，并由 Go 测试以子进程驱动 Node.js 运行器；差异经版本化分类账逐项归类，待处理项使对应协议的差分门禁失败。

## Considered Options

- 引用研究目录：零准备，但违反“CI 不依赖本机研究仓库”，且该目录缺生成数据。
- 在 barness 内 vendor 冻结源码并自建：可完全控制，但要引入整个 monorepo 构建链，数据仍需外部来源。
- 重新生成模型数据：可从源码重建，但数据为生成时点快照，无法复现发布时的目录与价格。
- 固定 npm 发布件：发布件 `gitHead` 为 `f07218c4`，与冻结 commit 在 `packages/ai`、`packages/telemetry` 仅差 CHANGELOG；冻结源码填入发布数据后的构建产物与发布件逐字节一致（770/770）。首期采用。

## Consequences

- 重建只需 `npm ci --ignore-scripts`；运行时无网络。运行器空环境启动、拒绝凭据/endpoint 变量、套接字限于 loopback、key 仅来自用例；每次运行核对版本与模型数据哈希。来源与核验步骤见 `node/PROVENANCE.md`。
- 比较不删除字段：仅对象键语义排序，数字按精确值比较，只对显式的消息时间路径做一一映射。barness 结果投影到 pi 形状时保留全部序列化字段；Code/Phase 等 Go/租户扩展不进入 pi golden。
- 分类只有已修复/扩展/待处理三类；未登记差异视为待处理，“已修复”项再次出现视为回归。扩展须引用规范条款或明确批准，不能由实现者推断。
- 差分依赖 Node.js 与已安装副本，因此通过 `BARNESS_AI_PIDIFF=1` 显式开启；未开启时证据记为 NOT_RUN，普通 `go test ./...` 保持离线。规范 §6 将受影响差分列入常规门禁：在 CI 中开启并决定待处理项清零前的处理方式，仍待维护者决定，故本 ADR 为 proposed。
- 更换冻结版本即更换兼容基线：先改规范，再重做等价核验、更新 `provenance.json` 与分类账。
