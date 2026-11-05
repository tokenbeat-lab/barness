# barness-ai 未发布变更

### 2026-11-04：采用 DeepSeek V4.1 Flash 1017

官方 2026-10-17 更新继续使用 `deepseek-flash` 调用 1017 版本，调用方式与价格不变。
现有 Responses 和 Chat 身份已具备该协议能力；本次按 Responses 优先顺序执行两条 CLI revise/check/apply/export。
仅将两处展示名称从 DeepSeek V4.1 Flash 更新为 DeepSeek V4.1 Flash 1017，并补充官方来源与字段决定；不虚构日期 wire ID，不增加重复身份。
[官方更新](https://api-docs.deepseek.com/updates/) · [官方模型与价格](https://api-docs.deepseek.com/quick_start/pricing/)。

目录仍为 81 条（75 chat、3 image、3 classifier），DeepSeek 仍为 4 条，内容版本依次 2 → 3 → 4。
全部 ID、能力、价格及相对顺序保持；文本/图片、1M context、384k output、已声明 low/high/max 继续使用。
峰值 input/output/cacheRead/cacheWrite 为 0.3/1.2/0.006/0 USD/百万 token；仅作固定保守估算，不按日历自动换价。
`deepseek-flash` 是滚动别名；展示 1017 不构成永久锁定该 checkpoint。

Catalog.Version：`catalog-34de22d2e31abb82108b6dde3dccb5def54fbb825e2f7281b862924b8f6860b4`。
Catalog.Hash：`sha256:fd23a12ac65a114116bcfbaf6f6ff335e3ab4cf5b2af255cc00d1cff49a85156`。
完整 JSON SHA-256：`672cb593ac0e3c6894bfbca8c753254162fdcba8c41adf8321a37108c6e31de4`。
维护 revision：`sha256:14525dac0d1c3505be8f3b330f7bdb25f9e29a609779618bcd9f0b9d61ac8ee1`。

SDK 核心、依赖与 schema 未改变；Go 改动仅相关测试与示例目录配置路径。
历史采用检查保留原展示名称来核对原运行契约；新专项独立核对当前名称，本批数据审计严格限定两处 name 变化。
ChineseEvaluation 新建 v10，仅改配置文档版本与目录 pin，v1–v9 原字节保持。
原预期 RED、Kimi 历史标签检查失败与修复重跑均留存；启动准备尚未完成时的一次 launcher 失败单独记录，缺少的精确起止不补造。
完整普通验证、生成、构建与静态检查的真实结果见完成索引。未执行厂商请求，不把滚动别名旧测量改为 1017 实测资格。
外部 JSON 仅对本次实际受测程序取得代码兼容依据，Host 显式加载并创建新 Client；现有 Client 不热更新。
BuiltinCatalog 默认更新须重建分发程序；其他旧版本兼容性未验证。未提交、未发布软件或部署。
[当前状态](verification/status.md) · [本次交付](delivery.md)。

完整普通实际 PASS/exit0；两段并集覆盖 937 个已发现顶层测试，100 项固定必需场景 PASS，原生 audit 均为 0。最终同源生成、构建、静态检查和 live 标签编译/静态检查均 exit0。
