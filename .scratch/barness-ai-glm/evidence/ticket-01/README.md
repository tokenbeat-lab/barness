# 工单 01：既有 Provider 协议能力重构证据

以 `b2ad0c0` 的现有公共 Client E2E 为重构前基线。先运行请求、流、结果、费用及回调场景，再修改生产代码；重构不新增调用语义，复用既有外部断言，不新增 helper 单测。失败方式已由 spec A14 和现有 E2E 枚举：Responses 的 cache/summary/include 字段 presence 与头耦合、DeepSeek thinking 开关及 usage 回归、Anthropic 缓存/beta/thinking/tool/header 丢失、回放或回调重写行为变化。

结果和内容哈希见 [run-index.json](run-index.json)。原始脱敏 bundle 保存在仓库主目录的 `.evidence/barness-ai-glm/ticket-01/`；其中每个 manifest 含逐场景重放命令、SDK/目录版本、请求、脚本化响应、事件、Result/Attempt/Usage、断言和 artifact SHA-256。归档不进入 Git，索引和比较工具进入 Git。前后 manifest 的 `git_commit` 是当时 checkout 的基线，重构后的精确生产文件由索引 `source_sha256` 标识。

## 结果

- 前后各 2,340 个案例：2,339 PASS，1 个冻结 oracle guard 因未打开差分开关而 NOT_RUN。此 NOT_RUN 没有计为 PASS。
- 比较 10,598 个 artifact，0 个未解释差异；双方全部原始 artifact 哈希先经校验。
- 冻结 pi-ai **1.1.0** 上的受影响 Responses、两条 Chat 路由和 Anthropic 差分 726 PASS、0 pending。DeepSeek Responses 仍按 ADR-0014 独立证明，不计作 pi 路由通过。Messages 的最终四组配置另外运行 6 个 cache/budget/adaptive/tool 差分案例，全部 PASS。
- `go vet ./ai/...`、`git diff --check` 通过。所有 bundle 脱敏审计均为零 findings。
- 完整 `go test ./ai/... -count=1` 中 8,002 PASS、1,050 opt-in NOT_RUN，46 个非协议失败来自新 worktree 未复制既有忽略 ZIP，以及 source-inventory 测试尚未看见未暂存的新文件。复制原有三个 ZIP 后重跑 45 个导入案例通过；暂存新文件后 tracker 案例通过。完整初次失败和两次恢复报告均保留，未将初次命令记为通过；没有修改既有测试或门禁。

## 比较边界

[compare.py](compare.py) 校验 manifest 中每个 artifact 的原始哈希，要求所有前置案例仍存在且状态一致，并比较最终请求、响应脚本、fixture/输入/选项、流事件、Result/error/Usage 及工具往返。仅移除不可复现的 `timestamp`、`durationMs`、随机 `retryDelay` 和本地替身端口；仅并发案例的请求列表按内容排序。

既有 8 个 `thinking-sampling-concurrent` 案例会以最后完成租户的结果覆盖同名 `result.json`；仅该文件不做跨运行对比，其请求仍逐租户比较、隔离断言仍须通过。索引显式保留这 8 项，不移除其他案例的租户、账户、catalogHash、RequestID、AttemptID 或事件顺序。比较是前后行为证明，不是 GLM 支持或真实服务证明。

## 重放

基线与最终版本分别执行相同命令；仅当使用新的隔离 checkout 时，先复制仓库已有的 `.scratch/barness-ai-openai-decisions/{text-choice-evidence/text-choice.zip,jev-baseline/baseline.zip,runtime-evidence/runtime.zip}`。依赖新源码清单的测试要求新文件已纳入 Git。

```sh
BARNESS_AI_EVIDENCE_DIR=/tmp/glm-01-replay go test ./ai/e2e -count=1 -run 'Test(Responses|Options|CacheKeyScope|Chat|DeepSeek|Anthropic|ManagedEffort|History|NativeState|ToolRoundTrip|ThinkingSampling|Callbacks)'
BARNESS_AI_PIDIFF=1 BARNESS_AI_EVIDENCE_DIR=/tmp/glm-01-diff go test ./ai/e2e -count=1 -run 'TestPiDifferential/(anthropic-messages|openai-completions|deepseek|openai-responses)'
go vet ./ai/...
python3 .scratch/barness-ai-glm/evidence/ticket-01/compare.py BEFORE_MANIFEST AFTER_MANIFEST --output /tmp/glm-01-comparison.json
```

差分要求冻结的 node 依赖已按 `ai/internal/testkit/pioracle/node/package-lock.json` 安装。每次运行均重新校验来源 pins；未升级 oracle、未修改差分账本。本票未读取真实 Key、未运行收费请求，也没有新增 GLM 路由、目录或支持声明。能力边界见 [ADR-0032](../../../../docs/adr/0032-barness-ai-provider-automatic-fields.md)。
