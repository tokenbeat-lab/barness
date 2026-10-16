# Host 统一图片输入与结果视图

此命令用真实 `Client.Classify` / `HookedClient.Classify`、本地受控 HTTP Provider
和合法合成 PNG/JPEG，重放统一输入、局部图片、视觉图例、Box、拒答及撤销。
Jev 使用正式内置目录中的固定 `jev-1.14.0`；`jev-1.13.0` 仍仅文本。
图片与 Box 仍需显式实验选择、策略和 Binding 授权，目录纳入不代表生产准确率。
受控结果只证明协议与 Host 流程，真实服务和视觉质量均为 NOT_RUN；不修改正式目录、
支持矩阵或生产可用性结论。[ADR-0030](../../../docs/adr/0030-barness-ai-unified-classifier-image-references.md)
保持 proposed，父规格不变。

```sh
go run ./ai/examples/classifierimages
# 指定一个新的目录；已有目录会拒绝，避免覆盖历史证据
go run ./ai/examples/classifierimages -out .evidence/my-unified-images
# 单场景重放；无 key、无远程网络
go run ./ai/examples/classifierimages -route jev -scenario hook -entry hooked
go run ./ai/examples/classifierimages -route decisions -scenario state -entry plain

go test ./ai/e2e -run '^TestClassifierImagesExample' -count=1
BARNESS_EXAMPLE_RACE=1 go test -race ./ai/e2e -run '^TestClassifierImagesExample' -count=1
```

默认 48 个场景覆盖两协议和两个入口：context、reuse、state、instructions、criteria、
box、image-revoked、binding-revoked、budget、canceled；另有 Jev 的 bad-box、
bad-legend、experiment-off、hook，以及 Decisions 的 refusal。预期拒绝也要求明确
Code/Phase、零发送与原子结果，不能仅凭进程成功认定可用。受控 Provider 按最终请求
返回固定答案及图例，不进行视觉判断；本例故意返回存在性零与合法 Box 来证明二者独立。

## 同一公共输入切换协议

`inputs.go` 的公共请求不含 assets、私有 key 或 native URL。图片来自 Host 完成预处理后
编码的文件；两协议共用同一 `ClassifierRequest`，只替换明确的 Target 与 Options：

```go
input := ai.ClassifierRequest{
    State: json.RawMessage(`{"task":"synthetic rectangle"}`),
    Images: []ai.Image{pngImage, jpegImage}, // Data 为严格标准 base64
    Questions: map[string]ai.ClassifierQuestion{
        "exists": ai.BoolQuestion{Instructions: json.RawMessage(`"Is the rectangle present?"`)},
    },
}
result, err := client.Classify(ctx, scope,
    ai.Target{BindingID: "jev", ModelID: "jev-1.14.0"}, input,
    ai.TypeSafeOptions{ExperimentalImages: true})
// 保持 input，替换为 Host 自己授权的 Decisions Target：
result, err = client.Classify(ctx, nextScope,
    ai.Target{BindingID: "decisions", ModelID: "gpt-6-luna"}, input,
    ai.OpenAIDecisionsOptions{})
```

每次新调用必须使用新的 RequestID；tenant 来自认证 Host 身份。显式复用/排序：

```go
input.ImageReferences = []ai.ImageReference{
    {ImageIndex: 1, Placement: ai.ImagePlacementContext},
    {ImageIndex: 0, Placement: ai.ImagePlacementContext},
    {ImageIndex: 1, Placement: ai.ImagePlacementContext},
}
```

省略/空引用时按来源次序各使用一次；显式引用是完整用途声明，所有来源必须被使用。
`value` 指向公共请求的现存 null 槽位，例如 `/state/photo`、
`/questions/exists/instructions/example`、`/questions/score/criteria/0/example`。
每个引用只在自己的位置生效，无法表达的位置整体拒绝。Decisions 对所有局部图片、Box
在 capability 阶段且凭据前拒绝，不提升作用域、不静默改写、不改走其他协议。

## 授权、能力与容量

三项授权缺一不可：型号声明 image/MIME/count/location，Binding 授予
`AllowClassifierImages`，`Policy.Classifier.Images` 显式启用。Jev 另外要求每次调用
`ExperimentalImages: true`；它选择实验 header，不授予权限。Host 候选才声明 context、
state、instructions、criteria 和 box。内置 Jev 目录及真实支持状态不因此改变。

本例只运行一个串行合成调用：最多 2 个来源、每图 64 KiB、总 decoded 128 KiB；最多
8 次引用、展开 decoded 512 KiB；完整 native 请求 1 MiB、响应 64 KiB、最多 4 题、
State/单题各 4 KiB。数字是示例负载预算，不能当作库默认。References 为 nil 时，
展开数/字节分别继承来源的有限上限；复用仍计每次出现，最终 native Hook 再校验。
Decisions 按最终图片 block 计数，Jev 按最终 asset 顺序发布身份；同一来源复用可只占一个
asset，但出现预算仍独立计数。全局 MaxRequestBytes 包括编码与 JSON 开销。

