# 工单 18 两轴审查

固定点：`a6c163af07663fed878de3ca4f18e75eeac2bc2a`，审查其后工作树变更；规格为
[工单 18](../issues/18-pressure-loopback-interruption.md)。按 code-review 技能，
Standards 与 Spec 由独立并行子代理检查，修正和新增 socket 边界后分别复核。

## Standards

当前无开放问题。曾发现 P2：无符号基线减法把堆下降展示为极大增长；已 clamp 到零。
文件均低于 500 行，原始错误只在合成测试审计文件，fixture 屏障、结果通道和等待有界。
未新增依赖。socket Write 串行，ENOBUFS 最多等待 1 秒，只续写未发送后缀；其他错误
直接返回，Close 唤醒等待，原子计数提供观测。七个公共 Client E2E 分支验证恢复及
失败、关闭、deadline/取消，并验证没有新增 HTTP 请求或 Attempt。

核验脚本两项实际问题已修正：审计 CLI 需要根 manifest，改用独立汇总 envelope；
fixtures 是脱敏前输入哈希，改为原输入/交付 artifact 分别检查。首次审计调用失败
保留在 retention-replay，原包/清单不改。上述修正经 Standards 复核无新问题。

## Spec

当前无开放问题。此前 P1 为缺少原始中断原因。追加自然失败中，15 路 unexpected EOF
全按相同 requestId 对到 fixture TCP Write ENOBUFS；客户端均未取消/关闭 Body，
32 个 handler outcome 齐全且审计零发现，已确认此次同症状失败的直接因果链。
原 P1 已解决；旧工单 17 没有逐路记录，不能倒推其九路具体 errno。

回放遗漏前序轮、按并发到达序号猜身份两项 P2 已修正：回放完整循环，使用统一合成
requestId，关联诊断有红/绿证据。socket 修复限于测试基础设施，不重放 HTTP/Attempt；
原 32 路负载、3 分钟、2 GiB、完整输出及失败断言未放宽；没有范围扩张。

最终说明复核指出两处表述：早期阶段仍称“最终源码/仍开放”，以及对早期包源码哈希
的过度声明。均已修正为阶段记录和明确的来源限制，早期包不冒充最终受测源码。
Standards：0 开放（最严重：无）；Spec：0 开放（最严重：无）。
最终源码完整普通/race 各 4,461 PASS，同进程 100 轮及 10 个独立进程 PASS，
两种 vet PASS。交付 14 份归档共 43 个 E2E 包，失败历史保留；归档哈希、用例产物、
最终四种运行的源码匹配及整包脱敏审计通过，实际输出见 verification.log。
