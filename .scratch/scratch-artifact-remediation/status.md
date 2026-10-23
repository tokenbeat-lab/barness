# 治理当前状态

日期：2026-10-24（Asia/Shanghai）。**工单01–04 COMPLETE；工单05 IN_PROGRESS（第5A项活动工具迁移）。** 第5B项尚未实施；本轮没有新的产品E2E/race事实，完整发布保持 false。

## 第4项：历史链接与原样快照追溯

- 基线提交：`19795e5decef6a9902e4adbf768a684378fd17ff`。只新增外部追溯索引/导航和本治理记录，没有修改原历史正文、脚本、快照、原生索引、产品源码、夹具、支持矩阵或授权。
- 入口：[历史追溯总览](historical-traceability.md)、[349处逐项处置](historical-link-dispositions.md)、[逐项JSON](historical-links.json)、[快照索引](historical-snapshot-index.md)、[快照来源JSON](historical-snapshots.json)、[固定Git版本](historical-git-targets.md)。这些展示索引不供产品或证据消费者执行，也不发放资格。
- 原349项保持顺序/内容：272项按已证实原基址可读，70项为已删除仓库路径的历史版本，2项临时地址有原哈希确认的永久副本，4项为网站根相对地址，1项在第2项已修正活动入口并保留原文。
- 40项来源记录包含33份原样文档、6份原来源清单哈希一致的来源捕获及1份路径脱敏历史报告。原样文档19份与Git原文相等，14份由原发布清单/状态回执pin确认，但文档Git版本未找到；不拿执行来源commit冒充文档版本。Kimi、Decisions、MiniMax及前轮治理快照均按原路径/基址解释。
- 旧`/tmp/barness-ticket17-current-source-applicability.json`未重建。[ticket09永久原件](../barness-ai-minimax/evidence/ticket-09/final-closure/current-source-applicability.json)与[ticket17审查原件](../barness-ai-minimax/evidence/ticket-17/reviews/current-source-applicability.json)各49,705字节，均等于两份原文记录SHA256 `29d7009a2f31e03199b17225b3f069b78d99777e4c0d8c21ea71f1c9e51f8ff0`。只恢复导航；原2306917适用范围、严格相等性false、原selector/矩阵/原生pin和既有结论不改。

