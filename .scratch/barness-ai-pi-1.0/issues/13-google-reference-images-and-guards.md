# 13: 完成 Google 参考图编辑与请求守卫

**What to build:** 宿主通过 Google 的同一图像入口提交提示和有序参考图编辑图片；可信回调、型号能力与总请求大小受到完整检查，不能引入厂商状态、外部资源或其他执行方式。

**Blocked by:** 12 — 通过 Google Interactions 完成图像生成。

**Status:** resolved

**依据：** 规格“图像：共同契约”“图像：Google Gemini Interactions”“图像回调边界”、实施设计第 8.1、12.3–12.9、13 节。

- [x] 先写参考图编辑、MIME、数量/字节超限及每个拒绝字段的回调失败 fixture，再扩展输入与守卫；构造前和回调后的场景均经公共入口验证。
- [x] 输入编码为 prompt 文本在前、参考图按原顺序在后，图片为内联 base64 与 MIME；复用现有图像类型校验并独立复制。首期公共输入不支持交错文本/图片或 mask，该收敛有扩展登记。
- [x] 参考图上限取宿主策略、型号能力和协议硬限制的较小值，协议最多 14 张；首批型号关于 10 个物体加 4 个角色的能力说明保留。每张图片受已有单图字节限额约束。
- [x] 含 prompt、全部内联图片与编码开销的最终请求不超过 20 MB，并遵守更小的宿主 MaxRequestBytes；先按已知长度/base64 膨胀估算检查再复制，回调插入或扩大图片后重新检查总量与单图限额。
- [x] 尺寸与宽高比均按最终请求的型号能力验证，首批型号只支持 1K/2K/4K 时拒绝 512；数量约束不变成精确 N 承诺，也不自动重发凑数。
- [x] 每个禁止项有独立失败证明：previous_interaction_id、background、agent、tools、environment、webhook_config、continuation_token、service_tier、stream、input URI 及 URI delivery；删除或放宽 store=false、type=image、delivery=inline 也被拒绝。
- [x] 回调改变型号、认证、目标或操作不能扩大授权，无效请求体为 callback_failed，扩大授权为 tenant_denied，资源超限为 resource_limit；所有前置失败不读取任意资源、不发 Provider 请求。
- [x] 编辑完整保留全部输出步骤的文本/图片顺序，严查坏 base64、MIME 不符、单图/总图字节和输出数量；失败原子清空输出、保留用量且不重放，响应体与许可均释放。
- [x] 补齐 P09 的边界追溯、公共输入收敛的扩展说明及协议 ADR；产出生成/编辑、各拒绝项、取消/释放、资源限额、观测脱敏和既有路线回归的可重复离线证据。


## Answer

2026-10-08：已完成并关闭。Google GenerateImages 通过同一 v1beta Interactions 入口，
发送 prompt 文本及原顺序的内联参考图，并独立冻结输入；数量取宿主、型号和 14 张协议
上限的交集。初始输入与最终回调均验证 MIME/base64、单图字节、完整请求编码开销，
协议总量为十进制 20,000,000 字节，并服从更小的宿主预算。已知普通容器及 raw JSON
在复制/编码前检查；自定义编码按已有契约冻结一次，再校验最终授权、能力与预算。
同步无状态固定项、全部禁止字段及 URI/目标/认证/操作扩大均独立拒绝；尺寸/宽高比按
最终型号检查。编辑复用全部输出步骤的有序解析、原子失败、合法用量保留、不重放、
取消与资源释放。ADR-0022、契约、扩展 ledger 与 P09 追溯已同步；真实目录/live 归工单 14。

验证提交 `b1ac4da6da3db8cb57183201062c0ec31de330bc` 通过两种 go vet、Google/OpenAI/Unary 定向 E2E，以及开启
冻结 pi 差分和压力场景的全量 `go test -race ./... -count=1`。4276 个场景全部 PASS，
Google 268 个、OpenAI 图像 267 个、pi 差分 623 个（pending 0）、压力 11 个；
Google 40 个资源场景的全部记录 gauge 归零。原始 run 与交付目录脱敏审计均为
0 发现，源码/逐场景/交付哈希验证通过，七个 P09 追溯项及 P0/差分/目录快照均 PASS。
已有路线回归包含共享编码守卫的自定义 byte 元素接受性证明。

Standards 各轮报告 7 项具体发现（含 IsZero 残余细化）、Spec 4 项，全部先补公共
E2E 红灯，再修复并独立复核；两轴最终均无遗留项。证据保留 17 个实际红灯场景，
以及最终脱敏请求/脚本/结果/断言、分配量、资源读数、Observer 审计与版本/哈希。
[可重复核验的离线证据](../google-reference-images-evidence/README.md)。
