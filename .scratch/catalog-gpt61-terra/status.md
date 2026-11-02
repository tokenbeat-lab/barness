# GPT-6.1 Terra 目录接纳与内置采用状态

当前状态：三组默认数据已采用。普通回归首轮中断后修订三处测试夹具，待复验。此前原件保持。
采用起点：`37a283ee4ef743f93d3a12132d1bc2c3211fb1e5`。
GPT-6.1 Terra 子 agent `/root/terra_builtin_verification` 负责受影响验证编排和执行；
root 负责完整审阅包导出、默认数据采用和交付。

本次影响：维护源 `catalog/data/providers/`、运行快照 `ai/catalog_builtin.json`、
完整采用基线 `catalog/data/baseline/`；必要的历史迁移断言与 Terra 公共 E2E。
不改变 Provider 协议或共享资源实现。旧 118 条定义、价格、身份和相对顺序保留，
新增一个 Responses 身份；内置目录共 119 条。

普通范围：公共 Builtin/Client 的 Terra 合约、历史投影、完整基线采用与新维护种子，
以及完整受影响普通包。Terra 子 agent 先保留公共 RED，再于采用后执行 GREEN；
具体 argv、必需场景和显式 inputs 由该子 agent 按测试 SOP 固定并留存。
另核对 `go generate ./ai`、构建/vet 和生成数据一致性。

live：NOT_APPLICABLE。仅验证受控代码兼容和默认数据采用，不需新厂商事实。
race：NOT_APPLICABLE。未改共享状态、锁或取消路径。
压力：NOT_APPLICABLE。未改复杂度、限额或设计负载；119 条在既有覆盖容量内。
本次无新厂商调用、发布或部署范围，不宣称完整发布门禁或真实型号资格。

审阅包已按 CLI dry-run → confirm 导出，revision 保持
`sha256:4047a1f3e757fd6d1f2a0378341478e2f034df05e40031e628112a0d9e78b5dc`。
完整包：
`.evidence/catalog-gpt61-terra-builtin/maintenance/20261102T005920Z-000b6471/approved-review/`。
19 个库存文件加 `review.json`，包含 12 份 Provider、State、snapshot、revision 和
4 份去重输入；4 个原外部引用保持。索引 SHA-256：
`b6342df484d6622b225c42431c04567e4d6c76794aeafdd18fcb70c54d136574`。
快照与本地已批准输出逐字节相同。

采用核对：12 份 Provider 文件均使用审阅包原字节；仅 OpenAI 发生语义变更，
其他 11 份仅为 CLI 导出后的确定性序列化差异。运行快照保留原批准字节，
完整基线 20 个文件（含完成索引），没有遗留 README 或未索引文件。
首次采用脚本已完成替换，但末尾误将“仅一份语义变更”当成“仅一份字节变化”，
断言 exit1；原 `03-adopt-approved-data` 命令/日志保留。只读 v2 分别核对原字节、
JSON 语义和完整库存后 PASS/exit0，见同 attempt 的 `adoption-result-v2.json`。

Terra 公共 RED 已保留：
`.evidence/catalog-gpt61-terra-builtin/e2e/20261102T005948.081733Z-hiffow6d/receipt.json`，
实际 FAIL/exit1，唯一失败为当时内置目录尚无目标；没有重写为成功。
采用后的最终定向 GREEN 在 `8a8dc08b97650bb31544ead014c0c0c69527bc70`
通过：3 个 Terra 测试、用量/价格/推理/授权等子场景与来源完整性，
回执 `.evidence/catalog-gpt61-terra-builtin/e2e/20261102T010455.785682Z-kafe5ryb/receipt.json`。
规范与需求两个独立 Terra 审查通过，见维护 attempt 的 `review-v3.json`。

完整普通首轮：
`.evidence/catalog-gpt61-terra-builtin/e2e/20261102T010549.256058Z-gxw82nfi/receipt.json`，
观察到 `TestCatalogCandidateFixedEntryReceipt` 和
`TestCatalogCandidateLastWithdrawalToEmpty` 失败后停止，
collector 为 CANCELLED、实际进程退出 -15；未完成项不算通过。
原因是旧单模型合成测试修改了 Provider，却沿用正式 baseline，完整性拒绝符合合同。
另扫描发现旧 README 留存断言和隔离采用 helper 覆盖目录的旧假设。
仅修订三个 `ai/e2e` 夹具/断言：合成历史 seed 使用自己的空初始基线；
源码归档逐项检查真实完整基线；隔离采用先移除旧 baseline 再复制新审阅包。
产品完整性校验、默认定义和原失败证据均不改。

定向 fixture 回归首轮：
`.evidence/catalog-gpt61-terra-builtin/e2e/20261102T010809.594560Z-u374xgby/receipt.json`，
FAIL/exit1。前两个失败场景和源码归档已通过；ReviewExport 的旧断言把继承引用
视作当前临时 store 的原件，并在第二轮强制覆盖已有 origin。按现有合同修订
这两处断言：继承引用独立对照原基线，保留已有 `originReview`；只有原为空时
补当前来源索引摘要。本轮失败原件保留，后续定向和完整普通复验待执行。

