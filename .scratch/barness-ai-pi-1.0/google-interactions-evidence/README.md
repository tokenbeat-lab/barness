# 工单 12：Google Interactions 图像生成离线证据

仅证明宿主自带目录的 P09 生成协议。参考图由工单 13 交付，真实模态形状、费率、
型号包含与本组合冒烟由工单 14 确认。本目录不构成 Google 图像已通过 live 的声明。

公共 Client/HookedClient 在本地受控 HTTP Provider 上执行 JSON fixture；响应状态、
交错步骤、坏图、用量缺项、思考包含/排除、回调授权与资源失败均形成独立场景。
每次 E2E 自动写 `.evidence/barness-ai/<UTC run-id>/`，含 manifest、SHA-256、逐场景
fixture/脚本/请求/结果/断言与 replay 命令，结束时进行脱敏审计。

复跑（不需要厂商凭据）：

```sh
go vet ./...
go vet -tags live ./...
go test ./ai/e2e -run '^TestGoogleImages' -count=1
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1
```

全量差分使用已冻结安装的 pi 1.0.0，P09 每个场景明确跳过差分；本工单也有 routes、
P0、追溯和原生目录豁免断言。模态 Observer 只保留计数与调用归属，结果、尝试、
观测各自拥有明细；body 与许可在完成/取消/超限/回调失败后归零。

## 验证记录

最终全量验证与双轴审查记录在 `verification.json`；`manifest.json` 固定所交付证据
场景和版本，逐文件 SHA-256 保存在 `sha256.json`。追溯断言由 `trace.json` 记录。
