# 工单 13：Google 参考图编辑与守卫离线证据

公共 Client/HookedClient 通过本地受控 HTTP Provider 完成 Google Interactions 生成与参考图
编辑；真实型号、价格、模态形状及本组合 live 由工单 14 确认。本目录不作 live 支持声明。

验证提交 `b1ac4da`：4,276 个场景全部 PASS，其中 Google 268 个、OpenAI 图像 267 个、
冻结 pi 差分 623 个、压力 11 个；两种 go vet 与完整 race 回归通过。双轴审查无遗留项。

复跑无需厂商凭据：

```sh
go vet ./...
go vet -tags live ./...
go test ./ai/e2e -run '^TestGoogleImages' -count=1
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1
```

每次 E2E 写 `.evidence/barness-ai/<UTC run-id>/`：manifest 包含 SDK/目录版本、测试提交、
逐场景断言与回放命令；结束时自动审计。fixture 显式跳过 pi 差分，P09 是已登记扩展。

本次验证包含：prompt 在前、有序参考图与独立输入快照；构造前和回调后 MIME/base64、
宿主/型号/14 张上限、单图与含编码开销的 20,000,000 字节总请求限制；每个禁止字段与
固定项删除/放宽的独立拒绝；最终型号尺寸/宽高比；全部输出步骤的顺序；坏图、超限、
取消和回调失败的原子结果、合法用量保留、无 2xx 重放；响应体/许可归零和 Observer 脱敏。
带类型 map、嵌入结构体、指针、raw JSON、普通 byte slice 和自定义 map 键的普通值均有
已知长度分配量断言；合法带类型图片和自定义键提示在精确宿主请求预算下通过。自定义 marshaler 表示
按既有契约在冻结后检查，不能扩大授权或超过最终资源预算。

`verification.json` 记录完整回归的实际退出码、场景数量和版本；`manifest.json`、
`google-cases.json` 保存 P09 回放与实际脱敏请求/脚本/结果/断言，附原始文件哈希。
`resources.json` 是资源归零读数，`observations-google-images.json` 是独立白名单审计的
真实观测记录。`trace.json` 保存现有 releasegate 对本工单追溯项的计算结果，
仅证明离线要求；现有 bundle 评估的未运行命令/live 发布门禁不属于此追溯结论。
`source-hashes.json` 与 `sha256.json` 校验源文件和交付内容；`red.json` 保留实现前及
审查修复前的公共 E2E 红灯。双轴审查与复核分列在 `review.md`。

核验：

```sh
python3 .scratch/barness-ai-pi-1.0/google-reference-images-evidence/verify.py
go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/google-reference-images-evidence
```

重新采集时先审计完整 run，再以 `releasegate -bundle <run> -out <ignored-dir>` 重算追溯；
提取七个 P09 项到 trace.json，以仓库相对 run 路径执行 capture.py，最后核验并审计交付。
每个工单的 capture/verify 保持独立，避免后续交付改变既有不可变证据的版本与校验语义。