采用前当前状态快照：同维护 attempt 的 `status-before-adoption.md`。
既有未跟踪 `.catalog/`、`local-catalog.json` 和
`.scratch/provider-integration-status-20261102/` 均保留；不将它们混入提交。

## 已完成的本地接纳阶段

状态：已通过 CLI 完成本地接纳和应用（2026-11-02）。
受测和当前提交：`37a283ee4ef743f93d3a12132d1bc2c3211fb1e5`。

影响：仅新增一个本地运行定义，不修改协议实现、维护工具、内置目录或
已有条目。Host 通过 `catalog.LoadFile` 和 `Config.Catalog` 显式选择输出文件。
开始时 `.catalog` 和 `local-catalog.json` 均不存在。
既有未跟踪目录 `.scratch/provider-integration-status-20261102/` 保留。

公共验证入口：`catalogctl check` 固定的候选目录受控 E2E，覆盖最终完整目录
的每条身份、请求/响应、用量、价格和适用能力。普通阶段 PASS/exit0：
119 条身份、1,032 个必需原生场景全部 PASS，audit 发现 0。
不新增测试，不修改冻结 oracle。

live：NOT_APPLICABLE。本次接纳仅确认代码兼容，不需要新的厂商事实；不发送
真实模型请求，不改变原有厂商资格。

race：NOT_APPLICABLE。仅维护数据新增，不改共享状态、锁、取消或资源路径。
压力：NOT_APPLICABLE。不改复杂度或设计负载；118 到 119 条属于既有目录规模。
不沿用旧功能的 release.pass 作为本次新增的厂商资格。

输入和操作原件：
`.evidence/catalog-gpt61-terra/maintenance/20261102T003945Z-dfba5cb9/`。
models.dev 抓取包含本型号，但 `pdf` 模态不在 barness 的文本/图片类型内；
`tiers.size` 未区分严格大于边界。按 SOP 选择官方补充入口，不静默丢弃这些字段。
价格仅表示 Standard、默认 30 分钟缓存写；12 小时、Batch/Flex、地区加价等
条件不混入基础费率。Responses 的现有显式缓存模式只使用默认或 30 分钟 TTL。

## 实际结果和原件

- 主源预览：exit1，`unsupported_definition`（`pdf`）；未创建维护库。
  `02-primary-preview.command.json` 与 stdout 留存，不覆写为通过。
- 官方完整定义预览：exit0，只有非阻塞人工确认提示；随后以同一候选 ID 冻结。
- 候选：`sha256:1ff8a740cd4a90dcf0e6be9f62bbdadbff649bf610b64c3758a43d663b004a37`。
- 实际 `catalogctl check --timeout 5m --confirm`：exit0，9.16 秒，
  `applicable/code_compatible`。检查引用：
  `.evidence/catalog-gpt61-terra/check-inputs/20261102T004145.325113000Z-2642263489/check.json`。
- 普通回执：
  `.evidence/catalog-gpt61-terra/e2e/20261102T004146.432669Z-oggu1wyq/receipt.json`，
  SHA-256 `784f3c78fcc7eef602534dc640e91c6b03bc46719028fc2340c7c104f371785b`。
  原生包为该 attempt 的 `bundles/20261102T004148.031028000Z/`；
  原生 Go 测试库存 1,151，选定输入 49 个文件，完整源码集合由回执单独固定。
- 应用预览和确认应用均 exit0；revision：
  `sha256:4047a1f3e757fd6d1f2a0378341478e2f034df05e40031e628112a0d9e78b5dc`。
  OpenAI 内容版本为 2，其他 11 个 Provider 保持 1。
- `local-catalog.json` 完整文件 SHA-256：
  `04107af45ba03e3a4e7660aa5255997303b8c1442b74fc6d905b37fba0ba80ca`。
  与冻结和受测字节相等，原 118 条定义逐字段相等，仅新增目标身份。
  `cataloggen --store .catalog --out <绝对路径>/local-catalog.json --check` exit0。
- 额外只读审查 v1 把原生 case ID 与 Go 测试名误当同一命名空间，断言 exit1；
  `post-apply-audit-v1-failure.json` 保留该错误。v2 改为分别核对 manifest case ID、
  Go 库存、audit 和摘要后 PASS/exit0，未改变候选或产品数据。

所有实际 argv、数字退出、起止时间和 stdout/stderr 摘要见上述 maintenance
attempt 下 `01` 至 `12` 的 `*.command.json`。最终不可变索引：
`.evidence/catalog-gpt61-terra/maintenance/20261102T003945Z-dfba5cb9/completion-v1.json`。
完成前状态快照为同目录 `status-before-completion.md`。

Host 使用根目录 `local-catalog.json` 装配新 Client；内置数据采用和线上发布是
独立操作。本次未签发厂商资格或完整发布结论。

完成后回滚方式：对新增四维身份执行 `withdraw`，形成新候选并重新 check/apply；
保留历史和原检查。当前未执行回滚。
