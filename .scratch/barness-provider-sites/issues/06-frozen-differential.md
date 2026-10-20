# 06 冻结基线地区差分与缓存 v2 账本

Status: resolved
Type: task
Blocked by: 01, 02, 03

## Context

[规格](../spec.md) R05 R06 R11 R18 R20，RP06/RP18。冻结 pi 数据实际含 zai Chat glm-5.3/5.2 与 minimax-cn Messages M3。当前 wrapper 没列出该数据不能作为“不在冻结基线”的依据。

## Ownership

冻结 pi wrapper 的已有 Provider export 选择、新地区实际差分 E2E/fixture、精确缓存 v2 账本条目与相关 wrapper 测试；不变更上游版本、模型数据、SDK 或完整来源 pins。与 04 的 supportmatrix/release/live 并行，各自提交，05 等待两者。

## Acceptance

公共 Client 对冻结 pi 的适用两个新路线及三个型号实际比较；其余四路线登记扩展并独立 fixture。所有旧8差分保留；精确记录 v2 字段变化，不免除其他 wire/归属/usage 比较。先失败再实现，保留实际 argv/exit/时间/来源/原生工件。无付费、无 race。

## Comments

2026-10-20：从票04拆出可独立并行的真实差分工作；额外测试/工具工作不降低原验收合同。


## Answer

2026-10-20：实现与定向普通差分完成。整体状态仍只在 [spec 当前状态](../spec.md#当前状态) 维护。

- wrapper 仅向原 APIS 表追加冻结包已经导出的 `ZAI_MODELS` 与 `MINIMAX_CN_MODELS`。版本、43 文件 model-data、827 文件源码、lockfile、SDK、provenance 及全部 pins 未变；先在公共子进程 Models 接缝得到三个身份缺失的实际 red，再添加既有 export，启动安全控制保持。
- ZAI Chat `glm-5.3` / `glm-5.2` 与 MiniMax CN Messages `MiniMax-M3`，分别执行 text、分片 tools、原生 tool-replay × stream / streamSimple：18 次真实双侧完整请求 / 事件 / Result 比较，并增加 3 次完整型号元数据比较。Case 前缀固定为 `PIDIFF-RP06-zai-chat-`、`PIDIFF-RP06-minimax-cn-messages-`，所有实际 Provider / model / API 保留，不使用 modelPatch、字段剔除或替换上游代码。
- 差异仅按明确 case × protocol × kind × path 分类；每个例外先检查独立字面 wire、header 长度、用量和按各自来源计算的成本。完整保留标准 API / Coding Plan、历史 / 当前官方主机、保守输出界限、text-only、协议配置、thinking / cache / eager 字段及当前独立价格的差异，不据 fixture 宣称资格。边界控制拒绝未知 case、字段、协议或种类。
- 四个真正无冻结路线的组合追加注册，六个完整型号身份 × 四种 Client 入口独立执行，共 24 个 extension 案例；明确 `piParity: NOT_RUN`，没有 `pidiff.json`，不能借作真实差分或 live 通过。两个已有冻结路线从不登记路线豁免。原八子集与原案例保留。
- 缓存例外改为 ADR-0049 的精确 v2 域、UTF-8 uint64 大端长度前缀 Tenant / 实际 Provider / Account / Session 算法；Responses 三个键 / 亲和字段限定原 6 案例，Chat key 限定原 3 案例；没有其他 wire / 内容 / 用量豁免。

实现提交 `1623db6`；已融合 integration `fc9c07d`，最终受测代码 `2b884cb46ae68906e707633b4d373a547d866d3e`。本关闭记录不改变已测代码或输入。[fixture 与差异说明](../../../ai/e2e/testdata/regional-pidiff/README.md) 解释有限来源与迁移约束。

[工件索引](../../../.evidence/provider-sites/differential/index.json) 集中保留 argv、退出码、UTC 起止时间、source Hash、原始 scripts / captures / events / Result / usage、请求与结果 Hash、原生审计及 red 原件：

- `final-new-routes/receipt.json`：47 个原生案例 PASS，21 个真实差分记录均 pending 0，包耗时 4.495s。
- `final-wrapper-controls/receipt.json`：Node 子进程 / 目录 / pin / 账户环境 / 网络隔离控制 10/10 PASS，退出码 0。
- `integrated-old-eight/receipt.json`：融合后原八差分子集的定向 15 个案例 PASS，包耗时 4.234s；更广的原案例及精确旧缓存 / 亲和案例在 `old-chat-routes`、`old-added-scopes`、`old-nonchat-controls` 回执中实际通过，范围以各自 native manifest 为准，未冒称全量发布。
- 所有实际 Go 差分阶段显式设置 `BARNESS_AI_PIDIFF=1`；Node runner 自身采用空环境，仅访问各例指定的 loopback。`frozen-inputs.json` 核对禁止变更的文件与原始提交逐字节相等。

仅使用临时 worktree 的 node_modules 符号链接访问 root 已安装冻结包；未安装或升级依赖，检查后已移除链接，原目录未修改。未读取真实 Key、未调用付费服务、未运行 race；正式门禁仍由 04 / 05 按各自实际证据判断。
