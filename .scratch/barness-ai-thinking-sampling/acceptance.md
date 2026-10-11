# barness-ai thinking level 采样映射：验收与证据合同

Status: ready-for-agent
Updated: 2026-10-11

关联[规范](spec.md)与[详细设计](design.md)。本矩阵是实现前制定的失败场景与验收合同，当前实际结果见[evidence.md](evidence.md)。先写场景，再写生产实现；用公共入口观察最终载荷，不用私有 helper 测试代替。

## 1. 固定测试边界与路线

用户已确认公共 Client E2E：Host 替身提供租户/Binding/凭据，受控 Provider 捕获请求并返回固定响应脚本；测试使用公开 Catalog、Client 与 WithHooks。Stream 消费到唯一终态，Complete 检查返回结果和 Attempt；两者都核对捕获载荷及资源释放。

| 路线 | 采样能力 | pi 差分 |
|---|---|---|
| OpenAI × Responses | 消费通用与分档 map | 固定 1.1.0 原生路线 |
| OpenAI × Chat Completions | 消费通用与分档 map | 既有兼容路线及有限 extension 边界 |
| DeepSeek × Chat Completions | 消费 map，保留原生 thinking 开关 | 固定 1.1.0 原生路线 |
| DeepSeek × Responses | 消费 map，保留既有 Responses 能力 | barness 扩展，自身断言；显式跳过原生差分 |
| Anthropic × Messages | 忽略采样 map | 验证原生载荷未改变；既有差分照常 |
| Google × Gemini | 忽略采样 map | 验证原生载荷未改变；既有差分照常 |

核心 S01、S07、S08、S09、S11 在四条消费路线分别覆盖 Stream、Complete、StreamSimple、CompleteSimple，并各有 WithHooks 无修改版本。其他场景按表中适用入口执行：共享采样边界至少覆盖两个 API，每条 Provider 特有分支独立覆盖；full/simple 差异不能以另一入口的结果替代。WithHooks 的行为和授权由 S26–S28 加强。使用合成模型与数值，不依赖内置目录恰好有某档默认。

## 2. 功能失败场景

| ID | 先暴露的失败 / 输入 | 必须观察的公共结果 | 需求 |
|---|---|---|---|
| S01 | general、high profile、call 同 key，具名 temperature 也赋值 | 具名 → general → high → call；call 最高；未覆盖 key 保留 | R04 |
| S02 | low/medium=null、high 支持，请求 low | sampler 用 high；按 full/simple 各自核对 wire | R02–R03 |
| S03 | 请求档位不可用，所有更高档也不可用 | 按序向下选择最近支持档；有更高档时不能先选更低档 | R02 |
| S04 | minimal→low，另给不同 minimal/low profile；两档映射同 wire 的变体 | 用 canonical minimal profile，wire 可为 low；不按 wire 反查 | R02 |
| S05 | xhigh/max profile 存在，level map 未显式支持；显式支持的对照 | 未支持时回落；显式非 null 时选择自身；profile 不声明能力 | R02 |
| S06 | off→none；off=null 且最低可用档变为 low/minimal；全部档 null | 依次选 off / 最低支持档 / off 兜底；不把 null 当 none | R02 |
| S07 | 省略 reasoning，off 支持及 off=null；具名选项仍存在 | 从 off 选档；off=null 可选最低支持档，但不额外写 effort | R03 |
| S08 | 非推理模型收到 high，off/high 有不同默认 | 只消费 off，不因为分档配置开启 thinking | R02–R03 |
| S09 | 无 profile、选中档缺失、选中档为空对象三个变体 | 保留 general/call；无全部 map 时与旧请求相同；空档不清空 general | R01、R04 |
| S10 | general 嵌套对象/数组，被 profile/call 覆盖 | 顶层整体替换，不递归补回旧嵌套字段；原配置不变 | R04 |
| S11 | call/profile 明确给 0、false、null，另有同 key 默认 | 值存在且正确；null 不删除字段；空 bytes 另由 S19 拒绝 | R04 |
| S12 | full low 但 low/medium=null；simple low 同配置 | 两者采样 high；full wire 原始 low，simple wire high 映射值 | R03 |
| S13 | Responses full 只有 summary；medium 支持与 medium=null 两种 | 采样从 medium 起算，必要时夹取；wire effort 仍 medium | R03 |
| S14 | DeepSeek Chat 省略 reasoning，off 支持与 off=null | off 支持发 disabled 并选 off；off=null 不发 disabled，采样向上 | R03、R05 |
| S15 | DeepSeek Chat 显式 reasoning 与 canonical→wire 映射 | 原 thinking enabled/effort 规则保留；profile 使用 canonical 档 | R02–R03 |
| S16 | call map 覆写 reasoning/reasoning_effort 或 DeepSeek thinking | 使用覆写前求得的 profile；最终 JSON 保留覆写，不二次选档 | R04、R07 |
| S17 | native Anthropic/Gemini 带合法新 map，含本协议未建模的键 | 新旧 map 均不泄漏到载荷；原 Managed Effort/budget 行为保持 | R05 |

