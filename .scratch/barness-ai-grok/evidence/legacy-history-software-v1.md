# Ticket14 离线软件交接 v1

Status: SOFTWARE_PREPARED_REVIEW_PENDING；不是发布回执，也不授予厂商调用权限。

软件 pin：`9ea3b9d3d72c37c538971c7fafa40ade5c883fa7`，已合根 `51eb2a2482a58030ef214d2aa4e4c3817d02422e`。源码自此未修改，后续仅文档/索引收尾。完整2,048项执行来源 map、85个本票软件/合同改动文件及当前 SHA、精确 argv、24次保留回执、120个主/辅助 manifest 与 audit、来源事实和156项有限新增 case 见[机器索引](legacy-history-software-v1.json)。该索引 SHA256：`44518976781d125038606a181cd205b71b124b3a8e6ace6ba090c31d2172d70a`。

## 已实现的公共行为

`grokhistory` 从公共 Host/Message 构建固定合成 xAI 历史：同轮两对长前缀/规范化碰撞的 opaque pair、后轮复用第一完整 pair、三个不同可见工具结果标记。22条实际目标 Client 路径验证三个合法唯一关联、结果配对和顺序、可见历史、两个 cross-model downgrade 与原始请求不变。另保存调用前后公开 `Envelope()` 必要字段，原生消费者在4096字节内独立校验，弥补 Message JSON 省略可信状态的证据缺口。安全编码的 MiniMax/Kimi Responses 关联不被误判为资源 ID；Responses 自身的 `include:["reasoning.encrypted_content"]` 保留，历史中的实际 foreign cipher/resource 字段仍拒绝。

所有完整原生合同都加新场景，不接受缺失、替换、篡改、UNSUPPORTED 或错误完成作为新增能力通过。既有 optional 语义和 FAIL/NOT_RUN 的部分报告语义保持。Generic6、GLM2、国际M3三条取得新固定 scope/单进程 grant，绑定 source、嵌入 build source、binary、SDK、目录、完整 models/scenes/budgets、账户/地区及期限；producer 在第一次HTTP前用 common Git 独占持久 claim。原始 grant 字节、claim、manifest source 和独立消费者互验。当前 M3 拒绝旧 ledger/closed renewal selector，不恢复代次或预算。

当前正式范围为 Grok+16已发布旧路线，共17条。六条 Kimi 保留 candidate/controlled，不晋升；旧9call Kimi grant 已拒绝，候选10call合同及131072输入预留证明已准备。本文所有新运行均离线/合成，无 `.env`、真实密钥或厂商请求。

## 测试事实与限制

- 合并后定向 **PASS**：[084150回执](../../../.evidence/barness-ai-grok/e2e/20261026T084150.094299Z-g5i5d3s_/receipt.json)，exit0，48.882秒；十个 `TestGrokHistory*` 根、ChineseEvaluation、Usage、Snapshot，265个测试/子测试名称，101项本票新增case均PASS，主包及audit有效。回执 SHA256 `26d1e59c2a11efa35e62086a620fd096dde8db6f29a033275be7d8a35857c5f1`。
- 固定同来源的有界历史正例 **PASS**：[084711回执](../../../.evidence/barness-ai-grok/e2e/20261026T084711.661876Z-hnsepeqb/receipt.json)，exit0，45.410秒；实际仅 `TestFinalRenewalAuthorityProcess/valid-three`、`TestRecovery4AuthorityProcess/valid-three` 与 `TestGrokHistoryProducerBuild` 主包。回执 SHA256 `9e336245f31f1e46ad2b2f8078b30ec3ddf8c3f8de566c74c99c097fd80afa57`。它不是完整 minimaxlive 包通过。
- 本票全部156新增case曾在[083305定向GREEN](../../../.evidence/barness-ai-grok/e2e/20261026T083305.519965Z-651c3los/receipt.json)通过，包括55个 regional/Kimi 完整消费者篡改case；其源早于最终根15合并，不能冒称当前完整冻结。根最终集成验证需覆盖此完整集合。
- [080940受影响普通尝试](../../../.evidence/barness-ai-grok/e2e/20261026T080940.049899Z-7kxgpk_x/receipt.json)保留 **FAIL**：`./ai`、supportmatrix、release 完整包PASS，E2E只有后来由15修复的ChineseEvaluation/Usage目录pin失败；minimaxlive 在显式15分钟Go时限 **TIMEOUT**，不能称整包GREEN。根最终普通冻结另用30分钟Go/2400秒外层，未改变任何厂商预算。
- [083431探索尝试](../../../.evidence/barness-ai-grok/e2e/20261026T083431.519506Z-3v9kpaly/receipt.json)保留FAIL：10分钟超时、期间源码合并且无主E2E包；它不属于验收GREEN。[080840](../../../.evidence/barness-ai-grok/e2e/20261026T080840.331048Z-b_exsapv/receipt.json)为停止格式修复前旧进程的CANCELLED记录。
- 真实正式 live、race、压力、发布 **NOT_RUN**，`release.pass=false`。本票不声明完整集成冻结或当前付费资格。

受影响普通包：`./ai`、`./ai/e2e`、`./ai/internal/testkit/{grokhistory,livegrant,supportmatrix,minimaxlive,release}`、`./ai/live/cmd/{processgrant,regionalverify,minimaxverify,glmformal}`。live-tag完整producer通过公开E2E子进程无凭据编译；这不是最终付费冻结binary。

## RED 原件

