# 工单 08：实现前的失败方式与公共验收边界

2026-10-08；开工基点 beda5c1ef835d7de6314596699628e5d6eeb830b。
接缝沿工单：Client.Classify → HTTP Provider；带双开关的 live 进程；
supportmatrix.Merge / CLI → 原子矩阵文件；写出的 evidence → audit.Bundle。

- 双开关未打开、未选组合或缺凭据误报 PASS；漏账户别名或混入其他 key 仍发请求。
- 分类错误跑聊天场景、错 operation/provider/API、替代版本或 alias 掩盖必交能力。
- 预算只数逻辑调用而漏问题、网络尝试、重试或失败请求；超限后仍发请求。
- 混合问题缺种类；bool criteria 错形状；响应型号覆盖授权型号；请求 ID 丢失。
- 漏报 usage 被当作零消耗；分布不满足独立约束仍通过；成功响应被隐式重试。
- 空问题被送到厂商；422 探针绕过 Client 或被本地字节限制拒绝，却声称验证了线上错误体。
- 有界 token 超限输入不能触发 422 时虚构确认；真实失败降为 NOT_RUN/UNSUPPORTED。
- 旧 schema、错误操作、重复 expected、空/缺能力或超预算报告被合并为完整通过。
- NOT_RUN 覆盖有效历史；只跑子集置 allPassedAt；同协议通过传播到其他组合。
- 一批报告后项失败时前项已落盘；拒绝改变输入或留下部分文件。
- 凭据、账户身份或环境泄入 wire/日志；Observer 含 state/questions/answers/错误正文。
- 目录价未用官方事实、输出非零、别名列入；内容变更未升版本/哈希/快照。

先复现再实现，沿既有隔离装置验证规则与公共 E2E 验证目录/协议；
最终产物包括脱敏 wire、live-report、矩阵合并前后、审计、离线/race/P0/差分追溯及可重放命令。
