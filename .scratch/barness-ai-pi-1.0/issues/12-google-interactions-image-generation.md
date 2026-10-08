# 12: 通过 Google Interactions 完成图像生成

**What to build:** 宿主通过 GenerateImages 使用 Google Interactions 生成图片，选择型号允许的宽高比和图像尺寸，取得全部有序输出、真实终态和正确的模态用量；调用固定为同步、无状态、内联交付。

**Blocked by:** 10 — 完成 OpenAI 内联 JSON 编辑与 mask。

**Status:** resolved

**依据：** 规格“图像：Google Gemini Interactions”“用量、计价与观测”、实施设计第 9、12–13 节，以及 ADR-0012/0014/0018。实现顺序沿用已确认的 OpenAI 图像先于 Google；11 的真实冒烟可独立执行。

- [x] 先用既有 generateImages/JSON 场景能力写 P09 生成、选项错误、状态错误、交错输出、坏图、用量和回调越权 fixture，再实现该路线。
- [x] 绑定为 Google × google-interactions × image，经唯一 HTTP 客户端向固定 v1beta interactions 发请求，型号使用裸 ID，认证只放 x-goog-api-key；不自动回退 v1、不改动现有 generateContent 聊天路线。
- [x] GoogleImagesOptions 只提供宽高比与图像尺寸，按型号能力校验；nil 为默认，拒绝其他协议选项。先交付无参考图的完整生成，参考图路径由 13 交付。
- [x] 请求固定 store=false、同步执行和图像 inline delivery；回调后再次核对固定项、授权型号、允许字段及最终请求大小。续接、后台执行、Agent、工具、环境、webhook、续读令牌、服务等级和 stream 不能引入，网址输入输出不能扩大权限。
- [x] 沿用图像及全局资源策略、尝试准入、取消、超时、可信回调和不重放规则；成功 JSON 按总输出限额读取，读取、解析和完整输出校验之前不释放许可。
- [x] 只接受 completed，遍历所有 model_output 步骤的全部内容，保留 text/image 出现顺序并整体校验图片；不使用只保留最后一张图的便利属性，不通过并发请求凑数量，也不承诺精确图片张数。
- [x] failed 与 cancelled 映射 upstream_error/response，in_progress、requires_action 或带 continuation_token 映射 protocol/response；只返回文本或没有有效图片为 protocol，有安全阻断诊断为 upstream_error。图片只有 URI 时失败，不去下载。
- [x] 响应 id 只作诊断 ResponseID，实际响应型号仅在报告时填写，不成为原生续接状态，不覆盖授权 ModelID；失败输出整体为空，已报告用量和调用元数据保留。
- [x] Usage.Input 取总输入，Output 取总输出加思考，Reasoning 为已计入 Output 的思考量，TotalTokens 优先取厂商报告；输入 7、输出 20、思考 22、总量 49 的 fixture 应得到 Input=7、Output=42、Reasoning=22、TotalTokens=49。
- [x] 按文本/图片映射输入输出模态明细；四个总量与两组明细齐全为 complete，缺项为 partial，缺 usage 为 unreported。思考按文本输出费率计价，采用带来源 fixture 的明确包含规则，模态明细已含思考时不能重复计价；真实形状由 14 确认。
- [x] 原生 Google 图像登记为扩展路线，全部 P09 fixture 明确跳过 pi 差分，并由离线用例断言登记；增加对应 P0、追溯、目录包含豁免与观测模态审计，未冒烟不宣称支持。
- [x] 修订 unary 协议 ADR 的 Interactions 取舍、v1beta 与无状态理由，以及 ADR-0012/0014、用量和契约说明；Gemini 聊天仍不调用响应回调。完成聊天、分类、OpenAI 图像回归，交付脱敏离线证据。



## Answer

2026-10-08：已完成并关闭。宿主目录可通过 GenerateImages 使用 Google ×
google-interactions × image，固定 v1beta、裸型号 ID、API-key 请求头、同步无状态
inline 交付；入口和最终回调均验证能力、固定项、允许字段与预算。响应逐步骤/内容块
有序解析并整体发布，失败保留合法用量、诊断身份和调用元数据；模态计价避免重复思考。
参考图仍归工单 13，真实型号/live 与模态形状确认归工单 14。

实现提交 `7bfd5d72dcb774614ffecb0302aa6a041183d6ab` 已通过：
`go vet ./...`、`go vet -tags live ./...`、Google 定向 E2E，以及开启冻结 pi 差分和
压力场景的全量 `go test -race ./... -count=1`。4,126 个证据场景全部 PASS，Google
119 个、OpenAI 图像 267 个、pi 差分 623 个（pending 0）、压力 11 个；Google 20 个
资源场景全部 gauge 归零。原始 run 与交付目录脱敏审计均 0 发现，源码/逐场景/交付
哈希校验通过；四个新增追溯项及 P0/差分/目录快照均 PASS。

Standards 和 Spec 各发现 1 项 P2，先补红灯 E2E，再修复并分别复核，均无遗留项。
ADR-0010/0012/0014/0018/0022、契约、扩展 ledger、P09/P0/trace 与领域词汇已同步。
[可重放离线证据](../google-interactions-evidence/README.md) 包含版本、fixture/输出/源码
哈希、实际脱敏请求与结果、资源读数、Observer 白名单审计和双轴审查。
