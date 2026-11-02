# 代码变化与分发判定

- 目录版本：catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6。
- 核心运行代码变化：NO；依赖变化：NO。
- 比较基准：37a283ee4ef743f93d3a12132d1bc2c3211fb1e5；比较目标：4831e3544945f667d9a817be8dfb92dd08856575。
- 其他 Go 变化：ChineseEvaluation 示例 CLI 默认配置 v3 → v4；测试另行修订。
- 外部 JSON：已在上述基准代码实际检查相同原字节；Host 有外部加载/注入能力时，
  可只分发 catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6.json，重新加载并创建新 Client；任意旧发布版未验证。
- 编译内置默认：需要重新构建并分发程序，因为目录由 go:embed 编入二进制。
- 示例默认、JSON 分发、软件发布与厂商真实资格分别记录，本次没有软件发布或厂商资格。

精确文件差异、条件及证据见 code-and-distribution.json；28 项必需场景、完整普通
回执与原件摘要见 verification/。当前目录不包含所有原生 case 胶囊。
