# 工单 14：实现前的失败方式与验证边界

验证边界由工单指定：公共 Client.GenerateImages、独立 live 进程、公开支持矩阵 CLI、目录查询与快照。
沿用已有 E2E / evidence 装置；不测试私有 adapter 函数。

- 型号更替：别名、弃用或替代型号不能证明 gemini-nano-banana-2.1；注册与报告固定裸 ID。
- 版本错误：v1、generateContent、代理地址或 URL key 均拒绝；固定官方 v1beta/interactions。
- 固定项失效：store=false、单 image response_format、1K/1:1 缺失不可通过；维护者批准的 delivery 省略必须固定，回调不能添加该字段。
- usage 误算：thought 单独报告，模态是否包含必须由真实响应等式证明；不能重复收费。
- 超预算：失败和重试也消耗调用与图片预留；最多四次、四张，1K；2K、数量扩大不得发送。
- 输出误判：遍历全部 model_output；不可读、MIME/尺寸/数量不符、URI、续读或非 completed 均 FAIL。
- 缺证据：缺 key/未执行为 NOT_RUN；必交生成/编辑不得记为 UNSUPPORTED；截断不能证明 usage。
- 矩阵污染：错操作/Provider/协议/型号、旧报告、配置失败、少能力/少形状证据不可部分改写矩阵。
- 证据泄密：来源处有界捕获、脱敏，独立 artifact 审计零发现；不保存凭据或账户真实标识。
- 审查新增失败方式（修复前）：请求截断不能跳过响应脱敏；响应超限、无效 JSON 或读取中断不能退回原文；短签名、续读令牌及签名 URI 也必须清除，并保留长度/哈希诊断或完整安全形状。

先取得本组合真实 PASS 再纳入内置目录；宿主 probe 与支持声明分开。
