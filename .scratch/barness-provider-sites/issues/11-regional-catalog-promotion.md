# 11 地区候选的受控晋升入口与两阶段目录

Status: claimed
Type: task

## Context

[规格 R03 / R04 / R16 / R18](../spec.md)、[07 最终验收](07-final-acceptance.md)。本工单预先实现并审查晋升机制；激活数据必须等待 08 修复后的六条真实候选全部通过自身资格。生产目录当前不含九身份，不能在资格之前注册。

只读变更图：根 `.evidence/provider-sites/final-acceptance-20261020/plan/promotion-map.md`。正式模型展示名也应在新 candidate 前固定，避免把内部 `own candidate` 状态带进 Builtin，或正式采后再次改变全局 Catalog Hash。

## Required behavior

为九身份增加一个初始明确 inactive 的独立数据载体，以及最小生产 loader hook；生产不能依赖 testkit 或运行时读取 `.evidence`。inactive 时保持原 Builtin 内容与版本。激活时只接受本服务九份已固定的独立 Model，给完整 BuiltinCatalog 新版本/Hash；不改旧厂商数据或现有 MiniMax loader，不做兼容 shim。

公共晋升入口消费六份自身真实 candidate qualification，严格核对全部九身份、source/catalog/build/authorization/seal 与自身原件；任一路失败、缺失、controlled/replay/兄弟站点或错来源均拒绝并保持生产数据不变。通过后生成独立不可覆盖 lineage receipt，再仅激活数据文件；实际成功只在获准真实 candidate 后执行。离线 controlled 流程只可产生明确的非生产提案，不能作为激活许可。

候选展示名、候选共同版本、snapshot及pin在付费前固定为稳定正式名称；不刷新官方原件。现有元数据、价格和协议能力不因命名变化扩大。候选 status/ownLive 作为固定资格输入快照保持，实际支持由 qualification/activation/formal receipts 和支持矩阵表达。

公共普通 E2E 应兼容 inactive/active 两态：九身份应全无或全有，active 必须与自身候选事实一致。Host fixture 应组合独立资格目录而不重复追加已在 Builtin 的身份；相同完整身份但不同事实必须明确拒绝，不能放松重复身份校验或默默忽略冲突。保留未通过不能晋升的负向边界，Pi extension 的 NOT_RUN 语义不变。

## Ownership / validation

负责 `ai/catalog.go`、新的独立地区 catalog loader/data、public regional promotion CLI / 消费器、候选九份展示名称与相应版本/snapshot/pin、上述两处 E2E fixture及直接公共 Client/CLI 验证。不改生产协议thinking、Minimax续验授权、finalbudget或GLM发布工具，不读取Key、不发真实请求、不执行race、不激活九身份。

先读取验证策略并列出失败方式，按已约定公共 Client/CLI 接缝 RED→GREEN，只有不可网络证明的真实资格留待新原生调用。每个代码文件不超过500行；保留命令、数字退出、UTC、源码和原件。实现本身不授予付费，也不表示资格已通过。

## Acceptance

晋升逻辑、原生证据拒绝、两态普通行为与名字/Hash边界完成独立审查；Builtin仍inactive。后续只在真实候选全部通过后执行data-only activation及正式目录pins，重建二进制，再采自身 formal。所有tracked派生矩阵/selector在完整ordinary/race冻结前发布。
