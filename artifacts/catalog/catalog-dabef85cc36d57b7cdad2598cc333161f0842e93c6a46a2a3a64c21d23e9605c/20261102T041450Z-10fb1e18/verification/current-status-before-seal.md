# 型号名称限制移除当前状态

当前阶段：验证已完成，正式交付候选准备完成，最终审查及索引封存待完成。
基准提交：9584d3d6f095aefa4ab73bdc75c39a705b7931d6。
工作批次：`artifacts/catalog/work/20261102T041450Z-10fb1e18/`。
根目录原有 .catalog/、local-catalog.json、provider-integration-status scratch 保留。

影响范围：Kimi/xAI chat、xAI image 和 TypeSafe 图片/Box classifier 的目录校验，
以及去掉名称判断后仍须保留的直接 Catalog 身份边界。其他能力和 Host 授权保留。
只读审计原件已归集至工作批次 inputs/model-id-audit-v1.md。

root 编排 GPT-6.1 Terra：审计 agent 只读分类，验证 agent 负责测试先行和各阶段原件，
产品 worker 只修改指定运行文件。共享当前 checkout，所有 worker 保留其他人修改。

已完成普通范围：新 ID 的外部加载/Hash/NewClient/受控调用、能力拒绝、空/重复身份、
未授权0HTTP以及合成 CLI JSON-only 新增闭环，完整受影响普通包与生成/构建/vet。
实际必需 case、argv、来源与输入按阶段独立冻结；未执行不写通过。

live NOT_APPLICABLE：本次只验证已有能力下的代码兼容，不取得新厂商事实。
race NOT_APPLICABLE：无共享状态、锁、取消或资源实现变化。
压力 NOT_APPLICABLE：无复杂度、限额或设计负载变化。
以上依据最终产品差异复核；测试新增的只读历史夹具不改变产品并发实现。

真实125条目录、Provider 版本、基线与历史快照本次保持；合成型号只存在于测试原件。
后续交付使用同一 Catalog.Version 的新批次，不覆盖 Kimi K3.1 Pro 原158文件包。

公共 RED 两次原件：
`.evidence/catalog-model-id-policy/e2e/20261102T041548.575496Z-zus0m6ad/` 和
`.evidence/catalog-model-id-policy/e2e/20261102T041617.254114Z-aj6ywbrd/`。
第二次实际退出1，原生11 FAIL/135 PASS、audit0，覆盖9新身份、CLI code_gap及
跨Provider/API同名身份。产品修改在 RED 封存后执行。

定向 GREEN 首轮原件：
`.evidence/catalog-model-id-policy/e2e/20261102T041721.865236Z-tplc27st/`，
实际 FAIL/exit1、183 PASS/8 FAIL、audit0。4个场景的响应别名断言超出旧合同，
另4个父场景随之失败。Kimi Chat/Responses 原有规则允许服务端别名，分别记录
请求目标身份。root修正规格和测试要求，产品原响应合同不改变。
该失败回执保留，后续使用新attempt重验。

第二次定向原件：
`.evidence/catalog-model-id-policy/e2e/20261102T041903.793107Z-ccwlmiya/`，
实际 FAIL/exit1、203 PASS/4 FAIL、audit0。仅 Kimi Responses 两个场景及父场景失败：
旧解析器不填写 ResponseModel，测试误将 Chat 的别名记录合同套到 Responses。
依据 ai/responses.go:413 与 ai/chat_stream.go:217 分别限定断言；产品响应逻辑仍不改。

第三次定向 GREEN：
`.evidence/catalog-model-id-policy/e2e/20261102T042005.753739Z-_exvvpl8/receipt.json`，
实际 PASS/exit0，207/207 原生场景通过、audit0。9个合成身份、3个目录入口、
受控请求/用量/计价、别名合同、严格协议、空ID/能力/授权拒绝及 CLI 闭环均覆盖。

独立旧负例审计发现 Jev Box 三条仅改ID的旧拒绝场景；原测试失败原件：
`.evidence/catalog-model-id-policy/e2e/20261102T042109.497968Z-9nmol3cw/`，
实际FAIL/exit1、3 FAIL/5 PASS。仅将这3条声明能力的同协议型号改为接纳；
Decisions、纯文本、不可见位置和重复位置的真实能力负例仍拒绝，默认目录未变化。

