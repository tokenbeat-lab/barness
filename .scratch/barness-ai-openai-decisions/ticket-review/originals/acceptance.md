# Decisions 分类验收矩阵与证据合同

Status: ready-for-agent
Updated: 2026-10-12

关联[spec R01–R20](spec.md)、[设计](design.md)、[实施计划](implementation-plan.md)。本文件是先于生产实现的失败目录；表中“必须结果”是期待，不是已经执行的 PASS。

## 1. 已确认的测试 Seam

用户已确认使用 Client.Classify / HookedClient.Classify。沿用 Host 替身、受控 Provider、Config.Transport、公共 Catalog/Lookup/NewClient/Hash 和实际 artifact 核验；不增加私有 parser/Adapter 测试入口。

新路线使用 P10；Jev 仍 P07。每个 case 展开 route、plain/hooked、错误变体，稳定 ID 为 P10-ODNN-<variant>-<entry>；Jev 对照沿用旧 ID 并增加独立明确的 regression case。Callback 用 public Hook 注入、响应错误用 HTTP body 脚本注入；不得从私有函数绕过 scope/authority。

本期 scope 形状检查继续先于 Binding resolver，Jev 原 Code/Phase 和优先级不变。资源场景只在自己给定的有限政策下断言；初始 input 错误和 Hook 后错误阶段分别验证。

## 2. 公共契约、目录与授权

| ID | 先失败的场景 | 必须观察的结果 | 需求 / 工单 |
|---|---|---|---|
| OD01 | 旧 Jev Catalog/Binding 不填写任何新字段 | 构造/Hash/Lookup/实际 Classify 成功；新增字段保持省略；text-only、单候选与宽说明不变 | R02/R05 / 01,02 |
| OD02 | 添加同 ID gpt-6-luna chat/classifier，另有跨租户同名 Binding | 完整身份独立；错误操作 tenant_denied/capability，零 credential/Attempt/request | R01/R14 / 02,03 |
| OD03 | 非白名单、未知目录型号、外国租户、disabled Binding、错 Provider/API | 原授权优先级、Code/Phase、归属不泄漏；零凭据/发送 | R01/R02 / 01,03 |
| OD04 | nil、空值、非 nil pointer、typed nil、错协议 Options | 合法默认相同；typed nil 与错协议 invalid_request/capability，零 Attempt，不 panic | R03 / 02,03 |
| OD05 | 未知/重复 Input 或 DescriptionFormats，非法 MinChoices/图片能力 | NewClient 与 Hash 一致拒绝；旧省略默认合法；不收紧旧字段已有可构造范围 | R05 / 02 |
| OD06 | 修改原 Catalog、Lookup/List/Client.Catalog 返回值的新 slices / tiers | 后续调用/查询不变；Hash 纳入所有新值，map 插入顺序不影响快照 | R05/R14 / 02 |
| OD07 | Classifier nil、Images nil；GenerateImages 策略启用/关闭作对照 | Classifier nil 禁全部分类；Images nil 只禁分类图片；两操作不能相互启用 | R06 / 02,05 |
| OD08 | AllowClassifierImages 默认 false、非 classifier Binding 为 true | 默认不授权图片；新字段错误操作 invalid_request/binding；旧 text 调用不受影响 | R07 / 02,05 |
| OD09 | 原公共无题、空名、nil question、坏 state、256 choice、11 score | 与旧 Jev 相同 scope 错误及拒绝顺序；两路线零 resolver/credential/Attempt/request | R02/R04 / 01,03 |
| OD10 | Jev 单候选成功，Luna 单候选；模型进一步收窄范围 | Jev 保持成功；Luna invalid_request/capability，零 credential/Attempt；模型范围不扩大 public 范围 | R04 / 03 |

OD09 的新增 resolver 次数断言固定原入口顺序，不接受通过“先解析再返回旧错误”伪装兼容。OD05 不把既有可构造但大于公共上限的 MaxChoices/score 元数据改成新 constructor 错误。

## 3. State、问题和 native 请求

