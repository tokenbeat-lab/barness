# 两轴 code-review：工单 06

基点：0a3ba58bdaccc32fb1925203134c35c397183161。
两个独立审查代理读取 git diff <base>...HEAD，不编辑代码。

## Standards

初次发现两项 P2：

1. encoding/json 的结构体字段名匹配不区分大小写。noul/Noul、score/Score 等不同 JSON 键可覆盖非法值，
   违反公共契约、ADR-0021 的无歧义字段与 AGENTS.md 原则 3。已在结构体解码前校验精确 JSON tag，
   任意状态/说明对象与问题/选项名仍独立保留自己的键。
2. 显式 bool criteria:null 与省略都被解码为 nil，跳过提供时的两种说明校验。
   已改用 RawMessage 分辨存在性，显式 null 拒绝，省略继续有效，嵌套 BoolCriteria 使用精确 tag 边界。

两项均由公共入口 E2E 先复现（review-fields-red.log），再修复并通过（review-fields-green.log）。
复核于 f8f5e42：两项均解决；原记录、回调与原子响应错误测试覆盖修复边界，没有修复回归。
没有实质性的 baseline heuristic smell；封闭类型独立 tagged encoder 的注释理由符合 AGENTS.md 原则 7。

## Spec

初次报告：No Spec findings. The diff implements issue 06’s mixed question types, native JSON precision,
final callback validation, atomic answer failure with usage retention, and scoped pi parity projections/extension
registration. No missing requirements, scope creep, or incorrect implementation identified within this issue’s scope.

修复复核：No Spec regressions. Exact field validation enforces the closed-record requirement while preserving
arbitrary state, description, question, and option keys. Separating omitted bool criteria from explicit null matches
the documented supplied-criteria rule. Public request, callback and atomic response regressions cover the boundaries.

最终未解决发现：Standards 0；Spec 0。初次 Standards 2 项 P2 已解决，Spec 0 项。

证据追溯复核于 6d435aa：分类目录用例补齐差分记录，完整 pi 目录进入比较；baseUrl/input 的例外
仅限目录用例、精确路径和 only_pi 方向，实际值仍单独断言，未知字段仍失败。
Standards 与 Spec 两位审查者对此补充均无发现。
