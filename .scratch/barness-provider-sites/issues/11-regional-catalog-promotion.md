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

## Answer

预装独立 `catalog_regional.go/json`，当前 payload 明确 `active=false`、空 models/版本，Builtin 内容与整体版本保持原值。六份 own live candidate 原生消费者全部通过并具有相同 candidate catalog、source/build/binary/commit，才可由固定公共 `regionalpromote -receipt <new> <six-native-dirs>` 生成九身份激活数据。receipt 限于本轮新 `.evidence/provider-sites/final-acceptance-20261020/`，拒绝输入原件树及符号链接父路径；独占 fsync intent 记录旧/新 SHA、目标、六路 report/seal/manifest/grant/authorization/source/catalog/facts pins 后，再原子替换数据并 fsync 父目录。中断的 intent 保留，不覆盖、不自动重试。生产代码不依赖 testkit 或证据目录。

离线 `-controlled -proposal <new> <six-native-dirs>` 仅生成明确 controlled、inactive 且无法作为生产 payload 载入的提案。真实 live 激活成功尚未执行；受控 positive 不代表自身 live 资格。缺失、重复、controlled 冒充 live、目标已变、锁、receipt 重用均拒绝。两态 Host qualification 组合严格接受完整且逐字段相同的九身份，拒绝 partial/name/price/duplicate 冲突；公开 NewClient 的重复身份校验保持。

九份候选显示名固定为 `Z.AI GLM-5.2/5.3`、`MiniMax-M3 中国站`；共同版本 `.2`，snapshot/pin 同步更新；INCOMPLETE/NOT_RUN、官方原件及 source-files 均保持。08 若增加已证兼容事实，须在此基础上另升 `.3` 后再开始 candidate。

公共 Client/CLI 逐片 RED→GREEN 原件和数字退出、UTC、source、argv 保留于 [普通验证索引](../../../.evidence/provider-sites/final-acceptance-20261020/promotion-dev/index.json)。最终受影响 regional 普通 E2E 717 PASS / 4 NOT_RUN（默认未开启的 PIDiff catalog/inventory），独立审计 0 findings；另行启用冻结差分补跑 22 PASS、21 份真实 PiDiff、审计 0（原始名称断言 RED 保留）；两条既有 result.name 说明仅改已选名称，路径/kind/cases/classification/有限ID集合完整相等回执见同一索引；四条非适用 extension 的 piParity 仍保持 NOT_RUN。原目录回归和限定 vet 通过。原生 bundle 和命令细节见索引。保留中间失败和审计纠正原文。未读取凭据、未调用真实服务、未运行 race、未激活九身份。最终真实晋升及整体验收继续由 [spec 唯一当前状态](../spec.md#当前状态) 维护。

## Comments

2026-10-20：上述Answer记录实施准备时的inactive状态。固定续验2六条自身candidate同commit/build/binary全部81/81且原生qualify通过后，根任务已执行原生晋升及catalog-pins；独占lineage Hash `d6c6e104f51df88d77481cb243fc561774a933ab4061e398fdb2a1b7b85028c1`，目录版本2026-10-20.1-regional，九身份metadata与自身候选完全一致。数据delta两轴复核零问题，晋升后重建并完成双vet/整个ai/live检查。新formal及完整发布仍未运行，当前状态只在spec维护。
