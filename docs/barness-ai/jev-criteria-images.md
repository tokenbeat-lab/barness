# Jev 视觉标准与公共 Score 图例

Host 显式候选目录中的固定 `jev-1.14.0` 可声明
`ClassifierImageLocationCriteria`。Choice 的单个候选说明、Bool 的 true/false
说明、Score 的每个有序等级都能使用根图片或结构化说明里的图片。
图片仍来自 `ClassifierRequest.Images`，用 `ImageReferences` 的 `value` Pointer
指向现存 null 槽位；Pointer 相对公共请求根，例如：

```go
req := ai.ClassifierRequest{
    State: json.RawMessage(`"待评分对象"`),
    Images: []ai.Image{lowExample, highExample},
    Questions: map[string]ai.ClassifierQuestion{
        "quality": ai.ScoreQuestion{
            Instructions: json.RawMessage(`"按示例评分"`),
            Criteria: []json.RawMessage{
                json.RawMessage(`{"label":"低","example":null}`),
                json.RawMessage(`{"label":"高","example":null}`),
            },
        },
    },
    ImageReferences: []ai.ImageReference{
        {ImageIndex: 0, Placement: ai.ImagePlacementValue,
            Pointer: "/questions/quality/criteria/0/example"},
        {ImageIndex: 1, Placement: ai.ImagePlacementValue,
            Pointer: "/questions/quality/criteria/1/example"},
    },
}
```

调用同时要求 Classifier 图片资源策略、目录的图片/MIME/位置/数量声明、
`Binding.AllowClassifierImages` 和 `TypeSafeOptions{ExperimentalImages:true}`。
目录声明不会自动授予租户权限。普通 object/array 说明仍受 DescriptionFormats
限制，明确引用的根图片独立受图片位置能力限制。

未引用的 Choice null 保留“仅以候选名说明”的含义；Score/Bool 根 null 必须明确
被引用。标准容器、题型字段和候选名称不能成为图片目标。图片保留在本题本标准；
复用来源只上传一次，每个实际出现单独计入引用数和展开字节预算。不能用局部标准
图片扩大其他题目的可见范围。Decisions 含任意局部标准图片时，在凭据前整体拒绝。

ScoreAnswer 的可选 Legend 先严格匹配最终 frozen Criteria，再把实际图片值还原为
null，并返回 `LegendImageReferences`。每条引用的 placement 为 value、pointer
相对该答案 Legend 根，如 `/0/example`；imageIndex 指向同一次结果的 ImageInputs。
因此本例可返回：

```json
{
  "legend": {
    "0": {"label":"低","example":null},
    "1": {"label":"高","example":null}
  },
  "legendImageReferences": [
    {"imageIndex":0,"placement":"value","pointer":"/0/example"},
    {"imageIndex":1,"placement":"value","pointer":"/1/example"}
  ]
}
```

ImageInputs 是最终 native 图片的 MIME 和 decoded 文件 SHA-256，索引不保证与
入口 Images 一致。遍历先 State，再按问题名排序，逐题先 Instructions 后 Criteria；
对象和 Choice/Bool 键排序，数组和 Score 等级保留原序。每个最终 Jev asset 首次
出现时建立一项，相同 bytes 的不同 asset 不合并。图例引用按等级、对象键与数组
顺序返回，可多次指向同一项。Host 不应从索引推断入口来源或 Provider 资源归属。

Hook 合法改题、改名、重排、增删或换图后，答案验证及图例引用只依赖最终快照。
图例缺项、错项、重复、改图或损坏分布都会令整次 Answers/ImageInputs 不发布；
唯一合法 usage 与 ResponseModel 保留，成功 HTTP 的坏答案不会自动重试。
无图或未返回图例时保持原行为。图片 bytes、私有 key 和图例路径不进入共享
Attempt/Observer；文件 hash 属于本次调用租户的结果内容。

这是受控 HTTP 验证的候选能力。真实服务、视觉质量和正式支持仍需各自证据。
内置目录和支持矩阵没有因此更新；Box 由后续工单验收。撤销 Binding/型号或图片
授权阻止新调用，已 pin 调用依固定快照完成。代码回滚无需数据迁移。
