# 工单 09：OpenAI 图像生成离线交付

Client 与 HookedClient 已提供同步 GenerateImages。OpenAI 原生生成复用共享 unary
生命周期；显式 ImagePolicy、封闭选项、最终回调校验、整体图像验证、模态用量和独立
Observer 快照见 ADR-0022 与公共契约。编辑/mask 属于工单 10，真实型号/费率/live
属于工单 11；此次 Provider、型号、能力和价格均为合成数据。

失败方式先于实现记录于 [FAILURES.md](FAILURES.md)。两轴复审的两个 P2 先由公共 E2E
复现（[review-red.json](review-red.json)），修复后再专项验证与复审；工程标准 0 项、规格
0 项遗留，见 [review.md](review.md)。没有新增单元测试或内部测试接口。

最终全量在 `5922ebd6a38292412bf4ae606c05ead903c43374` 上通过：3,872 个证据场景全部 PASS，其中图像
132 个；冻结 pi 差分 623 个，pending 0；压力 11 个。
图像 16 个资源场景的 body/call/permit/event/waiter 全部归零；
4 个超限回调分配场景最高 13,104 字节（验收上界 4 MiB）。
run 和交付目录审计均为 0 发现，目录快照未加入未经真实冒烟的图像型号。

[report.json](report.json) 保存测试版本、SDK/目录版本、源文件哈希与统计；
[image-cases.json](image-cases.json) 保存逐场景回放命令、fixture 和输出哈希；
[image-resources.json](image-resources.json) 保存实际资源及分配读数；
[catalog-snapshot.json](catalog-snapshot.json) 保存目录快照。
完整原始 bundle 位于仓库忽略目录 `.evidence/barness-ai/issue09-final/20261008T055212.761348000Z`，可按以下命令重新生成。

    BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai/issue09-final" go test -race ./... -count=1
    python3 .scratch/barness-ai-pi-1.0/openai-generation-evidence/verify.py
    go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/openai-generation-evidence

发布数据评估见 [release-evaluation.json](release-evaluation.json)：此 -bundle 调用仅评估
已有证据，不导入上述成功命令和真实 live bundle，因此整体发布结论为 false、退出码 1；
并非本工单测试失败。P0 离线、差分、审计、追溯、快照及工单 09 新增追溯项均已 PASS。
实际命令退出码见 [commands.json](commands.json)。最终提交只保存工单状态及证据。
