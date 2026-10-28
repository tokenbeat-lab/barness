Status: resolved
Type: task

# 修复账本留存副本字段读取类型

实际preflight.py把preservedCopy当成字符串；国际Responses/Messages的旧记录存的是{path, sha256}对象，产生PosixPath/dict TypeError。此前仅offline_review.py有修正，实际入口未接入。

失败方式先列出：两种真实格式均须接受，无效类型、缺path/hash、副本哈希不符、权威账本哈希不符均须拒绝。先执行Python子进程回归，运行实际入口账本读取段，再修改实现。篡改只用独占夹具，真实12条记录只读。完整preflight会写STOP/scope，本次不复跑；付费采集不在本次请求范围。

变更不涉及生产源码、冻结工具、并发或负载。保留失败/STOP/授权/账本，独立证据在.evidence/kimi-history-paid-restart-20261028/type-error-fix-v1。

## Answer

实际入口与离线审阅入口现共用verify_ledger_record。先失败、后修复；34项实际读取段子进程回归通过，六路线12条原账本均接受，篡改/缺字段仍拒绝。无厂商请求，旧STOP保留。详见../status.md及type-error-fix-v1/verification.json。
