# 工单 13：十组合发布证据

本目录记录实际执行，最终状态以 `release-report.json` / `release-report.md` 为准。
发布 CLI 实装十组合、固定 P10 第七差分、R01–R20、normal/race 各自容量与撤销、
Jev/Luna 双效果、完整文件 inventory 和逐包审计。未执行/失败不会写成交付完成。

`collect.py --live --out <新目录>` 从 `.env` 每次仅取选定厂商 key，分别启动十个
smoke 进程及一个 Jev 评估进程；每个进程只持有自己的凭据。固定显式 loopback
HTTP proxy `127.0.0.1:7890`，官方 endpoint 与 TLS 验证不变。账户、区域、权限、
key 专用性、厂商支出限额均未独立核实，记录 unknown。每次失败保留且输出目录不可复用。
Luna 使用工单 12 的独立完整真实包，按当前执行源码集合验证适用性，不改写原 Hash/时间。

2026-10-13 本次十个 schema 3 smoke 均成功，各自 merge 成功；Jev 为 COMPLETE，
24 Call / 24 Attempt / 0 retry，24/24 正确，已知估价 $0.000539406。
这些是固定测试任务的结果，不保证生产质量、不等于厂商账单。完整 gate 正在核验。

从仓库根重跑（真实采集会产生厂商调用费用）：

```sh
python3 .scratch/barness-ai-openai-decisions/release-evidence/collect.py --live --out .evidence/new-ten-route-live
python3 .scratch/barness-ai-openai-decisions/release-evidence/run-gate.py --collection .evidence/new-ten-route-live/collection.json --out .evidence/new-ten-route-gate
```

`run-gate.py` 本身不持有 key 或访问厂商。它执行现有发布命令，normal/race 同时开启
固定差分与压力，执行两种 vet，导入所有 live 与两个 evaluation 包，并保存实际退出码。
`-bundle`/`-race-bundle` 仅重放证据，不把旧命令结果冒充新的执行。

来源适用性：保留原始完整 code-snapshot 与 CodeHash；当前 gate 比较 runtime、
评估引擎/CLI、requestid、testkit 依赖及 go.mod/go.sum 的全部非测试源码集合（含新增/删除）。
测试、live harness、release 工具及无关 examples 不参与模型效果适用性比较；它们不改变
原评估的执行结果。目录/config/dataset 独立固定。公共 verifier 同时重放真实捕获并核验
manifest、Observer、所有文件 Hash 和 audit；历史 Jev 仍按 recorded snapshot 独立验证。
