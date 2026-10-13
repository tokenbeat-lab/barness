# 09 授权未知信息路径：两轴复审

固定点 `fc6937e` 至 `abdb733`：新增显式 unknown 核验策略，保持缺 key NOT_RUN、
错误证明/跨组合 key/远端代理拒绝，候选报告 schema=2、nullable 权限证明。
Spec 以 issue 09 和用户 2026-10-13 明确授权为准。沿用项目 AGENTS、领域/issue
标准及 code-review smell baseline，两轴分别由独立子代理复审。

## Standards

0 项可操作发现。新增策略显式 opt-in，unknown 与已核验状态区分，受控故障先于
实现；隔离凭据、预算、代理、有限超时和敏感信息边界保持清晰。正式目录、矩阵
与领域模块没有改动。

追加固定点 `abdb733` 至 `1c8afce`：0 发现。新夹具仍通过公共进程触发 inventory
读取拒绝，并验证失败撤销验收；避免了大文件审计导致的监督超时。注释保留原因
和历史证据位置，未改变运行时、预算或真实验收条件。

## Spec

0 项可操作发现。独立核验真实 bundle 的 83 项 inventory Hash 全部匹配，九个
scenario 及 result/exchanges/四个 Observer 记录完整，八个实际 Model 与完整
usage、取消的空答案及 unreported 用量、预算与执行源码 Hash 均匹配。用户授权
unknown 不阻塞已落实，未将 unknown 解释为权限证明，未代替正式目录/矩阵。

追加固定点 `abdb733` 至 `1c8afce`：0 项 Spec 发现。仅替换故障夹具，仍通过
非普通文件触发 inventory 读取拒绝，并保持 FAIL/evidence、Complete=false。
线上运行时、预算与 unknown 授权路径未变。代理独立运行 integrity read/write
两项 race 进程测试，均通过，未修改文件或调用线上 API。

Standards 0 / Spec 0；两轴均无剩余发现。原初审发现和修复记录留在上级历史包。