| ID | 先失败的场景 | 必须观察的结果 | 需求 / 工单 |
|---|---|---|---|
| OD11 | string/object/array State，空值和大精度数 | string 原文；对象数组 compact text，数字保留；Jev 原 JSON 不变，不多加引号 | R02/R08 / 03 |
| OD12 | Jev structured instructions/criteria；同 req 换 Luna | Jev 继续合法；Luna capability 拒绝，不静默 stringify；choice null 说明省略可合法 | R04 / 03 |
| OD13 | choice、score、bool 混合、bool 无/有/空 string criteria | wire 分别 choice/score/predicate；true/false 拼接固定；公共 bool 概率和 ordered score 保持 | R08 / 03 |
| OD14 | string "true"/"false" 候选；numeric-looking 问题/候选 key | string 值不变成 bool；冻结数组次序确定，不依赖 map 插入；name 匹配不依赖 index | R08/R15 / 03,06 |
| OD15 | image-only 判断使用空对象 State；遗漏 State / 无图和有图 | 有图仍要求合法 State；无图 string，有图单 user / text-first / ordered images | R03/R08 / 03,05 |
| OD16 | 原 native TypeSafe 与 Decisions 捕获字段混入另一协议 | Jev 不出现 input/images/predicate；Decisions 不出现 state/noul；无 SDK retry/host/env header 泄漏 | R01/R02/R08 / 01,03 |

## 4. 图片、资源和最终回调

| ID | 先失败的场景 | 必须观察的结果 | 需求 / 工单 |
|---|---|---|---|
| OD17 | 真 PNG/JPEG/WebP 资产，两个有区别图片顺序 | public Classify 捕获顺序和 data URL 正确；asset hash 可独立验证；只做编码/header 校验 | R06/R08 / 05 |
| OD18 | 坏 MIME、错 magic、非法 base64、padding/CRLF、URL/path | 初始 invalid_request/scope；零凭据/Attempt/发送；不自动修补/下载/resize | R06 / 05 |
| OD19 | decoded 每图/总图/count 边界、base64 长度溢出和 JSON escaping | policy 超限 resource_limit/scope；完整 native JSON 超限 request；known size 在复制/解码前拒绝 | R06 / 05 |
| OD20 | 128/129 张，Host MaxImages 高于 API 作对照 | 128 的有界离线 case 可发；129 capability 拒绝；policy 更小则先 scope 拒绝 | R05/R06 / 05 |
| OD21 | Jev 收图片；Luna policy/model/Binding 三个授权条件分别缺失 | 不静默丢弃；政策 scope、型号 capability、Binding tenant_denied/capability；无 creds/Attempt | R06/R07 / 05 |
| OD22 | Hook 合法原地改题或 replace，改名/类型/顺序/候选/levels | 据最终集合解析；原 req 与 Hook map 后续同步修改不改变冻结字节/结果 | R09/R14 / 03,04 |
| OD23 | Hook 加 model/endpoint/auth/tools/operation/permission/external URL | tenant_denied/request，零 Attempt；普通未知字段 callback_failed/request；Jev 原规则不变 | R07/R09 / 03,05 |
| OD24 | Hook 无图变有图，原 Binding flag false/true，两次对照 | false 拒绝；true 仍须 policy/model/格式/大小均合法；不能用新 body 授权自己 | R07/R09 / 05 |
| OD25 | Hook 多 message、非 user、坏 type、boolean choice、detail/safety_identifier | callback_failed/request；外部引用 tenant_denied；零发送，无隐式扩宽原生子集 | R08/R09 / 03,05 |
| OD26 | Hook 深层/后位图片超限、累计图字节、巨大 name/question、custom replacement | 对最终整个 JSON 和每题/每图有界检查；不只检查入口或第一图；错误后资源归零 | R06/R09 / 05,06 |

OD17 是有效资产测试，不能使用仅声明 MIME 的 arbitrary base64 冒充图片。OD19 同时保存 public 下界、encoded body 和 decoded image 读数，不能只检查其中一种大小。

## 5. 答案、拒答、分布及用量

