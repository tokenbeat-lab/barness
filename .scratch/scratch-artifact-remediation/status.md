# 普通测试输入治理当前状态

日期：2026-10-24（Asia/Shanghai）。**工单01：COMPLETE；普通验收 PASS。** 其他治理阶段尚未实施；本功能完整发布判定仍为 false。

## 实现与来源

- 实际受测提交：`214cff42fdf72d6cd00dfd76bdd80f57999d8b1c`。收尾仅更新本状态、工单、说明链接和验证指针；当前提交与受测提交的完整执行来源/选定输入相等性见本轮只读报告及独立核对。
- 三处原证据测试改用 Git 跟踪的固定输入：Jev/Decisions 原字节 ZIP+manifest；runtime 保留原首个完整包的六个撤销场景及全部原件，派生 manifest 与历史身份明确区分。
- 干净全包额外发现 GLM 六份原归档只在宿主 .evidence。现按原 SHA 交付到 ai/live/testdata/glm-formal/archives，公共消费者和测试 helper 直接读取；原 index/pin/报告不改、无宿主路径兜底、无重新生成历史证据。
- 新固定输入552文件、13,551,018字节；1909份执行来源、8757份保守选定历史输入。历史 scratch 字节全部不变；原生 GLM index/usage/pin 字节不变。[依赖清单](input-dependencies.json)记录调用与存储边界。
- 分支：`codex/retained-ordinary-inputs`。基线571ca112、首片640476c、GLM补片214cff4；原有未跟踪 SOP 回归与审查目录保留，未纳入提交。

## 实际验证

- 独立 Git clone 只检出受测提交的跟踪文件；可共用只读 Git 对象与 Go 依赖缓存，没有 .env、宿主历史 .evidence 或原三份忽略 ZIP。所有552个固定输入均受 Git 跟踪且与主工作区同字节，测试前后 git status 均为空。
- 实际 argv：`go test -json ./ai/e2e ./ai ./ai/internal/testkit/release ./ai/release/cmd/releasegate ./ai/internal/testkit/supportmatrix -count=1 -failfast -timeout=20m`。SOP采集器固定GOENV=off、空GOFLAGS，隔离HOME并清除凭据/live/race/压力配置。
- 退出0，墙钟446.40秒；14,943个Go通过节点。710个预发现E2E顶层入口全部有实际结束事件：705 PASS、5 SKIP。四个实际测试包PASS；releasegate包无测试文件，真实事件为SKIP，不冒充该包有测试通过。
- 唯一主原生包：14,071 PASS / 214 NOT_RUN / 0 FAIL；58项预选必交case全部PASS，未缩减断言；全部辅助原生包及audit通过。214 NOT_RUN与5个Go SKIP单独保留，不计为完成。
- [最终普通回执](../../.evidence/scratch-artifact-remediation/e2e/clean-final/.evidence/scratch-artifact-remediation/e2e/20261023T222836.128993Z-nq6l0f8o/receipt.json)，SHA256 `c08a2e7fa6e4d480bc2b3a9fd051afaa8dd4ef56d04352b0c0f5f7a06368d270`；[主manifest](../../.evidence/scratch-artifact-remediation/e2e/clean-final/.evidence/scratch-artifact-remediation/e2e/20261023T222836.128993Z-nq6l0f8o/bundles/20261023T222841.672172000Z/manifest.json)，SHA256 `0f5d390e9095756a4994216b9268371080e98165d1cf59720279e03aac4445af`；[audit](../../.evidence/scratch-artifact-remediation/e2e/clean-final/.evidence/scratch-artifact-remediation/e2e/20261023T222836.128993Z-nq6l0f8o/bundles/20261023T222841.672172000Z/audit.json)，SHA256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。
- [完整统计与干净性核对](../../.evidence/scratch-artifact-remediation/planning/clean-final-summary.json)、[执行前库存](../../.evidence/scratch-artifact-remediation/planning/clean-final-preflight.json)及[原件/引用检查](../../.evidence/scratch-artifact-remediation/planning/integrity-check.json)分别保存来源、实际case集合、包退出和输入哈希。
- 两份ZIP经原公共verifier --audit，runtime在独立副本审计，四命令exit0；六GLM归档在正确原生bundle根做独立审计，七命令exit0。原固定夹具未被审计器写回。回执见input-audit-qnvshf7a与glm-input-audit-7zl4hw8f的summary。

## 失败与先红后绿

- 基线干净clone定向RED：20261023T222133.479214Z-yv2tja9a，45FAIL/1PASS/audit0；同46个case定向GREEN为20261023T222316.990694Z-7ufuosop，46PASS/audit0。
- 首候选640476c的完整五包FAIL：20261023T222503.230183Z-9fwf2u8f；首个TestGLMFormalReports/chat报immutable archive unavailable，原件保留，三个其他测试包成功不冒充全包完成。
- GLM独立定向RED为20261023T222711.380852Z-lqtkxntt；同12case定向GREEN为20261023T222744.357964Z-shw5hgib，12PASS/audit0；之后取得上方新冻结来源完整干净五包PASS。
- 补充GLM审计首次选错ZIP外层目录，因无manifest实际exit1；原失败glm-input-audit-0cen1zpl保留。修正为各report所在原生根后新attempt全部通过，不改写首次退出或原件。

## 阶段与只读判定

- live：NOT_APPLICABLE（本次普通输入可交付性验收）。改动仅固定历史字节和消费位置，不改变厂商wire/型号/价格/目录/晋升规则，不请求厂商。GLM消费者源码有变化，旧历史证据对当前来源仍被拒绝；没有续期或授予新的正式资格。
- race：NOT_RUN，未有人类对本次范围的选择。候选为上述五包加 `-race`、`-timeout=75m`，覆盖本轮1909来源/8757输入/58必交case及完整包；实际耗时未测，须独立选择后执行。
- 压力：NOT_APPLICABLE；不改变运行时资源策略、并发/复杂度限额或设计负载，普通消费者只读固定夹具。真实清理、付费调用均NOT_RUN。
- release.pass=false：缺人工race。新的[不可变回执索引](../../.evidence/scratch-artifact-remediation/record-20261023T223511Z-pkea1jy5/record.json)及[只读报告](../../.evidence/scratch-artifact-remediation/record-20261023T223511Z-pkea1jy5/report/release-report.md)消费既有事实，不补跑阶段；判定实际命令与数字退出另留在本轮目录。
- 地区recovery7暂停、历史账本和支持矩阵保持；不启动第2项治理。回滚按三处读取/固定输入及GLM存储读取成组回退；新旧失败和授权原件均保留。
