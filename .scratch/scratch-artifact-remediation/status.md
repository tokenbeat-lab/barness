# 治理当前状态

日期：2026-10-24（Asia/Shanghai）。**工单01–06 COMPLETE；第5B项顶层导航完成。** 第5A项活动工具迁移与普通验收事实保留。真实live资格缺口及race未执行，完整发布保持false；recovery7继续暂停。

## 第5B项交付

- [顶层导航](../README.md)覆盖当前工作区22个目录，每目录一行，列出类型、spec/唯一当前状态入口、活动工具、历史资料和普通测试直接依赖。研究/展示/缺陷/回归/审查/治理按实际用途登记；未设集中状态的目录如实指向工单或报告，不增造状态源。顶层不复制阶段结果、轮次、预算或授权。
- SOP工具及显式进程测试链接到tools/verification，Jev链接到现行jevformal核验器；GLM/Decisions/MiniMax仍供普通测试读取的源码与Pi旧资料中的静态数据分别标明。地区工具仅导航到唯一状态/工具README及分类，不提供恢复命令；旧分类版本按迁移映射读取，不修改原封存说明。
- 基线e0ddbb649169e09db328976ce46ab734a573c331，库存含两个原有未跟踪目录barness-ai-sop-test-20261022、scratch-artifact-review。本次只链接这些本地材料，不修改或纳入Git；没有重命名或补造spec。修订前状态/spec/plan及库存见[本轮基线](../../.evidence/scratch-artifact-remediation/docs/20261024T002612Z-bc0058a0/before.json)、[旧状态原字节](../../.evidence/scratch-artifact-remediation/docs/20261024T002612Z-bc0058a0/previous-status.md)。
- [导航检查](../../.evidence/scratch-artifact-remediation/docs/20261024T002612Z-bc0058a0/navigation-check.json)与[实际命令回执](../../.evidence/scratch-artifact-remediation/docs/20261024T002612Z-bc0058a0/navigation-check-command/command-receipt.json)为exit0；最终交付复核及其命令回执存于[本轮检查目录](../../.evidence/scratch-artifact-remediation/docs/20261024T002612Z-bc0058a0/)。检查覆盖目录集合、六栏/单行、全部本地链接/锚点、格式、无动态结论副本、8813个历史/暂停保护文件、14份原有未跟踪材料及前轮原件。首轮检查将历史文件名日期误作正文状态，真实exit1的[原报告](../../.evidence/scratch-artifact-remediation/docs/20261024T002612Z-bc0058a0/initial-check.json)/日志保留；检查改为扫描可见正文后另起成功attempt，没有改旧报告。
- 本项不改执行来源或实际消费数据；1912项完整来源与8804项实际选定输入集合/字节仍等于第5A受测4e291311。新导航仅展示，另在交付索引固定哈希，不冒充既有回执的实际测试输入；没有重跑Go/Node测试、构建、历史脚本或工具合成测试。此前真实资格缺口/race未执行保持，压力影响无新增；本项未执行付费、race、压力或真实清理。
- 新的[不可变索引](../../.evidence/scratch-artifact-remediation/docs/20261024T002612Z-bc0058a0/release-record/record.json)绑定本状态、原普通回执及其完整实际输入；[只读报告](../../.evidence/scratch-artifact-remediation/docs/20261024T002612Z-bc0058a0/release-record/report/release-report.md)与[判定命令回执](../../.evidence/scratch-artifact-remediation/docs/20261024T002612Z-bc0058a0/release-evaluate/command-receipt.json)独立保存。旧第5A状态绑定由本轮旧状态原字节验证，不回写前轮record/pin。回滚仅撤回顶层导航，原件/回执继续保留。

## 第5A项交付

- 迁移前基线：`f286690ebf30c7dd7bf1908276dc1ff78f4c9bf0`。先记录[ADR-0053](../../docs/adr/0053-verification-tool-ownership.md)与失败清单/公共seam，再实施两个RED→GREEN切片；测试/ADR提交6239f358先于迁移实现8221f106、0d5e866b。
- 当前入口：[tools/verification工具合同](../../tools/verification/README.md)、[Jev独立核验器](../../ai/internal/testkit/jevformal/verify_protocol.py)及[归属/历史后继映射](tool-migration.md)。8个SOP模块、8个原进程测试/支持和README移入tools；Jev364行Python保持原字节，归入拥有它的Go消费者旁。
- 活动旧PY/README路径直接删除，无wrapper/shim/双写；[18份原字节/原路径/基址/模式/Git](history/phase5-before-20261024/index.json)保留。14份迁移后文件字节完全相同，另4份仅为Python来源目录、测试根定位及README导航调整。其余历史脚本、原生索引、回执/pin、支持矩阵与控制器不变。
- Python采集器与Go只读发布消费者独立枚举ai/、tools/verification/、go.mod/go.sum。Jev原生producer/consumer显式固定新Python路径及字节；Go调用点/测试/活动SOP文档使用新入口。第3/4项原索引保持其审查版本，18个后继路径由新映射解释，不重写旧pin。

## 实际普通验收

