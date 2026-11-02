# 型号名称限制移除当前状态

当前阶段：产品、215项定向与9身份CLI证明完成，准备最终冻结并执行完整普通。
基准提交：9584d3d6f095aefa4ab73bdc75c39a705b7931d6。
工作批次：`artifacts/catalog/work/20261102T041450Z-10fb1e18/`。
根目录原有 .catalog/、local-catalog.json、provider-integration-status scratch 保留。

影响范围：Kimi/xAI chat、xAI image 和 TypeSafe 图片/Box classifier 的目录校验，
以及去掉名称判断后仍须保留的直接 Catalog 身份边界。其他能力和 Host 授权保留。
只读审计原件已归集至工作批次 inputs/model-id-audit-v1.md。

root 编排 GPT-6.1 Terra：审计 agent 只读分类，验证 agent 负责测试先行和各阶段原件，
产品 worker 只修改指定运行文件。共享当前 checkout，所有 worker 保留其他人修改。

普通计划：新 ID 的外部加载/Hash/NewClient/受控调用、能力拒绝、空/重复身份、
未授权0HTTP以及合成 CLI JSON-only 新增闭环。完成后全受影响普通包与生成/构建/vet。
实际必需 case、argv、来源与输入按阶段独立冻结；未执行不写通过。

live 计划 NOT_APPLICABLE：本次只验证已有能力下的代码兼容，不取得新厂商事实。
race 计划 NOT_APPLICABLE：无共享状态、锁、取消或资源实现变化。
压力计划 NOT_APPLICABLE：无复杂度、限额或设计负载变化。
上述判断依最终差异复核；若出现适用范围，按 SOP 明确列出并等待人工选择。

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
未发现更多必须改动项；源码修订、测试适配、SOP、changelog草稿和ADR将在完整普通前冻结。

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

最终完整普通计划改用不可覆盖 v3 脚本/输入清单/必需清单，必需225项，
80项清单加脚本与清单自身共82实际输入。旧v1/v2材料保留，不将其当成已执行v3事实。
