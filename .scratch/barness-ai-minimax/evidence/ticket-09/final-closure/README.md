# 最终审查与交付证据

当前交付 **PARTIAL**。本入口区分代码问题闭合、历史原件留存、实际验收结果与当前来源适用性。它不产生新的厂商报告、付费授权或发布认证。

## 实际结果

用户批准固定六项新增最多108calls/132HTTPAttempts。六项全部关闭，五项collection PASS、一项FAIL，实际102/102；连此前545/545累计647/647。中国区GLM直连，国际protocol114和两项视觉经127.0.0.1:7890。M3未重跑，未用额度不转移，没有下一轮授权。

- [GLM原件](../../ticket-15/new-actual/glm/README.md)：Chat、Responses各22/22 PASS；Messages21PASS/1FAIL，cancel在可观察内容前触发调用期限，120001ms。其`live-report.json`未生成，严格续期exit1。
- [效果原件](../../ticket-15/new-actual/effects/README.md)：protocol114、visual Jev、visual Decisions各12/12 PASS。
- [装配原件](../../ticket-15/new-actual/assembly/README.md)：派生Jev0、校验0、候选四报告导入1，无默认导入和完成标记。原17身份矩阵16PASS/1FAIL是旧Responses结果，不能改写成最新Messages失败。
- [完整门禁](../../ticket-15/new-actual/final-gate/README.md)：来源7931的首次实际完整CLI结束，八输入校验0，vet/live-vet/ordinary/race为0/0/1/1，outer1、报告false、post-guards闭合。两主套各12445PASS/6FAIL/0NOT_RUN，审计0、868真实PIDIFF PASS、29个pressure命名PASS（固定规则选择15项）。

## Standards

固定基点`ecfd4b4d6ec54db30453f4b5112989e9fdbd0782`至原审查点7931：原1项P2文档问题见[原始报告](initial-source-standards.json)，[票17文档复审](ticket17-docs-standards.json)闭合。[票17源码复审](ticket17-source-standards.json)亦为零发现。

留存工具的原始问题和后续负向夹具构造器问题也分别保留；[最终构造器Standards复审](retention-factory-final-standards.json)零发现，原ordinary、票17、B门禁及补正中间报告可按索引逐一核对。

原始审查报告不被最终零发现结果覆盖。各留存原件的原始发现、补正和复审另列在对应目录；`review-files.json`逐文件记录SHA和大小。

## Spec

原2项P2见[原始报告](initial-source-spec.json)：MiniMax Messages最终payload形状校验与过时文档。[源码复审](ticket17-source-spec.json)和[文档复审](ticket17-docs-spec.json)均零发现。[票17完整离线历史](../../ticket-17/README.md)保留真实RED、错误夹具、失败GREEN和环境准备分支；不将这些分支都称作缺陷RED。

44个新公共用例中42个非法输入零Provider/Attempt拒绝，2个合法业务JSON保持原值发送。定向普通/race各1663PASS/8NOT_RUN；同8项mandatory以显式开关另各8PASS，双vet0、审计0。测量为ca67+WIP，1027个编译/JSON/mod/sum文件等于2306；不是提交后重跑，也不是完整1056运行来源冻结。

原两项留存Spec问题在补正复审中关闭；新夹具构造器的后续读取边界由[最终Spec复审](retention-factory-final-spec.json)单独闭合，零发现。原报告与原件不改写。

代码审查的零发现不替代必交真实验收。当前GLM Messages失败、严格续期/装配拒绝、完整门禁false及修复后的严格当前来源证明缺口均保留，09、10和spec继续开放。

## 来源适用性

[当前来源核对JSON](current-source-applicability.json)与[说明](current-source-applicability.md)以源码2306917为对象。M3三组263→264、GLM三组217→218、文本两组执行261→262、protocol862→865、视觉两组323→324均不满足原严格来源相等性；文本原始860文件未改，派生Jev继承protocol/visual约束。

原报告继续证明其测量来源的历史结果；不刷新原hash/map/selector/catalog/矩阵时间或授权。generic10依据原受影响组合规则，核对SDK/catalog/models/live producer及非MiniMax路径字节不变，不以缺map作为豁免。来源7931的完整门禁不认证后续2306源码。[最终源码字节相等封印](current-source-byte-equality.json)覆盖共1117个文件（ai源码/数据及go.mod/go.sum），完整集合含新增/删除检查，并核对原1027编译/JSON/mod/sum证明；只排除独立审查的两个README。它扩展2306来源适用性核对，不更新旧native资格。

## 保留与复核

大型原件位于主仓库`.evidence/minimax-retained/`。各原件入口记录完整清单、公开有界无Key复核、实际退出码与私有副本拒绝分支；不会执行归档中的任意代码来取信其声明。全量reader的256MiB压缩、2GiB解码、64MiB单member等上限保持。

该入口的报告是原始审查输出的字节副本；独立复审、最终合入与清理记录在同目录索引中。用户原有session-analysis及其他功能human-review目录保持。

源码/文档/留存最终未修复发现：Standards0，Spec0；必交实际验收仍未通过。
