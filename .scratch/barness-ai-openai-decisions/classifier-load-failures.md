# 08 失败方式与验证接缝

接缝沿用批准工单：公共 Client.Classify/Complete/GenerateImages 的实际设计负载，及
LoadBundle → DesignLoadEvidence 对写出 artifact 的文件系统导入。起点 c4cc495。

- 旧 Jev mixed-pressure PASS 缺 Luna 图片报告，仍通过导入。
- Host 示例缺 classifier Images，或 count/每图/总图/完整 JSON 容量未明确配置。
- 新报告删任一负载、身份、count、decoded/base64/data URL/JSON、ratio、预算、并发、
  GC/TotalAlloc、Provider 副本、吞吐、取消或资源字段；错误值、旧报告或错误路线通过。
- ratio 不是 used/limit，limit 未匹配 Host 实际政策，缩小负载或扩大 256/512 MiB 预算。
- 实际未占满 4/8 许可，未等完整 unary 字节读取、未持有完整结果时采样；单次 GC 漏暂存峰值。
- 取消后不归零、重放已收到 2xx 的请求、相同图片负载无法恢复；普通或 race 缺适用报告。
- manifest 不含实际配置/脚本/捕获/结果/Attempt/Observer/量测/命令，或审计发现敏感内容。

原混合批次保持每租户 chat、OpenAI image、Google image、Jev 各一次及原 sizes；
新增批次每租户四次 Luna 图片分类（128 KiB state、八个 4000 字节问题、两张 256 KiB PNG）。
新增成功、完整读取处取消、恢复三批各自占满声明并发，各批独立适用相同原内存预算；
顺序批次之间保留 Provider 捕获，报告每批新复制字节及固定共享脚本，不能借分批缩小输入。
