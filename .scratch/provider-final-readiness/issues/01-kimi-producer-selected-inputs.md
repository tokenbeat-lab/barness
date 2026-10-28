# Kimi producer必须同时核对source.files和selectedInputs

Status: resolved
Type: task

真实六份旧run.py回执中，kimi/README.md和regional/README.md按工具合同在selectedInputs；原生执行来源包含二者。现有verifyKimiProducer只查source.files，误拒完整真实baseline。此前受控producer夹具把二者都放source.files，漏覆盖真实布局。原件无缺口，不需要厂商重测。

公共进程反例先通过分析overlay建立，source与selectedInputs的原字段/哈希对齐应通过，缺失、改动、两栏冲突均拒绝。修复只合并两个实际输入域进行逐文件精确比较，不放宽原生完整来源、不变更产品/SDK/权限。保留所有失败与当前未批准提案；最后用真实六份旧producer离线验证，再重建正常live-tag冻结二进制并更新六份proposal。不能激活目录、请求厂商或恢复暂停。

## Answer

已修复：逐原生来源文件精确核对source.files与selectedInputs；缺失、任一域哈希不符或两域冲突仍拒绝。公共CLI布局反例先FAIL后PASS，六份真实旧producer最终全部PASS。消费者修正后已重建正常live-tag二进制并形成六份v2不可执行提案，最终完整普通回归PASS/exit0；不需要厂商重采。详见 ../status.md 及 ../../../.evidence/provider-final-readiness/verification-and-retention-v1.json。
