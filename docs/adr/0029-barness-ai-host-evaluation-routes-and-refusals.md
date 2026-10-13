---
status: accepted
date: 2026-10-13
---

# Host 评估按完整 route 与 recorded snapshot 分离核验和适用性

工单 11 在 ADR-0024 的固定 24 样本任务上新增显式 Jev/Luna route。原无 route 配置、Jev
默认命令及 schema 1 意义保持；带 Operation/Provider/API/Model 的配置产生 schema 2。选择
显式版本避免把历史 schema 1 的 usage_incomplete 排除评分含义静默改成新报告语义。schema 2
普通答案和逐题拒答独立计数，质量按合法答案统计，partial/unreported 消耗单独阻止 COMPLETE；
protocolStatus、任务 status 和质量分别呈现。全拒答可完整，条件指标 null，Host 承担后续业务动作。

核验从包自己的数据集、配置和完整目录快照出发，再对账价格、全部样本、Observer、代码内容 Hash、
route Hash、文件 inventory 和审计。原 Jev 历史包早于 catalog 文件，以原 commit
`a12c037356751301f292f770188bd44bb629dabf` 导出的不可变目录精确匹配自身 recorded Hash；
不覆盖它的配置、日期或原结论。这是历史证据合同，不是跨协议兼容转换。对当前目录强制核验
会使合法历史证据失效，因此把核验与当前发布适用性分离：发布消费者继续明确检查当前预期
route/config/catalog/code/日期，历史可核验不表示当前可发布。

所有默认离线运行保持 NOT_RUN；fixture 使用无网络能力的 synthetic transport，带独立身份、
raw captures 和 Observer，不能成为 live 质量证据。24 Call/Attempt、零重试、串行、600 秒
上限，取消停止后续请求。Luna 的 alias 权重不因目录与价格 Hash 固定而称为固定。协议库
和支持矩阵不拥有中文任务的阈值、转人工或业务 fallback。
