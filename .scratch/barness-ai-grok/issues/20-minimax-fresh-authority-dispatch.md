# MiniMax 正式资格器的新进程授权分派

Type: task
Status: resolved

## Contract

[Spec](../spec.md) R21/R23；[测试 SOP](../../../docs/agents/verification.md)。只修复正式资格消费将 fresh ProcessGrant 误送入 legacy origin 校验的问题。保持真实原生场景、模型、SDK、预算、调用/重试、源/二进制/账户/地区与完整产物校验；不改真实旧回执，不把不完整或失败场景标成通过，不发付费权限。

## Failure

v6 第8条 MiniMax 国际 Chat 和第16条 Responses 各自14场景原生PASS（含自身非空推理），正式 minimaxverify -formal -publish 实际exit1：MiniMax actual origin launcher receipt invalid。独立诊断精确定位 Verify(mode=live) 在既有 fresh ProcessGrant 验证前无条件调用 legacy verifyOrigin；旧one-shot receipt SourceCommit89000346…与真实本批50760138…不等。本批 argv/cwd/native SHA/grant均匹配批准配置，env launcher SHA不是此拒绝谓词。不存在可选旧origin的合法invocation修复，也不得手写/覆盖旧launcher receipt。

[完整实际批次检查点](../../../.evidence/barness-ai-grok/live-batches/gf6-20261026T233415Z-ceb8d8c1/completed-checkpoint-v1.json)；[只读诊断](../../../.evidence/barness-ai-grok/live-batches/gf6-20261026T233415Z-ceb8d8c1/read-only-diagnoses/barness-grok-v6-minimax-launcher-diagnosis.md)。第13条 Messages另有真实流读取失败，不能靠本票修复或降低门槛消除。

## Ownership

单一implementer在独立现有native-proof工作树，先同步集成分支507601386083f1d21d83875a8ce46c892117ed54。拥有最小MiniMax正式资格authority dispatch、公开CLI/进程E2E与必要受控夹具/契约说明/局部证据摘要。你不是代码库中唯一工作者，不得回退他人编辑；按集成tip协调。不拥有root spec/map/05/06/20状态、真实授权/grant/账本/原件、目录或其他Provider。root负责合并、独立双轴、完整冻结及后续人类边界。

## Failure modes and verification

先列全失败方式，在公开CLI/进程边界提交实际RED，再实现。fresh正确ProcessGrant/claim自身证据完整时不依赖legacy旧origin；fresh grant/claim/源/二进制/scope/账本/claim身份/目录/完整场景任何缺失、篡改、错配仍拒绝；fresh不合法不得退回legacy；无fresh授权的legacy报告继续要求原origin，不新增兼容shim；controlled不得冒称真实live或正式发布。保持现有完整性、终态、推理、资源与用量门槛。先调查现有消费合同和夹具，仅按真实第二权威用例建立清晰分派，避免复制验证逻辑或并行宽松路径。

测试使用既有受控公共流程/私有Git common授权目录，无真实Key/厂商/用户grant或共享真实store；不要用真实旧receipt/账本写入RED。允许必要目标普通run.py及受控原生编译，保存命令/数字exit/source/manifest/audit。不得启动完整普通、live、race、压力或付费消费者复跑；不得清理/重置任何已消耗权限。500行文件边界与AGENTS/SOP适用。

## Comments

- 2026-10-27：v6已一次执行全部17条，14资格通过、2资格拒绝、1Messages原生流失败；270逻辑/268HTTP。当前无新的付费重跑权限，所有历史成功/失败原件保留。此票仅授予原实现任务所需离线修复；修复后来源适用性须独立评估，不能预报旧live已对新源合格。

## Answer

已完成最小新/旧授权分派及foreignmodel防绕过，公开三协议CLI真实RED后45场景GREEN；既有53主/6辅助回归PASS，v15两轴0/0。实际10326/2064来源完整四包16,033 Go测试、1568主/129入口/38辅助PASS，642.833秒/exit0/audit0。[完整核验](../../../.evidence/barness-ai-grok/final-ordinary-validation-v10.json)及[当前状态](../spec.md#唯一当前状态)。软件关闭不授予付费或当前新源正式资格；旧v6原件和3条失败保留。
