# 02: 冻结 pi 差分判定器

**What to build:** 维护者可以用同一份逻辑输入与响应脚本，分别驱动冻结 pi-ai（`0.87.1`、commit `898ab804050730e9dcefb4443875d5a932aa6a32`）与 barness-ai，自动比较两者实际发出的请求和得到的事件/结果，并输出首个差异位置与分类。它是兼容判定器，复用 E2E 场景而非第二套测试体系（spec Testing Decisions §2、§5、§6）。

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] 在 barness 自身可重建的位置建立冻结 pi 的独立副本：补齐 `src/providers/data` 生成数据与依赖，记录来源、哈希与重建步骤；不修改原研究目录，CI 不依赖开发者本机研究仓库
- [ ] pi 侧运行器只连接本地受控 Provider，不读取环境 key、不访问线上服务；不执行整套 pi 测试
- [ ] 比较维度：请求路径、headers、认证来源、JSON 字段 presence；Provider/API/model、消息块、工具关联、原生状态；事件类型/块索引/delta/顺序；StopReason、errorMessage、Usage 与成本
- [ ] 只对 JSON 对象键做语义排序；数组与事件不排序；时间与生成 ID 使用明确字段的一一映射并保持后续引用；不得统一删除 null、零值、空数组、未知字段或全部 ID
- [ ] 租户元数据、Code/Phase、计量完整性、资源限制、同步访问与脱敏等扩展不进入 pi golden
- [ ] 差分记录至少包含 case_id、pi_commit、sdk_versions、model_catalog_hash、原始请求与帧哈希、首个不同 JSON path / 事件位置、分类（已修复 / 规范定义或批准的扩展 / 待处理）、处理决定
- [ ] “待处理”差异使对应协议的差分门禁失败
- [ ] 以 01 的 Responses 文本场景作为首个差分用例跑通，结果写入证据包
