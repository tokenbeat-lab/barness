# Kimi K3.1 Pro 中国站官方来源研究

抓取原件及其 UTC 时间、HTTP 状态、最终 URL、响应头、字节数、SHA-256 分别保存在 `inputs/cn/*.body` 和 `*.json`。`platform.moonshot.cn/docs/pricing/chat` 的实际最终地址是 `https://platform.kimi.com/docs/pricing/chat`。这些是公开文档读取，不是模型调用或资格验证。

- 官方 `quickstart.body` 与 `models.body`：wire ID `kimi-k3.1-pro`；K3 系列上下文 1,048,576；始终思考且保留思考，low/high/max 默认 max。`quickstart.body` 第 472 行明确输出默认 131,072、可设置的最大值 1,048,576；不能把默认当厂商最高能力。目录能否表达两种上限是另一个边界问题。
- `chat.body`、`responses.body` 和 `messages.body` 均明确提供此 ID；Chat/Responses 的文档分别明确默认及最高输出。官方声明可用不等于本项目三路线已完成厂商实测。原件见 https://platform.kimi.com/docs/api/chat.md、https://platform.kimi.com/docs/api/responses.md、https://platform.kimi.com/docs/api/messages.md。
- `pricing.body` 是官方 HTML，包含完整服务端嵌入价格表；其 Markdown 版本 `pricing-md.body` 的表为空，不可单独用它取得费率。中国站标准按量每百万 token：输入 ¥13、输出 ¥68、缓存命中 ¥1.30、5m 缓存写入 ¥13、1h 缓存写入 ¥26。1 MTok 为 1,000,000 token；Batch 60% 另属批量规则，未混入标准费率。官方页面 https://platform.kimi.com/docs/pricing/chat。
- `cache.body`：自动缓存默认 5m，可选择 1h；Chat/Responses 用 prompt_cache_options，Messages 顶层 cache_control。计费独立写入，不能把缓存写入视为免费。此 K3.1 Pro 在文档中明确支持 cache write。https://platform.kimi.com/docs/guide/context-caching.md。
- 此次中国站定价原件为 CNY，未发现中国站该 serving 的 USD 定价；不能把国际站价格当中国站价格。当前目录四费率 USD 要求额外有明确换算依据。

## FX 决策边界

`docs/adr/0052-barness-ai-kimi-cost-estimates.md` 第 17 行将仓库 2026-10-16 CFETS 1 USD = 6.7220 CNY 历史快照明确用于中国**初批**，基础费率按十位半偶舍入；1h 使用已舍入 input × 2，更新快照须换版本并取得适用验证。原件索引 `ai/internal/testkit/kimi/source-files.json` 的 fx-2026-10-16，来源 https://www.chinamoney.com.cn/r/cms/www/chinamoney/data/fx/ccpr-notice2026-10-16.json，原始 SHA-256 80922f3101d7f18366502258e4f93ce418b2369044aded64c8a650eea0e87020。ADR 未明确将该固定汇率授权扩展到 K3.1 新型号，也未要求每新模型必须抓实时汇率。

可供决策的历史基准估算（不是当前汇率或官方 USD 报价）：依上述精度换算 Input 1.9339482297、Output 10.1160368938、CacheRead 0.1933948230、5m Write 1.9339482297 USD/百万；1h 按已舍入 Input×2 为 3.8678964594。需明确接受旧汇率及记录原币、来源、日期和新价格版本；另一选项是取得并冻结适用的 CFETS 官方新快照，再按同规则换算。是否采用哪项仍是未决价格政策，不应静默选择。

## 2026-11-02 官方 CFETS 冻结快照

已直接取得中国外汇交易中心 [2026-11-02 人民币汇率中间价原件](https://www.chinamoney.com.cn/r/cms/www/chinamoney/data/fx/ccpr-notice2026-11-02.json)：1 USD = 6.7146 CNY。UTC 2026-11-02T02:37:05.412214+00:00 请求，HTTP 200，原文件 1,377 字节，SHA-256 `1f112864a09d7ff8dc45ef64033f1aaaa115d9a9b33ecd2421726790395aef2f`。原件和响应元数据分别为 `inputs/cn/fx-2026-11-02.body` 和 `inputs/cn/fx-2026-11-02.json`。

对新型号单独固定此日期和汇率，按现有十位半偶舍入规则：Input ¥13 → USD 1.9360795878/M；Output ¥68 → USD 10.1271855360/M；CacheRead ¥1.30 → USD 0.1936079588/M；5m Write ¥13 → USD 1.9360795878/M。1h Write 依 ADR 已舍入 Input × 2 得 USD 3.8721591756/M。该中间价是固定换算估算，不是实时结算价；K3 原有价格和原汇率保持不变。
