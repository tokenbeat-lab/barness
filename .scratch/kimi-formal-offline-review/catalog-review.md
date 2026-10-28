# Kimi 正式目录离线审阅

2026-10-28。六个候选身份的字段和费用来源通过冻结资料核对；正式目录继续 inactive。交付仍有一个软件接线缺口：目录晋升入口没有消费固定9＋1组合，不能把六次补采成功直接等同于目录已可晋升。

## 审阅结论

| 项目 | 结论 | 依据及边界 |
| --- | --- | --- |
| 当前生产目标 | 保持不变 | `ai/catalog_kimi.json`仍为schema1、active=false、空version和models；内置目录不提供Kimi身份 |
| 候选身份 | 六个独立声明齐全 | moonshot-cn、moonshot各自Chat/Responses/Messages；型号均为kimi-k3，不继承兄弟站或另一协议的资格 |
| 正式数据版本和来源 | 既有约束一致 | Kimi子目录预定版本kimi-k3-formal-2026-10-22.1；候选、原币价格及来源索引合成的SnapshotHash与loader硬pin一致。未写入正式数据 |
| 完整官方原件 | 17份校验通过 | 压缩哈希、解压原字节哈希和长度均匹配source-files.json；仅使用仓库已留存的2026-10-22资料及2026-10-16汇率原件 |
| 上下文和输出 | 保持现有受支持子集 | 定价原表明确1,048,576上下文；Chat/Responses原文默认输出131072、可设置更大。本项目只支持131072，不推广百万输出，不把4096补项当容量实测 |
| 推理和模态 | 保持已有硬约束 | 仅text、reasoning=true、默认max；low/high/max，其他档位null；requiresThinking=true，不开启额外compat/采样能力。Messages“不承诺关闭推理”是本子集边界，不改写成厂商没有该能力 |
| 费用 | 冻结快照一致 | 两站分别保留原币，六个协议各有独立记录；中国站换算与十位半偶舍入一致，1h估算沿用rounded Input×2。实际账户费率仍待可信Host确认 |
| 旧真实证据 | 沿用已核验事实 | 六份旧九项/54项保持原stage/source/时间/预算；新xai-origin-history六项仍NOT_RUN |
| 当前晋升入口 | 软件阻塞 | 仍要求六个同包完整十项、同source/build/binary/commit，不能表达六组9＋1的十二份独立测量 |

完整逐身份字段、原币价格重算、17份来源哈希、读取文件与冻结输入见[机器审阅回执](../../.evidence/kimi-formal-offline-review/review-v1.json)。源码证据见[kimi_promotion.go第37行](../../ai/internal/testkit/supportmatrix/kimi_promotion.go#L37)、[第45行](../../ai/internal/testkit/supportmatrix/kimi_promotion.go#L45)、[单包读取入口](../../ai/internal/testkit/supportmatrix/regional.go#L55)。本轮为静态源码审阅，没有运行晋升CLI或新测试。

## 目录晋升的最小衔接

下一项实现限于[具体问题01](issues/01-kimi-promotion-composition.md)：让Kimi专用晋升入口消费现有 `VerifyKimiComposition` 和完整输入闭包，六个combo分别核验，每组保留两次测量及十项coverage。替换组合路径中“六路统一source/build/commit”的过宽条件，继续核对实际产品、SDK、所用型号及账户相容性。受控数据仍只能生成不可激活的审阅材料。

数据目标只涉及独立Kimi六身份；不修改其他地区九型号，不新增通用组合平台、续期或恢复控制器，不拼造一个十项Report，不向旧包补写新场景或预算。既有exclusive intent和原子目录替换边界可继续复用。

建议先补此接线并完成相应受控/普通检查，然后一次重新冻结受影响二进制、更新尚未批准的七份绑定。消费者来源变化不要求重采旧54项、MiniMax旧13项或其他Provider。七份v3在当前代码下仍与冻结一致；未来代码实际变化后才更新它们。

## 后续执行边界

账户和费率确认、后续独立付费批准齐备后，各Kimi路线只采xai-origin-history一项；先得到六份真实候选9＋1完整性，再以审阅过的相同型号事实完成正式数据晋升。原测量标签不改，无需额外购买六组full formal。

目录晋升改变聚合Builtin目录哈希，不改变已有Provider型号事实；应逐型号比对并对目录变更选择相应普通范围，不要求所有Provider重新付费验证。若型号能力、SDK、实际费率或账户发生变化，单列该组差异及适用范围，不自动扩大到所有路线。

正式目录、七项真实补采、人工race/压力和发布继续未完成。当前release.pass=false，旧暂停轮次保持暂停。