预算真RED为 [064342](../../../.evidence/barness-ai-grok/e2e/20261026T064342.026216Z-ql5vpgd3/receipt.json)，六个candidate模板仍9/9/36864；更早063518/063829是既有formal Kimi未晋升的正确拒绝，不是预算RED。完整原生计划RED为 [064724](../../../.evidence/barness-ai-grok/e2e/20261026T064724.228156Z-392m666w/receipt.json)，公开Envelope证据RED为 [080218](../../../.evidence/barness-ai-grok/e2e/20261026T080218.328564Z-xkckgxoh/receipt.json)。最后两项 [083025](../../../.evidence/barness-ai-grok/e2e/20261026T083025.429638Z-tc_yjqev/receipt.json)复现 public Merge 接受新必交场景UNSUPPORTED，以及错误禁止Responses自己的include selector；精确RED测试源码SHA `69ac8e149c38fe21862292a461ef58d2bb35c44477a25a34ea978d1789339b88`，旁存 `red-test-source.go.txt` 和 `red-test.patch`。早期未跟踪文件导致空 `red-source.patch` 的事实保留，不把空patch当成功留存。所有已恢复RED源码均按原回执hash核对，见索引。

## 有限入口与原生case

十个新公开根及156个完整case ID在索引逐项列出：WIRE22、KIMI-GRANT6、FULL-PLAN16、FRESH-GRANT11、PREHTTP39、MINIMAX-FRESH3、NATIVE-MUTATION55、BUILD/GLM-CONTROLLED/MANDATORY-PROJECTION/FULL-SELECTOR各1。负例和所有受控捕获明确标为synthetic，不产生formal资格。

按路线计26个新增原生场景：Generic `LIVE-P01/P05/P04/P06/P02/P03-xai-origin-history`（依OpenAI Responses、DeepSeek Responses、OpenAI Chat、DeepSeek Chat、Anthropic、Gemini）；GLM两路线各 `GLM11-glm-5.3/5.2-xai-origin-history`；国际M3三路线各 `MM07-xai-origin-history`；Regional/Kimi使用 `REGIONAL-<route>-<model>-xai-origin-history`，精确列表见索引。历史harness子case `trusted-formal` 改为 `historical-formal-refused`；它证明当前入口拒绝旧authority。

原生新场景除完整既有inventory外，增加 `immutable-input.json`、`history-provenance.json`；必须关联原请求、wire、实际Result/metadata、downgrade、assertions、完整报告/manifest/audit/source与seal。新授权关联原始 `process-grant.json`、`process-authorization-receipt.json`、commonGit两个claim、`process_build_source` 和独立新账本。120个主/辅助manifest按路径和SHA绑定；包含故意失败/空case的子进程，不能将其全部解释为通过能力。

## 原生预算、来源与build

| 族 | 场景/路线 | 固有上限 |
| --- | --- | --- |
| Generic6 | 9/9/11/12/12/7 | 32calls/≤32HTTP；每scene至多1环境retry，Client自动retry0；4096/call；call180s/scene240s/process1800s |
| GLM2 | 24 | 24calls/32预留attempts、0retry；4096/call；call120s/process1800s/finalize20s |
| ZAI Chat/Messages | 20 | 24calls/24attempts、98304output/120000input；call120s/process1800s |
| MiniMaxCN3 | 10 | 同上述regional上限 |
| Kimi6候选 | 10 | 10calls/10attempts、40960output/131072input；完整编码预留proof保留 |
| 国际M3 standard3 | 14 | 20calls/24attempts、40000output/120000input；原1capability rerun和3transient request rerun有限规则保留；call120s/process1800s/finalize20s |

未受影响的ZAI Responses仍18，Zhipu Responses仍22。缺失usage保持unknown，没有合同的累计USD/输入硬cap不虚构。闭合历史M3保留独立冻结13场景oracle（39calls/36attempts三路线）；它仅供既有闭合读取，当前Plan/VerifyCapture、新授权一律14，不延长旧执行权。真实历史原件、pins、账本未改写。

固定9ea3下：Generic/Regional `cefc0218cc5bb7561ae0b7996272a9f5a1860a1ddf4fdefa374fbeeebf0cb5bb`；GLM `9f4a0bf2093810ebb1c75215c63c71bb3a4abbfbfcb230de801f8985d9c8e6ea`；国际M3 `b410b050fa909ed039684e5ff4359faa0496bf247fc373d5928bd99fd03108f9`。它们来自受控candidate只读scope/source命令，源码hash公式见[合同](../../../docs/barness-ai/legacy-history-qualification.md)；candidate GLM/M3的catalog不冒称formal。最终formal scope和binary必须在最终来源另行生成。Builtin为 `2026-10-26.1` / `sha256:98c6a6f0657ad7338fd6520411c6503c211ba2fd336282f4459d099db73091e4`。

冻结build模板：`go test -c -tags live -ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.legacyBuildSource=SOURCE_SHA256 -o FROZEN_PRODUCER ./ai/live`；显式 `GOENV=off`、空 `GOFLAGS`。`processgrant -source ROUTE -stage formal`无密钥输出精确flag和机器scope。Generic仅选定 `BARNESS_AI_LIVE_KEY_<ROUTE>` 加PERMISSIONS_FILE/ACCOUNT_ALIAS/REGION_ALIAS（均完整 `BARNESS_AI_LIVE_` 前缀）；GLM使用专用KEY/GRANT/ACCOUNT_ALIAS/REGION_ALIAS/PROXY；M3使用专用KEY/GRANT与原standard/entitlement/account/region/proxy字段，不继承ambient代理。详细完整ENV及source公式见合同，SDK/完整models在索引逐路线scope记录。未读取用户本机密钥。

[未授权正式计划v1](../formal-live-plan-v1.md)现有场景238/calls444/attempts472和本票预算吻合，仍待最终source/review/fullordinary/build/各路线新授权，不因本文变为获批。