| ID | 先失败的场景 | 必须观察的结果 | 需求 / 工单 |
|---|---|---|---|
| OD27 | 正常 answers 乱序、混合题、重复 score label | 按 name 回配；全部验证后发布；score indices/labels/distribution 保留 | R10/R11 / 04 |
| OD28 | 单题 refusal、混合 refusal、全 refusal | 完整 Answers 名称集；typed refusal 无编造原因；stop/nil error，一 Attempt 不重试 | R10 / 04 |
| OD29 | TypeSafe native refusal、Decisions unknown discriminator / wrong kind | protocol/response；Jev 不放宽；清空全部 Answers，明确 usage 保留 | R02/R10/R11 / 04 |
| OD30 | 缺/多/重复 answer name，null/no name，重复/case-alias JSON 字段 | protocol/response；无 last-wins，正常部分也不发布 | R11 / 04 |
| OD31 | choice 未声明/boolean value，概率缺多重复，选项概率非最大；合法并列 | 非法整体拒绝；字符串集合精确；最高概率并列合法；confidence 不重算为 chosen probability | R11 / 04 |
| OD32 | probability/confidence/score null、NaN/overflow、越界、非 1 概率和 | finite/range/sum 严格，1e-6 边界有合法和非法对照；无归一化 | R11 / 04 |
| OD33 | score index 非整数/重复/缺多、label 不符、expectation 不符 | 拒绝；完整 0..n-1 和最终 rubric；保持原数值 | R11 / 04 |
| OD34 | 真实 Jev 百分位舍入 replay；同数值发 Decisions；高精度反例 | Jev 原有 bounded exception 保持；Decisions 不套该容差，无无界新例外 | R02/R11 / 01,04 |
| OD35 | usage 缺/null、部分、明确全零；无 model 或 response alias | unreported/partial/complete 分清；Reasoning unset/0 分清；授权 ModelID 不替换 | R12 / 04 |
| OD36 | duplicate usage/token、负数/小数/overflow、缓存/总数/reasoning 矛盾 | protocol/response，不伪造 complete；无选择最后值；合法独立 usage 在答案错误时保留 | R11/R12 / 04 |
| OD37 | full input 含非零 cache details，非零免费 output / reasoning | Input 不扣缓存，canonical cache=0；total/reasoning 保合法 presence；成本只按 Input | R12 / 04 |
| OD38 | input=271999/272000/272001 和输出/缓存组合 | 阶梯严格大于阈值；全请求 .1/.2；input 不重复计，cache/output Cost=0 | R12 / 04 |
| OD39 | 额外 vendor metadata 与已知字段大小写别名/非法语义字段 | benign additive 字段有界忽略；别名/重复/改变答案含义拒绝；Jev 严格未知字段不变 | R02/R11 / 04 |

已计费结果与合法 refusal 是不同维度，不能以 refusal 给零 usage。源码例子中的缓存统计是合成协议证据，不能据此宣称核过真实账单。

## 6. 生命周期、租户和故障

| ID | 先失败的场景 | 必须观察的结果 | 需求 / 工单 |
|---|---|---|---|
| OD40 | 429/503 后成功、504、x-should-retry/Retry-After、延迟超预算 | Binding 0/显式 retry 规则不漂移；504 与 Pi 的差异独立记录；有限 Attempt/等待 | R13/R15 / 06 |
| OD41 | 2xx 后 malformed/refusal/read failure/response hook error | 不重试；合法 refusal 成功；其余错误阶段与唯一 usage 保留；body/permit 释放 | R10/R13 / 04,06 |
| OD42 | connect/header/read-idle/call timeout；resolver/header/payload/image/read/backoff 取消 | 唯一终态，有限用量、正确 attribution/Attempt、body/permit/waiter 归零 | R13 / 05,06 |
| OD43 | 成功/错误 body 过大、深度超限、全 question/probability 大响应 | MaxOutputBytes/MaxErrorBodyBytes 有效；不套 SSE Frame limit，不重复尝试限额故障 | R11/R13 / 06 |
| OD44 | 共享 Client 的 tenant A/B，同名 Binding/Model；Jev/Luna/chat/image 混合 | 所有捕获、结果、Attempt、Observer 归属独立；无 global current tenant，race 无发现 | R14 / 06 |
| OD45 | 首次故障后 Host 修改 Binding/Credential/Catalog；回调保留 mutable reference | 重试同 bytes/headers/key/version/image permission；无重读；hook 次数固定 | R09/R13/R14 / 06 |
| OD46 | Observer 拥塞/异常、secret/error body/图片正文 | 只含 metadata，结果不变，有界队列；证据和日志审计无发现 | R19 / 06 |
| OD47 | classifier 图片政策设计负载占满，随后取消和恢复 | 实际内存/并发/限额余量/资源释放有读数；不通过降低负载得到 PASS | R06/R13/R19 / 06 |

并发输入各自独立；在明确 barrier 后修改已复制数据，不并发写构造器正在读取的 map。压力目标使用现有 MixedPolicy 的真实数值依据加新图片负载，在 source 中记录；要求<=75%限额占用及有界 peak heap，若数值不适用先报告，不填通用默认。

