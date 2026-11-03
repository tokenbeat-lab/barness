# barness-ai 未发布变更

### 2026-11-03：OpenAI chat 下架旧型号

按维护者要求，仅保留 GPT-5.6、GPT-6 及以上的现有 OpenAI chat 身份：
`gpt-5.6-luna`、`gpt-5.6-sol`、`gpt-5.6-terra`、`gpt-6-astra`、
`gpt-6-luna`、`gpt-6-sol`、`gpt-6.1-sol`、`gpt-6.1-terra`。
其余 20 个型号（含 Daybreak 别名）的 37 条四维身份逐条通过
`withdraw` 预览/冻结、`check`、`apply` 预览/确认下架；完整身份清单保存在交付记录。
剩余 OpenAI chat 为 Responses 8 条、Chat Completions 7 条。
总目录从 125 条缩减为 88 条（82 chat、3 image、3 classifier）。
下架定义、来源和追加历史保留，其他 Provider、image/classifier、保留模型定义、
价格及相对顺序不变。OpenAI 内容版本由 2 升至 39，其他 Provider 版本不变。

| 目录标识 | 值 |
| --- | --- |
| Catalog.Version | `catalog-21027c1121f250321e83789b00daa0b4573b372f72258256da1d0378d21fa5b2` |
| Catalog.Hash | `sha256:9dff935e34e42b96adc8afeb34c2db689447c936f1c584928ea08b3ed0a7b65a` |
| JSON 文件 SHA-256 | `e706fdc1dd13aed8799478da342467a810005e769420abce1ed0493dd8c0d9e6` |

**代码与分发影响：** 比较基准为 `b041eff5ca74393a88f66916b13ce1aaeec12ee8`。
核心运行代码、模型字段和运行依赖未改变。变更包括目录数据与完整审阅基线、
相关 E2E 历史夹具装配、文档、OpenAI 验证工具默认目标，以及 ChineseEvaluation 默认配置 v6（只改配置版本
和目录 pin，v1–v5 原字节保留）。有外部加载能力且通过本次检查的受测程序可只
接收版本化 JSON，Host 加载后创建新 Client；现有 Client 不热更新。编译内置默认
使用者需要重建并分发程序。其他旧程序版本的兼容性未验证。

OpenAI 验证工具默认目标改为 `gpt-5.6-luna`、`gpt-5.6-sol` 和 `gpt-6-sol`，
quiet 使用已声明支持的 `off`；删除会重新加入已下架 pro 型号并请求厂商的旧探针。
这涉及 `ai/internal/testkit/supportmatrix/` 与 `ai/live/` 的验证合同和目标，不改变
SDK 核心。旧报告、测量和 pin 不重写，新目标尚未取得本次新的厂商实测资格。

**验证：** 37 次候选代码兼容检查均实际 PASS/exit0。内置下架 E2E 先 RED，
同步后目录/保留调用/37 条拒绝/完整基线及当前配置定向普通 PASS/exit0。
完整受影响普通范围及生成、构建、vet 的实际结果、源码输入和回执见
[本次唯一状态](verification/status.md)；未执行不计通过。
live 不适用：无新厂商事实；race/压力不适用：并发、资源及设计负载未改变。
本次没有软件发布、部署或新型号厂商资格声明。
版本化 JSON、完整审阅包、changelog 快照和完成索引见
[本次交付](delivery.md)。

最终完整普通实际 PASS/exit0，墙钟 1109.84 秒；固定必需场景 42 项全部通过，全部 audit 为 0。最终同源生成、构建、vet 均 exit0。实际受测工作区源码/输入与结果以本次状态及不可变完成索引为准。

