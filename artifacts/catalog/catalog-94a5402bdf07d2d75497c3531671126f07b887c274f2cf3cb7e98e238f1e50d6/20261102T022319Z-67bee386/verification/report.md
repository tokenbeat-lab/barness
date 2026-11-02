# 内置目录采用验证

受测提交：`d4dbafaa1537ada3f6a56ade817dc63ad20d9f9d`。
完整普通验证 PASS，实际退出 0，墙钟 1234.91 秒。命令由
`run-ordinary.py` 执行，精确子进程 argv、来源和选定输入见回执。

- 回执：`.evidence/catalog-gpt61-terra-builtin/e2e/20261102T013403.060605Z-n5h34x6a/receipt.json`，SHA-256 `2cf3d64eac39e3fc0120f938a8321a7088e4f2e005aa82ce1c108115cfbdb74e`。
- 唯一主 E2E bundle：`bundles/20261102T013405.293044000Z/`，manifest SHA-256 `0a54c500d83f1229d5ca92b5e8d7dd2c2c82d5d701ba1b609d0b4f82b3ec50f5`；16,889 PASS、226 NOT_RUN，audit 0。
- 三个受控 live-tag 子进程 bundle 分别 32、77、6 个 PASS，audit 均 0。这是离线 harness 验证，不是厂商 live 资格。
- 必需 28 个原生 case 全部 PASS：14 个 Terra、9 个目录 fixture/版本及 inclusion、5 个中文评估配置。逐项 ID 与状态见 `result-v1.json`。
- 源码集合 2,159 文件、选定输入 87 文件，逐文件当前字节与回执相等。批准快照、本地输出、内置数据和基线快照 SHA-256 均为 `04107af45ba03e3a4e7660aa5255997303b8c1442b74fc6d905b37fba0ba80ca`。
- `go generate ./ai`、受影响包 `go build` 和 `go vet` 均退出 0；实际 argv、日志与摘要见 `build-d4dbafa/`。

三个完整普通 attempt 因观察到测试失败而停止，实际状态 CANCELLED/exit -15；它们及定向 RED、后续通过的单独 attempt 均未覆盖。此前编排脚本修改导致 7 个旧字节未在执行时独立保存；现按会话编辑重建，逐个摘要与原回执匹配，存于 `script-versions/`。当前实际脚本自身是普通回执选定输入。

本次 live、race 和压力均 NOT_APPLICABLE：没有新厂商事实需求，也没有共享状态、资源实现或设计负载变更。普通测试中的 skip 不表示执行过压力。没有发布或新厂商资格结论。

机器可读索引 `result-v1.json` SHA-256：`4f2332c9b42251c21b3591d1312f516c617d4f58f8fae2577da7b512fa0316bd`。它列出所有回执路径、摘要、退出、各 bundle、必需 case 与适用性。
