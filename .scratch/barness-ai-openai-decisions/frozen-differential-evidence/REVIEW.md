# 工单 07 双轴审查

固定起点 `724f09ea17ff9fbfe0b1716cc8aca12637ab8371`，最终实现 `6acff6d`。
命令：`git diff 724f09ea17ff9fbfe0b1716cc8aca12637ab8371...6acff6d`。
Standards 与 Spec 由 code-review 技能要求的两个独立子代理并行审查。

## Standards

初审 2 项：Node 异步 subprocess 缺时限（P2）；相同 Luna 配置重复且无解释（P3）。
修复：spawn 使用 10 秒 timeout 并 await close；两处真实用例共享有说明的 configureDifferentialLuna。

复审：无剩余文档规范违例或可操作 smell。操作身份、受控 endpoint、源码/version/SDK lock
固定、capture hash 与独立诊断/capability 断言均保持；新增代码文件均小于 500 行。

## Spec

初审 4 项 P2：可同时修改 package/provenance version；SDK lock hash 未强制；
capability 扩展只断言四字段；errorMessage 扩展缺独立诊断断言。
另由完整测试发现旧 ID-only Costs 查询选中同名 classifier 价格。

修复：固定启动版本并加 paired-tamper 反例；固定 lock checksum 并在 Open/import 校验，
missing/wrong-lock 反例先红后绿；完整 capability/input literal；每 fixture 的 Pi/Go
诊断 literal 与公开 Code/Phase/HTTPStatus/ProviderRequestID；旧 Costs 显式查 chat，
Decisions 使用独立 classifier Models 接缝。完整 raw Barness result 也写入 captures。

复审：OD14/OD48/OD49/OD50 在本票范围内完整，完整 wire/命名答案、有限 case/path/kind
与自身断言、共同字段反例、Jev 保留和 P10 Required 均符合要求。

Standards 0，Spec 0；两轴在实现提交后复核。全套/race 和证据打包的实际收据另见 validation.json。
