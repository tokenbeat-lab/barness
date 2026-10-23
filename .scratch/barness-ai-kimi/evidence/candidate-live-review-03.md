# 修订后 Kimi 六路线候选付费验证提案

本次范围尚未批准。用户确认的可选cached_tokens默认0及有限成本边界已实现；原始中国站SSE离线回放现为完整用量/确定成本。双轴审查均PASS/0发现，冻结完整普通验收14045必需用例PASS、214重负载NOT_RUN；14913个Go PASS节点，主和五辅助审计零发现。原付费失败及已消耗授权永久保留，新来源使用新构建和新grant，不续用原范围。

| 路线 | Provider | Account / Region | 凭据变量名 | 估算预留上限 |
| --- | --- | --- | --- | --- |
| kimi-cn-chat | moonshot-cn | KIMI / CN | `KIMI_CN_KEY` | US$1.50 |
| kimi-cn-responses | moonshot-cn | KIMI / CN | `KIMI_CN_KEY` | US$1.50 |
| kimi-cn-messages | moonshot-cn | KIMI / CN | `KIMI_CN_KEY` | US$1.50 |
| kimi-intl-chat | moonshot | KIMI / GLOBAL | `KIMI_KEY` | US$1.50 |
| kimi-intl-responses | moonshot | KIMI / GLOBAL | `KIMI_KEY` | US$1.50 |
| kimi-intl-messages | moonshot | KIMI / GLOBAL | `KIMI_KEY` | US$1.50 |

每路线独立一次候选运行：最多9个Logical Calls、9个HTTP Attempts、零重试；每调用输出≤4096、每路线输出预留≤36864、输入预留≤131072；单调用≤120秒、单进程≤1800秒。六路线共最多54调用，固定价格快照保守预留合计≤US$9。每型号九项：full/simple各stream/complete、原生历史、工具调用/结果、取消和恢复。首个失败、取消或中断即停止整个范围，已执行授权永久占用；不自动重买。

本次草案到期：**2026-10-24T01:15:22+08:00（Asia/Shanghai）**。仅候选阶段；正式验证、目录晋升、旧18条受影响路线、race和压力不在此范围内。账户/地区和变量名沿用用户已确认的信息，本轮未读取Key。

冻结普通受测提交 `2ffa761c437cdb6403cbaf05c7f1846e11965f49`。native来源摘要 `22b841baaf4b6a61f982a29cef8ed4199bdb05748a413705944955911f406eea`，冻结二进制SHA-256 `5c9be68b18ebef4f2a5e7f164ca6421fbe24821643a89c71297a8f9b6062ca3d`。构建于02aea23，完整1351份执行来源与冻结普通来源相同；后续仅文档/索引收尾，不刷新构建或旧证据pin。

- [机器范围和六份新grant哈希](../../../.evidence/barness-ai-kimi/preparation/03/candidate-live-proposal.json)
- [实际构建argv/退出及来源](../../../.evidence/barness-ai-kimi/preparation/03/build.json)
- [有限普通清单与135项输入](ordinary-selection-05.json)
- [最终普通核对与辅助审计](../../../.evidence/barness-ai-kimi/acceptance/05/verification.json)
- [当前来源/输入集合相等性](../../../.evidence/barness-ai-kimi/closure/03/source-equality.json)

获准后的精确原生命令：

```text
/Users/cyber/RestoX/barness/.evidence/barness-ai-kimi/preparation/03/kimi-live.test -test.run=^TestRegionalCandidate$ -test.count=1 -test.timeout=31m
```

每次只通过SOP run.py转发所选路线的一个Key、对应grant与同一135项选定输入。随后对精确原生bundle执行regionalverify -qualify，再用consume_live.py取得schema2外层只读消费者回执。捕获完整脱敏留存；候选通过不产生正式资格。

本提案由[当前状态](../spec.md#唯一当前状态)引用。SOP第4节要求源码改变后使用新来源绑定并取得本次明确付费范围；此前批准的首失败停止范围不能转授新源码。
