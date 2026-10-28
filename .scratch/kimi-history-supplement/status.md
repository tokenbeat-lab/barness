# Kimi 补项实施当前状态

2026-10-27：本次实施与受控、普通验证完成。六路线旧源码差异已审，真实旧54项当前离线校验全部通过；固定单场景补采入口、只读两来源组合消费者已实现。真实路线仍各为9/10，缺失场景未执行，正式目录仍inactive，`release.pass=false`。

当前索引：[current-qualification-index-v4.json](../../.evidence/kimi-history-supplement/current-qualification-index-v4.json)；本次判定：[decisions-v1.json](../../.evidence/kimi-history-supplement/decisions-v1.json)；完整验证与保全：[verification-and-retention-v1.json](../../.evidence/kimi-history-supplement/verification-and-retention-v1.json)。旧索引和全部历史事实保留；23条是路线/型号证据条目数，不是Provider路径交付比例。

## 已实现

- 按原三个提交核验六路线源码原字节，保留完整diff，并追加78条具体Provider/API/文件/前后SHA审阅。只支持已审旧九项；新增xAI来源行为不由旧证据代证。未来产品差异、SDK变化、未知运行时代码、错误来源/Provider继续拒绝，见[source-review.md](source-review.md)。
- 真实旧54项：当前编码器重建54个Host请求，54/54 wire相同；48个完成响应相同、6个取消捕获前缀有效。当前原生baseline校验保留原9-call预算、候选标签、时间、授权和可信账本；完整CLI仍拒绝旧九项独立取得十项资格。
- `processgrant`新增固定purpose的未完成机器模板；`^TestKimiHistorySupplement$`只允许schema2、固定场景、原baseline manifest、一次调用/一次HTTP/零重试、4096输入预留/4096输出、USD0.10、240s/单次120s。真实模式首发等待60s；精确次数为1。旧标准授权/新补项授权不能交叉启动，失败或取消不退还授权。模板不签发执行权限。
- `kimi-composition.json`只接受固定原九项加新一项的两个非嵌套引用。校验原件hash、同路线/型号/账户/地区/租户/SDK、真实进程argv/退出/日志/二进制和可信账本，保留两次独立测量及十项覆盖。当前Host编码不使用捕获payload覆盖。`releasegate -inspect-bundle`、原生live消费者与release record共用此判定；外层消费者绑定全部传递输入。
- 单项报告始终`complete=false`，受控组合只得到`CONTROLLED_COMPLETE`。真实组合、正式资格分别要求真实模式及自身正式目录事实；不伪造一次完整采集、不激活目录、不新增调度/续期/恢复控制器。
- ADR-0055、测试SOP和[REGIONAL.md](../../ai/live/REGIONAL.md#kimi-原九项与单项补采)已同步。旧九项受控夹具使用独立`live,kimi_baseline_fixture`标签，实际执行九次loopback，普通live二进制不含该入口；该夹具拒绝live，所有新合成授权写前断言私有Git authority。

## 六路线事实

| 路线 | 真实旧九项离线适用性 | 受控9+1 | 新增真实xai-origin-history | 当前真实十项 |
| --- | --- | --- | --- | --- |
| kimi-cn-chat | PASS，9项 | PASS | NOT_RUN | 9/10 |
| kimi-cn-responses | PASS，9项 | PASS | NOT_RUN | 9/10 |
| kimi-cn-messages | PASS，9项 | PASS | NOT_RUN | 9/10 |
| kimi-intl-chat | PASS，9项 | PASS | NOT_RUN | 9/10 |
| kimi-intl-responses | PASS，9项 | PASS | NOT_RUN | 9/10 |
| kimi-intl-messages | PASS，9项 | PASS | NOT_RUN | 9/10 |

后续最小厂商事实仅为这六项；每路线USD0.10、合计USD0.60是拟议上限，本次授权未签发。原54项无需因本次消费者/测试变化重新付费采集。其他Provider既有适用性及MiniMax历史缺口沿用旧索引，不由Kimi受控结果改写。

## 验证与来源

实际受测与当前HEAD均为`69ff4af880e68ef3ca245e5128bdae5e59a2ef7b`，另以2029个实际源码文件和181个选定输入的完整哈希绑定未提交工作区。相对本次开始仅37个工具/测试/消费者/harness来源文件变化，产品适配器、目录与go.mod/go.sum变化为0；三份文档按选定输入核对。当前来源、普通回执来源及实际选定输入集合完全相等。

- 完整普通：[receipt.json](../../.evidence/kimi-history-supplement/e2e/20261027T123725.813300Z-pzo24_qg/receipt.json)，退出0，总墙钟654.379s。实际argv及181输入见[selection](../../.evidence/kimi-history-supplement/full-ordinary-selection-v1.json)。覆盖`./ai/e2e ./ai ./ai/internal/testkit/{release,supportmatrix,minimaxlive,jevformal} ./ai/examples/visualeval`，`-count=1 -failfast -timeout=20m`。
- 主E2E 15528项：15314 PASS、214 NOT_RUN；新增必需KHS用例91项全部PASS（源码36、模板及拒绝12、采集6、组合6、入场拒绝22、失败不可重用6、Host重编码3）。辅助原生库存25+32+77+6项全部PASS，所有audit无发现。
- 六组实际独立九次及单次受控进程通过；覆盖重封Host输入、重pin进程证明、错误entry/count/timeout、跨路线/型号、模式升级、单包伪完整、预算、缺省/重复/未知授权字段、503零重试与失败后不可复用。RED/失败attempt和后续GREEN均保留。
- [最终公共回放](../../.evidence/kimi-history-supplement/public-replay-v2/receipt.json)、[最终原生旧九项](../../.evidence/kimi-history-supplement/native-baselines-v2/receipt.json)均退出0；[真实旧九项单独申请完整资格](../../.evidence/kimi-history-supplement/standalone-nine-refusals-v1/report.json)六次均预期拒绝、退出1。
- 工具Python过程56项PASS，普通影响包及`live,kimi_baseline_fixture`标签vet通过；编译的只读CLI留存在[build-v1](../../.evidence/kimi-history-supplement/build-v1/receipt.json)。每次独立运行均清理live/压力/凭据环境。

## 保全与未执行范围

[最终保全](../../.evidence/kimi-history-supplement/history-protection-final-v1.json)：8740个受保护历史文件全部相同；六个公共存储根共8156文件，无新增、删除或内容改变。原grant、授权、claim、账本、旧失败、暂停记录与以前的索引保持；新受控原件、私有账本和实际运行二进制也留存于普通attempt的bundles旁，组合描述符固定引用它们。只读重验受控组合时使用该次留存的私有Git authority，不能改用公共账本。

- 付费授权签发0、厂商请求0、真实claim0；六条补项live均NOT_RUN。
- race与压力均NOT_RUN，用户明确禁止。本次新schema入场、既有一次性claim接入及进程取消/资源范围留待将来按影响选择，不推断人工授权，也不扩大到所有Provider复验。
- 目录激活、真实支持发布、暂停轮次恢复均未执行；本次实施通过不等于所有Provider交付或完整发布通过。

回滚范围仅为本次源码/文档差异及新入口的启用；保留此前未提交改动和全部新旧证据、授权与账本，不回滚账本或复活旧权限。
