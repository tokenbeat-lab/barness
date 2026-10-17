# 14：当前目录发布整合验收

2026-10-17。整体发布 **BLOCKED / FAIL**，工单保留 **needs-info，未 resolved**。
整合代码和受控验证已交付；当前整版仍有两个实际测试失败、一个 Decisions NOT_RUN 包清单缺陷、
未运行的真实/效果证据，以及原 #07 尚未满足的因果验收。本次真实 Calls/Attempts 均为 0。

## 实现与版本

当前 `main` 的实现提交为 `ceb4625`、`1ded61e`、`0f5d99a`、`823cc4a`、`4310a53`；
`b9f1574` 更正追溯来源，`f3f6ebb` 保存收集器及独立复核。比较起点为
`bf9e8570cc4d3bff97bf37076f0e8321335c6692`。最终完整运行的执行提交为 `4310a53`，
之后仅修改验收文档与证据收集脚本，Go 源码保持一致。

正式目录为 `2026-10-17.1`，Hash：
`sha256:ae3315ce504af2019ce6476f3b021093e13428a5992b3e5fc7d3208576eb49c7`。
目录快照、支持矩阵、Pi 锁文件/来源和差分账本与比较起点逐字节一致。

- 11 个独立路线身份保留旧 Jev 1.13、新 Jev 1.14 与 Decisions `gpt-6-luna`。
- R01–17、T01–16、17 条相关用户故事固定完整必需用例；50 条追溯在两阶段均无缺项。
- normal/race 分别导入 plain/hooked、四并发、双租户、双协议图片负载，完整对账输入、
  图片 Hash、原生请求/响应、命名答案、usage、Attempt/Observer 和资源释放。
- 原 Pi 1.1.0、P07/P10 投影及账本保持；新语义以有限 case/path 登记，不扩大豁免。
- Jev/Decisions 分别要求当前正式目录的独立视觉任务；数据、评分、每路线
  12 Calls/12 Attempts、零重试预算不变。受控 COMPLETE 的真实效果仍为 NOT_RUN。
- 原 #07 的状态由原工单决定；临时源码副本已覆盖 unresolved/resolved 两种状态，
  原工单、父规格及 ADR 状态保持。本轮没有部署。

## 实际执行结果

证据根目录：`.evidence/ticket14/final-4310a53/`。

| 检查 | 实际结果 |
| --- | --- |
| vet / live-tag vet | 均 PASS |
| 完整 normal | 9,094 用例：9,093 PASS、1 FAIL；约 487.6 秒 |
| 完整 race | 9,094 用例：9,093 PASS、1 FAIL；约 1,860.6 秒，最终审计已完成 |
| UC14 公开入口/CLI 专项 | normal/race 均 33 PASS |
| 冻结 Pi 差分 | 两阶段各 861 PASS，pending 0；包含在上述总数中 |
| 两入口图片设计负载 | normal/race 共四组均 PASS；资源全部归零 |
| 主 normal/race 包脱敏审计 | 均 0 发现 |
| live-tag 受控 harness | PASS |
| 11 路线独立真实预检 | 全部 NOT_RUN；真实 Calls/Attempts 0 |
| 两文本、两正式视觉效果包 | 独立 verifier 四次 exit 0；真实效果仍 NOT_RUN |
| 最终发布命令 | exit 1，`release-report.json` 的 pass=false |
| 聚合清单/审计门禁 | FAIL：Decisions NOT_RUN 场景缺少对应清单 case；脱敏发现数仍为 0 |

race 命令使用 `-timeout=30m`；Go 在 TestMain 收尾阶段还有工具退出宽限，实际墙钟耗时
约 31 分钟。没有变更任何业务场景的负载、预算、截止时间或断言。

图片负载的 TotalAlloc 分别为 2,843,520 / 2,991,952 / 2,996,736 / 3,308,048 bytes，
均低于 64 MiB；最大限额利用率 0.75，保留至少 25% 余量。完整数据见
`multimodal-capacity-summary.json`。

## 未满足的发布条件

