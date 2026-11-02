# 目录定义决定型号集合，运行库校验执行能力

用户授权删除 Kimi 型号 ID 硬编码，并核实和删除其他按具体型号名称决定目录接纳的
运行逻辑。后续同类模型通过审核 JSON 更新，保留能力校验和 Host 授权。

## 本次范围

删除四处名称准入：Kimi chat、xAI Responses chat、xAI image、TypeSafe SystemOne
image/Box classifier。只删除具体 ID 相等条件；协议路由、服务端点、输入输出模态、
容量、推理档位、兼容标志、图片位置/MIME 和计价结构仍按已实现能力校验。
TypeSafe 的局部命名随协议能力改名，不再将图片能力归因于某个 Jev 名称。

保留非空型号 ID 与四维身份唯一性。外部 LoadFile/ValidateDefinitions 和直接
Catalog.Hash/NewClient 都必须验证受影响身份；不能因移除名称条件而放入空 ID。
不得无依据扩大其他字段的旧直接构造合同，例如强制原来可省略的显示名称。

保留目录精确查找、目标/请求模型一致性、原生历史同模型限制，以及
Binding.AllowedModels 的独立租户授权。响应按各协议原合同解析：Kimi Chat 记录
服务端别名到 ResponseModel；Kimi Responses 不填写 ResponseModel。两者都保留请求的
Model 与 Metadata.ModelID；现有严格比较的协议继续拒绝不匹配响应。
不新增跨协议统一同名规则。新目录条目不会自动授予调用权限。
保留现有证据、捕获与资格检查中的冻结真实型号身份，不重写历史或扩大厂商资格。

Gemini 名称驱动的请求格式、历史编码和 OpenAI gpt-5.5 服务等级倍率不属于名称
准入校验。本次保留其原行为并记录原因。新协议或不同计价语义仍依赖相应代码能力。
不为这些行为新增 JSON 字段，不借本次删除校验改变旧请求或估价。

## 可观察结果

受影响四类使用未在代码中出现的新 ID，JSON 具备相同已实现能力时，公共 LoadFile、
Hash 和 NewClient 接受定义，受控调用发出精确新 ID，保留正确用量和成本。
缺少 Host 授权时，调用拒绝且 HTTP 次数为0。
非法能力、空 ID、重复身份和错误协议继续拒绝。有效但厂商不存在的 ID 不取得实测资格。

公共 CLI E2E 使用明确标记的合成目录，完成 preview/freeze/check/apply，证明后续
同类新增仅修改 JSON 即可生效。合成身份不加入真实内置目录，不使用厂商凭据。
旧125条运行目录、维护源和完整基线保持原字节；目录 Version/Hash 不因程序策略改动而升版。

## 验证与交付

按测试 SOP 先保留真实 RED，再修改产品，再定向 GREEN。独立审查后提交冻结。
完整受影响普通包包含 ai/e2e、ai、catalog/...、两个目录 CLI；覆盖当前 ChineseEvaluation
与旧 Provider 的回归。随后顺序核对生成、构建和 vet。禁止阶段运行期间提交或改输入。
live/race/压力按最终影响分别判定；真实付费调用和适用的昂贵阶段必须按 SOP 独立选择。

更新 changelog、Catalog SOP 和 ADR。版本化运行 JSON、完整原 review、执行与审查
索引归入 artifacts/catalog/<现有Catalog.Version>/<新run-id>/，保留旧交付包不覆盖。
新包明确核心代码 YES、依赖变化的实际判断，以及当前/旧程序的外部 JSON 兼容证据。

本次须发布包含策略改动的 SDK 并由使用者重建一次；后续同能力的外部 JSON 更新不再
依赖型号名代码修改。Host 必须显式加载并创建新 Client；编译内置默认仍须重建程序。
目录接纳、代码兼容、厂商实测和 Host 授权分别记录，不宣称本次已发布软件版本。