最终定向 GREEN：
`.evidence/catalog-model-id-policy/e2e/20261102T042122.569488Z-ulu96n25/receipt.json`，
实际 PASS/exit0，215/215 原生场景通过、audit0，receipt SHA-256
`85c02a2ca87ddd3ece3c132d40f5c0582c9195fae3bbf6eb0bf84a594915f2ef`。

产品仅改5份运行文件，删除4处名称条件，新增共享空ID保护，保留能力与授权合同。
Gemini/OpenAI已有名称驱动行为保留并列在审计报告。其他旧名称负例已独立扫描，
该轮局部审计未再发现项，但后来仍遗漏两组旧断言，已在下方失败与补审记录说明。
源码修订、测试适配、SOP、changelog草稿和ADR在各轮完整普通前分别冻结。

完整普通使用 WORK/verification/run-ordinary-v2.py、ordinary-inputs-v2.json
（80项跨默认范围输入）和required-cases-v2.json（215项），三者保持原字节。
普通开始后不得提交或修改受测源码/输入；当前状态Markdown不作实际输入。

初次冻结：65a1ecbb2cb97f51c37cfee849925712187e5acc。最终定向工作树与其
2168项源码、80项输入逐项相等，证明见 WORK/verification/green-applicability-v1.json，
SHA ea473ccc51eba68561357eb176a6335bef792546056eccc537aca7b232ff955f。
独立Spec和Standards审查均无发现，原件为WORK/review-spec-v1.md与review-standards-v1.md。

验证agent补充说明：215项中，公共Client覆盖9身份，但CLI只新增1个Kimi身份。
root决定在完整普通前补齐9身份的CLI闭环，避免以单类CLI证明替代其他操作的维护路径。
现扩展测试，产品和125条真实目录保持原样；之后使用新的测试/输入/脚本版本冻结。

9身份CLI矩阵原件：
`.evidence/catalog-model-id-policy/e2e/20261102T042409.321856Z-84w6k84i/receipt.json`，
PASS/exit0、10/10原生场景通过、audit0，SHA-256
`eb2dac25a4a4930ee908daa11f1c3ea1d8d4f6f865bc93f4e2d8b62f643649b1`。
一个CLI二进制供9个独立store使用，每条保留preview/freeze/check/apply原件；
check-reference.json 指向仓库根 .evidence 内各自完整检查，不依赖已清理的临时store。
每条应用输出与候选snapshot原字节相等，原125身份和顺序保持，只有所选合成身份新增。

该轮完整普通改用不可覆盖 v3 脚本/输入清单/必需清单，必需225项，
80项清单加脚本与清单自身共82实际输入。旧v1/v2材料保留，不将其当成已执行v3事实。

最终冻结：807afbc548ad9410a090121a8671bdb189f50128。矩阵定向工作树与最终
2169项源码、80项所选输入逐项相等，证明为verification/matrix-applicability-v1.json，
SHA 851484eae5d4d8dd23ec8e9e8922b02477ec9db8c60414e8f2b14f92765aa6bc。
独立增量review-spec-v2.md与review-standards-v2.md均无发现。

完整普通v3原件：
`.evidence/catalog-model-id-policy/e2e/20261102T042820.344950Z-_6n8y39z/receipt.json`。
实际来源807afbc5、所选输入82项、必需225项；终态CANCELLED/exit-15，原因见下方。
命令为go test -json ./ai/e2e ./ai ./ai/examples/chineseeval/... ./catalog/...
./cmd/catalogctl ./cmd/cataloggen -count=1 -timeout=40m，外层collector上限2700秒。

新版本化交付批次为原Catalog.Version下的20261102T041450Z-10fb1e18目录；
运行125条快照与34文件review复制自原批准包，均与当前内置/基线逐字节相等。
这是新程序策略的验证批次，不是新型号数据版本，不把历史check冒充本次策略证明。
delivery.json尚未封存；普通及构建的最终结果已归集，见文末。

完整普通v3观察到TestCatalogMaintenanceSourceReview/reviewed-current失败后停止，
collector CANCELLED/exit-15，原件保留。旧测试要求历史xAI Images来源适用于当前
image_pricing.go，来源审核器按合同拒绝本次变化。没有启动生成/构建/vet。

