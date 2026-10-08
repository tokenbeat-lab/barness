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

测试提交 `7bfd5d72dcb774614ffecb0302aa6a041183d6ab` 的全量 race 回归成功退出：
4,126 个场景全部 PASS，Google 119 个、OpenAI 图像 267 个、冻结 pi 差分 623 个、
压力 11 个。Google 20 个资源场景的全部 gauge 归零；四个新增追溯项、P0 离线、
差分与目录快照校验均通过，差分 pending 为 0。原始 run 与交付目录脱敏审计均为 0 发现，交付哈希校验通过。
原始完整证据保留在 `.evidence/barness-ai/20261008T080243.953470000Z`。

最终全量验证记录在 `verification.json`；`manifest.json` 固定 P09 场景和版本；
`google-cases.json` 保存逐场景回放、fixture/原始输出哈希，以及实际脱敏请求、响应脚本、
结果与断言；重复 fixture 正文留在原始 run 和源码中。`resources.json` 保存 gauge，
`observations-google-images.json` 保留实际观测并单独应用计数字段白名单审计。
源文件 SHA-256 在 `source-hashes.json`，交付文件 SHA-256 在 `sha256.json`。
`trace.json` 保存现有 release 工具计算出的四个 P09 追溯结果；它仅证明此工单离线要求。
`review.md` 分列 Standards/Spec 的发现和复核，`red.json` 保留修复前失败记录。

交付核验与脱敏复查：

```sh
python3 .scratch/barness-ai-pi-1.0/google-interactions-evidence/verify.py
go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/google-interactions-evidence
```

重新采集时，先审计新 run；使用现有 `releasegate -bundle <run> -out <ignored-dir>`
重算 traceability 并提取四个 Google 项到 `trace.json`，然后以仓库相对路径传给
`capture.py`。该现有 bundle 评估的命令/live 发布门禁不在本离线追溯结论内。
交付核验检查源码、交付内容与仍存在的原始逐场景输出；原始 run 不存在时可用 manifest
中的独立命令复跑。本工单只新增宿主目录的生成协议，未改动真实型号支持矩阵。
