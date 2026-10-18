# Ticket16 GLM 专用响应头上限

代码 `dcd30f7ce6bf65f5bd1101697c8e5c98155ceb98` 只在 GLM 专用 Client 装配把 `ResponseHeaderTimeout` 从继承的90秒设为120秒，与已有120秒整次调用上限一致。共享 live policy 的90秒响应头/90秒读空闲、15秒连接和其他厂商策略不变；GLM场景、模型、断言、4096输出上限与重试/预算不变。

预先约定的公共 Client seam 通过真实本地 TLS 服务延迟91秒返回 Responses SSE。旧代码 RED1 精确报 `ResponseHeaderTimeout (1m30s)`；只加入一条生产赋值后，同一测试 GREEN0，响应头91,002ms、整次调用91,009ms、一条 transport 请求/一个 Attempt，完整文字和终态保留。原有双模型控制各PASS；较早 Host/call/headers/read-idle/cancel/settlement 检查 ordinary/race 各179PASS，两条 vet 均0。没有第三轮91秒 race等待：边界已按真实计时完成 RED/GREEN，竞态验证覆盖179项短控制与取消路径。

这是本地受控计时证明，**不代表厂商将在120秒内成功响应**。合成 usage 仍标记 partial；其 token/费率不作厂商消费或账单。历史实际90秒失败及旧原始收据完全保留，不刷新旧 pins、不授予付费权限，此次 Key读取/厂商请求/付费调用均0。

RED测量 HEAD `06229f68` 加明确 test WIP；GREEN测量 HEAD `d8e71006` 加明确工作区源码 SHA，最终源字节等于上述代码提交。设置初期一次拷贝与测试编辑重叠，原 `assembly-before.go` 作为 WIP投影保留；另外从 exact Git blob 恢复的 `assembly-committed-before.go` 匹配原始 source-before SHA。该差异在执行RED前发现并写入原始注释，不把WIP说成旧提交源码。

逐项比较三份 M3正式263文件 map 全部不变；三份既有 GLM217文件 map 仅装配文件改变，因此全部过期，需要未来新GLM3采集。263中仅 `ai/internal/testkit/minimax/README.md` 不在较宽1048执行源集合内，另保留其代码提交字节及正式报告中的独立SHA，复验仍检查全部263项；不删除 map 项。

```sh
python3 -B .scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment/verify.py --common-root /Users/cyber/RestoX/barness
python3 -B .scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment/check-boundaries.py --common-root /Users/cyber/RestoX/barness
```

只读归档 `.evidence/minimax-retained/ticket-16-glm-header-alignment/deadline-alignment-raw.tar.gz`：3,785普通文件、10,434,366 payload字节、1,572,266压缩字节，SHA `fa7d7673b6708cab627e62cd9355b168aaca8a5c851ea63d6365327175e446ae`。保存全部TDD原始bundle/命令/逐阶段源和历史map，前后逐文件SHA一致；3,724个原生 artifact 引用逐项验证。

复用固定已pin的 implementation reader：32MiB压缩/192MiB解码tar/32,768条目，8MiB inventory、1MiB单观察元数据；本胶囊另固定最多512个观察成员/16MiB观察总量。实际解码tar17,397,760字节。读者逐级拒绝链接、重复/越界成员和hash漂移，不解压、不执行保存的代码、不自动重跑本地或厂商调用。