修订仅涉及E2E：历史迁移正例读取首个RED回执的完整2168文件source map，附原件及
hash；新增当前策略来源拒绝。原审核器与reviewed_sources.json均未修改。
7/7来源定向PASS/audit0/exit0，回执
`.evidence/catalog-model-id-policy/e2e/20261102T043954.709247Z-z1d6pacr/receipt.json`。

独立审计又定位xAI Images complete/paths/history的同一历史来源假设。其原失败分别
保留在044038.592509Z-t6e4vjhm、044049.119679Z-bpern134、044055.566644Z-y55tgz9q。
只为测试从真实Git9584d3d6导出ai/、go.mod/go.sum，releasegate子进程在该隔离目录
核对历史合同，packet放唯一可写目录。当前SDK来源另作拒绝；不改全局cwd/env，
不增加产品注入参数，不改旧grant/ledger/native source/capture或授予当前资格。

初次解包失败原件044400.075903Z-n5qp4g98：Git PAX元数据头，历史比较尚未执行；
仅修测试解包器。044420.502936Z-kzec2ztp命令只覆盖当前拒绝；历史完整合同随后在
044437.329507Z-bixafe6t通过6case。044458.962904Z-8ekb361y通过148case后，
进一步将历史ai目录设0500、文件及模块文件设0400，保留独立packet可写目录。
以上均在本feature e2e原件根，失败/范围不足/中间通过分别保留。

最终只读历史定向：
`.evidence/catalog-model-id-policy/e2e/20261102T044646.946229Z-0wet3qqo/receipt.json`，
PASS/exit0、148/148原生case PASS、audit0，SHA
`b0052626a279389fde95bc9633591e063070f4b26bb33b6d243c85694d8de7bf`。
原始Git archive、导出argv/退出、2189项文件清单和逐项hash保存于原件；
它是历史比较上下文，不是当前SDK获得厂商资格或旧SDK重新执行的证明。

新完整普通使用v5脚本/清单，327项固定必需case，82项跨默认输入加脚本和清单
共84项实际输入。53个动态fixture case按完整原生清单核对；最终还核对历史archive
与树的内容hash，不能仅用主bundle audit替代。

最终来源4f2764bb3392d17e21820cf960573fe913998a6f，产品5文件与65a保持相同。
历史148定向工作树的2171项源码、81项输入与冻结版逐项相等，证明
verification/historical-applicability-v1.json，SHA
424ca3af0c55e0401391bf14c891e5c56ab33b9b42c744479534ee6d954fc69a。
独立review-spec-v3.md、review-standards-v3.md均无发现。

完整普通v5运行原件：
`.evidence/catalog-model-id-policy/e2e/20261102T045214.516113Z-zroapojr/receipt.json`。
实际source固定4f2764bb，327项固定必需、84项实际输入。
该轮终态CANCELLED/exit-15，原因及后续修订见下方。

完整普通v5观察到TestJevContextCandidateContract与TestKimiCatalogBoundary/model
失败后停止，collector CANCELLED/exit-15。旧断言仅将ID改为alias/旧Jev或kimi-k2.5，
却期待拒绝；新策略接受这些人工声明的同能力定义，符合需求。仅修测试预期，不修改
真实型号能力或产品行为。旧失败原件保持，不以取消掩盖观察到的断言失败。

修订定向36case回执：
`.evidence/catalog-model-id-policy/e2e/20261102T050450.914556Z-e9cc3t05/receipt.json`，
PASS/exit0/audit0。补充旧名称断言审计在WORK/inputs/old-name-negative-audit-v2.md，
记录此前漏项、逐类ID写入/替换/表驱动分支及明确排除范围，不把静态检查当运行证明。

完整相关边界子集：
`.evidence/catalog-model-id-policy/e2e/20261102T050516.128659Z-301xoyc0/receipt.json`，
PASS/exit0、主bundle audit0，SHA
`f8b994c2eb970e9d71d84548b32fada4198f91b713143c192d0de4362e220ca4`。
范围含Kimi/Jev/classifier/LocalCatalog/xAI及future，两个实际失败测试均执行并通过。
Go 3947 pass、178 skip；skip均为未启用的classifier pressure，不算压力通过。

