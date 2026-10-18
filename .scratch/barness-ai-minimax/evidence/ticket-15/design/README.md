# 固定六进程设计审查

本目录保留最初两份设计字节、原始Standards/Spec发现，以及明确GLM专属120秒响应头预算和实际六项无Key检查后的v2。原发现不改写；v2的设计审查不授予源码、实际支持或付费批准。

固定三GLM、protocol114和两视觉进程，新上限108calls/132HTTPAttempts；原545/545保留，合计最大653/677。拟复用的10项generic与2文本报告只按既有相关来源规则使用，不刷新旧pins或绕旧launcher guard。GLM专属Host调整由[票16](../../../issues/16-glm-host-deadline-alignment.md)独立TDD，未来三路须同源完整实测。

[index.json](index.json)列出精确文件Hash；本目录执行`shasum -a 256 -c SHA256SUMS`可无Key复核保存字节。最终源码/二进制/计划、六项实际无Key前置退出码、独立两轴源码审查与新的人类预算批准仍须分别完成。