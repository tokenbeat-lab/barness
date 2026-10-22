# 验证策略

决定测试范围、准备付费验证或执行发布门禁时读取本文件。当前决定见 [ADR-0051](../adr/0051-scoped-verification-receipts.md)；本次迁移的实际状态见[唯一当前状态](../../.scratch/barness-ai-test-process-review/status.md)。pi 只作为架构与设计参考，行为以 barness 公共契约和自身证据验证。

## 开发阶段

行为变更先在公共 Client 或 CLI/进程边界建立可重复的失败场景，再实现与运行受影响普通 E2E。必须隔离测试时，先列出所有失败方式。纯文档、展示清单和审查记录检查引用、格式、状态与来源相等性；参与运行、价格、追溯或证据消费的数据按实际行为验证。

| 变更 | 普通验证范围 | 后续人工选择参考 |
| --- | --- | --- |
| 生命周期、队列、取消、共享状态与资源策略 | 受影响公共 E2E，含隔离、释放与受控时序 | 对应并发路径、旧 Provider 回归及受影响设计负载 |
| 协议编码、解析、历史与用量 | 对应协议 E2E 与脱敏捕获回放 | 受影响自身 live 路线及异步/资源路径 |
| 发布判定、证据消费者与维护工具 | 公共 CLI/进程、取消和文件系统失败场景 | 受影响的执行、取消与并发读取路径 |
| 纯文档及展示数据 | 引用、状态与完整来源/选定输入集合相等性 | 来源或输入未变时引用已有事实 |
| 模块、SDK、构建标签或无法限定的共享变更 | 扩大普通 E2E 和编译/vet 范围 | 按影响确定真实路线、race 与压力范围 |

测试范围由影响决定。合并本身不触发全量重跑；相关行为、测试、依赖或选定输入变化后重验受影响范围。并发场景用受控服务、hook 和 channel 确定阶段及取消时点，不用真实压力代替确定性回归。

在仓库根使用[回执工具](../../.scratch/barness-ai-test-process-review/tools/README.md)，每轮写入独立 attempt：

```sh
python3 .scratch/barness-ai-test-process-review/tools/run.py --feature <slug> --stage e2e -- \
  go test -json ./ai/e2e ./ai -count=1 -timeout=20m
```

定向验证用 `-run`，CLI/消费者另覆盖受影响包。采集器固定 ai 与维护工具执行来源（包括 Git 忽略的源码），清除继承的重负载及凭据配置。实际消费的仓库外层脚本、历史夹具、目录、授权与文档等另以可重复的 `--input <仓库相对文件>` 固定；不能只依赖 HEAD。密钥不能放进 argv 或输入快照。

## 审查、冻结与大功能验收

先整合审查最终 payload、原生历史、取消、超时、用量、资源释放和租户隔离；修复后复审，再提交并冻结源码、测试、依赖、工具和选定输入。按顺序完成：普通 E2E → 受影响自身 live 或有依据的不适用 → 人工 race → 人工压力或有依据的不适用。前项未通过就停止；各阶段独立保留原件。阶段完成、修复、合并与收尾均不自动启动下一阶段。

付费 live 必须另有明确路线、资格与预算授权；计划实施批准不是付费许可。一个进程只持有对应路线凭据，继续遵循各原生生产者的一次性授权、预算、完整用例、来源与资格规则。离线回放和受控 loopback 不产生真实支持资格。

race 和压力分别要求人类明确选择实际 argv 与范围。实施批准、普通通过和文档收尾不包含这种选择。race 不顺带开启压力；压力只在独立阶段启用 `BARNESS_AI_PRESSURE=1`。失败重跑与扩大范围按新的明确选择处理。等待时交付建议命令、范围、来源、输入及已有证据，实际未执行保持 NOT_RUN，`release.pass: false`。

## 并发与容量入口

普通并发覆盖包括 MiniMax 的共享租户、准入、取消、Observer、重试，队列限额与释放，协议超时，以及视觉发送前后取消。旧 Provider 对应路径按影响加入。

手动容量集合包括 `TestPolicyPressure`、`TestMixedPressure`、`TestClassifierImagesPressure` 及其 Import、`TestToolArgumentsScaling`、`TestAdmissionBurstPressure`、`TestPressureFixtureWriteBackpressure`、`TestPressureToolWriteIntegrity`、GLM/MiniMax/Multimodal 的 `SharedDesignLoad`/`RuntimeDesignLoad`，以及 `TestMiniMaxReleasePressureLoadConsumer`。`TestPressureLoopback` 还要求显式循环数。它们默认不运行；普通 MiniMax 消费者只读取留存夹具，不能启动负载子进程。小响应的 loopback 故障诊断仍属普通 E2E。

## 发布判定与收尾

`releasegate -record <json> -out <新目录>` 只读判定提交/来源、E2E、受影响 live、race、压力五项事实，不执行测试、vet、Node 或厂商请求。schema 与用法见[工具合同](../../.scratch/barness-ai-test-process-review/tools/README.md)和[发布指南](../barness-ai/README.md)。不适用必须与影响分析一致，缺失、失败、NOT_RUN 和过期不得当作通过；保留各原生消费者的完整性、脱敏、资格和资源断言。

每阶段保存真实 argv、数字退出、起止和墙钟时间、范围、工具身份、完整执行来源/输入、日志及原生 `bundles/<run-id>/manifest.json` 和 audit 的精确路径与哈希。启动失败没有原生退出码；中断或采集失败如实记录，重跑不能覆盖。父子测试耗时不能相加当作墙钟。

每功能只维护一份当前状态，报告与审查链接它。文档收尾只有在完整相关来源与选定输入相等时才能沿用普通结果，并注明实际受测 SHA 与当前提交的关系。真实清理单独按已审计划执行；历史失败、当前依赖、授权账本和未知证据不因成功重跑而删除。维护工具的临时合成清理自测不代表开始真实清理。
