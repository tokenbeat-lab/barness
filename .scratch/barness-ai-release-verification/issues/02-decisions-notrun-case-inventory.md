# 02：对齐 Decisions NOT_RUN 生产者与逐场景清单合同

Status: needs-triage

来源：Jev 图片整合票 14 的当前正式目录零凭据预检。没有真实请求或收费调用。

`.evidence/ticket14/final-4310a53/preflight/openai-decisions/20261017T000555.185866000Z/`
的 `live-report.json` 声明全部 9 个 `LIVE-P10-*` 场景为 NOT_RUN，Calls/Attempts 均 0；
`manifest.json` 却不含这 9 个 case。当前 `verifyLiveManifest` 逐场景要求唯一捕获 case，
因此报告 `live artifact integrity verification failed`。脱敏发现数为 0，清单完整性仍失败。

这是原生产者与验证合同的不一致，不能靠改写旧证据、伪造已执行场景、跳过清单校验或把
NOT_RUN 升为 PASS 消除。原始报表与清单均保持不变。

调查入口：`ai/live/decisions_harness_test.go` 的未运行分支，以及
`ai/internal/testkit/release/live_manifest.go` 的逐场景检查。先在独立零凭据进程经公开
生产者/发布 CLI 复现拒绝，再按合同修复生产与验证的一致性；保留每个场景未运行原因、
零实际预算、完整清单与审计。NOT_RUN 仍不能满足真实服务发布门禁。

完整结论及原始字节回执见 `.scratch/jev-1.14-multimodal-box/14-verification.md` 与
`14-evidence.json`。不得借其他任务凭据或未授权发起真实服务请求。
