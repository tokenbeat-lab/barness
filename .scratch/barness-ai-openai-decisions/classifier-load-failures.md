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

复审补充：null gauges、所有内存/分配伪造为 1、重复一个租户的结果、空 Attempt、
错答案名均通过实际 artifact 先红再修复。资源值必须存在且非 null；已知 request/body 副本
构成量测下界；每租户结果、独立 RequestID、Binding/目录/Attempt/Observer 和 q0..q7 完整集合
必须匹配。非零、缺项和矛盾均拒绝，原/新增校验共用规则。

审查前全仓 fd1b05a 的旧 P07-E08-unary-timeout-body-call-deadline 出现一例失败：
`unary_lifecycle_test.go:27-29` 将 CallTimeout 和 ReadIdleTimeout 同设 100ms；实际返回
读取空闲超时、预期 aborted 调用 deadline，体现接近同时到期的分类竞争。
该用例/运行时不属于图片容量政策；本票不改变时限、负载或断言，保留原失败完整
artifacts，专项原样重放与最终完整运行分别记录，不把重放通过改写为原运行通过。

最终量测补充：MaxQuestionBytes 同时约束公共 BoolQuestion 与实际 native predicate。
公共题目为 4033 字节，native 多 name/type 后为 4050 字节；先以实际捕获断言复现
原少计 17 字节的 RED，再令报告与导入按实际较大值核验。8192 限额、负载未改变，
早期报告原样保留，只有最终 0856b2b 的 4050 读数作为当前容量证据。

最终普通全仓 0856b2b 的旧 E08-policy-pressure-local-design-load 保留一例失败：
两个 loopback TCP write 返回 `no buffer space available`（ENOBUFS），Provider 侧
记录对应 reset/broken pipe。发生于普通/race 全仓进程同时运行期间；现有 transport
错误记录定位到 Host 操作系统 socket buffer 资源。该旧场景独立原样重放通过，
未改四路线调用、并发、图片/文本尺寸、预算、限额或断言，也未吞错/自动重试。
原 FAIL 全部 artifacts 与重放各自保留，不把图片容量通过称为全仓普通 PASS。
