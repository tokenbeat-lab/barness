# 05: 打通分类图片、三重授权与完整资源边界

**What to build:** 已明确配置的 Host 和租户通过 Classify 单选判断文字及有序内联图片；图片政策、型号能力和 Binding 权限任一缺失都在网络前拒绝，Hook 后同样执行授权与资源检查。

**Blocked by:** 02: 打通 Decisions 文本单选与最终请求闭环

**Status:** resolved
**Type:** task
**Source tickets:** 02, 05
**Requirements:** R03, R05, R06, R07, R08, R09, R14, R19
**Acceptance cases:** OD05, OD06, OD07, OD08, OD15, OD17, OD18, OD19, OD20, OD21, OD22, OD23, OD24, OD25, OD26, OD42
**Updated:** 2026-10-12

## Responsibility

负责本切片所需 Images JSON、图片目录元数据、ClassifierImagePolicy、Binding.AllowClassifierImages、入口验证/复制、native message 与最终 guard。复用纯 base64/header 校验，共享全局单图/请求限额；GenerateImages 继续独立。文本单选足以验证，不依赖混合题或计费扩展。协作时保留他人改动。

## Failure scenarios first

- nil 政策、text-only、默认 false 权限仍能带图；TypeSafe 丢图后返回成功。
- 图像生成政策意外授权分类，非 classifier Binding 打开新字段，目录修改扩大租户权限。
- 坏 MIME/magic/base64/padding/CRLF，URL/path/file ID/detail 被修补、下载或隐式放行。
- 在复制/解码巨大图之后才判限额，数量/decoded 总量/JSON 编码字节任一漏计。
- Hook 添加第二条 message、后位图片、巨大问题名，或用 payload 自报授权逃过检查。

## Acceptance criteria

- [x] 全部格式/计数/字节/权限和 Hook 失败场景先写公共 E2E，再落地公共字段与完整图片路径。
- [x] Images 省略/null/空数组为无图；严格公共 JSON 拒绝非对象、null/非字符串、未知/case-alias 字段；无图 Jev wire 与省略态不变。
- [x] Classifier.Images nil 关闭、非 nil 的 MaxImages/MaxTotalImageBytes 正且有界；GenerateImages 子政策独立；AllowClassifierImages 默认 false，错误操作 invalid_request/binding。
- [x] 图片目录 Input/MaxInputImages/MIME 声明、默认、矛盾与 Hash/clone/Lookup/List/Client.Catalog 深复制均验证；新协议支持声明不得超出已实现范围。
- [x] State 仍合法且必填；真 PNG/JPEG/WebP 有 hash 可辨资产，只做严格编码/header 校验，不分配 raster，不拉取资源。
- [x] scope 阶段先计 known size 和 checked arithmetic，再验证和复制；每图 decoded、总 decoded、count 与整段 JSON 编码分别测界。
- [x] 无图 input string，有图一个 user、text-first、ordered data URLs；128 张有界离线可发送，129 capability 拒绝，较小 Host 政策先 scope 拒绝。
- [x] 三授权条件缺失有确定 Code/Phase 与零 credential/Attempt/request；TypeSafe 收图明确拒绝；Hook 加图按已 pin 权限检查。
- [x] 最终 guard 检查全部 message/content/图片/问题及完整请求；越权 tenant_denied/request、坏 schema callback_failed/request、超资源 resource_limit/request。
- [x] 图片验证可取消；原 slice/回调引用在同步 barrier 后改变不影响最终字节；错误后 body/permit/waiter 归零。
- [x] 保存图片原资产/MIME/hash、decoded/encoded/whole JSON 读数及最终捕获，附 SHA-256 manifest、整包审计与准确重放命令；旧文本 Jev 和图像生成对照通过。

## Completion

独立图片单选闭环可演示。三题型由混合题切片拥有；压力与自身 live 仍需各自证据。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 02

2026-10-12：待实施，公共字段目前尚未增加。

### 原票 05

2026-10-12：图片是本功能必交能力，三个开关/能力检查缺一不可。


### 本票实施（2026-10-12）

已完成独立图片单选离线闭环：严格 Images JSON、ClassifierImagePolicy、Binding.AllowClassifierImages、
目录图片元数据与独立复制/Hash，以及单 user/text-first/有序 data URLs 的原生转换。图片政策、
型号能力和租户权限缺一即拒绝，入口在凭据前执行 known size/strict base64/header 检查；Jev 收图明确拒绝，
无图 wire 和 GenerateImages 政策保持独立。

Hook 后按 pinned Binding 检查整个 message/content 树和最终问题；所有图片、每图/总 decoded、数量、
状态及完整 encoded JSON 都有界。typed/native struct、raw JSON、byte slices 和 custom marshalers 遵循
实际 JSON encoder 的边界，不通过请求自报权限。同步 header/payload/response barrier 后修改原 slice、
问题、Binding 或回调 map 不改变最终发送字节；取消与错误后的 body/permit/waiter gauges 归零。

实现与审查修复至 `926ed8a`。失败场景先记录于 [image-failures](../image-failures.md)，公共 E2E 保留
格式/授权/资源失败，以及审查发现的 message type、raw text 分配、Unicode、typed 整请求和 byte-slice
编码反例的红→绿。最终 131 个图片 case 全 PASS；最终专项 race 连同 Jev/Classifier 对照共 431 PASS。
两轴 code-review 最终 Standards 0、Spec 0，详见 [REVIEW](../image-evidence/REVIEW.md)。

完整执行状态、编译版本、最终整套/差分/vet 及整包审计以 [validation](../image-evidence/validation.json)
和 [证据 README](../image-evidence/README.md) 为准。包保存真 PNG/JPEG/WebP 原资产/MIME/hash、
decoded/base64/public/native JSON 读数、最终 HTTP 捕获、九次红色历史、每文件 SHA-256、资源/Observer
读数和准确重放命令。其他既有工作区改动保留；本票未新增内置型号或厂商真实调用。
设计负载、新 Pi Decisions 差分、自身 live 和发布结论继续由对应后续工单提供。
