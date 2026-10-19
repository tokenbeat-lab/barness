# 运行库模块与逐文件审查说明

固定版本`f5cf536`；实现差分起点`6b893c4`。本页覆盖全部44个运行库Go文件。行号为当前文件位置，完整变更块见files-code.md。

把图片源当成“文件柜”，把ImageReferences当成“放置说明”。不同Provider如何放置、最终发出了什么、回答如何解释，分别由独立边界负责。

## 公共契约与编解码

| 文件与起读行 | 改了什么，如何理解 | 人工核对重点 |
|---|---|---|
| [ai/classifier.go L12](/Users/cyber/RestoX/barness/ai/classifier.go:12) | 请求新增ImageReferences；结果新增ImageInputs；TypeSafeOptions增加显式实验选择。L63起组织完整调用，L127起深复制，L156起能力和授权检查。 | 先看L72–107是否在复制前限额，再看L191–231是否在取凭据前拒绝不支持/未授权图片；失败L109–119清空答案和图片身份。 |
| [ai/classifier_references.go L20](/Users/cyber/RestoX/barness/ai/classifier_references.go:20) | 定义context/value两种引用。L53严格读JSON；L84把省略/空引用归一为按序context。 | 索引0合法但JSON缺失不可默认为0；显式声明须用完所有源；context重复合法。 |
| [ai/classifier_reference_values.go L12](/Users/cyber/RestoX/barness/ai/classifier_reference_values.go:12) | 只遍历State、单题说明、具体criteria内容。L99解析RFC6901；L140要求唯一且已存在的null槽。 | ~0/~1的解码顺序、数组前导零、逃逸键、重复目标、字段容器和普通文字都需重点看。 |
| [ai/classifier_json.go L65](/Users/cyber/RestoX/barness/ai/classifier_json.go:65) | 请求解码纳入引用，L97将题型JSON解为封闭变体，包含Box；沿用精确字段与重复键拒绝。 | JSON输入与Go结构体直传的校验是否会得到一致领域语义。 |
| [ai/classifier_questions.go L45](/Users/cyber/RestoX/barness/ai/classifier_questions.go:45) | Choice/Score/Bool序列化与native模型解耦；Box由独立文件提供。 | 阅读实际差分，避免把未变的旧类型定义误认为本任务新建。 |
| [ai/classifier_answers.go L17](/Users/cyber/RestoX/barness/ai/classifier_answers.go:17) | ScoreAnswer增加LegendImageReferences，索引指向最终ImageInputs，Pointer相对Legend。 | 返回图例不暴露Jev私有素材key或Base64；零值/空引用的序列化合同。 |
| [ai/classifier_image_inputs.go L21](/Users/cyber/RestoX/barness/ai/classifier_image_inputs.go:21) | 最终图片身份是MIME+解码文件的SHA256。L68用4KiB流式缓冲逐图计算。 | 基于Hook之后真正发送的图片；同字节不同来源不自动合并；取消点和错误时不发布身份。 |
| [ai/classifier_box.go L11](/Users/cyber/RestoX/barness/ai/classifier_box.go:11) | 新增BoxQuestion、归一化四边Box和带Nullable置信度的BoxAnswer；L55验证正面积。 | 边界须在[0,1]且left<right/top<bottom；confidence缺失、显式null、0是三种状态；不推断存在性或imageId。 |
| [ai/classifier_box_visibility.go L6](/Users/cyber/RestoX/barness/ai/classifier_box_visibility.go:6) | 每个Box必须能看到context/State或它自己的Instructions中的图片。 | 另一题的局部图片、只有criteria图片，不能给该Box授权可见性。 |

## 能力、预算与不可变快照

