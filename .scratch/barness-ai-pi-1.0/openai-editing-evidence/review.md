# 工单 10 两轴审查

固定基点：`3b31ec0edb01d2a0a64828fe36acbe8dbb0b7e6b`；最终实现提交：
`2e86ea2`。规格来源：`issues/10-openai-json-image-editing.md`、规格的图像契约与
回调边界、设计第 8.1、11、13 节。按 code-review 技能由两名独立只读子代理并行审查。

全部问题先通过公共 Client/HookedClient E2E 复现，再修复；红灯证据见
[review-red.json](review-red.json)、[review2-red.json](review2-red.json)、
[review3-red.json](review3-red.json)、[review4-red.json](review4-red.json)。没有新增内部单元测试接口。

## Standards

最终复审：此前容量预检、自定义编码、TextMarshaler key、嵌入字段及循环
pointer/interface 的 findings 均已解决。深度错误已沿调用链返回，公共 E2E
覆盖参考图与 mask。未发现新增可行动问题。

初审及后续复审共四项已修复：

- P1：pointer/struct/raw JSON 的已知 URL 曾在单图限额检查前产生大分配。
  现在遍历普通表示，并以字节视图及固定缓冲区检查 raw JSON。
- P2：具备自定义 MarshalJSON 的命名 slice/map 曾按底层形状误判。
  现在尊重编码器的方法选择，最终 JSON 仍验证权限、schema 和预算。
- P2：实现 TextMarshaler 的非字符串 map key 曾被误拒绝。
  现在委托编码器处理键，并保留普通值的容量预检；不重复执行宿主编码函数。
- P1：struct 内的循环 pointer/interface 曾使预检永久循环。
  现在按已有 JSON 深度上限有界展开；参考图及 mask 的循环和超深表示在发送前失败。

## Spec

最终复审通过，0 项可行动问题。此前报告的分配前检查缺口均已修复，包含 raw
参考图、mask、URL 字段值、指针、嵌入字段、字段遮蔽及自定义键的普通值；最终修改
未发现新增问题。当前实现符合工单 10 的范围。真实 JSON 编辑接受性、首批型号能力
与目录纳入仍由工单 11 验证。

初审及后续复审共四项已修复：

- P1：pointer/struct/raw JSON 的已知容量检查迟于编码分配。
  普通值、raw URL、raw mask、escaped URL 和大量参考图均在编码前受限。
- P1：匿名嵌入字段的 image_url 曾绕过容量预检。
  现在按 JSON 名称及提升深度选择字段，Go 名称遮蔽不会跳过有效 JSON URL。
- P1：自定义 map key 的普通字符串值曾绕过容量预检。
  现在提前检查已知 URL 大小，合法自定义键仍可生成封闭的内联 JSON 请求。

- P1：image_url 字段中的 RawMessage 字符串值曾被视为未知编码而跳过。
  现在以字节视图及固定缓冲区预检；普通键、指针、结构体、自定义键均有公共 E2E。

Standards：4 项已修复、0 项遗留；Spec：4 项已修复、0 项遗留。两个轴最高级别均为 P1。
