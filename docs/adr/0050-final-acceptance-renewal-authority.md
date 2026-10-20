---
status: accepted
date: 2026-10-20
---

# 最终验收使用独立且绑定完整前序的续期授权

新增地区 Provider 后需要对最终源码重新取得国际 MiniMax 的三份自身正式证明；[ADR-0047](0047-current-source-formal-renewal.md) 的首轮续期已关闭。采用第二个实际用例 `provider-sites-final-acceptance-1`，只承载国际 `minimax`、`standard`、`MiniMax-M3` 的 Chat / Responses / Messages。旧 ID 永久拒绝活动执行，仅保留历史报告和来源的只读验证；不覆盖旧 grant、账本、claim 或 launcher，也不提供调用方自选的新轮次入口。

新 schema 2 grant 绑定原正式 39 Calls / 36 Attempts 的 SHA / 尾链，以及已关闭首轮续期的完整固定 canonical 文件集合：8 份 authority 文件、3 份 launcher；同时包含原正式账本和原 3 份 launcher，共 15 份事实。原生导出完整路径到 SHA 集合；来源审查和人类批准共同绑定其摘要及计划，不能用提交 bundle 自带的 manifest 选择历史。新轮次仍限制为总计 60 / 72、每身份 20 / 24、单次输出 4096、每身份累计请求输出 40000 与输入预留 120000。前序预算及失败不会转移或重置。

父 launcher 在读取 `.env` 前使用自己的唯一 intent 和只读 `minimaxverify -renewal-preflight=<route>` 校验；该校验不能领取子进程的 claim。子进程 Host 在读取所选 Key 前调用 `BeginRenewal`，以同一 Git common owner 的独占锁领取带进程随机 nonce 的持久 route claim。首笔预留只允许该进程消费一次 pending 状态；崩溃、换进程、复制相同 session、部分写入和删除账本均不会恢复可用性。claim / 账本和目录均同步到磁盘；不确定写入保留占用，代价是失败后必须检查原始事实并另行批准有限轮次。

活动执行必须匹配冻结 HEAD 和完整执行来源；历史消费仅允许未选中数据提交后的完整集合相等性。调用所依赖的固定美元 guard 及其区域候选价格目录均进入 MiniMax 执行来源集合。新授权机制和离线合成测试不授予真实请求、账户资格或付费预算。实际批准、证据与发布状态只记录在[规格当前状态](../../.scratch/barness-provider-sites/spec.md)；[工单 09](../../.scratch/barness-provider-sites/issues/09-final-round-minimax-authority.md)保留实现与验证指针。

直接替换旧常量或重封旧 grant 会复活已消费的权限；无限轮次注册表则会扩大目前只有两个真实用例的授权面，因此均不采用。

## 首个候选批次关闭后的固定续验 2

首个地区候选批次发生真实失败，`provider-sites-final-acceptance-1` 的原生 M3 grant/账本/nonce 从未发放。固定活动入口改为 `provider-sites-final-acceptance-2`，旧1显式选择和无 selector 的 canonical 账本路径均永久拒绝。旧1只允许已封存的 operator authority 和空锁；任何新 grant、nonce、账本或其他 intent 都令新原生预检失败。

schema3在上述15份原生历史之外，单独绑定 canonical `closed-cohort-1/seal.json`。grant、人类批准和独立审查各自 pin 该摘要，并保留旧批次实际40/40、已认领52/52及1.5USD整笔占用。新33进程仅722/758、28.4USD，跨轮774/810、全额29.9USD，不增加原30USD。原生 M3自身60/72和原15份历史39/36×2不变。固定路径从 common Git owner解析，合成独立owner可形成自身封存；不把真实本机SHA常量写进SDK。

封存中的旧二进制以稳定文件流式散列验证，限制4096文件、64MiB/文件、512MiB总量，源码归档遍历限制8192条目、深度32；禁止符号链接和未pin的归档文件。该具体历史校验与一般运行时授权分离，避免为一次续验引入通用轮次框架。