| 文件与起读行 | 改了什么，如何理解 | 人工核对重点 |
|---|---|---|
| [ai/classifier_policy.go L34](/Users/cyber/RestoX/barness/ai/classifier_policy.go:34) | 题数/State/题目尺寸及引用根图例外；L91检查题型、数量和描述格式能力；L167检查各题结构。 | 根null只在明确图片引用下有效；不能借引用放松无图文本的旧合同。 |
| [ai/classifier_capabilities.go L10](/Users/cyber/RestoX/barness/ai/classifier_capabilities.go:10) | 目录能力合法性扩展到位置、格式与Box；L72检查带图片根的说明格式。 | 能力描述和图片放置是不同轴；支持JSON说明不自动意味着支持局部图片。 |
| [ai/classifier_images.go L12](/Users/cyber/RestoX/barness/ai/classifier_images.go:12) | 图片策略挂接References子策略，保留源图数量/字节及MIME/model限制。 | 源图预算与展开引用预算分别消耗；Binding图片权限继续单独控制。 |
| [ai/classifier_reference_policy.go L7](/Users/cyber/RestoX/barness/ai/classifier_reference_policy.go:7) | 新增MaxReferences和MaxExpandedImageBytes；L12未显式配置时继承有限源图上限；L21在复制/展开前检查。 | 同一图重复100次不能按一次计资源；inline Decisions还要预估重复data URL体积。 |
| [ai/classifier_reference_sizes.go L8](/Users/cyber/RestoX/barness/ai/classifier_reference_sizes.go:8) | 为JSON文本投影计算可安全拒绝的长度下界，另处理每种题目内容。 | 原JSON转成字符串后转义会增长；下界只用于提前拒绝，最终编码仍须独立检查。 |
| [ai/classifier_payload.go L12](/Users/cyber/RestoX/barness/ai/classifier_payload.go:12) | Hook后的普通分类payload仍执行State、题目与总request预算。 | 不要只审公共请求；回调修改后的最终payload是第二个不可信边界。 |
| [ai/policy.go L165](/Users/cyber/RestoX/barness/ai/policy.go:165) | 深复制Classifier.Images.References，使调用配置快照不借用Host可变指针。 | NewClient后外部修改策略不能扩大已固定调用的限额。 |

## Jev原生边界与投影

| 文件与起读行 | 改了什么，如何理解 | 人工核对重点 |
|---|---|---|
| [ai/typesafe_request.go L12](/Users/cyber/RestoX/barness/ai/typesafe_request.go:12) | 先重构为私有typeSafeRequest/typeSafeQuestion；L34由公共类型单向编码，L77封闭解码native变体。 | native bool仍叫noul；answer验证使用Hook后的native问题，而非原Host问题。 |
| [ai/typesafe_guard.go L11](/Users/cyber/RestoX/barness/ai/typesafe_guard.go:11) | 冻结最终native请求，重验模型、State、题数、大小和能力，避免重建公共请求来推断结果。 | 保持旧错误优先级、HTML转义后的规范化每题限额；不能让Hook改模型或注入额外authority。 |
| [ai/typesafe_images.go L29](/Users/cyber/RestoX/barness/ai/typesafe_images.go:29) | JSON assets投影：每源生成私有key，L58每源最多8次碰撞尝试，L96/L101投影State和题目。 | key不得误中原始业务文字/对象键，也不能互为substring；支持子集只走JSON assets，未新增multipart公共接口。 |
| [ai/typesafe_state_images.go L12](/Users/cyber/RestoX/barness/ai/typesafe_state_images.go:12) | 先预算替换增长再展开。L53有context时用{state:原State,images:引用序列}包裹；无context保持业务根。 | L101记录包裹来源，不能把Hook新加的同名state/images业务对象误认成系统信封；数字保持json.Number。 |
| [ai/typesafe_question_images.go L8](/Users/cyber/RestoX/barness/ai/typesafe_question_images.go:8) | 单题Instructions和Choice/Score/Bool的criteria只替换声明null槽；逐题先预留增长。 | 同题可多图；其他题不自动看见；图片根例外不放松其他文本容器格式。 |
| [ai/typesafe.go L10](/Users/cyber/RestoX/barness/ai/typesafe.go:10) | 显式设置实验header，Hook后冻结请求/图片，发送固定body，最终全部答案成功才发布ImageInputs。 | L30–57是核心安全段；L100拒绝Hook擅自启用、关闭或改实验header；decode失败不得新发付费请求。 |

## 最终图片校验与资源防护