- 来源捕获另注明格式：reference.txt为HTML正文提取、pi-classifier.json为npm单条目，捕获文件SHA不冒充原HTML/完整tgz。原HTML位置及完整tgz字节本轮未确认；[格式说明](historical-snapshot-index.md#捕获格式不等于原始下载字节)保留转换和未知范围。首轮第4项记录/状态/索引原字节另存[capture澄清前快照](../../.evidence/scratch-artifact-remediation/docs/20261023T234325Z-t6bj9jrp/capture-clarification/before.json)，不覆盖其回执。

## 仍未确认的部分与查找边界

- 70处旧源码引用涉及28个当前缺失路径，已取得49个固定Git版本并校验blob/SHA。原链接无目标版本或哈希，故逐项保留expectedOriginalSHA256=null及originalRevisionVerified=false；不能声称确切原字节全部找回。若原结论依赖该确切版本，仍须原pin或独立的新来源证明，原接受条件与资格不放宽。
- 14份原样文档的Git版本未知，但原pin对应的文档字节仍在；另1份脱敏派生报告的原完整路径/原字节哈希未知。4处网站根链接只按原sources.json映射来源URL，远端内容/当前可用性未验证。没有把这些项写成当前PASS。
- 只查349项的明确目标、已证实原基址、原发布/回执/来源清单和有限本地Git。快照最多80个原路径变更commit、单blob最多8MiB；旧源码最多40个唯一路径、每文档/目标4个commit、200个blob、总64MiB，单Git命令20秒。源码实际49个blob/8,268,529字节，约2.98秒。
- 临时库存先命中两个原哈希相等的随Git副本；其余缺失路径均有有界Git版本，故停止扩大查找。没有穷尽TAR/ZIP、全盘/tmp、用户目录或网络，也不声称归档不存在其他副本。[查询命令/范围与限制原件](../../.evidence/scratch-artifact-remediation/docs/20261023T234325Z-t6bj9jrp/search-log.json)保持。未定位只能称“本范围未找到”，不能推断永久丢失。
- 本轮没有已确认永久丢失的原件，也没有完全未定位的本地候选；这不等于历史原件全部完备，以上版本/转换/远端限制继续可见，相关资格保持受限。

## 本轮验证与适用范围

- 349项覆盖、原文href/片段、40项快照来源、19份Git相等原文、14份原pin、49个旧源码指针、两份临时库存原SHA及全部新引用核对通过。原失效列表、8757份旧选定scratch输入的集合与字节、345项保护对象及1909份完整执行来源保持；前轮已pin的输入除本治理status/spec/plan外均不变。
- 第3项record绑定的原状态完整保存在[修订前状态](../../.evidence/scratch-artifact-remediation/docs/20261023T234325Z-t6bj9jrp/previous-status.md)，前轮spec/plan也同attempt保存。历史record、报告、脚本及所有pin不改，不能把旧状态hash套用到修改后的当前路径。
- [真实文档检查命令回执](../../.evidence/scratch-artifact-remediation/docs/20261023T234325Z-t6bj9jrp/capture-clarification/final-check-command/command-receipt.json)保留argv、实际数字退出0、起止/墙钟、日志和检查器SHA；[检查结果](../../.evidence/scratch-artifact-remediation/docs/20261023T234325Z-t6bj9jrp/capture-clarification/final-document-check.json)列全检查和限制。[本轮不可变索引](../../.evidence/scratch-artifact-remediation/docs/20261023T234325Z-t6bj9jrp/capture-clarification/record.json)及[只读报告](../../.evidence/scratch-artifact-remediation/docs/20261023T234325Z-t6bj9jrp/capture-clarification/report/release-report.md)另存。
- 本轮来源检查复用现行SOP纯读source_snapshot API，只读文件/Git；没有执行历史脚本或捕获命令，未运行Go/Node测试、vet、构建、厂商、race、压力或真实清理。未修改供运行/资格消费者使用的追溯数据；新索引仅展示。
- 新选定输入含当前外部索引、原pin与检查事实；本轮E2E NOT_RUN、race NOT_RUN，live/压力有纯文档不适用依据；Source通过，release.pass=false。原第1项实际E2E仍只适用其原来源和输入，不因本次导航恢复重新绑定为当前全输入PASS。

## 已完成的前三项

- 第1项：受测`214cff42fdf72d6cd00dfd76bdd80f57999d8b1c`，干净独立Git clone五包普通命令exit0、446.40秒；58项必交PASS，主包14,071 PASS/214 NOT_RUN/0 FAIL，audit0。552份固定输入随Git交付；[原普通统计](../../.evidence/scratch-artifact-remediation/planning/clean-final-summary.json)、[原索引](../../.evidence/scratch-artifact-remediation/record-20261023T223511Z-pkea1jy5/record.json)及[依赖清单](input-dependencies.json)保留。缺输入RED、全包FAIL与误选原生根的exit1原件不改。
- 第2项：统一基础spec/合同/追溯/地区README；[四份原文/基址](history/phase2-before-20261024/index.json)、[原检查](../../.evidence/scratch-artifact-remediation/docs/20261023T224146Z-wd6mrs8d/final-document-check.json)、[原索引](../../.evidence/scratch-artifact-remediation/docs/20261023T224146Z-wd6mrs8d/record.json)保留。三份旧scratch Markdown变化如实单列，没有新Go执行。
- 第3项：[320项脚本身份及边界](script-boundaries.md)完成，86活动工具/57测试/159历史源码/18诊断装置；3212检查、1539引用通过。[原不可变索引](../../.evidence/scratch-artifact-remediation/docs/20261023T225126Z-5olxg8hl/record.json)及[检查结果](../../.evidence/scratch-artifact-remediation/docs/20261023T225126Z-5olxg8hl/final-document-check.json)保留。唯一活动工具README导航变化已有原字节快照，脚本未变、未迁移或执行。

## 暂停、未执行与回滚

- recovery7仍为PAUSED_AFTER_NATIVE_ORDINARY_TIMEOUT，用户暂停true；paidContinuationAuthorized/actualHuman7Created/nativeGrant7Created/paidClaims7Created/releasePass均false，formalREADY/race为NOT_RUN。没有恢复prepare/check/构建、付费或正式race。
- live与压力NOT_APPLICABLE依据：只新增外部展示追溯，无厂商wire/目录/支持资格、资源/并发/容量行为变化。E2E/race为NOT_RUN，不冒充不适用。其他功能唯一状态、原授权/grant/claim/账本及余额边界保持。
- 工具迁移、第5项目录治理和真实清理尚未开始；原有两个未跟踪目录保持。回滚第4项只回退外部索引/导航及本治理说明，保留查询/核验原件，不回写归档或旧结论，不恢复关闭轮次。