最终完整普通v6计划329项固定必需，明确包含实际失败对应UC04-candidate-contract和
KIMI-INTL-CHAT-K01-catalog-reject-empty-id；82项manifest加脚本/清单共84实际输入。
该计划经后续独立补审及新提交冻结后执行，结果见文末。

最终补修冻结8d97cb6002722aceeb7fb3ccfaf5eb27e0991f11。产品与65a后保持一致。
完整旧边界回执与其2171项源码、82项实际输入逐项相等；记录为
verification/boundary-applicability-v1.json，SHA
ee7d709a3254b677fdee21bc7ebd57f519a65b3ccd4a2b752a9978948986c9e5。
独立review-spec-v4.md与review-standards-v4.md认可测试修订；唯一状态旧措辞在此修正。

完整普通v6来源固定8d97cb60，329项必需、84项实际输入。
运行期间未提交或改实际输入；生成/构建/vet在该轮实际通过后顺序执行。

## 最终验证结果

实际受测提交：8d97cb6002722aceeb7fb3ccfaf5eb27e0991f11。
最终普通原件：
`.evidence/catalog-model-id-policy/e2e/20261102T051207.821665Z-1z37t3sn/receipt.json`，
PASS/真实exit0、墙钟1361.622秒，SHA
55b5b268324169db71bcbab89e53d202abf70453598301ce09295a10ecbefd73。
主bundle `20261102T051209.995718000Z` 为17232 PASS、226 NOT_RUN；
329固定必需case全部实际PASS，包含之前实际失败项。四bundle audit均无发现。

go generate ./ai、go build ./ai/... ./catalog/... ./cmd/catalogctl ./cmd/cataloggen、
同包范围go vet按顺序真实exit0。执行脚本、argv/工具/起止/日志/每步源码输入核验
保存在WORK/verification/build-v1/和对应原字节脚本，不将构建当成新测试测量。

机器结果WORK/verification/result-v1.json，SHA
3eca7c5fc610e086afaa473559c1d013ffaf341886e5235b359dfe6b1bf29c9c。
嵌套检查/归档核验ordinary-inspection-v6c.json，SHA
c6a051d80fd3287c38af6b64e4028188572fb46332ee0d4248f2fd65421f98ba；
9条CLI各自完整check/required/audit/来源与应用原字节均通过。

历史Git archive SHA
32bac88ed05080baa36bbb21dff06bb373a329ab7068f93dcfb21983547d2d74，
2189项完整文件集合/每项SHA及只读权限相符。53条动态记录由48个qualification-fixture
和5个path-fixture组成，后5条复用5个迁移packet；实际独立packet为48，不是少测5项。
两轮case→packet→producer及复用映射相等、所有实际producer exit0。完整映射
fixture-coverage-v1.json，SHA
6fa94c37bc2689d2dffffc39fcdaef2aab68206ffc57eed9ee272965d6269ccf。
两次只读汇总脚本因目录定位漏项而失败，其原件保留，最终v6c完成核验；不改普通PASS测量。

skip-index-v1.json逐条保留252个Go skip和226个原生NOT_RUN，SHA
a32eefcb8db6e4a81381e9b60956a8ee28b5d160d647157585c7025be7bc8992。
329必需case无skip；普通未启用的压力不算压力通过。未执行真实厂商/live/race/压力。
当前source gate拒绝旧xAI Images资格；历史比较不表示旧SDK重新执行或当前SDK取得资格。

文档后source/input核验post-docs-equality-v1.json，SHA
52e505495c0c11c89f42c096b96fae602138fa06e0155995c425b1b2b3c3ac3b：
2171源码文件完整集合/每项SHA相等，84实际输入相等，HEAD仍为8d97cb60。

分发判定：SDK核心5文件变化、运行依赖不变，必须先升级一次SDK以取得新准入策略。
此后相同已实现能力的新增通过审核外部JSON/CLI检查后装配新Client，Host仍独立授权。
本包125条原JSON与9584基准相同，旧程序继续使用原125条无需因相同JSON升级；
旧程序拒绝的是新增未来名称。编译内置默认仍必须重建，当前125条数据与版本未变化。
本次没有发布软件版本、部署或取得新厂商资格，也没有完整产品发布门禁通过的声明。