1. normal 的 `E08-policy-pressure-cloud-interactive-design-load` 失败，捕获到 socket 写入
   `ENOBUFS`（no buffer space available），31 条断言失败，其中 10 条直接记录该写入错误。
   原压力流完整性工单 #07 的既有因果验收仍未满足；本轮没有修改或关闭它。
2. race 的 `P07-E08-unary-timeout-body-call-deadline` 失败：预期 stop reason `aborted`，
   实际 `error` / `ReadIdleTimeout (100ms)`。相关测试源码与比较起点一致，原因待定位。
   跟踪：`.scratch/barness-ai-release-verification/issues/01-unary-race-timeout-stop-reason.md`。
3. Decisions 零凭据预检声明 9 个 `LIVE-P10-*` NOT_RUN 场景，但清单无对应 case，严格
   完整性校验据实拒绝。跟踪：`.scratch/barness-ai-release-verification/issues/02-decisions-notrun-case-inventory.md`。
4. 当前没有本次获准的真实服务凭据与有限预算。旧 Hash、候选、受控 HTTP、另一模型或
   历史 PASS 不能替代当前各路线真实/效果证据。原工单 13 的旧真实包另由既有 verifier
   拒绝为 `formal source set drift`，见 `historical-applicability/result.json`。

所有失败保持原始结论，没有以专项重跑覆盖本次失败。复查命令可能受调度/资源状态影响：

```sh
BARNESS_AI_PRESSURE=1 go test ./ai/e2e -count=1 -run '^TestPolicyPressure$/^cloud-interactive$/^design-load$'
go test -race ./ai/e2e -count=1 -run '^TestUnaryTransportTimeouts$/^body-call-deadline$'
```

## 保留、复现与核验

`14-evidence.json` 固定完整 Go 源码集合、目录/配置/fixture/基线、当前主包清单及审计、
预算、实际命令和完整存储字节索引。索引位于 `.evidence/ticket14/integrity/`。
嵌套捕获和故意损坏的夹具属于原始证据，不作为独立运行的 PASS；符号链接仅记录目标字符串
的 Hash，不跟随或读取目标。主包清单继续拒绝链接。

live-tag 故障 harness 的两个故意中止子记录器保持 MISSING 审计；它们的逐字节相同副本
另经公开 auditbundle 审计，均 0 发现，见 `.evidence/ticket14/child-audits/`。
原子包没有补写 audit.json。历史失败、缺审计、编译诊断和中止运行均保留。

此前 `final-1ded61e` 的 normal 为 9,081 PASS，race 用例也有 9,081 个终态 PASS，但
审计未完成即在旧 20m 参数的退出宽限后被终止，整条 race 命令 FAIL。
`final-823cc4a` 在修正原工单状态转换断言前主动中止，作为部分历史保留。

```sh
# 输出目录必须不存在；只继承工具链必需环境，不读取任何路线 key。
python3 .scratch/jev-1.14-multimodal-box/run-14-verification.py .evidence/ticket14/NEW_RUN --gate
python3 .scratch/jev-1.14-multimodal-box/seal-14-evidence.py --record .evidence/ticket14/NEW_RUN
python3 .scratch/jev-1.14-multimodal-box/seal-14-evidence.py --verify
```

runner 为收集完整失败结论可以正常退出；应查看内部 `commands.json` 和发布报告，
不能将 runner 退出码视为发布通过。其整体命令上限为 60 分钟，超时会终止并回收进程组。
当前回执针对本次零凭据 BLOCKED 运行，不是通用的真实验收签名工具。

## 独立复核与后续

`14-code-review.md` 分别记录 Standards / Spec 两轴复核，剩余代码发现均为 0。
语义篡改、单例伪全组追溯、视觉包删除、超时子进程、链接目标变化及原工单状态转换
均保留失败先行与修正后的公开进程证据；没有在实现后新增单元测试。

公开迁移、实验选项、格式/位置拒绝、图例/索引和版本化撤销见
`docs/barness-ai/current-catalog-release.md`。回滚按本轮提交整体恢复合同与消费逻辑。
取得各路线明确授权和有限预算、解决上述实际失败及清单缺陷、满足原 #07 验收后，
再生成当前源码/目录证据并执行完整门禁。本票不声明功能真实支持或整版发布完成。
