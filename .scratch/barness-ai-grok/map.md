# Grok 实施工单图

## Notes

集成分支：`codex/barness-ai-grok`。契约：[spec](spec.md)；唯一当前状态：[spec](spec.md#唯一当前状态)。

依赖：01 → {02, 03, 07}；{02, 03} → 04；{04, 07} → 08 → 09 → 10 → 11 → 12；10 → 13 → {14, 15}；{12, 14, 15} → 05 → 06。07为独立候选入口兄弟工单。每工单在独立分支/工作树实现，合并到集成分支。

## Decisions-so-far

- [01 已解决](issues/01-foundation.md)：`04474e07` 交付候选/目录/目标授权；保留公共 E2E red/green，正式目录仍未激活。

- [04 实现已解决](issues/04-host-verification.md)：`c924c311`补齐48个公共Host/runtime及验收断言，定向普通回执PASS；已有生产实现通过，未虚构行为Red。全局两轴审查、完整来源冻结/普通包与后续资格仍由集成阶段完成。04不依赖07，真实资格05等待两者汇合。

- 01–04 交付候选实现与普通验证；05/06 按 SOP 另取 live/race/压力授权，未满足前不能关闭或宣称正式支持。

- [08 审查修复已解决](issues/08-review-fixes.md)：`51c95922`修复两轴审查五项发现，逐项保留公开 red/green；冻结定向范围605测试/564case通过。独立复核与完整集成普通回执仍待完成，05还需后续冻结及授权。

- [09 规范证据标签已解决](issues/09-canonical-evidence.md)：真实原生消费者发现空 effort 产生中间空目录段；标签改为default并加实际producer→collector→CLI回归。`8f77ecc8`冻结定向65顶层/565case通过，native source/e2e均true；旧错误原件保留，05等待新的集成冻结。

## Fog

- 历史候选v3已确认真实流形状、非零reasoning、图片与request-ID；当前最终冻结来源仍需17条正式live，六条Kimi仅保留受控/候选，不借旧pin完成正式资格。

- [10 JPEG夹具已解决](issues/10-live-jpeg-fixture.md)：真实proxy运行暴露8×8像素过小；`00e7612c`改为32×32，用户/工具JPEG受控捕获先RED后GREEN，两轴v4复核0/0，新冻结完整普通565/565通过。修复后付费运行待新授权；05/06仍未关闭。

- 自身候选v3真实18/18已通过；05余项拆为11公共真实捕获回放与12自身正式合同。12等待[晋升边界顺序选择](formal-qualification-plan.md)，11可独立推进；不以05整个工单阻塞其子任务形成环。

- [13 已解决](issues/13-request-id-adoption.md)：真实24响应确认x-request-id后接入，34新增公共case先RED后GREEN，全部xAI568case通过；最终全局审查/冻结仍待12/14完成。12 producer已合入，consumer与14旧目标授权/来源历史合同并行实施；未启动新付费或人工阶段。

- [12软件已解决](issues/12-formal-native-contract.md)、[14已解决](issues/14-legacy-xai-history-proof.md)、[15已解决](issues/15-formal-review-fixes.md)：最终源码f89e59a6、v7两轴0/0、完整受测97697bb2，16184 Go测试与1456必交/119入口通过；原件全部逐SHA归档。当前frontier为05等待[17路线付费计划](formal-live-plan-v1.md)的明确授权；06仍被05及后续人工阶段阻塞，release.pass=false。

- 真实收尾消费者随后拒绝新GLM包装case的4350项库存（原上限512），[15已重新打开](issues/15-formal-review-fixes.md)。保留完整Go PASS与实际CLI source=true/e2e=false原件；采用既有辅助bundle边界修复，导入规则不放宽，05继续被15阻塞。

- [15最终已解决](issues/15-formal-review-fixes.md)：辅助bundle归属fix421efc23、v8复核0/0、新受测812782ec完整三包15837测试/1457主必交/120入口及32辅助均PASS；真实readonly-v6 source/E2E均true。05现等待[计划v2](formal-live-plan-v2.md)的明确付费授权；不自动启动17路线或人工阶段。

- 正式批次已按授权停止：xAI正式18/18通过；OpenAI9/9实际调用通过但非空消息证据解码被消费者拒绝，后15条NOT_RUN。[16修复前沿](issues/16-generic-capture-decoding.md)只处理离线消费，不启动新的付费调用；[当前状态](spec.md#本次正式17路线已授权并按失败规则停止)。

- [16已解决](issues/16-generic-capture-decoding.md)：局部DTO修复5f933af9、mandatory-history回归fd02b56a，12项CLI回归与v10两轴0/0，受测81bdaeaf完整三包15,850测试/1469主必交/121入口及32辅助PASS。当前frontier回到05等待[新来源计划v3](formal-live-plan-v3.md)的新一次付费选择；旧批次停止原件与39次调用保持，不自动续跑。

- v3正式资格在OpenAI Chat声明的UNSUPPORTED/零请求语义处拒绝，后13条停止。新[17](issues/17-declared-unsupported-qualification.md)负责原生合同一致性；前三条资格与65调用原件保持，05继续blocked17，未授权新付费重跑或人工阶段。

- [17已解决](issues/17-declared-unsupported-qualification.md)：保留唯一原生UNSUPPORTED零活动合同、三份公共CLI接线RED/61项定向GREEN、v12两轴0/0；最终完整四包15,982测试/1518必交/125入口/32辅助PASS。05等待[新来源计划v4](formal-live-plan-v4.md)明确付费选择，06仍待有效live后人工race/压力；旧正式104调用及失败原件不续跑、不改pin。

- v4限定续跑后7条QUALIFIED，第8国际MiniMax Chat实际13 PASS/1零推理正例FAIL，后9停止。新[18探针维护](issues/18-minimax-reasoning-probe.md)只改善thinking-on任务，保持自身实际推理正例门槛，先受控RED后实现；真实137逻辑calls/136HTTP及所有原件保留，未授权新付费或人工阶段。

- [19隔离修复](issues/19-controlled-grant-store-isolation.md)：18补充验证漏查既有FreshLedger helper的Git store隔离，实际追加3个synthetic ID/6个合成guard，未读真实Key/未请求厂商；偏差记录保留不删除。与18同一实现者在私有namespace先RED后修复受控调用，不改真实授权合同或扩大产品范围。

- [18](issues/18-minimax-reasoning-probe.md)/[19](issues/19-controlled-grant-store-isolation.md)已解决：M3仅thinking-on有界任务，私有Git store/真实第二进程拒绝补齐，v14两轴0/0；完整普通15,987测试/1522主必交/127入口/38辅助PASS。当前frontier回到05等待[计划v6](formal-live-plan-v6.md)新一次明确付费与失败策略选择；06人工阶段仍后置，旧241/240正式调用与全部失败/偏差原件不改。

- v6一次17全部完成：14资格通过、2国际MiniMax旧origin资格拒绝、1Messages真实流失败；270calls/267HTTP，238必交236PASS+1声明UNSUPPORTED+1FAIL。新[20](issues/20-minimax-fresh-authority-dispatch.md)离线修权威分派，05被20及实际失败阻塞，06后置；不自动重跑或复用本批权限。[当前状态](spec.md#唯一当前状态)。

- [20已解决](issues/20-minimax-fresh-authority-dispatch.md)：fresh/legacy分派保持严格原生验证，公开CLI45新case和既有53主/6辅助PASS，v15双轴0/0；实际10326/2064来源完整16,033测试/1568主/129入口/38辅助PASS。05等待[计划v8](formal-live-plan-v8.md)新一次17付费授权；06人工阶段后置。旧v6及全部失败/guard保持，不自动重跑。

- v8获17一次授权但165匹配普通实际FAIL：两GLM公共包装10秒截断，native32断言PASS仍无完整定稿且有晚写。新[21](issues/21-glm-controlled-process-budget.md)修测试时限/子树终止；真实路线0启动/0调用/0claim，原权限未用且不得自动重绑，05被21阻塞。

- 21局部GLM/取消回归已取得实际通过；完整工具49项出现旧Google合成夹具缺Grok历史证明的公开CLI失败，拆[22](issues/22-tool-fixture-grok-history.md)仅修夹具并保持消费者严格拒绝。05同时被21/22及后续完整冻结验证阻塞，当前正式资格仍见spec唯一状态。
