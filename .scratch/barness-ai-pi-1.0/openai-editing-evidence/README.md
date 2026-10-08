# 工单 10：OpenAI JSON 图像编辑离线交付

GenerateImages 根据参考图选择 generations 或 edits；编辑只发送有序的内联 JSON
参考图，可选 mask 和型号允许的 input_fidelity。输入、可信回调及最终冻结请求均受
宿主、协议和型号能力约束；尺寸规则带来源并在目录快照边界独立复制。编辑复用生成
路径的整体图像发布、用量、请求 ID、失败不重放及资源释放。详见 ADR-0022、公共契约
与 P08 追溯。此工单使用合成 Provider/型号，不修改真实型号目录或支持矩阵。

[FAILURES.md](FAILURES.md) 在实现前记录失败方式。公共 E2E 从 Client/HookedClient
入口验证，无新增内部单元测试。审查发现先补红灯，再修复；最终 Standards 和 Spec
各 0 项遗留，见 [review.md](review.md)。红灯记录仅保留脱敏的场景名称和统计。

最终全量在 `2e86ea2d4472b23a459cf2688c7499d1ed77d139` 上通过：4,003 个场景全部 PASS，
其中图像 263 个、编辑 132 个；冻结 pi 差分 623 个，pending 0；压力 11 个。
图像 61 个资源场景的全部记录 gauge 归零；23 个容量场景最高分配
1,622,256 字节（验收上界 4 MiB）。竞态检查与静态检查通过，原始 run 与
交付目录脱敏审计均为 0 发现。完整原始 bundle 保留于仓库忽略目录
`.evidence/barness-ai/issue10-final/20261008T064250.334556000Z`。

[report.json](report.json) 保存最终提交、版本、源文件哈希与统计；
[image-cases.json](image-cases.json) 保存逐场景回放命令、fixture 和输出哈希；
[image-resources.json](image-resources.json) 保存实际资源与分配读数；
[catalog-snapshot.json](catalog-snapshot.json) 保存未纳入新图像型号的目录快照。
[commands.json](commands.json) 记录检查退出码。

完整回放与交付校验：

    BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai/issue10-final" go test -race ./... -count=1
    python3 .scratch/barness-ai-pi-1.0/openai-editing-evidence/verify.py
    go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/openai-editing-evidence

重新采集时，先审计新 run bundle，再将仓库相对路径传给 capture.py；verify.py 会验证
交付源文件和目录内所有内容哈希，并在原始 bundle 存在时核对逐场景输出。

真实 JSON 编辑接受性、首批 2.5 型号的保真度/能力及目录纳入由工单 11 验证；运行时
没有 multipart 回退。图像功能的离线交付与真实型号发布验收分别留有明确证据。

[release-evaluation.json](release-evaluation.json) 的现有 bundle 评估确认 P0、差分、审计、
追溯和目录快照均 PASS；工单 10 的三个新增追溯项均 PASS。此调用不导入已成功运行的
命令及真实 live 输入，因此整体发布结论为 false、退出码 1；成功测试及静态检查的
实际退出码单列于 commands.json，不能把该评估声明为真实型号发布验收。
