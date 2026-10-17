# barness-ai 统一图片、引用与 Jev Box 设计

Status: proposed
Updated: 2026-10-16

本轮只交付文档。本文是接口及映射的初步设计；[spec.md](spec.md) 是交付规格，[research.md](research.md) 保存一手协议事实。任何代码形状均为拟议契约，尚未实现。本文取代研究报告最初的公共 Assets 草案。

## 1. 目标与领域边界

Host 以同一种图片数据、引用身份和放置规则调用 Jev / Decisions。图片不是 Provider 资源，引用不是授权，也不是自然语言路径。Adapter 负责协议投影，只有能够保持作用域与结构的投影才允许发送。

沿用 OperationClassifier、Client.Classify / HookedClient.Classify 和一次状态、多题、原子答案的语义。Box 是类型化图像定位答案，不新增操作、BoxClient 或聊天模拟。统一图片本体复用 Image；本期新增引用仅用于分类，聊天和图像生成继续使用各自已有内容契约。

统一不要求两协议具有相同能力：全局上下文图片可跨两者使用；State / Instructions / Criteria 内的局部图片首期由 Jev 提供。无法表达的放置在 capability 阶段拒绝，禁止上移为全局图片、转成普通字符串、拆请求或切换 Provider。

## 2. 一个图片来源、一种引用记录

### 2.1 拟议类型形状

```go
// 以下为设计形状；实际 tagged JSON 编解码须显式实现。
type ClassifierRequest struct {
    State           json.RawMessage
    Questions       map[string]ClassifierQuestion
    Images          []Image
    ImageReferences []ImageReference
}
type ImageReference struct {
    ImageIndex int            // 引用本请求 Images 的零基下标
    Placement  ImagePlacement // "context" 或 "value"
    Pointer    string         // 仅 value 使用；相对公共请求根
}
type ImagePlacement string
```

Image 沿用 Data（标准 Base64 文件内容）与 MimeType（规范 MIME）。不新增公共 Assets、format、part、image_url、文件路径或远端 URL。相同 bytes 放入 Images 两次仍是两个来源条目，不按内容静默去重；同一条目可被多次引用。

公共 JSON 新增 imageReferences；每条记录必须有整数 imageIndex 和 placement。零下标是合法值，缺少该字段不能按零补齐。context 记录必须省略 pointer；value 记录必须有非空 pointer；null、未知 placement、额外字段、大小写别名和重复字段拒绝。Image 继续采用 {data,mimeType}，不增加 discriminator。

### 2.2 引用默认与顺序

- imageReferences 省略、null 或空数组：每个 Images[i] 恰好产生一个 context 引用，顺序为 i 递增；无图时为空。这是统一契约的简写规则，延续既有公共语义。
- 非空数组为完整引用声明；只采用声明中的位置与顺序，不为未引用来源补全上下文；每个 Images 条目必须至少被引用一次。
- context 引用按其在完整声明中的先后顺序排列；value 引用由目标位置决定，不按 map 迭代顺序。允许同图多个 context 引用或 context 与 value 同时引用，分别计入展开预算。
- 两种输入在进入 Adapter 前均形成同一个不可变引用计划；没有旧 / 新执行栈或兼容 shim。旧 Decisions 无显式引用时保留已有 text-first、images-in-order 的 native wire。

### 2.3 value 引用的位置

