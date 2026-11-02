# CLI 矩阵增量 Spec 审查

固定差异 65a1ecbb...807afbc5：未发现规格缺失、未授权扩展或实现错误。规格要求「为上述9个身份完成 preview/freeze/check/apply」「每条身份使用独立或有序前置基准，保留各自真实原件」：矩阵复用一份二进制、逐身份建立独立临时 store，执行四步，并校验应用与候选原字节相等、旧集合及顺序不变、仅增加所选身份。共享案例实际包含 Kimi 两站各三协议、xAI chat/image、SystemOne image/Box 共九身份。

每条来源明确标记合成、使用 synthetic.invalid、临时 store 和输出；应用快照留作测试 fixture，不写真实维护目录或声称厂商资格。定向回执实际 PASS/exit 0；九个 check-reference 分别指向存在的仓库 .evidence 原件。保留旧单身份测试。完整普通及交付仍待执行，本次未运行测试。
