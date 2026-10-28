# Kimi 晋升入口消费固定九加一

Status: ready-for-agent
Type: task

## 问题

`ai/internal/testkit/supportmatrix/kimi_promotion.go:37` 仍调用 `regionalReport(dir,false,mode)`。后者只读取 `regional-report.json` 并要求同包完整十项，不认识仅有 `kimi-composition.json` 的固定9+1目录。第45行还要求六路单包同source/build/binary/commit；第51行的 `RegionalPromotionProof` 只能记录一个测量。实际旧九项来自多个历史提交，新单项有自己的二进制与时间，不能压成统一来源。

`releasegate` 的inspect、native消费者和record已支持 `VerifyKimiComposition`，目录晋升入口没有分派。即使六份真实补项都成功，直接把六个组合目录交给旧入口仍会拒绝。此为静态调用链确认的软件接线缺口，不是厂商产品变化、缺原件或新增付费场景。与ADR-0055允许固定两来源的决定不一致。

## 最小后续实现范围

只修改Kimi专用晋升路径及其公共CLI测试：消费现有 `VerifyKimiComposition(dir,false)` 和完整输入闭包；六个combo各一次、mode必须live，检查各路线当前相关型号/SDK/账户事实。每路线保留两次原测量、coverage和producer/账本引用，以Kimi自己的晋升证明表达十二份测量，不伪造一个十项Report或六路统一commit。受控模式只生成不可激活审阅材料。

复用现有独立Kimi目标、预期inactive哈希、独占intent/原子replace边界；不改其他地区九型号的激活合同、不新增通用N+M框架/恢复控制器、不借旧权限。优先在付费补采前完成这项修正与普通受控验证，之后按实际来源变化重新冻结并更新尚未批准的七份绑定。旧54项、MiniMax原13项和其他Provider的厂商事实不因此失效。

## 实现前失败清单与验收

- 六个真实布局的受控9+1输入经公共 `regionalpromote -kimi -controlled` 入口先RED后GREEN；证明输出仍不可用于生产激活。
- 旧九项或新单项独立、五条/重复/跨路线、混合mode、篡改基包/producer/可信账本、来源或SDK/型号事实不适用均拒绝且零生产目录写入。
- 不因两次合法测量的commit/source/build/binary/时间不同而拒绝；仍分别核验其原时刻授权及完整原件，不能用全局相等代替适用性。
- 保留每路线两次测量及10项coverage；真实尚缺时不伪造live正例、不签付费授权。
- 相应普通/CLI回归及来源核对通过后，仅重新绑定未批准二进制和scope。实际激活、真实补采、race/压力继续独立按用户选择。

本轮只完成离线审阅和问题规格，未修改上述代码、未执行测试、未调用ActivateKimiCatalog。