| 文件与起读行 | 改了什么，如何理解 | 人工核对重点 |
|---|---|---|
| [ai/typesafe_image_guard.go L41](/Users/cyber/RestoX/barness/ai/typesafe_image_guard.go:41) | 重新校验Hook后assets的权限、模型、数量、MIME、源图/展开预算及实际文件；按最终首次使用建立key→结果index。 | L91要求所有assets被实际使用；L112检查私有key命名域与非内容字段；不能沿用Host旧索引。 |
| [ai/typesafe_image_occurrences.go L17](/Users/cyber/RestoX/barness/ai/typesafe_image_occurrences.go:17) | 确定性遍历最终State/题目内容；只认完整字符串值等于key，拒绝文字中的substring插值；L132重新检查Box可见性。 | 最终context被删但合法局部图仍在时能继续；题目按名称排序，Instructions先于Criteria，数组保留顺序。 |
| [ai/typesafe_image_limits.go L15](/Users/cyber/RestoX/barness/ai/typesafe_image_limits.go:15) | 独立协议界限：图16MiB、体32MiB、key1KiB/合计64KiB；L22检查PNG/JPEG文件头、尺寸，L77扫描静态PNG chunks。 | 不解压像素、不自动裁剪/旋转/缩放；拒绝APNG；Host更紧的限额仍独立生效。 |
| [ai/typesafe_image_payload.go L22](/Users/cyber/RestoX/barness/ai/typesafe_image_payload.go:22) | 在Hook借用对象被JSON编码前，检查assets已知源图大小；支持map/struct/RawMessage和可预知[]byte编码。 | 编码后才发现过大可能已分配巨量内存；自定义encoder输出未知，最终guard继续负责，不提前调用用户encoder。 |
| [ai/typesafe_reference_payload.go L19](/Users/cyber/RestoX/barness/ai/typesafe_reference_payload.go:19) | 预检最终Go内容可能形成的素材出现次数和展开字节，处理RawMessage、Unicode规范化、[]byte及,string。 | L186非法UTF8与L195孤立surrogate须匹配encoding/json；不能因Go值表面不同而漏计最终同一key。 |
| [ai/image_payload.go L10](/Users/cyber/RestoX/barness/ai/image_payload.go:10) | 共用的反射边界仅剥离安全可读值、定位实际JSON资源字段。 | 该辅助层服务两条真实图片边界，未创建新的通用公共框架。 |
| [ai/image_payload_fields.go L57](/Users/cyber/RestoX/barness/ai/image_payload_fields.go:57) | 按encoding/json的匿名字段提升、tag优先级、冲突省略和omitempty/omitzero挑字段；L21处理,string。 | 这是本任务最复杂的预算防护之一：声明类型和运行时类型不同、双重指针、interface、数值素材key均需看回归。 |

## 答案转换与原子发布

| 文件与起读行 | 改了什么，如何理解 | 人工核对重点 |
|---|---|---|
| [ai/typesafe_box.go L5](/Users/cyber/RestoX/barness/ai/typesafe_box.go:5) | Box native解码，边用指针区分缺失/null与数值0，confidence用Nullable区分三态。 | 少边、反向/零面积、越界、错误类型、非有限数值整体拒绝。 |
| [ai/typesafe_answers.go L23](/Users/cyber/RestoX/barness/ai/typesafe_answers.go:23) | Score校验改为对比最终native criteria；L58校验完整回显图例，L78保留精确JSON数字比较。 | 必须先证明legend与最终rubric一致，再恢复公共图片引用，不能用“修正图例”掩盖服务回显错误。 |
| [ai/typesafe_legend_images.go L13](/Users/cyber/RestoX/barness/ai/typesafe_legend_images.go:13) | 把已验证legend中的私有图片key还原为null，并形成相对Legend的公共value引用。 | 多个层级/数组和转义键的Pointer；无图图例原字节尽量保留。 |
| [ai/typesafe_response.go L10](/Users/cyber/RestoX/barness/ai/typesafe_response.go:10) | 先记录usage再逐题解码，全部通过才一次性发布Answers；加入Box和视觉legend转换。 | 一题坏不能返回其他半成品；失败仍保留已经确认的usage与响应元数据。 |

## Decisions适配

