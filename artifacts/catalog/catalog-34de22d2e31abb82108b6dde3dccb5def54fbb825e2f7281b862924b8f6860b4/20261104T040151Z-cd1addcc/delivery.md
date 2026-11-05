# DeepSeek V4.1 Flash 1017 目录交付

官方 1017 升级继续使用 deepseek-flash。现有模型协议能力与峰值价格不变，因此使用 revise 更新两处展示名称和官方字段追溯，先 Responses、后 Chat。
不创建日期 wire ID 或重复条目。目录仍为 81 条（75 chat、3 image、3 classifier），DeepSeek 4 条、内容版本 4。
全部 ID、能力、价格和顺序保持；仅两处 name 改为 DeepSeek V4.1 Flash 1017。该滚动别名不永久锁定 checkpoint。
文本/图片、1M context、384k output 和已声明 low/high/max 保持。固定峰值估算 0.3/1.2/0.006/0 USD/百万 token，不按时段自动换价。

版本：`catalog-34de22d2e31abb82108b6dde3dccb5def54fbb825e2f7281b862924b8f6860b4`。Catalog.Hash：`sha256:fd23a12ac65a114116bcfbaf6f6ff335e3ab4cf5b2af255cc00d1cff49a85156`。
完整 JSON SHA-256：`672cb593ac0e3c6894bfbca8c753254162fdcba8c41adf8321a37108c6e31de4`。维护 revision：`sha256:14525dac0d1c3505be8f3b330f7bdb25f9e29a609779618bcd9f0b9d61ac8ee1`。
[运行 JSON](catalog-34de22d2e31abb82108b6dde3dccb5def54fbb825e2f7281b862924b8f6860b4.json) · [变更记录](CHANGELOG.md) · [普通验证](verification/ordinary-summary.json) · [候选检查](verification/candidate-checks.json) · [状态](verification/status.md)。

HEAD/实际受测 HEAD：`665862ff90074eecda0a573050114de9ce88c80d`；起点含此前未提交变更，实际基准由 before-source/before-builtin/before-code 固定。
本次 SDK 核心、运行依赖、schema 变化均 NO；其他 Go 文件仅相关 E2E 和示例默认配置 v10。旧 v1–v9 不改写。
两次完整候选与最终定向/完整普通/生成/构建/静态检查均通过。原预期 RED、Kimi 历史标签比较失败与 launcher exit2 保留。
历史比较保持原展示标签，其余运行字段/顺序继续核对；新专项和独立完整数据审计核对本次恰好两处 name 变化。
原 serving 关联不改写，官方 update 是显式字段补充。Responses 过渡快照单独留存，未作为编译内置默认采用。

外部 JSON：仅本次受测程序取得最终原 JSON 的兼容依据，Host 显式加载后创建新 Client；现有 Client 不热更新，其他旧版本未验证。
BuiltinCatalog 更新须重建并分发程序。live/race/压力按影响不适用；未提交、未发布、未部署，不产生 1017 新厂商实测资格。
完整 review/ 与采用 baseline 相等；官方补充原件见 verification/supplemental-sources。工具/CLI/测试胶囊仍依赖外部原路径，不声明全部证据自包含；运行 JSON 可以独立使用。
回滚需新 revise 恢复 name，重新 check/apply/export、同步三处并另建配置 pin；旧历史不重写。本次未回滚。
