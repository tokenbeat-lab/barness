# 工单 16：独立中文分类评估

固定任务和协议宿主见 [示例说明](../../../ai/examples/chineseeval/README.md)，
边界决策见 [ADR-0024](../../../docs/adr/0024-barness-ai-chinese-classification-evaluation.md)。
[失败方式](FAILURES.md) 在实现前记录；既有受控 Provider、公用 Classify 和文件 CLI
接缝验证输入漂移、缺答/错键、协议/HTTP 错误、缺用量、预算、取消、NOT_RUN、报告不完整及审计。
没有新增库内测试探针或 Provider 模拟接口。

[真实报告](live/evaluation-report.json) 于 2026-10-08 18:51:11–18:51:18（北京时间）
独立执行，24/24 条正确，四类各 6/6。全部 24 条 scored，0 失败、0 未执行；
24 次调用/问题/尝试、0 重试，输入 12843、输出 1080 token，全部完整上报，
按目录 2026-10-08.5 输入 $0.042/百万 token、输出零计价，估算成本 $0.000539406。
请求与实际响应型号都是 jev-1.13.0，逐样本厂商 request ID、目录/凭据版本和时间均保留。
[NOT_RUN 包](not-run/evaluation-report.json) 的分数为 null，全部样本未运行、零调用。

真实本轮没有误判，errorIDs 为空；按标签混淆表仍全部保留。
confidence 和所选类别概率的高置信段均包含 24 条，均值 0.9995833、正确率 1，
ECE 都是 0.000416667，多类 Brier 0.00000833333。空段为 null。
这份 24 条明确目标的合成数据不足以评估生产总体或低置信样本的校准质量；
100% 不能作为通用中文业务准确率保证，也没有用它设定上线阈值。
受控 fixture 另含四个明确误判，验证 20/24 分数、标签转移、置信区间和错误样本追溯，
不把它的结果称为真实效果。

真实启用与支持矩阵冒烟隔离：[run-live.py](run-live.py) 只提取根 .env 的 TYPESAFE_KEY，
向独立评估二进制注入本 key 与账户别名，不继承其他 Provider 凭据或冒烟开关。
账户区域和权限级别仍未核实。本工单没有写支持矩阵；真实业务包和 NOT_RUN 包都使用
既有逐文件审计，零发现。原始调用结果包含所需合成评估内容，Observer 不含状态/问题/答案。

复现（仓库根）：

    go test ./ai/e2e -run '^TestChineseEvaluation' -count=1
    go run ./ai/examples/chineseeval/cmd/chineseeval -verify .scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/live
    go run ./ai/examples/chineseeval/cmd/chineseeval -verify .scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/not-run
    python3 .scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/run-live.py
    python3 .scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/run-live.py --live

最后一条会真实消耗已授权账户预算；每次保存新包，生成不确定性允许答案不同。
固定数据集在首次联网之前已经编写和冻结，任务与哈希保留在报告，未以真实预测反改标签。
真实运行提交指向任务开始基线（源码当时在工作区），交付 manifest 另固定实现源码哈希。
首次文件回放发现答案视图 `type` 字段的严格解码不兼容；先加序列化 E2E 红断言，
独立转换视图后绿并成功核验原真实包，没有丢弃结果或为此追加付费调用。

两轴 [code-review](review.md) 的四项 P2 均以新 E2E 红断言复现后修正；
导入报告验证非负用量/固定价格/尝试身份，并核对 manifest 与完整的 Observer 生命周期。
首次调用前取消为 NOT_RUN/null，命令错误保留安全阶段标识。原真实包通过修正后的核验。
