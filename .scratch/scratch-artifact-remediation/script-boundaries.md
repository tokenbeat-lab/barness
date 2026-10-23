# Scratch 脚本身份与执行边界

本轮逐项登记318份独立脚本（310 Python、8 MJS）和2份压缩Python源码快照。原件全部保留；没有独立业务TS/TSX/JS文件，MJS是两组固定pi发布件核验程序；归档内部源码不作为独立活动入口。本次通过源码/AST和调用关系界定身份，help和import均不用于探测；历史脚本没有执行或导入。来源相等性检查复用现行SOP的receipt_files.source_snapshot纯读API，不启动测试或历史命令。

唯一进度及检查原件见[当前状态](status.md)。现行操作按[测试SOP](../../docs/agents/verification.md)及[工具合同](../barness-ai-test-process-review/tools/README.md)；地区操作只到[当前暂停入口](../barness-provider-sites/tools/README.md)。此分类不续期旧授权、补买路线或恢复recovery7。

## 身份与使用边界

| 身份 | 数量 | 含义 |
| --- | ---: | --- |
| 活动工具 | 86 | 活动工具/消费者及仍受测试消费的模块；activity决定真实执行边界 |
| 测试 | 57 | 测试/fixture/support；controlled与historical分开 |
| 历史源码 | 159 | 历史执行/生成/封存/核验源码及快照 |
| 诊断装置 | 18 | 原问题诊断或探针，不产生正式资格 |

“活动”表示现有调用链仍需要其源码；其中只有8个SOP工具/模块属于通用当前流程。旧控制器即使仍被合成测试加载，其真实轮次仍关闭；6个recovery7模块的真实操作保持暂停。33份现有工具测试/支持文件须显式选择，不能声称都在普通Go套件自动运行；另24份测试属于历史失败清单。

每份记录包含身份理由、调用方、CLI/父模块输入、输出与写点、网络/厂商/构建/测试/race/压力能力、适用Git/hash/原pins与替代入口。[机器清单](script-boundaries.json)保存全部读写/子进程调用点、导入与字面引用、顶层候选、字面pins、相邻原件及压缩/解压双哈希；各分册给出逐文件说明。静态关键字只是定位线索，非发生事实或“安全可运行”认证；未能静态证实的间接行为明确留在子命令/父工作流边界。

## 按目录阅读

| 目录 | 独立脚本 | 压缩快照 | 分册 |
| --- | ---: | ---: | --- |
| barness-ai-glm | 14 | 0 | [逐项说明](script-boundaries/barness-ai-glm.md) |
| barness-ai-minimax | 133 | 2 | [逐项说明](script-boundaries/barness-ai-minimax.md) |
| barness-ai-openai-decisions | 16 | 0 | [逐项说明](script-boundaries/barness-ai-openai-decisions.md) |
| barness-ai-pi-1.0 | 32 | 0 | [逐项说明](script-boundaries/barness-ai-pi-1.0.md) |
| barness-ai-release-verification | 2 | 0 | [逐项说明](script-boundaries/barness-ai-release-verification.md) |
| barness-ai-test-process-review | 19 | 0 | [逐项说明](script-boundaries/barness-ai-test-process-review.md) |
| barness-ai-thinking-sampling | 16 | 0 | [逐项说明](script-boundaries/barness-ai-thinking-sampling.md) |
| barness-provider-sites | 64 | 0 | [逐项说明](script-boundaries/barness-provider-sites.md) |
| jev-1.14-multimodal-box | 22 | 0 | [逐项说明](script-boundaries/jev-1.14-multimodal-box.md) |

## 容易误用的入口

| 原件/场景 | 边界与当前入口 |
| --- | --- |
| thinking-sampling/fixture-cases.py | 顶层直接覆盖4份fixture；保留一次性作者源码，不运行修补当前fixture。新需求另开失败场景与显式输入输出。 |
| thinking-sampling/update-ledger.py | parse_args前读取已删除旧ledger，help也会失败；旧差分退役，当前用自身契约/留存fixture。 |
| Jev run-14-verification.py | 零凭据仍编译、vet、11路线预检、效果/harness，--gate可启动旧全量/race/差分/压力；由分阶段SOP替代日常使用。 |
| 两组release-verification/*.mjs | 1.0.0与1.1.0来源等价历史；run联网下载、npm安装、TS构建/审计，failures/faults会修改临时输入；不恢复活动oracle。 |
| GLM ticket13封存程序 | preserve/index/seal/reconcile含固定本机/临时路径且会写原归档/索引；只保留原轮次，不能当可移植通用命令。 |
| Jev verify-11-evidence.py | Go jevformal和live受控测试仍执行，活动保留；Python结构通过不足以产生正式资格。 |
| Jev/Decisions verify.py | 普通测试仍消费Git内retained ZIP；--audit会写临时副本并调用Go审计，不执行ZIP内旧Node源码。 |
| GLM/MiniMax rollback.py | 普通测试仍执行固定Git基线的临时恢复/构建/E2E证明；不是生产回滚指令，输出和临时清理均有副作用。 |
| provider-sites/minimax tools | CLI/helper/测试分开登记。prepare/check可能构建，run-selected可能付费，finalize可能写矩阵；测试存在不复活关闭轮次或recovery7。 |
| native-pressure、replay-red、retention/review探针 | 依赖历史source guard/overlay/OS时序与私有fixture，可编译或产生loopback负载；仅诊断，不产生发布资格。 |

## 封存说明与原件绑定

barness-ai-release-verification/README.md及其native-pressure/README.md已被原evidence.json逐字节绑定，故保留正文；其中“当前”“零凭据重现”和整轮命令只属原轮次。两组上游README也按固定版本材料阅读。活动导航以本页、SOP和当前工具README为准，不从旧README复制整轮命令。分类本身不改变原hash/pin、失败结论或来源适用性；[封存入口记录](script-boundaries.json)同时列出原记录哈希。

原路径仍被源码的parents[n]/本机常量、测试和封存哈希依赖，尚未迁移。历史来源不匹配时保留拒绝/缺失，不能通过重封存、补写PASS或重建差分环境修复。压缩快照单独列为历史源码，本次没有解压执行；链接缺失的全面追溯属于第4项。