S02–S06 的能力规则以完整 canonical 顺序生成表驱动 fixture，包含普通档 null、xhigh/max null、向上优先、向下回落和全部 null；不得只检查研究里三条 simple probe。S12/S13 是有意保留的兼容行为，不以“wire 与 sampler 必须相等”为断言。

## 3. 配置、目录与快照失败场景

| ID | 先暴露的失败 / 输入 | 必须观察的公共结果 | 需求 |
|---|---|---|---|
| S18 | 新 map key 是 none、MINIMAL、未知或空字符串 | NewClient 和 Catalog.Hash 拒绝同一错误位置；零网络 | R01、R06 |
| S19 | 未激活档含非法 JSON、空 RawMessage、nil 档对象；有效空档及值 null 对照 | 非法配置构造/Hash 失败；合法对照成功；不能等选中才校验 | R06 |
| S20 | 每个消费 API 的每个既有保留键位于任意 inactive profile | 两个 Provider 都拒绝，错误不回显值；覆盖完整保留键列表 | R06–R07 |
| S21 | Client 构造完成后改原 outer/inner map 及 bytes | 后续公开调用仍用原快照；每一层独立；Hash/目录内容一致 | R08 |
| S22 | 修改 Catalog 查询、枚举与 Client 返回目录的 profile/bytes | 原 Catalog 与 Client 后续查询/调用不变；各公开出口独立 | R08 |
| S23 | 两个 map 内容一致但插入顺序不同；只改一档一个值 | 前者 Hash 相同，后者 Hash 不同；Hash 与构造规则一致 | R06、R09 |
| S24 | 核验后的 target 数据包含或不包含模型默认 | 所有既有型号逐字段比较通用/分档 map；无值不补推荐值 | R09–R10 |
| S25 | Anthropic/Gemini 的新 map 有坏 key/JSON；另有合法保留键 | 坏结构拒绝；合法 map 保存且忽略，不套用 Chat/Responses 保留键 | R05–R06 |

S21/S22 在同步后进行调用方修改，禁止并发修改正在被构造器读取的同一个 map。需要覆盖外层、内层和 RawMessage 引用，不能只修改一个数字证明“深拷贝”。S24 不允许只检查新字段存在，必须检查已核验数据的真实内容，并保留既有模型字段与硬约束验收。

## 4. 生命周期、授权与并发失败场景