## 7. 差分、真实来源、宿主效果与发布

| ID | 必须实际证据 | 完成规则 | 需求 / 工单 |
|---|---|---|---|
| OD48 | Pi 1.1.0 runner Decisions classify/catalog 新路线 | version/commit/hash 仍固定、外网/key 不泄漏、完整共同请求/基本答案比较 | R15 / 06 |
| OD49 | refusal/score detail/usage/JSON precision/strictness/retry/format/order 扩展 | 每条理由与有限 case/path 自己通过；无全 body/result 豁免，零 pending | R15 / 06 |
| OD50 | 原 Jev 全部 E2E、diff、官方 wire replay；新旧 catalog 对照 | 原断言不删、不改成 UNSUPPORTED；新 Capability metadata 有明确新增字段投影 | R02/R15 / 01,06 |
| OD51 | first-party source fixture，candidate 自带目录/自己 live | 实际 route/model/format/rubric/usage 有 captures 和日期/hash；922K/服务 score max 未测不冒称 | R16 / 07 |
| OD52 | 正式 Luna classifier 目录版本/Hash 和旧 Jev 内容 | 候选验收后才纳入，旧省略态保存；同 ID chat 价格不串；没有造日期型号 | R05/R16 / 07 |
| OD53 | 第十组合完整报告，错身份/旧 Hash/缺项/预算超额/跨行合并反例 | merge 原子拒绝非法报告；自己正式 hash 的 live；旧九行历史与 required set 保留 | R17 / 07,09 |
| OD54 | 无 key / 未执行、自然 429/504/厂商失败 | NOT_RUN 或 FAIL；拒答可完整有效，不为制造拒答循环；不沿用其他组合 PASS | R10/R17 / 07 |
| OD55 | 同固定中文数据集 Luna 独立 route/config/report，Jev 老报告核验 | 拒答纳入分母/状态，准确率/条件准确率分开，校准/延迟/消耗独立；旧 verifier 保持 | R18 / 08 |
| OD56 | Host 旧 Options/结构化说明/普通答案断言迁移示例 | 共同子集成功，不兼容输入正确拒绝；typed refusal 必有分支；无业务 fallback 藏库内 | R03/R04/R10/R18 / 08 |
| OD57 | 完整 release gate、十组合/效果适用性、正常/race/diff/压力/vet/audit | 固定合同不缩小，全部证据当前且哈希匹配，所有失败和未执行诚实保留 | R17/R19 / 09 |
| OD58 | 停用新 Binding/AllowClassifierImages/catalog authorization，活动调用对照 | 新调用拒绝、已 pin 调用按固定 snapshot 结束；Jev/聊天不自动重路由，已有费用保留 | R07/R20 / 09 |

OD51 的 candidate 与 OD53 正式目录不是同一次证明。价格 Hash 变更后既有报告是否仍适用按 gate 明确规则核验，不能只替换历史报告 hash。没有真实凭据时工单保留 NOT_RUN，交付结论未完成。

## 8. artifacts、执行与审计

每次 case 写出：fixture/配置、请求/响应脚本与 raw capture、result/error/metadata、Attempt、Observer、assertions、图片文件及 SHA-256、差分记录（适用时）、versions、limits/heap/admission 读数、完整文件 manifest、audit 和重放命令。成功与失败都保留，同名执行不得覆盖旧失败。

无密钥/正文进 Observer，测试正文只限已声明合成数据或固定测试资产。真实 account/region 只记 alias，header 白名单沿用；注册实际 key 后来源脱敏和整包审计，发现秘密则 FAIL，不复述它。

研发先运行新增目标公共 E2E 与 Jev 对照；只有新增失败/改动才重复扩测。阶段六完成后运行固定差分、mixed/race/pressure。最终沿用实际仓库命令，不要求尚未实现的参数：

~~~sh
go test ./ai/e2e -count=1
BARNESS_AI_PIDIFF=1 go test ./ai/e2e -count=1
go test -race ./...
go vet ./...
go vet -tags live ./...
~~~

真实执行由各工单记录准确命令和独立 key 环境变量；命令/artifacts 只含变量名或引用，不写 secret 值。发布继续使用已有 releasegate Interface，若为十组合/两个效果报告新增参数，在工单 09 明确实装并记录。NOT_RUN、FAIL、UNSUPPORTED、PASS 仅由真实执行状态生成；必交能力不能以 UNSUPPORTED 放行。