pointer 使用 [RFC 6901 JSON 字符串形式](https://www.rfc-editor.org/rfc/rfc6901.html)，相对 ClassifierRequest 的公共 JSON 根。仅解析位置，不执行赋值、路径创建或通配查询。支持 ~0 / ~1；不支持 URI fragment、百分号解码、JSONPath、数组追加符号或负下标。数组下标禁止前导零；对象数字键仍按对象键处理；Unicode 不归一化。

| 允许位置 | 示例 | 可见范围 |
| --- | --- | --- |
| State 的值或其后代 | /state/photo、/state/frames/0 | 所有问题共享 |
| 某题 Instructions 的值或后代 | /questions/q/instructions/1 | 仅该题 |
| Choice 某候选说明的值或后代 | /questions/q/criteria/yes/example | 仅该题 |
| Score 某等级说明的值或后代 | /questions/q/criteria/0 | 仅该题，等级下标保持 |
| Bool 两侧说明的值或后代 | /questions/q/criteria/true/example | 仅该题 |

目标必须存在且为 JSON null；它与该引用共同形成一个图片节点。其他 null 保持原业务含义，尤其不改变 Choice 的无说明 null。一个位置只能有一条引用，不能指向字段名、问题名、type、整个 Questions / Criteria 容器、Images、ImageReferences 或任何授权字段。Bool 两侧说明和 Score 等级集合仍须满足原结构与数量。

State / Instructions 根可在显式引用时以 null 表示单张图片；没有相应引用时，仍执行已有 string/object/array 与非空说明规则。图片节点本身算非空内容，嵌套图片不使其父容器失去 object/array 类型。DescriptionFormats 约束普通文字容器，ImageLocations 另行约束图片位置；不能把裸 null 当作新增文字格式。

一句话夹图片用有序内容数组表达，例如 ["请对照这个样例", null, "判断 State 中的产品"] 并引用第二个元素。首期不支持字符串内字节 / 字符偏移，不解释任意文本中的 imageIndex、UUID、URL 或特殊对象。这是公共支持范围的主动缩窄，不是 Jev 原生能力限制。

### 2.4 公共请求示例

以下 Base64 为说明占位，未发送。模板说明 q 只供 Jev 使用；改为两个 context 引用且说明保持字符串即可用于两协议共同子集。

```json
{
  "state": {"photo": null},
  "questions": {
    "q": {"type":"bool", "instructions":["photo 是否与样例一致？",null]}
  },
  "images": [
    {"data":"<PNG_BASE64>","mimeType":"image/png"},
    {"data":"<JPEG_BASE64>","mimeType":"image/jpeg"}
  ],
  "imageReferences": [
    {"imageIndex":0,"placement":"value","pointer":"/state/photo"},
    {"imageIndex":1,"placement":"value","pointer":"/questions/q/instructions/1"}
  ]
}
```

## 3. Jev / System One 的具体投影

### 3.1 私有引用、assets 与内容

Adapter 为每个已使用来源分配私有高熵 key（同一来源多位置共用 key），只在已解析的 null 槽位填入 key；不对公共文本做 find-and-replace。映射为 native assets[key]={type:"image",format:"png"|"jpeg",data_base64:Image.Data}；MIME 决定 format，不按文件名猜测。首期 JSON + POST /systemone，Bearer 保持已有运行时。

生成 key 后须确认它既不等于、也不作为子串出现于宿主任何原始字符串（对象键和值）；各 key 也不能互为子串。碰撞时最多重选 8 次，仍失败则以不含正文的 invalid_request/request 终止且不创建 Attempt。生成长度和 key 总量遵守 Jev 上限。key 及最终 bytes 在重试前固定；每次新的 Logical Call 可不同。

上例 native 投影如下；<K0>/<K1> 仅为私有 key 的展示简写，实际运行使用通过碰撞检查的完整值：

```json
{
  "model":"jev-1.14.0",
  "state":{"photo":"<K0>"},
  "questions":{"q":{"type":"noul","instructions":["photo 是否与样例一致？","<K1>"]}},
  "assets":{
    "<K0>":{"type":"image","format":"png","data_base64":"<PNG_BASE64>"},
    "<K1>":{"type":"image","format":"jpeg","data_base64":"<JPEG_BASE64>"}
  }
}
```

模型收到的是服务端替换后的图片节点；普通文本、对象键、数值原拼写、候选键与等级顺序保留。无图请求不带 assets，维持既有 TypeSafe wire。原生 substring 引用只记录为研究事实；本期公共接口与最终 Hook guard 仅支持 whole-value 图片节点，避免不可逆的字符串切分。

### 3.2 context 引用投影

没有 context 时，native state 就是替换 value 槽位后的公共 State。有 context 时，native state 采用完整包裹 {state:<替换后的原始 State>,images:[<按引用顺序排列的私有 key>]}。不向原始对象添加保留字段，原始 State 可为字符串、对象、数组或被引用的根图片，均作为 state 子树保留。

这是新增 Jev context 图片路线的结构投影；用户对象中的 state / images 字段不会冲突。Host 依赖公共 State 业务字段和 context 图片次序，不需要填写 Jev 包裹路径。设计公开投影供审查，但公共契约不承诺 native 绝对路径稳定；库不改写自然语言路径，也不保证两 Provider 对自然语言的解释相同。显式 JSON Pointer 永远指公共根，在包裹之前解析，不暴露私有包裹位置给调用者。

例如公共 State={ticket:"退款"}、一个 context 图，native state 为 {state:{ticket:"退款"},images:[<K0>]}；问题数量不导致 state 图片重复插入。题内 value 引用仍只供所属题，不能挪入 images 数组。需要精确结构定位的 Jev 请求使用 value 引用。

### 3.3 实验、格式及限制

TypeSafeOptions 增加 ExperimentalImages bool，默认 false。图片 / Box 调用同时要求显式实验选择、图片资源策略、绑定图片授权、型号位置与 MIME 能力，Adapter 才发送 TypeSafe-Danger-Zone-Images: true。该选项不授予权限。Long Context 首期不开放，禁止 Hook 注入；不发明 Box 专用 header。

Jev 仅声明静态 PNG/JPEG；APNG、WebP 直传拒绝，PNG/JPEG 内容与 MIME 需一致。检查单图 decoded ≤16 MiB、每边 256–1568、面积 78,400–2,458,624、最终 body ≤32 MiB，以及 key 长度 / 总量；Host 更小预算优先。API 的处理后图片及展开体各 32 MiB 无法由原文件大小精确预测，服务端仍做最终判定，不把本地资源预算当作精确 token / 重编码预测。[官方限制](https://docs.typesafe.ai/images)

图片像素旋转、缩放、裁剪、格式转换由 Host 完成并显式记录；不在公共数据中添加自动处理标志。Multipart 后续只更换编码：同一来源一个 asset / file part，首 request JSON，固定 part 名及 boundary；不能新增第二个公共素材接口或重试执行栈。

## 4. Decisions 的具体投影

首期仅接受 context。只要引用计划中有 state / instructions / criteria value 位置，整体以 invalid_request/capability 在读取凭据前拒绝；即使该图也有 context 引用，仍不能丢掉局部用途后发送。Box 同样拒绝。限制是本次已实现协议子集，不对厂商未来能力作判断。

没有图时沿用 input 字符串。State 字符串去 JSON 引号；object/array 转保留数字原拼写的 compact JSON 文本。有图时生成唯一 role=user 消息：content 第一项 input_text（State 文本），随后按 context 引用顺序生成 input_image，image_url 为 data:<MIME>;base64,<Data>。重复引用须重复输出 block；不按来源索引重新排序、不按内容去重，不增加图片 caption 或 label。

例如 Images=[PNG0,JPEG1]、显式 context 引用顺序 [1,0,1]、State={ticket:"退款"}：

```json
{
  "model":"gpt-6-luna",
  "input":[{"role":"user","content":[
    {"type":"input_text","text":"{\"ticket\":\"退款\"}"},
    {"type":"input_image","image_url":"data:image/jpeg;base64,<JPEG_BASE64>"},
    {"type":"input_image","image_url":"data:image/png;base64,<PNG_BASE64>"},
    {"type":"input_image","image_url":"data:image/jpeg;base64,<JPEG_BASE64>"}
  ]}],
  "questions":[{"type":"predicate","name":"q","instructions":"图片中有商品吗？"}]
}
```

问题转换保持已有行为：choice map → choices[value,description]；score 有序说明 → levels[label]；bool → predicate，两侧非空文字按既有规则追加说明。说明仍只支持字符串，choice 至少两个，score 仍 2–10；不把局部图塞入 question 字符串。PNG/JPEG/WebP、最多 128 个 native image blocks 继续作为本项目支持子集；128 按引用展开次数计算，不按 Images 条目数。[现有映射](../../ai/openai_decisions_request.go)、[既有图片 guard](../../ai/openai_decisions_images.go)、[官方指南](https://developers.openai.com/api/docs/guides/decisions)

| 统一语义 | Jev 1.14 初期 | Decisions 初期 |
| --- | --- | --- |
| context 图片 | state 包裹中的 images 序列 | user content 中 State text 后的图片序列 |
| State value 图片 | 原位置换为私有引用，由服务端展开 | capability 拒绝 |
| Instructions value 图片 | 保持该题内位置 | capability 拒绝 |
| Criteria value 图片 | 保持候选 / 等级 / Bool 说明内位置 | capability 拒绝 |
| 同图重复引用 | 一个 asset、多处 key | 每引用一个 image block |
| WebP | 拒绝，由 Host 转换后重提 | 已有支持子集 |
| Box | 新 typed 答案 | capability 拒绝 |

共同子集是上下文 PNG/JPEG + 字符串说明及两协议共同问题约束。同一个公共请求仅切换 Target 和协议对应 Options，图片及引用不需重写；调用不会自动转换 Options 或回退。

## 5. 结果也使用公共引用

### 5.1 图例还原

Jev Score legend 回显 private key。先严格比较其与最终冻结 native Criteria 的值、键集及位置，再还原为公共 Legend：图片节点变回 null，并在 ScoreAnswer.LegendImageReferences 中附同一种 ImageReference 记录。此处 placement 只能为 value，pointer 相对该答案的 Legend 根，例如 /0/examples/1；不相对请求根。普通字符串不得被当作图片。

ClassifierResult 新增 ImageInputs 元数据数组，类型为 []ClassifierImageInput，JSON imageInputs；每项字段为 MimeType（JSON mimeType）和 SHA256（JSON sha256），后者是 decoded 原文件 bytes 的 SHA-256，编码为 64 个小写十六进制字符。ScoreAnswer 新增 LegendImageReferences []ImageReference，JSON legendImageReferences。均不含图片 bytes、私有 key 或厂商资源 ID。答案中的 imageIndex 引用该最终数组；数组仅在全部答案验证成功后发布，文本调用省略。原始请求图片索引与最终输入索引不作恒等承诺：Hook 可增删 / 换图，Host 按元数据核对最终图片身份。

最终数组按最终 native 请求第一次使用图像的确定顺序建立：先 state（对象键按既有确定键序、数组按原序），再按确定问题键序遍历其 instructions/criteria；同一 Jev asset 复用一项。Decisions 按最终图片 content 次序逐项记录，重复 block 可产生相同 hash 的不同项。hash 只说明文件内容，不授予权限；不能跨调用直接回放引用。当前 slice 不承诺不同投影下最终索引相同。

Example：原 Score 第 0 等级是 {example:null} 且引用一张图，公共 Legend[0] 仍为 {example:null}，LegendImageReferences 给出 imageIndex 与 /0/example；ImageInputs 给出该图 MIME/hash。Host 无需认识 TypeSafe key，亦不会把图片说明当作真正空值。没有图片的 Legend JSON、分布和有界舍入保持不变。

这两个结果字段解决当前真实的多模态图例与 Hook 最终归属问题；不引入图片存储服务、持久化资产注册或返回原始图像。元数据构造受图片 / 引用数量预算约束；它不是厂商 output_tokens 或响应 body 的一部分。

### 5.2 Box

BoxQuestion 只有非空 Instructions（允许含图片节点），实现 sealed question variant；BoxAnswer 含 Box 四边与 Nullable[float64] Confidence，公共 type 为 box。四边显式存在、有限、[0,1]、left<right、top<bottom；confidence 必须存在，null 或有限 [0,1] 数字，Unset 非法。没有虚构 image_id、criteria、not-found 或 refusal。

每个 Box 至少能看到一张图片：来自 context、State 或该题自己的 Instructions；图片仅出现在其他问题中不能满足前提。无可见图为 invalid_request/scope，Hook 后丢失可见图为 callback_failed/request。该检查只证明有图片可用，不推断搜索图或目标存在。

图例引用位置只能定位图片内容，不能证明自然语言指定哪张图是 Box 搜索图。BoxAnswer 不附推断的 imageIndex；Host 维护题目目的与目标图关联，需要无歧义时限制搜索图数量。存在性用同批 BoolQuestion 另问，阈值与是否采纳框由 Host 决定；不建立条件执行，confidence 不是存在概率。像素坐标转换使用 Host 实际显示图片的方向、尺寸与裁剪变换。[Box](https://docs.typesafe.ai/primitives/box)

## 6. 能力、资源与不可变数据流

ClassifierCapabilities 新增 ImageLocations 列表，取 context/state/instructions/criteria；图片型号必须至少有一个位置，未知 / 重复位置拒绝，文本型号不能声明位置。省略时对已有 Decisions 图片声明解释为 context，对 text-only 为无位置；查询不写入省略默认，不改旧目录 Hash。Jev 新图片型号显式声明实际支持位置，Kinds 加 box 仅限已确认支持型号。位置与 MIME 能力取协议实现、目录声明、资源和 Binding 的交集。

MaxInputImages 继续约束来源条目数；Decisions 另将展开引用数限制到 min(128,MaxInputImages)，防止借复用绕过原容量。TypeSafe 不发明原生图片数量上限，使用 Host / 型号明确支持的有限子集。整个目录校验、clone、Hash、查询都贯通新能力，目录变化必须新版本。

ClassifierImagePolicy 新增可选 References 子策略，含正数 MaxReferences / MaxExpandedImageBytes。nil 时沿用既有有限数量/总字节预算作为引用上限：分别取 MaxImages / MaxTotalImageBytes；这是已有预算的继承，不表示无限制。需要更大复用规模时 Host 显式配置。旧策略下每图一次的默认 context 自然满足相同预算；新增复杂输入不自动取得更大容量。

| 预算 | 计数口径 |
| --- | --- |
| MaxImages / MaxInputImages | Images 来源条目数，重复 bytes 也分别计数 |
| MaxImageBytes | 每来源 decoded 原文件 bytes |
| MaxTotalImageBytes | 所有来源 decoded bytes 总和 |
| MaxReferences | 规范化后的实际图片出现次数，含默认引用 |
| MaxExpandedImageBytes | 按每次出现累计 decoded 原文件 bytes；state 共享只算一次 |
| MaxStateBytes / MaxQuestionBytes | 公共文本树与最终 native 对应文档含生成 key / 包裹结构；assets bytes 单独计 |
| MaxRequestBytes | 公共输入已知长度先查，最终实际 native 编码再次查，包含引用表 / Base64 / JSON framing 各自所在阶段的实际成本 |
| MaxOutputBytes | 现有完整 Provider 成功响应 body 限制，不因多模态放宽 |

计数使用剩余预算减法避免溢出，复制前拒绝已知超限；严格 Base64 与格式检查流式执行、检查取消，尺寸读取不解压完整栅格。References 配置、Images、原始 JSON 与引用计划均独立拥有。厂商重编码和 tokenization 无公开精确公式，本地只保证自身资源上限；usage.input_tokens 按原报告估价一次，不加图片次数附费。

顺序：scope / 结构与已知大小 → 复制 / 解析引用 → Binding / Catalog / 能力及图片授权 / Options → pin 凭据快照 → 独立 native build → header Hook → payload Hook → 最终严格 guard / freeze → sendUnary → usage → 全部答案验证和图例还原 → 一次发布结果。

## 7. Hook、错误、观测与撤销

header / payload Hook 的时序沿用既有协议实现；最终 header 与 body 必须交叉检查图片权限、实验选择和 Long Context。Jev 最终支持全值 key、JSON Base64 PNG/JPEG；part、外部 URL、native substring 引用及 schema 外字段拒绝。素材 key 若出现在对象键或非图片用途的文字里也拒绝，避免回显中出现不可解释的私有引用。Decisions 最终保持唯一 user、text-first、随后图片、仅 data URL 的闭合输入树。不能按公共旧请求而漏检 Hook 新增素材。

Jev 最终 asset 的允许字段明确为 data_base64、format、可省略的 type；type 存在时必须为 image，format 接受 png/jpg/jpeg，jpg 与 jpeg 都还原为 image/jpeg。不得夹带 part，即使同时有合法 Base64；字节、格式与 MIME 一致性仍校验。Builder 总是输出显式 type 和规范 png/jpeg，不以 Hook 的合法省略扩展公共契约。

每协议拥有独立最终 native 记录，重建题目、图片用途、预算和结果引用依据；Hook 合法换题、换图、重排或添加素材均按最终 native 接受子集复核。私有 key 的改名不作为公共行为，任何 key 替换仍需唯一、全部使用且无歧义。所有回调一次，重试固定最终 bytes、headers、key / 数据、凭据与目录版本；2xx 后不重试。

| 失败 | Code / Phase；副作用 |
| --- | --- |
| 坏引用、缺图、非法 null / 位置、非法图像 | invalid_request/scope；零凭据读取、零 Attempt |
| 原始资源超限 | resource_limit/scope；已知量在复制前检查 |
| 协议 / 型号不支持位置、格式或 Box | invalid_request/capability；零凭据读取、零 Attempt |
| Jev 专属图像限制 | APNG / 最小尺寸等非法为 invalid_request/capability；文件/尺寸/面积最大值超限为 resource_limit/capability；均不读凭据 |
| Binding 未授权图片 | tenant_denied/capability；零凭据读取、零 Attempt |
| 缺实验选择或错误 Options | invalid_request/capability；零凭据读取、零 Attempt |
| Hook 损坏形状 / 引用 | callback_failed/request；不发送 |
| Hook 扩大认证、目标或图片权限 | tenant_denied/request；不发送 |
| 构建后 / Hook 后资源超限 | resource_limit/request；不发送 |
| 坏 Box / 答案 / 图例 | protocol/response；Answers、ImageInputs 不发布，唯一合法 usage 保留 |

通用 Base64 / 文件签名 / 引用与 Host 已知预算先在 scope 检查；需知道协议 / 型号的 Jev 专属检查在 resolve 之后、pin 之前完成，不把新的 Jev 尺寸 / 静态图限制施加给 Decisions。Hook 后相同专属形状失败归 callback_failed/request，超最大容量归 resource_limit/request。上述新增规则不改变无引用旧请求的已有错误优先级。Observer 仅现有元数据；不记录图片 bytes、引用路径、hash、私有 key、框、正文或错误文本。错误正文只回当前租户，遵守现有脱敏规则。

撤销 AllowClassifierImages、移除 AllowedModels 或禁用 Binding 只阻止新调用；已 pin 调用按原快照及有限重试完成，usage 保留。回滚不改走文本 / Jev 1.13 / 其他协议。

## 8. 验收接缝与实施边界

唯一业务 Seam 是现有 Classify Interface：用真实 Client、真实解析/转换/运行时接可控 HTTP Provider；Client 与 HookedClient 是该接缝的两个入口。通过返回结果、捕获请求、凭据读取次数、Attempt、取消及资源释放验证外部行为，不直接测试私有替换 helper。用户已在 2026-10-16 的 to-spec 接缝核对中明确确认使用这个既有公共接缝。

先列失败方式并写公共 E2E，再实现；不在实现后补单元测试。既有 Decisions 图片 / Hook / 字节 / 隔离测试、TypeSafe 混合题 / legend / 用量及真实 wire replay 为先例。迁移旧例子到同一公共图片来源，不建第二套镜像测试。

必测组：默认与显式 context 等价；[1,0,1] 顺序/复用；嵌套 value、RFC 转义和数字对象键；null 与普通 null 区分；越界/缺失下标/路径、重复或禁区位置、未引用来源；Jev 各位置作用域、普通相似文本不替换、包裹字段不冲突；Decisions 整体拒绝局部用途；MIME/APNG/尺寸/展开预算；最终 Hook 换图换题、字节与 header 越权；Score 图例及最终 hash 不泄露 key；Box 空 confidence/坏边/缺目标；并发同名租户与取消/重试/2xx 错误。所有组覆盖两个公开入口。

每个 E2E 生成可重放请求/响应、合成合法图片及 hash、结果、错误、Observer / Attempt、资源断言、运行配置和完整性清单；脱敏审计通过后才算证据。race 验证同一语义，不以运行时间或随机 key 固定值作通过条件。

真实验证另外使用固定 jev-1.14.0 候选目录和严格预算，记录 RequestID / ResponseModel / Catalog Hash；Jev 与 Decisions 各自形成图片能力证据。Box 用人工标注 IoU 与目标缺失样本分开评估存在误报，不能以类型正确代替定位效果。此轮不发请求、不改内置目录 / 支持矩阵 / 冻结 Pi，也不产生 PASS 声明。

## 9. 决策影响及待验证项

[拟议 ADR-0030](../../docs/adr/0030-barness-ai-unified-classifier-image-references.md) 记录公共引用与原生投影的取舍；扩展 ADR-0020 / 0021 / 0026 的模型能力、问题集合和图片路径，不覆盖其快照、重试、usage 和拒答保证。已有 proposed ADR 不因写文档而变为 accepted。

实施需同步公共说明、目录 / 资源配置、结果消费者、Pi 扩展账本与证据 schema；新结果字段为 additive，Host 解析器需接受。原 Images 字段及默认语义保留；未落地的 Assets 草案直接删除。无图 wire 与旧 Decisions 图片公共行为由回归锁定，目录语义变化另按 Hash 合同发布。

待真实验证：Jev context 包裹对任务效果的影响、多图 Box 自然语言归属、仅题内图的 Box 行为、重编码后容量和图片 tokens、null confidence 的出现频率、multipart 收益。这些不阻止已明确语义的离线实现，但不能被记为已通过；若损害预期效果，应修订投影及自己的证据，不能无声改变公共引用语义。
