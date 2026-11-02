# 独立 Spec 审查

基准 9584d3d6f095aefa4ab73bdc75c39a705b7931d6；目标 65a1ecbb2cb97f51c37cfee849925712187e5acc。

(a) 规格要求但缺失或部分：已实现范围未发现。规格要求「保留非空型号 ID 与四维身份唯一性」；共享 Catalog.validate 新增空白 ID 检查，既有四维去重保留，外部定义校验仍在。规格要求「公共 CLI E2E……完成 preview/freeze/check/apply」；合成来源测试覆盖这些步骤。完整普通与交付归集尚待执行，不能视为已通过或完成。

(b) 未授权范围扩张：未发现。规格要求「只删除具体 ID 相等条件」，四类能力门槛的差异限于 ID 判断、能力命名及诊断；共享空 ID 防护是规格明确要求。真实目录、维护源、基线不在 diff 中；Gemini 格式和 GPT-5.5 费率分支未修改。

(c) 表面实现但实际错误：未发现。规格要求「目录精确查找……Binding.AllowedModels 的独立租户授权」和协议原有响应合同；这些运行路径无差异，新合成 E2E 覆盖精确身份、无授权零 HTTP、hook 防篡改及 Kimi 别名区别。九身份分别覆盖两站三 chat 协议、xAI chat/image 和 SystemOne image/Box；保留能力、协议、模态与计价校验。合成身份与受控 CLI 不声称厂商资格。

依据：spec/status、ADR-0059、GLOSSARY、Catalog 与测试 SOP、固定 git diff、WORK/inputs/model-id-audit-v1.md。未运行测试；定向回执状态以 status 原件为准。
