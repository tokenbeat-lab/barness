# 离线适用性依据与限制

本轮产品 SDK、adapter、目录数据均未修改。实际实现变更只有目录原件恢复消费者与普通测试及共享xAI夹具的命名空间隔离。目录恢复从报告固定的 40 位原提交读取 ai/release/catalog-snapshot.json，同时验证原 version、文件内 hash 和当前 Catalog.Hash；不更改原包，不改旧目录身份，不发放执行权限。

## Kimi 原九项

六份原目录与当前各自候选完整型号数据相等；自身 Kimi 编码、解码、用量、费用和历史文件字节没有相关变化。原始来源映射与 grant/build source 三方相等，实际冻结二进制仅读取哈希。完整 producer 回执/实际输入、九项 manifest/断言/捕获/原生 seal、原候选消费退出0、原账户/地区/grant/两份可信账本及留存副本均核验。

当前保守文件比较仍看到 11（Chat/Messages）或 19（Responses）个运行文件差异。本次审阅的关键路径：binding 的 xAI 新常量；call、history、preflight、headers、catalog_validate 中显式 ProviderXAI 守卫；retry 中对 xAI 新增的估算分支；Responses 中 xAI 独立配置、nil 状态分派与选项分支；其余相关注释与目录组合。Kimi 原协议、价格、默认参数和 SDK 保持。opaque_history_ids 的新增行为针对 xAI 来源；旧 54 项 immutable-input 中均没有 xAI 来源，因此不能覆盖新 xai-origin-history。

除历史回执核对，还直接使用当前现有 verifyRegionalCase 逐场景解释全部 54 项，涵盖实际 wire、原始响应、归属、资源、usage/费用、Observer 和公共 Client 内存回放；54/54 无错误。读取函数通过明确标记的分析 overlay 暴露，只返回逐项检查结果，保留完整原函数及构建/运行输入。**该诊断不运行完整 regionalReport，也不产生正式资格或补造第十项。** 所以新结论为 PARTIALLY_APPLICABLE，不是完整候选/正式通过；内置目录仍 inactive。

## 三条正式记录的阻塞

TypeSafe 1.14、智谱 Responses、Z.AI Responses 的完整当前消费者实际拒绝产品来源差异；目录恢复后并非缺原目录或原件。当前目录投影均一致，独立 audit 均零发现。TypeSafe 1.14 经现有 ReplayProtocol 公共 Classify 内存回放12项；智谱经现有 verifyGLMCall 检查22项 wire/Result/Observer；Z.AI 经现有 verifyRegionalCase 检查及内存回放18项，均无错误。

这些事实足以保留并复用捕获及场景事实，尚不足以越过完整消费者的来源检查。保守集合含 cache_key、聊天主入口、历史分配，以及实际共享的 Admission/快照等；不能笼统说成 test-only，也不能因为第一项 cache_key 对分类路径不可达就忽略其余差异。差异清单在 projection-replay-v1/report.json，完整消费者拒绝在 current-consumption-final-v2/report.json。后续先做逐差异可达性与语义审阅，必要时加当前公开回放；尚未证明缺新增厂商事实，所以这三项不进入确定付费清单。

本次既没有把这三条拒绝改成 PASS，也没有扩大到前次16条重测。新的标准完整资格回执目前只增加4份（TypeSafe 1.13 属于分类路径的一个型号版本，不能把它等同于整条分类路径的两个版本都已完成）。
