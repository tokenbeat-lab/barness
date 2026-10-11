# 03：增加模型分档采样元数据、校验与隔离快照

Status: resolved
Type: task
Depends on: 02
Updated: 2026-10-11

## What to build

在 Chat Model 增加可选 `SamplingParamsByThinkingLevel`；完整接入 Catalog 校验、各公开返回边界的深拷贝、Hash、内置数据和版本快照。调用级、图像和分类类型不扩张。Host 继续负责目录组装。

## Ownership and context

负责 `ai/catalog_models.go`、`ai/catalog_validate.go`、`ai/catalog_snapshot.go`、内置目录数据及正式目录快照；E2E 负责新增元数据/发现/快照场景与字段差分。使用已核验的 1.1.0 生成数据，不新增依赖或 modelOverrides 加载器。协作时不要回退他人修改。

依据：[规范 R01、R06、R08–R09](../spec.md)、[设计 §2/4/5/6](../design.md)、[验收 S18–S25](../acceptance.md#3-配置目录与快照失败场景)。

## Failure scenarios first

- 错 canonical key、inactive 档非法 JSON、nil 档对象被忽略；Client 和 Hash 使用不同规则。
- inactive profile 可以写 model、tools 或其他既有保留键，绕过协议授权。
- 外层复制但内层或 RawMessage bytes 共用；公开 Lookup/List/Client.Catalog 返回值能污染后续调用。
- 新字段不进入 Hash，map 插入顺序使 Hash 漂移；目录内容变化却不升 Version 或更新快照。
- 全部原生协议套用 Chat 保留键，或把生成数据缺失误读为没有采样默认。

## Acceptance criteria

- [x] 先写 S18–S25 公共 Client/Catalog 失败与对照场景，再实现字段及边界规则。
- [x] 七个 canonical key，outer 省略有效、空档有效、nil/null 档无效、参数 null 有效；不扩大 SimpleOptions.Reasoning。
- [x] 新 map 在所有聊天型号全档验证结构/JSON；消费 API 全档复用保留键，原生忽略 API 不套用其他协议的保留键。
- [x] 各输入和公开输出边界深拷贝 outer/inner/bytes；构造后及返回后受控修改不改变 Client。
- [x] Hash 插入顺序稳定，任何值变化改变内容 Hash；共享构造验证，错误不含原始值。
- [x] 已有型号的两种采样默认均与已核验来源比较；保留既有硬约束、非聊天数据和独立扩展来源，不自动新增型号。
- [x] 实际内置目录变化升级 Version，同步正式 Catalog/价格快照和受影响 pins；没有数据变化时不伪造默认或目标数量。

## Verification and handoff

保存每个构造失败与成功对照的 fixture、错误类别、Hash 断言、快照修改证据和来源字段比较。受控 Provider 可作为后续请求的观察点；本工单尚不要求新 profile 改变请求，相关行为由工单 04 完成。

先运行新元数据/快照 E2E，再运行既有 Catalog/discovery/operation 隔离场景，开启目标目录字段差分。未实现请求映射仍是已登记 pending，不报 B03 完成。

## Comments

2026-10-11：方案阶段创建，未修改当前 Model 类型或目录版本。

2026-10-11 实施：公共元数据、全档校验、三层深拷贝和确定性 Hash 已实现，公开构造/Hash/查询/真实请求 E2E PASS。目录 2026-10-11.1 五项价格变化来自核验数据，没有合成内置默认。 命令、实际 artifacts、失败历史和未完成项见[专项证据](../evidence.md)与[来源核验](../release-verification/README.md)。
