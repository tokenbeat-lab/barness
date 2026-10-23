# 治理当前状态

日期：2026-10-24（Asia/Shanghai）。**工单01 COMPLETE；工单02 COMPLETE；工单03 COMPLETE（脚本边界文档验收通过）。** 第4项及后续尚未实施；E2E/race没有本轮新执行，完整发布判定保持 false。

## 第3项：脚本身份与执行边界

- 基线提交：`99d074964aa844a3ed0b3b9a047db08af18bc46f`。只新增分类/导航/治理说明，并为活动工具README增加分类入口；没有修改或迁移任何脚本、产品实现、依赖、夹具、原生索引或支持矩阵。
- 入口：[脚本边界总览](script-boundaries.md)、[逐文件机器清单](script-boundaries.json)及总览链接的九份目录分册。覆盖318份独立脚本（310 Python、8 MJS）和2份压缩Python快照；活动工具86、测试57、历史源码159、诊断装置18，合计320项。
- 每项记录身份理由、当前调用方、CLI/父模块输入、输出/写点、联网/厂商/构建/测试/race/压力能力、适用源码/原pins、原路径/基址/哈希与替代入口。静态候选与实际执行事实分开；间接子命令能力保留适用限制，不因关键字、main guard或help断言安全。
- 8份现行SOP工具/模块、5份现有公共消费者、67份仍供合成测试消费的旧控制器/模块、6份真实recovery7暂停模块分别标注。33份现有工具测试/支持与24份历史失败测试分开，不声称它们全部由普通Go套件自动执行。
- 特别保留Jev verify-11、Jev/Decisions留存ZIP verifier、GLM/MiniMax rollback及GLM launcher拒绝测试的现有接线。旧生成器、run-14、两组MJS、硬编码宿主封存程序和诊断overlay不推荐作日常入口；不重建旧差分环境。
- 已被原evidence.json绑定的旧release-verification README及其诊断README原字节保留，边界由外部分类说明；现行工具README新增导航，其[原文及基址/哈希](history/phase3-before-20261024/index.json)保留。上轮状态/spec/plan另存本轮attempt；旧record、pin和历史verdict不改。

## 本轮检查与适用范围

- 3211项文档/完整性检查通过，1538处引用/锚点有效；320项覆盖与字段、压缩/解压双哈希、现有调用方/静态引用、封存入口均核对。所有原脚本、332项保护对象、5份另列封存入口及地区暂停字节不变。
- 完整执行来源1909份的集合/字节等于本轮基线及第1项实际受测来源。旧保守输入8757份路径集合相等；相对第2项结束，仅现行工具README的导航文字变化，其余8756份字节相等，旧README精确快照可读。相对第1项原输入，仍有此前3份Markdown变化，不能重新绑定原E2E为当前全输入PASS。
- 本轮源码检查复用现行SOP receipt_files.source_snapshot纯读API；脚本身份分析只用AST/文本/调用关系，没有执行/导入历史源码，也未作help探针。Go测试、vet、构建、Node/npm、厂商、race、压力和真实清理均未运行。
- [实际文档检查命令回执](../../.evidence/scratch-artifact-remediation/docs/20261023T225126Z-5olxg8hl/final-check-command/command-receipt.json)固定argv、实际数字退出0、时间、日志及检查器SHA；[完整检查结果](../../.evidence/scratch-artifact-remediation/docs/20261023T225126Z-5olxg8hl/final-document-check.json)记录输入差异与引用。静态清单生成器/分析原件保留在同attempt，不是新的日常工具框架。
- [本轮不可变索引](../../.evidence/scratch-artifact-remediation/docs/20261023T225126Z-5olxg8hl/record.json)及[只读报告](../../.evidence/scratch-artifact-remediation/docs/20261023T225126Z-5olxg8hl/report/release-report.md)分别记录source、E2E、live、race、压力事实。Source通过；E2E NOT_RUN、race NOT_RUN，live/压力有纯文档不适用依据，release.pass=false。文档交付不等于完整发布通过。

## 已完成的前两项

- 第1项受测来源为`214cff42fdf72d6cd00dfd76bdd80f57999d8b1c`。干净独立Git clone五包普通命令真实exit0、446.40秒；58项必交case PASS；主原生包14,071 PASS / 214 NOT_RUN / 0 FAIL、audit0；710个顶层入口705 PASS/5 SKIP、14,943个Go通过节点。无测试文件的releasegate包不冒充有测试通过。
- 552份固定输入、13,551,018字节随Git交付：两个ZIP/manifest、六场景runtime派生bundle、六份原SHA GLM归档。[原普通统计](../../.evidence/scratch-artifact-remediation/planning/clean-final-summary.json)、[原不可变索引](../../.evidence/scratch-artifact-remediation/record-20261023T223511Z-pkea1jy5/record.json)和[依赖清单](input-dependencies.json)保留原适用范围。缺输入RED、首候选全包FAIL和审计误选原生根的exit1原件均保留。
- 第2项已统一基础spec、公共合同、研究追溯与地区README，50条User Stories、E01–E11、P01–P06身份保留；[四份原文/基址](history/phase2-before-20261024/index.json)、[原检查](../../.evidence/scratch-artifact-remediation/docs/20261023T224146Z-wd6mrs8d/final-document-check.json)及[不可变索引](../../.evidence/scratch-artifact-remediation/docs/20261023T224146Z-wd6mrs8d/record.json)保留。24项检查、112处引用、71保护对象通过；三份旧scratch Markdown变化如实记录，没有新Go执行。
- 第2项record绑定的状态原字节见[本轮修订前状态](../../.evidence/scratch-artifact-remediation/docs/20261023T225126Z-5olxg8hl/previous-status.md)；第1项record的原状态仍在第2项attempt。旧绑定不能套用到修改后的当前status路径。

## 暂停、未执行与回滚

- recovery7权威index仍为PAUSED_AFTER_NATIVE_ORDINARY_TIMEOUT，用户暂停true；paidContinuationAuthorized/actualHuman7Created/nativeGrant7Created/paidClaims7Created/releasePass均false，formalREADY/race为NOT_RUN。没有恢复prepare/check/构建、付费或正式race。
- live与压力NOT_APPLICABLE依据：只改展示分类/入口，无厂商wire/目录/资格、资源/并发/容量行为变化；不将NOT_RUN混写成不适用。地区及其他功能唯一状态、原授权/grant/claim/账本和余额边界保持。
- 真实清理、工具迁移、第4项历史目标追溯和第5项目录治理均未开始。原有未跟踪SOP回归和审查目录保留。
- 回滚第3项时成组回退新增分类/治理说明及活动README导航，保留新旧快照和检查原件；不改原脚本/历史pin，不恢复已关闭轮次或recovery7。