示例先装配 Client，再撤销 resolver 中的图片许可或 Binding；下一次调用读取新版本，
不读凭据、不发请求。预算超限、实验未选和预取消也在发送前失败。真实云端应用应由
自己的授权存储原子更新版本；本地不可变装配在权限改变后重新装配。不会自动 fallback。

## 最终身份、图例、拒答与 Box

`ClassifierResult.ImageInputs` 是最终 MIME 与 decoded 文件 SHA-256 的列表。
Decisions 的重复 block 各有一个条目；Jev 每个最终 asset 一项，遍历 State，再按题名排序，
每题 Instructions 先于 Criteria，对象键排序、数组保序。不要拿入口 Images 下标解释结果。
`hook` 场景通过可信 native Hook 发现 Adapter 生成的 key 并互换 PNG/JPEG 内容与格式；
公共输入无需知道这些 key。`box/hooked` 也换图，并显式更新 Host 的搜索图关联。

`ScoreAnswer.Legend` 图片位置还原为 null；`LegendImageReferences.Pointer` **相对 Legend 根**，
如 `/0/example`。`ImageIndex` 索引最终 `ImageInputs`。Host 按最终 MIME/hash 查找自己已持有
的文件，按 Pointer 在 UI 中放图；hash 不表示访问权或 Provider 资源归属。`interpret.go`
只处理本例明确的图例结构，其他业务结构应由对应 Host 视图转换，而非发明通用公共 AST。
错误清空全部 Answers 与 ImageInputs；独立合法 usage 与 Attempt 元数据仍保留。

Host 用封闭 switch 处理 Choice/Score/Bool/Box/Refusal。Refusal 是明确的逐题状态，
不能当作 false，也不虚构原因。Box 的四边为左上原点、归一化且正面积；Confidence 必填，
`IsNull()` 区分 null，`Get()` 保留数字零。存在性是独立 Bool 概率，不因它为零而删除 Box。
本例不设业务阈值；业务阈值、条件执行及后续重问预算归 Host。

多图搜索关联是 Host 责任：问题 `location` 明确指定第一张 context 为搜索图，第二张仅为
参考；Host 保存该文件的最终身份，并在自己的换图 Hook 中更新关联，不能推断模型返回了
image_id。实际文件预处理是原图 640×480 → 顺时针 90° 得 480×640 → 裁剪
`[40,80,360,560]` 得 320×480。框 `[left=.25,top=.25,right=.75,bottom=.75]`
依次变为裁剪图像素 `[80,120,240,360]`、旋转图 `[120,200,280,440]`、原图
`[200,200,440,360]`。图片没有未处理的 EXIF 方向；现实 Host 必须先正确处理实际方向及裁剪，
不能只拿原始文件宽高相乘。边缘逆变换使用 `H-x`，栅格像素旋转使用 `H-1-x`。

## 真实模式：仅由 Host 显式注入

此票没有运行真实服务。真实模式只运行一个指定场景，要求 Host Binding 明确零重试；不会自动找其他 key、catalog、
endpoint、Proxy 或协议。Host 提供 JSON `hostConfig`：`binding`、`target`、`catalog`、
`policy`，字段为对应的 ai 公共类型。可参考受控输出 `cases/.../config.json` 的 `host`
结构，自行填入可信 tenant、版本、账户、正式 endpoint、允许型号和明确核验的候选目录。
配置不能包含凭据；`-key-env` 是唯一凭据来源，秘密值不会存盘。可信配置不要来自用户 body。

```sh
# MY_JEV_IMAGE_KEY 的值由 Host 在启动前注入；这里不读取任何其他 key。
go run ./ai/examples/classifierimages -live -route jev -scenario context -entry plain \
  -config /path/to/trusted-host.json -key-env MY_JEV_IMAGE_KEY \
  -out .evidence/explicit-live-unified-images
```

仅允许 context/reuse/state/instructions/criteria/box；entry 可为 plain/hooked。
真实捕获按 Host 请求/响应预算有界读取（本例捕获器至多 16 MiB）；HTTPS 正常校验证书。
真实成功标为 OBSERVED，真实服务结论为 OBSERVED_NOT_EVALUATED，视觉质量仍为 NOT_RUN，
不冒充支持矩阵 PASS。该命令不承担工单 11/12 的完整真实协议证据或独立视觉效果评估。

## 证据与重放

每次运行生成原图和处理后 PNG/JPEG、方向/尺寸/裁剪/hash provenance、每场景完整公共
input、Host config/options/scope、最终 HTTP requests/response、result/error/view、Attempt、
Observer、资源归零记录与断言。`manifest.json` 封存所有产物 hash，`audit.json` 使用现有
脱敏审计；输出仅限调用归属 Host，不应上传至共享日志。审计失败/断言失败命令退出非零；
原目录保留便于诊断。每场景 config 包含单场景重放命令，动态 RequestID/时间不要求一致，
合成图像内容及 hash 可重复。

[工单 10 验证](../../../.scratch/jev-1.14-multimodal-box/10-verification.md)保留首红、行为红、
完整普通/race 包和独立完整性核验。撤销、取消和拒绝不会消耗额外重试或制造结果。