- 冻结受测提交：`4e29131189a28df43f923f1a1d792d53fac7926a`；执行来源1912项，选定输入8804项。本轮顶层展示导航和治理收尾不改变这些实际输入；完整执行来源与全部实际选定输入仍等于受测集合/字节，不把HEAD标签代替受测SHA。
- 新run公共入口先因不存在取得真实exit1，迁移后exit0；新Jev路径/source pin先RED再GREEN。公共collector→releasegate交接验证新目录、忽略PY及工具/测试/Jev源码的增删改，不能靠漏收文件取得来源通过。
- [最终完整普通回执](../../.evidence/scratch-artifact-remediation/e2e/20261024T000626.339050Z-pncc23t9/receipt.json)：真实exit0、443.79秒。命令为go test -json ./ai/e2e ./ai ./ai/internal/testkit/release ./ai/release/cmd/releasegate ./ai/internal/testkit/supportmatrix ./ai/internal/testkit/jevformal -count=1 -failfast -timeout=20m。四个有测试包PASS，releasegate/jevformal两个无测试包SKIP；不冒充测试通过。
- 主原生包14,071 PASS / 214 NOT_RUN / 0 FAIL、audit0；58项既有输入必交场景全部PASS，14,943个Go通过节点。41份ai Markdown另以--input固定，覆盖原生选择器实际读取但默认来源策略排除的文档。
- 同一冻结来源的[48项工具进程测试](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/final-sop-process-suite/command-receipt.json)真实exit0、28.54秒：环境/持久Go配置、隐藏负载阻断、stdout/真实退出、取消/超时/管道、最终写失败、只读资格与合成清理边界。清理仅在私有临时repo，没有真实清理。
- [五个显式Jev受控live-tag公共入口](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/final-controlled-jev/command-receipt.json)真实exit0、88.51秒；94个父原生case PASS、audit0。包括新UC14归属/pin与结构通过/拒绝真实资格，及既有正式管线、篡改、非普通文件、候选过程。只用合成凭据/loopback，属于普通受控验收，不计真实live通过。
- [检查结果](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/final-migration-check.json)及[真实检查命令回执](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/final-check-command/command-receipt.json)覆盖原件、334保护对象、8739份非迁移旧输入、当前源码/8804输入、调用路径、全部测试库存和活动链接。原状态完整保存在[本轮修订前快照](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/previous-status.md)，不修改第4项record的旧hash。

## 原生资格与禁止阶段

- [原生来源影响](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/native-source-impact.json)：GLM公开-source的264项集合/字节未因迁移改变；MiniMax原生执行选择器本体不变，迁移增删改均在其producer/capture选择之外。这只评估本次来源影响，不对既有失败/过期发新PASS。
- 地区公开-source的完整集合由481变482，6项Go内容变化、1项新增；[原生摘要对照](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/regional-qualification-impact.json)确认绑定不同。共用规则的12条路线（ZAI3、MiniMax CN3、Kimi两站6）不能沿用旧来源绑定；没有为它们生成新授权、构建恢复轮次或付费续期。
- Jev显式Python源路径改变，其原严格集合不能刷新迁就新位置。[当前只读消费者对旧Jev正式包的实际拒绝](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/old-jev-readonly-qualification/command-receipt.json)为exit1，原因wrong fixed formal protocol identity or provenance。原包也可能已因目录/此前来源过期，不能把该拒绝全部归因于此次搬迁。原历史结论和原pin保持，未重新封存或借用旧PASS。
- 因此本轮live为NOT_RUN，13个来源绑定相关路线均缺当前真实资格；不能一概写NOT_APPLICABLE。用户明确禁止付费，本单元不补买、不晋升、不恢复旧预算。
- race NOT_RUN，用户本次明确禁止。压力有依据NOT_APPLICABLE：仅归属/路径/来源选择范围变化，原进程取消/超时、资源上限、并发与负载算法未变；进程边界已普通验证，没有运行真实设计负载。真实scan/plan/move/verify/purge均NOT_RUN。
- recovery7权威index仍PAUSED_AFTER_NATIVE_ORDINARY_TIMEOUT，用户暂停true；paidContinuationAuthorized/actualHuman7Created/nativeGrant7Created/paidClaims7Created/releasePass均false，formalREADY/race为NOT_RUN。64个地区控制器、pause-handoff、closed6 seal、支持矩阵与原账本不变。
- [新不可变发布投影](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/release-record/record.json)使用实际1912来源/8804输入与58项必交case；[只读报告](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/release-record/report/release-report.md)为Source/E2E通过、live/race不通过、压力有依据NA，release.pass=false。补充进程/原件/资格说明另在delivery索引固定，不向schema1塞自造阶段事实。

## 保留的失败与偏差

- 两个预期RED均保留真实exit1及源码/输入；不会改为PASS。
- 首个完整采集包装器缺少runpy模块路径，实际exit1、Go尚未启动；[原包装器/失败](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/collector-start-failure.json)保留，修复只涉及本轮补充包装器。
- 首个已启动普通attempt因缺41份实际Markdown输入而主动停止，[原回执](../../.evidence/scratch-artifact-remediation/e2e/20261024T000455.638218Z-4ce4ypaq/receipt.json)保持CANCELLED、实际Go退出-9，保留部分输出/包。补齐输入后另起最终attempt取得PASS，没有回写原件。
- 首次Jev RED的相对输出落入包目录；[位置修正清单](../../.evidence/scratch-artifact-remediation/migration/20261023T235818Z-62ptjy95/red-jev/output-location-correction.json)记录164份本轮原件全部哈希相等地收进证据目录，仅移除新建空目录。后续均用绝对输出，不属于真实证据隔离/清理。

## 已完成前项与回滚

- 工单01干净工作区固定输入、工单02规范入口、工单03脚本分类、工单04历史追溯均完成。其原回执/失败及适用范围保留；第4项40条来源/349候选和未知原版本不因本项改写。历史入口后继只按新映射或原固定Git解析。
- 回滚第5A项将代码/调用点/来源选择器/测试/活动文档作为同一单元回退；保留18原件及所有新旧回执，不复活已消费授权或以任意新状态套用旧来源PASS。第5B仅回退导航；真实清理尚未开始，原有两个未跟踪目录保持。
