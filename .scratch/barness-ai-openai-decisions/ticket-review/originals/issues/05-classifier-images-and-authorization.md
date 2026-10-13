# 05 — 分类图片输入、租户授权和资源检查

Status: ready-for-agent
Type: task
Depends on: 04
Updated: 2026-10-12

## What to build

启用内联图片分类的完整 public path：Request/Model/Policy/Binding 三条件、有效格式、大小/count/clone，native text-first input 和 Hook 后图片检查。依据 [R03/R05–09/R14](../spec.md)、[设计 §2/4/5](../design.md)、[OD07–08/15/17–26](../acceptance.md)。

## Ownership and context

负责 classifier 的图片 policy/shape/byte/capability 接缝、Decisions input 构造与 final guard、图片 fixture 和 artifact。复用已有纯 base64/header 校验；不借 GenerateImages 的 mask/16 张/20 MiB/权限路径。协作不回退他人编辑。

## Failure scenarios first

- policy nil/model text-only/Binding flag false 仍能带图，或 TypeSafe 静默丢图。
- GenerateImages 的开关启用 Classify 图片；共享 Catalog 能力变化授予未授权租户权限。
- 只数 public Images，不检查 Hook 后嵌套 message 图片或额外后位图片。
- 忽略 decoded 总字节、base64/escaping 请求字节或 count；先复制巨大图再拒绝。
- MIME/编码不合自动纠正，URL/file ID 被拉取，图顺序变更、资源验证不可取消。

## Acceptance criteria

- [ ] 先写 OD07–08/15/17–26 的全部失败/边界公共 E2E，再实现图片。
- [ ] Binding.AllowClassifierImages 默认 false 且 pinned；final hook 加图同样授权，authority 不能从 payload 提供。
- [ ] policy 的 MaxImages/MaxTotalImageBytes 和 ResourcePolicy.MaxImageBytes/MaxRequestBytes 都有实际读数与边界。
- [ ] inline PNG/JPEG/WebP 的有效资产/编码/header、原数组和返回图信息独立；无下载/path/detail/resize。
- [ ] 捕获单 user、text-first、ordered image data URLs；128 的有界离线通过，129 的 API/model 边界正确拒绝。
- [ ] 拒绝阶段/Code 按设计；Jev 无图完全不变，收到图明确拒绝，图像生成不获权限。
- [ ] 字节/图片验证取消和 over-limit 后不泄漏 permit/body，不分配完整 raster。

## Verification and handoff

保存每图 asset/MIME/hash、decoded/encoded/whole JSON limits 和 hook 最终请求，重放有效与失败例。同步 marker 后修改已复制 slice，证明请求不漂移；不制造并发读取同 map 的测试竞态。将完整图片与文本 path 交给 06，尚无自己的 live 不声称厂商已通过。

## Comments

2026-10-12：图片是本功能必交能力，三个开关/能力检查缺一不可。
