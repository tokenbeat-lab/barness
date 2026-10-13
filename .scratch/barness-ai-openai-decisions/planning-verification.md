# 设计与已发布工单的文档核验

Updated: 2026-10-12
Scope: planning documents and published ticket structure only

用户已批准 13 张草案并替换原 9 张；当前正式票各自为 ready-for-agent，未执行的生产功能、live、业务效果和发布仍未完成。主要测试 Seam 继续为公共 Client.Classify / HookedClient.Classify。

本次核验包括：父 spec 的原字节与 56 条 User Story、R01–R20；OD01–OD58 连续唯一；正式 01–13 独立工单与批准草案的行为/验收一致；阻塞号/标题、前序存在、无环及无冗余直接边；每个验收 ID 在票、计划和矩阵归属一致；正式状态和原 Comments 来源保留；相对链接存在；单文档最多 500 行。

原 9 票和原配套文档完整保存于[历史快照](ticket-review/originals/README.md)，逐文件原始 SHA-256 核对；spec 和 design 保持原字节。批准前[草案文档核验](ticket-review/originals/ticket-review/validation.json)保留当时结果，新发布状态由[发布记录](ticket-review/publication.json)明确记录。

从 repo root 重放：

~~~sh
python3 .scratch/barness-ai-openai-decisions/ticket-review/verify.py
~~~

脚本只读核验文档，写出[validation.json](ticket-review/validation.json)及 SHA-256 manifest。结果仅是文档结构 PASS，不运行生产 E2E、差分、race、真实 API、压力或发布 gate；这些运行须各票独立形成实际证据。

初始可实施 frontier 只有[01：Jev 回归基线](issues/01-jev-regression-baseline.md)。后续按各票 Blocked by 和实际 resolved 状态推进；ready-for-agent 不表示所有阻塞项已完成。