| ID | 先暴露的失败 / 输入 | 必须观察的公共结果 | 需求 |
|---|---|---|---|
| S26 | OnPayload 读取采样结果后原地改值或替换整个载荷；覆写 reasoning | 回调见三层合并结果，修改最后生效；无 reselection；四个 Hooked 入口 | R04、R07 |
| S27 | 回调在合并后尝试换 model/引入未授权托管资源/超过既有请求限额 | 原最终载荷授权与限额拒绝，零发送；错误、Observer、证据不泄密 | R07 |
| S28 | 首次可重试失败；之后修改外部目录；回调带计数 | 按既有有界策略重试，相同请求字节/profile；回调一次，Attempt 独立 | R08 |
| S29 | 共享 Client/型号的租户使用不同 call override；独立 Client 同 ID 不同 profile 作对照 | 请求各归所属租户；call override/回调不污染目录或其他调用；独立目录不串扰；race 无发现 | R08 |
| S30 | 请求被拒绝、发送前取消、重试等待取消及既有发送失败 | 唯一终态/正确 Attempt 与有限次数、无泄漏，资源归零；不重新解析 profile | R07–R08 |

测试复用既有合法 retry、取消和资源限制场景，增加能证明采样状态不漂移的输入；不为每个 helper 重建一套生命周期测试。S29 使用同步屏障和独立调用输入，避免时间 sleep 或隐式共享“当前租户”。

## 5. 基线与发布验收

| ID | 失败场景与完成条件 | 工单 / 需求 |
|---|---|---|
| B01 | official 1.1.0 commit/npm/integrity/依赖/模型数据有来源；固定源码填入发布数据后两环境 dist 逐文件等价；差异清点完整；不得线上重生成 | 01 / R10 |
| B02 | 单一 oracle lock/provenance/数据哈希与运行 guard 一致；版本/commit/hash 不匹配拒绝；运行无凭据/外网；无旧 1.0 运行副本或 guard | 02 / R10 |
| B03 | 全部原有差分子集及新采样场景均运行；固定比较规则不删字段；零 pending；extension 原因与 case 有限；DeepSeek Responses 独立 skip+自身证据 | 02、05 / R05、R10 |
| B04 | 文档/目录/价格/追溯/固定合同同步；完整 release gate 含九组合匹配当前 hash 的 live、race/差分/压力和独立效果证据；审计零发现 | 06 / R09、R11–R12 |

工单 02 完成施工阶段的 oracle 更新不等于 B03 最终通过；新字段/映射未实现时 pending 必须关联工单 03/04/05。只有 B01–B04 全部满足才可将该版本标为正式交付。没有凭据或其他外部条件时，专项离线进展可以记录，完整发布不得填 PASS。

## 6. 可重复 artifacts 与结果判定

每个新增逻辑 case 有稳定 ID，并展开路线、入口、Hook 变体以避免同名证据覆盖。沿用现有 evidence 包：fixture、模型/选项配置、捕获请求及脚本/帧、结果与 Attempt、Observer、断言、pidiff（适用时）、每文件 SHA-256 与审计。选档期待值可在断言中记录，仅为合成测试信息，不新增生产 Observer 字段。

case verdict 必须来自执行：PASS/FAIL/NOT_RUN 不由需求映射或研究结论填充；原生差分豁免另有明确原因与自身断言，不能以 skip 得到 PASS。差分记录包含固定 pi 版本/commit/model-data hash、SDK、正式目录 hash 和请求/帧 hash。不得仅比较 temperature 或采样子对象而删除其他 payload 字段。

建议复用现有运行命令，实施时由实际 case 名替换筛选条件并记录完整命令：

```sh
go test ./ai/e2e/...
BARNESS_AI_PIDIFF=1 go test ./...
go test -race ./...
go vet ./...
go vet -tags live ./...
```

当前公共 E2E 位于 ai/e2e。最终发布运行现有 releasegate 命令并提供其要求的全部 live/evaluation 输入；门禁会打开差分与压力，不以上述日常 race 命令替代完整 gate。凭据不写入命令记录或 artifacts。

本次实施执行目标测试及完整发布门禁；未通过的外部真实能力保留 FAIL，不能用历史或合成结果填 PASS。
