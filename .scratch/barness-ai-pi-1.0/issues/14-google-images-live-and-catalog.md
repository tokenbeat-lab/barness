# 14: 完成 Google 真实冒烟与目录纳入

**What to build:** 维护者用 Google 图像组合自己的真实证据确认首批型号、v1beta 请求与输出配置、终态和 usage 形状，再将已验证能力、模态费率和支持状态纳入目录快照。

**Blocked by:** 13 — 完成 Google 参考图编辑与请求守卫。

**Status:** resolved

**依据：** 规格“图像：Google Gemini Interactions”“支持矩阵”“追溯与真实冒烟”、实施设计第 9.4–9.5、12、16 节，以及 ADR-0012/0016/0017/0018。

- [x] 先列型号更替、版本错误、固定项失效、usage 误算、超预算、缺证据和错误矩阵合并的失败方式，再扩展本组合冒烟；只经公共 Client 调用，复用独立进程、双开关和本组合低权限凭据。
- [x] 固定真实 v1beta 目标、裸型号 ID 和首批 gemini-nano-banana-2.1；真实冒烟覆盖一张生成及带一张参考图编辑，记录 store=false、维护者批准的 delivery 省略、输出配置与请求认证所在位置；响应只接受内联图片。
- [x] 实施时重新核对官方型号状态、参考图能力、宽高比、尺寸、输入 MIME 和价格；保存官方来源及抓取日期。不列入规格排除的弃用或已关闭型号，不自动换版本、型号或协议来取得通过。
- [x] 图像预算按调用、张数与分辨率独立执行，首期本组合不超过 4 张、最大冒烟分辨率 1K；检查可读图片、格式、尺寸、数量约束和协议终态，生成可重复校验产物，不作像素相等断言。
- [x] 真实响应遍历全部 model_output 内容，保存经过有界捕获与脱敏的形状证据；确认响应 ID 和型号的实际存在性、完成状态、内联返回及思考 token 的报告位置，URI 或续读不能当作同步成功。
- [x] 明确证明 output_tokens_by_modality 是否已含 thought：根据本组合证据固定 fixture 和计价规则，使 Output 含思考、Reasoning 为其中分项、TotalTokens 保留厂商总量，费用不漏算也不重复计算思考；同时验证 partial/unreported 语义。
- [x] 首批图像费率按文本/图片输入输出及实际适用的缓存读取分模态记录，带官方来源与抓取日期，分辨率/token 关系只作能力说明，不叠加按张价格。
- [x] 报告和支持矩阵行明确为 image/Google/google-interactions，只由本组合报告合并；全部预期必交能力完整覆盖，错操作、错协议、旧报告或配置失败不会部分改写矩阵。
- [x] 缺凭据或未执行为 NOT_RUN，真实失败为 FAIL，不以长期接口占位、替代型号或 UNSUPPORTED 代替必须交付的生成/编辑能力；通过后才完成本工单并纳入支持声明。
- [x] 目录升版本并重建能力、价格和内容哈希快照，修订协议 ADR、ADR-0012/0016/0017/0018、契约和追溯；交付真实报告、脱敏审计零发现、矩阵合并及已确认形状的离线回归证据。


## Comments

2026-10-08：完成。Google 真实生成及一参考图编辑均 PASS，固定官方 v1beta/interactions、
gemini-nano-banana-2.1、store=false、x-goog-api-key，最终进程 2 调用/2 张 1K JPEG/0 重试。

初次显式 delivery=inline 被真实 API 拒绝。维护者明确回复“允许省略 delivery，修订规格后继续真实验证”，
已同步规格、设计和 ADR-0023；字段固定省略，回调添加任意 delivery 均拒绝，URI/续读输出仍 FAIL。
两个早期真实 FAIL 报告和无凭据 NOT_RUN 也保留，不替换为成功声明。

真实响应没有 ID/请求 ID 响应头，仅有型号；原样记录缺失。图片模态明细仅 image=1120，
文本 thought 单独报告并只计价一次，缺失分项保持 partial；已知分项成本合计 0.075354 USD，
不代表完整账单。内置能力仅一参考/一输出、1K、1:1，价格来源与抓取日期已保存，目录升至 2026-10-08.5。

完整 pi 差分/pressure/race 回归 4,308 案例 PASS（Google 300、差分 623、pressure 11），
188 份资源记录全部归零；live harness race 39 案例 PASS；两种 go vet 通过。完整运行和交付脱敏审计零发现。
双轴审查发现 1 项 Standards P2 脱敏失败路径，已先 RED 后 GREEN 修复，复核遗留 0；Spec 遗留 0。

详见 [交付证据](../google-images-live-evidence/README.md)、[审查](../google-images-live-evidence/review.md)
和 [验证记录](../google-images-live-evidence/verification.json)。本工单只证明自身 Google 图像组合；
九组合完整发布门禁属于工单 17。凭据权限等级和实际区域不能由 API 响应证明，CI 由宿主提供准确账户别名。
