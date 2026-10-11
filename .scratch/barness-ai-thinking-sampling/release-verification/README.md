# pi-ai 1.1.0 发布件来源核验

2026-10-11 两个独立干净临时环境完成核验：AI **827/827**、telemetry **24/24** dist 文件逐字节一致，零差异；依赖 lock、43 个生成模型数据文件、全部来源/输出哈希、命令与脱敏审计均保存。

固定源码为 [v1.1.0](https://github.com/earendil-works/pi/tree/v1.1.0) 的 commit `abe508e1b89912adde45528136c3221eb69acdd7`。npm 的 AI 和 telemetry 1.1.0 元数据均没有 gitHead，因此 provenance 留空，没有将 tag commit 伪装成 npm gitHead。使用官方 tag API 核对 SHA，再下载固定 commit 的 codeload archive；其字节哈希由 inputs.json 固定。Git 端点连接失败的第一次尝试保留在 evidence-a，未计入通过。

| 材料 | 执行结果 |
| --- | --- |
| [evidence-b/report.json](evidence-b/report.json) / [evidence-c/report.json](evidence-c/report.json) | 两次独立来源等价 PASS |
| 各包的 dist-ai.sha256、dist-telemetry.sha256 | 完整 dist 文件集合及内容哈希 |
| dependencies.json、packages.json、inputs.sha256 | 固定依赖与官方发布件元数据 |
| failure-checks.json | 9 个误版本/虚构 gitHead/改坏数据/产物/依赖/完整性输入均拒绝 |
| boundary-checks.json | 下载超限和证据写入失败均拒绝 |
| audit.json、SHA256SUMS | 零发现和所有核验文件完整性 |
| [model-inventory.json](model-inventory.json) | 全部 87 个既有聊天条目逐字段清点 |
| [catalog-changes.json](catalog-changes.json) | 与旧目录比较，仅五处价格阶梯变化 |

只将官方发布的 dist/providers/data 填入固定源码，不运行线上 generate-models/hydrate。按 source lock 安装依赖后依次构建 chord、tui、telemetry、ai build:offline。比较所有隐藏文件、JS、声明、source map 和数据；不是只比较采样源码。内置 85 个发布聊天条目（另两项是重复协议映射）均没有 samplingParams 或 samplingParamsByThinkingLevel，因此不添加生产默认值。已列型号的 thinkingLevelMap、已承载 compat、容量及名称无变化，也没有新增硬约束；未支持型号不自动列入。

重建需要 npm / GitHub 官方下载访问，执行：

```sh
node .scratch/barness-ai-thinking-sampling/release-verification/run.mjs .scratch/barness-ai-thinking-sampling/release-verification/evidence-fresh
```

输出目录必须全新，失败证据不覆盖。边界核验和来源等价不表示 barness 的 E2E 或发布已经通过。运行时只有 ai/internal/testkit/pioracle/node 这一套 oracle；这些来源核验环境不作为第二套业务 oracle 使用。