| 文件与起读行 | 改了什么，如何理解 | 人工核对重点 |
|---|---|---|
| [ai/openai_decisions_request.go L35](/Users/cyber/RestoX/barness/ai/openai_decisions_request.go:35) | 公共context计划转成一个user消息：先text，再按引用次序逐个input_image；重复引用重复block。 | [1,0,1]必须三块；没有自动去重，也不能把局部图提升成context。 |
| [ai/openai_decisions.go L11](/Users/cyber/RestoX/barness/ai/openai_decisions.go:11) | Hook前后都冻结native请求；L33从最终input取图，L37计算身份，L60仅成功时发布。 | Hook换图/重排后结果身份必须随最终请求变化；其他结果原子规则与Jev一致。 |
| [ai/openai_decisions_guard.go L65](/Users/cyber/RestoX/barness/ai/openai_decisions_guard.go:65) | 独立native封闭schema守卫返回最终图片，保留题目能力/模型/权限检查。 | 最终schema不是只看原Host输入；不支持的Box、说明格式及局部位置不伪装支持。 |
| [ai/openai_decisions_images.go L22](/Users/cyber/RestoX/barness/ai/openai_decisions_images.go:22) | 从最终消息严格读取inline data URL图片，并核对权限、MIME、模型及展开限制。 | 只接受受支持消息拓扑；URL/远端资源不能混入此分类图片入口。 |
| [ai/openai_decisions_image_payload.go L31](/Users/cyber/RestoX/barness/ai/openai_decisions_image_payload.go:31) | Hook前序列化资源预检接入共用JSON字段选择和[]byte大小检查。 | typed payload不能绕过重复图片/解码大小预算；该协议的URL树检查与Jev assets保持各自职责。 |

## 正式目录

| 文件与起读行 | 改了什么，如何理解 | 人工核对重点 |
|---|---|---|
| [ai/catalog_operations.go L78](/Users/cyber/RestoX/barness/ai/catalog_operations.go:78) | 新增Box kind及context/state/instructions/criteria位置类型，并加入ClassifierCapabilities。 | 能力是可表达性，不是租户权限；位置、格式、题型和图片格式各自约束。 |
| [ai/catalog_typesafe.go L7](/Users/cyber/RestoX/barness/ai/catalog_typesafe.go:7) | 保留jev-1.13.0文本行，新增独立固定jev-1.14.0，24源图、PNG/JPEG、四位置与Box。 | 24是本库有限支持子集；目录加入后依然需要Binding允许、policy有限和ExperimentalImages显式选择。 |
| [ai/catalog_snapshot.go L55](/Users/cyber/RestoX/barness/ai/catalog_snapshot.go:55) | ClassifierModel clone深复制新增ImageLocations，与已有能力切片一起进入不可变目录快照。 | 外部改slice不能影响已创建Client；Hash可反映新增能力。 |
| [ai/catalog_validate.go L147](/Users/cyber/RestoX/barness/ai/catalog_validate.go:147) | 目录验证允许Box同时仍拒绝未知/重复或与协议不相容的能力。 | 本次差分很小，应和classifierModelProblem一起读，不能只看此入口。 |
| [ai/catalog.go L53](/Users/cyber/RestoX/barness/ai/catalog.go:53) | 内建目录版本随正式Jev纳入变化，触发目录Hash和证据当前性更新。 | 目录版本变化不是把1.13那行改名；两型号保持独立身份。 |

## 整版续修：共享超时

| 文件与起读行 | 改了什么，如何理解 | 人工核对重点 |
|---|---|---|
| [ai/timeouts.go L141](/Users/cyber/RestoX/barness/ai/timeouts.go:141) | 仅在Attempt超时确实早于Call deadline时启动timer；L181读闲复用与L199关闭共同处理nil timer。 | 解决取消回调传播延迟导致较晚ReadIdle抢占较早Call deadline的归属竞态；影响共享HTTP，不仅Jev。 |

## 六个需要特别保留的区分

1. 图片源数量与引用出现次数不同：Jev上传一个asset可被重复使用，但重复仍消耗展开预算。
2. 原Host图片索引与最终ImageInputs索引不同：Hook可以改变native内容；输出必须按最终内容计算。
3. Decisions每个最终图片block保留一条身份；Jev按最终使用顺序为每个asset保留一条。同Hash不代表来源合并。
4. 图片权限、目录能力、实验选择及资源预算是四道独立条件，一道满足不能代替另外三道。
5. Box只表示位置。存在性来自独立Bool问题，Host决定是否消费Box；库不自行丢弃低存在性的原始Box。
6. schema正确、协议真实PASS与业务效果好是三个不同结论，分别由运行库、协议探针和固定效果任务证明。
