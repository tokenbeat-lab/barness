# Kimi 六路线候选付费验证提案

这是待批准的具体范围，尚无付费执行授权。

双轴审查的五项问题及命名判断均已修复并复核关闭。最终完整普通回执 exit0/PASS：13956 个原生 PASS 用例、214 个重负载 NOT_RUN 用例；全部预声明必需 ID、非空断言与哈希核对成功。主审计及五个辅助包的实际 Go 事件库存审计均零发现；两个 CLI 包只有编译事实，其行为由主公共进程 E2E 覆盖。

| 路线 | Provider | Account / Region | 凭据变量名 | 估算预留上限 |
| --- | --- | --- | --- | --- |
| kimi-cn-chat | moonshot-cn | KIMI / CN | `KIMI_CN_KEY` | US$1.50 |
| kimi-cn-responses | moonshot-cn | KIMI / CN | `KIMI_CN_KEY` | US$1.50 |
| kimi-cn-messages | moonshot-cn | KIMI / CN | `KIMI_CN_KEY` | US$1.50 |
| kimi-intl-chat | moonshot | KIMI / GLOBAL | `KIMI_KEY` | US$1.50 |
| kimi-intl-responses | moonshot | KIMI / GLOBAL | `KIMI_KEY` | US$1.50 |
| kimi-intl-messages | moonshot | KIMI / GLOBAL | `KIMI_KEY` | US$1.50 |

每路线独立一次候选运行：9 Logical Calls、最多9 HTTP Attempts、零重试；每调用输出≤4096、每路线总输出预留≤36864、输入预留≤131072；单调用≤120秒、单进程≤1800秒。六路线合计最多54次调用，固定价格快照的保守美元预留总上限 US$9。失败、取消和中断消耗该次授权；失败后停止，不自动重买。账户/地区与变量名来自本对话，尚未读取 Key 值。

授权草案到期：**2026-10-23 14:12:29，Asia/Shanghai**。只批准候选阶段；目录审阅/晋升后的正式验证、受影响旧路线、race 和压力仍按各自阶段另行处理。

受测冻结提交 `0c85bcb716d2e372a55817f39af5c2e1dbc11b4f`；实际运行来源摘要 `8c7c89142c129fd9412073287e5be29a908d7ca5d7462e520fae7f5607f7a7f1`；二进制 SHA-256 `4a7c61e5c86e398a69deb3ec2a5a1fbe9a95e9c601903343ca0612d477cecde8`。二进制构建于 `33640dd7af3221ea70c1762e8d0a1ce8262658b4`；之后只有 E2E import 排序改变，实际 native 来源文件集合和哈希逐项相等，旧二进制及六份 grant 的原 pin 没有改写。

- [完整机器范围及逐路线 grant 哈希](../../../.evidence/barness-ai-kimi/preparation/01/candidate-live-proposal.json)
- [构建 argv、实际退出及来源](../../../.evidence/barness-ai-kimi/preparation/01/build.json)
- [native 来源相等性](../../../.evidence/barness-ai-kimi/preparation/01/native-source-equivalence-02.json)
- [冻结普通验证核对与辅助审计](../../../.evidence/barness-ai-kimi/acceptance/02/verification.json)
- [普通原件](../../../.evidence/barness-ai-kimi/e2e/20261022T222143.475608Z-ssn7nws_/receipt.json)

获准后的精确原生命令为：

```text
/Users/cyber/RestoX/barness/.evidence/barness-ai-kimi/preparation/01/kimi-live.test -test.run=^TestRegionalCandidate$ -test.count=1 -test.timeout=31m
```

每次仅转发所选路线的一个 Key；由 SOP `run.py --feature barness-ai-kimi --stage live --route <该路线>` 采集。其余选定输入与冻结普通回执中的129项保持相同。逐路线再执行只读 `regionalverify -qualify <精确原生 bundle>`；受控/候选材料不冒充正式支持。
