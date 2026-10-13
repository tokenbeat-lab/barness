# Pi 1.1.0：按 thinking level 选择采样参数

研究日期：2026-10-11。仅调研；未修改 barness-ai 生产代码、测试、兼容规范或冻结基线。

## 结论

`samplingParamsByThinkingLevel` 是 **Model 的分档默认值**。Pi 用统一的 `clampThinkingLevel` 求出受支持的 Pi 档位，再按 key 合并：`Model.samplingParams → Model.samplingParamsByThinkingLevel[effectiveLevel] → 请求 samplingParams`。合并结果最后覆盖请求体的具名字段，随后执行 `onPayload`。[核心 helper][helper]、[请求体应用][chat-apply]

这里的 effective level 是 **客户端按模型元数据夹取后的 Pi 档位**，不由 `thinkingLevelMap` 的厂商字符串反查，不等待服务端反馈。三个支持协议的 simple 入口会夹取推理选项；full 入口的采样 helper 也夹取，但 full 的 wire reasoning 字段没有统一夹取，所以不能把该特性概括成“永远按最终发出的 reasoning 值选采样”。[夹取逻辑][clamp]、[simple 入口][chat-simple]、[Responses 请求构造][responses-build]

适用范围是 `openai-completions`、`openai-responses`、`azure-openai-responses` 三种 **API 协议**，不是按 Provider 名称开关。DeepSeek、OpenRouter、本地兼容服务等只要走支持的 adapter 就能使用；Anthropic、Gemini 和 `openai-codex-responses` 等协议忽略该 map。[官方配置文档][docs]

版本需要校正：该特性首次发布于 **1.0.2（2026-10-04）**，1.1.0 包含它；引入 PR 为 #9776。[1.0.2 changelog][changelog]、[PR][pr]

## 1. 固定研究基线

| 项目 | 已核对的第一方证据 |
|---|---|
| 目标 tag | `v1.1.0` → `abe508e1b89912adde45528136c3221eb69acdd7`；[GitHub tag API][tag110] |
| 首次发布 tag | `v1.0.2` → `cd32f7725fdbddbaecdff5b1e68491563394e0ca`；[GitHub tag API][tag102] |
| barness 冻结对照 tag | `v1.0.0` → `a13d35a742c6ef8462812a28fbe1d8c8b7431c32`；[GitHub tag API][tag100] |
| 引入提交 | #9776，merge commit `76dfb88f63ce51ff2e3fe2ead4fcf1f65f71f121`，合并于 `2026-10-03T23:29:01Z`；[PR API][pr-api]、[固定提交][feature-commit] |
| 读取方式 | 下载官方 tag tarball，在 `/tmp/pi-1.1.0` 读取源码；报告源码链接均固定到 1.1.0 commit；未以 main 的现状代替目标版本 |

PR 解决的用例是开放模型对思考/非思考模式推荐不同采样参数；配置以同一套通用默认值为基础，为特定档位覆盖部分参数。[PR 的设计背景][pr]

## 2. 数据模型与入口

公开类型定义如下；`off` 加六种思考档位，共七个 canonical key。[类型定义][types]

```ts
type SamplingParams = Record<string, unknown>;
type SamplingParamsByThinkingLevel = Partial<Record<ModelThinkingLevel, SamplingParams>>;
```

`Model` 新增可选的 `samplingParamsByThinkingLevel`；调用选项仍只有单次请求的 `samplingParams`。Pi 没有新增调用级 `samplingParamsByThinkingLevel`。[Model 字段][model-types]、[StreamOptions][stream-types]

| 名称 | 职责 |
|---|---|
| `thinkingLevelMap` | 描述某 Pi 档位是否支持，并将它转换为厂商值；`null` 标记不支持 |
| `samplingParams` | 对所有档位生效的模型通用采样默认值 |
| `samplingParamsByThinkingLevel` | 以 Pi 档位为 key，覆盖该档位的部分默认值 |
| 请求 `samplingParams` | 单次调用覆盖，优先级最高；与 thinking level map 无关 |

上述职责由类型、helper 和配置文档共同界定。[Model 字段][model-types]、[helper][helper]、[文档][docs]

例如 `thinkingLevelMap.minimal = "low"` 时，effective Pi 档位仍为 `minimal`，选 `samplingParamsByThinkingLevel.minimal`，请求协议字段可以是 `low`。也可以令多个 Pi 档位映射到同一厂商值，同时保留不同采样默认值。[夹取逻辑][clamp]、[wire 映射][responses-build]

## 3. 档位选择和合并算法

`resolveSamplingParams(model, thinkingLevel, requestParams)` 做三件事：[helper][helper]

1. 通过 `clampThinkingLevel(model, thinkingLevel)` 求 effective Pi 档位。
2. 查找该 key 下的 `samplingParamsByThinkingLevel`；没有该档时不加入分档默认。
3. 以浅对象 spread 合并通用模型默认、分档默认、调用默认；三者都不存在时返回 `undefined`。

`clampThinkingLevel` 的具体规则如下。[源码][clamp]

| 条件 | 结果 |
|---|---|
| `reasoning: false` | 只支持 `off`，任何输入均夹取为 `off` |
| reasoning model 的普通档位 | `off/minimal/low/medium/high` 默认支持；对应 map 值为 `null` 时不支持 |
| `xhigh/max` | 只有 map 显式给出非 `null` 值才支持；遗漏 key 即不支持 |
| 请求档位已支持 | 保留该 Pi 档位，不使用 map 的字符串作为采样 key |
| 请求档位不支持 | 按 `off < minimal < low < medium < high < xhigh < max` 先向上寻找，再向下寻找 |
| 所有档位不支持，或非法运行时值且无可用档 | 最终兜底 `off` |

因此 `low: null, medium: null` 且 `high` 支持时，请求 `low` 使用 `high` 的采样参数；`xhigh/max` 未声明时会回落到较低支持档。`off: null` 表示不能选择 off 档，会向上选择最低可用档；它与 `off: "none"` 的含义不同。[源码][clamp]

合并只看 key 是否存在，不用 truthiness 过滤参数值：`0`、`false`、`null` 都会覆盖前值。`null` 会作为值进入请求体，不是删除指令；JS 的 `undefined` 属性在 JSON 序列化时通常省略。嵌套对象是整体替换，非递归合并；顶层返回新对象，但嵌套值可以保留原引用。[helper][helper]、[最终 Object.assign][chat-apply]

完整请求体优先级是：

```text
具名请求字段（temperature 等）
  → 模型通用 samplingParams
  → effective Pi 档位 samplingParamsByThinkingLevel
  → 调用 samplingParams
  → onPayload 的修改/替换
```

所以调用的 `temperature: 0` 仍可能被模型默认或分档默认覆盖；要稳定覆盖，放到调用 `samplingParams.temperature` 中。`onPayload` 在 `buildParams` 返回后执行。[具名字段][chat-named]、[采样覆盖][chat-apply]、[回调调用点][chat-callback]

## 4. Full 与 simple 的实际调用路径

| 入口/协议 | 采样档位的输入和处理 |
|---|---|
| simple：三个支持协议 | `buildBaseOptions` 先对 `options.reasoning ?? "off"` 解析采样；adapter 对显式 reasoning 夹取，将 `off` 转为未提供 `reasoningEffort`，调用 full stream |
| full：Chat Completions | `buildParams` 最后对 `options.reasoningEffort ?? "off"` 再调用 helper，并覆盖请求体 |
| full：Responses / Azure Responses | 先求 `reasoningEffort = options.reasoningEffort ?? (options.reasoningSummary ? "medium" : undefined)`，再对其或 `off` 调用 helper |

来源：[共用 simple options][helper]、[Chat simple][chat-simple]、[Chat full][chat-apply]、[Responses][responses-build]、[Azure][azure-build]

simple 的实现是 **两次解析/合并**：第一次合并后的结果作为 full 的 `options.samplingParams`，第二次再与模型默认、分档默认合并。当前三个支持 adapter 的两次 effective Pi key 一致；第一次产生的结果不会因第二次合并而丢失 caller 覆盖。离线 probe 核对了 low→high、未提供 reasoning 且 off 不支持、非推理模型收到 high 三种路径。[simple helper][helper]、[simple adapter][chat-simple]

这不能推导为 full 与 simple 在所有输入上生成同一推理请求。以下是需要保留的边界；前三项由固定源码推导，未运行厂商 API。[夹取源码][clamp]、[wire 构造][responses-build]、[Chat thinking 分支][chat-thinking]

| 场景 | 采样选择 | wire / 限制 |
|---|---|---|
| full 传 `reasoningEffort: "low"`，map 的 low、medium 为 null | helper 夹取到 high，选择 high 采样 | Responses 的 wire 映射使用原始 low；`null ?? "low"` 得到 low，可能被服务端拒绝。simple 会先把推理选项改为 high |
| Responses full 只有 `reasoningSummary: "auto"` | 默认以 medium 调用 helper；medium 支持时选 medium | wire effort 为 medium；若 medium 不支持，helper 会夹取到另一档，但 summary-only wire 仍是 medium |
| simple 未提供 reasoning，而模型 `off: null` | 两次 helper 都从 off 夹取到最低支持档，如 minimal | simple 没有显式 `reasoningEffort`；Responses 不写 off effort，DeepSeek 不写 disabled。实际服务端默认强度未由该功能获知 |
| 请求 map 或 `onPayload` 另写 `reasoning` / `reasoning_effort` | 采样已按此前计算的 Pi 档位选择 | 不重新选档；这两个阶段没有反馈闭环 |
| compat 不发送 reasoning effort 的模型 | helper 仍按 model 元数据和输入选档 | helper 不检查 endpoint 是否实际执行该强度 |

这解释了“实际采用的 thinking level”应限定为 Pi 正常 simple 路径下经模型能力夹取的 canonical level；不应扩大为最终 JSON、厂商执行强度或返回的 token 用量。两个参数 map 也没有互相反推的逻辑。[helper][helper]、[clamp][clamp]

## 5. Provider 与 API 协议范围

| API 协议 | 支持情况 | 固定版本中的 Provider 例子 |
|---|---|---|
| `openai-completions` | full、simple 应用通用与分档 map | DeepSeek；OpenRouter 的 Chat 路由；Azure 的 Foundry Chat 路由；兼容协议的自定义 Provider |
| `openai-responses` | full、simple 应用；支持 summary-only 选择 medium | OpenAI、xAI；OpenCode / GitHub Copilot 的 Responses 路由 |
| `azure-openai-responses` | full、simple 应用；支持 summary-only | Provider ID 为 `azure` |
| `anthropic-messages`、`google-generative-ai` 等 | 忽略这些 sampling map | 原生 Anthropic、Google 路由；通过兼容 Chat 路由访问相同厂商时按所用 API 判断 |
| `openai-codex-responses` | 忽略 map，即使 simple 先构建 base options | 不属于前三种支持协议；请求体只消费其建模的具名参数 |

来源：[官方协议范围][docs]、[DeepSeek Provider][deepseek-provider]、[OpenRouter Provider][openrouter-provider]、[OpenAI Provider][openai-provider]、[Azure Provider][azure-provider]、[xAI Provider][xai-provider]、[Codex 请求体][codex-build]

新参数只是透传默认值机制，不验证服务端是否接受 `top_k`、`min_p` 等，也不替厂商维护统一的温度推荐值。请求类型明确允许自由 key 并按原样加入请求体；实际参数语义由目标服务决定。[StreamOptions 约定][stream-types]

## 6. models.json 加载、校验与 override

`ModelDefinitionSchema` 和 `ModelOverrideSchema` 均新增该字段。分档 schema 声明七个可选 Pi 档位，每档为 `Record<string, unknown>`；`thinkingLevelMap` 允许 string/null，而采样档位值本身应是对象。没有温度区间或厂商采样 key 白名单校验，不能把“通过配置 schema”理解为“厂商一定接受”。[schema][config-schema]

加载过程是读取文件、移除 BOM 和 JSON comments、JSON.parse、schema 检查，然后对 Provider 配置 `structuredClone` 并递归 `deepFreeze`。失败会记录具体配置路径和错误；这与采样 helper 的浅合并是两个不同边界。[加载与冻结][config-load]

`modelFromJson` 将字段写入 Model。`modelOverrides` 按 **档位内的参数 key** 合并：保留没有覆写的档位，同一档保留没有覆写的参数 key；不是用一个 override 档位对象替换整档。[Model 构造与 override][composer] 配置组装先加入 `config.models`，再在最外层对全部 chat model 按 id 应用 `modelOverrides`，所以同一份配置刚定义的 custom model 也会收到 override。[配置层次][composer-layers]

以下是基于官方示例的完整自定义配置形状；数字仅演示覆盖关系。[文档例子][docs]

```json
{
  "providers": {
    "local-qwen": {
      "baseUrl": "http://127.0.0.1:8000/v1",
      "apiKey": "local-unused-key",
      "api": "openai-completions",
      "models": [{
        "id": "qwen-thinking-model",
        "reasoning": true,
        "thinkingLevelMap": { "off": "none", "minimal": null },
        "samplingParams": { "temperature": 1.0, "top_p": 0.95 },
        "samplingParamsByThinkingLevel": {
          "off": { "temperature": 0.7, "top_p": 0.8 },
          "high": { "top_k": 20 }
        }
      }],
      "modelOverrides": {
        "qwen-thinking-model": {
          "samplingParamsByThinkingLevel": { "high": { "temperature": 0.8 } }
        }
      }
    }
  }
}
```

最终 high 档默认为 `temperature=0.8, top_p=0.95, top_k=20`；off 档为 `temperature=0.7, top_p=0.8`。请求 `minimal` 会夹取为 low，因此没有 high 的 top_k。请求 map 仍可覆写其中任意 key。[夹取][clamp]、[运行时合并][helper]、[配置合并][composer]

未把这组数字称作 Pi 内置模型统一默认。tag 的 Provider 目录通过生成的 `src/providers/data/*.json` 引入型号数据，该生成目录未包含在下载的源码 tag 中；本轮没有核对 1.1.0 npm 发布件的模型数据。[生成目录引用][deepseek-models]

## 7. 与 barness 冻结的 Pi 1.0.0 的差异

| 位置 | 1.0.0 | 1.1.0 |
|---|---|---|
| Model 类型 | 只有 samplingParams | 新增分档 map 类型和字段 |
| simple base options | 直接转交调用 samplingParams | 提前调用 resolveSamplingParams |
| 三个 full 请求 builder | 末尾 `Object.assign` 通用模型默认、调用 map | 调用共用 helper，加入 clamp 后的分档默认，再 Object.assign |
| Responses summary-only | wire 可默认 medium；没有分档配置 | 采样也以 medium 作为初始档位 |
| coding-agent models.json | 只有通用 sampling 配置 | schema、Model 构造、modelOverrides 加入分档字段 |

来源：[1.0.0 simple 源码][old-helper]、[1.0.0 Chat 源码][old-chat]、[1.0.0 Responses 源码][old-responses]、[引入 PR 的变更文件][pr-files]、[1.1.0 helper][helper]

该机制没有新增统一采样数值，没有按任务内容自动调参，也没有将 Anthropic/Gemini 纳入通用采样 map。[helper][helper]、[范围说明][docs]

## 8. 验证与可重复证据

上游已有 `sampling-options.test.ts` 通过公共 `stream` / `streamSimple` 和 `onPayload` 捕获请求、在发网前抛错。覆盖：参数透传与 0、无配置省略、模型/调用覆盖、low→high、off、三个 full API 的分档顺序、Responses/Azure summary-only、具名字段被覆写、Anthropic 忽略 map。[上游测试][sampling-tests]

`model-registry.test.ts` 验证 custom models 和 modelOverrides 携带 map、档位内按 key 合并，以及未配置型号保持 undefined。[配置测试][registry-tests]

本轮没有运行整套上游 Vitest，也没有运行真实 Provider API。本轮执行的是 **固定源码函数的无网络 probe**：从官方源码动态提取 getSupportedThinkingLevels、clampThinkingLevel、resolveSamplingParams，在 Node v24.15.0 上运行；13 项 helper 断言和 3 项两阶段断言全部通过。它验证 helper 语义，不代替协议 E2E。[被执行的源码][helper]、[clamp][clamp]

| helper probe | 实际结果 |
|---|---|
| low/medium 不支持 | low→high；`temperature=.8, top_p=.95, top_k=64` |
| xhigh 未声明 | xhigh→high；`top_k=64` |
| minimal 映射为 low | 档位仍 minimal；选择 `temperature=.1`，不选 low 的 `.2` |
| off 映射为 none | 档位仍 off；`temperature=.7` |
| off 为 null | off→minimal；`temperature=.8` |
| 非推理模型请求 high | high→off；`temperature=.7` |
| 调用的 null/0 覆盖 | `top_p=null, top_k=0, temperature=0` |
| 缺少当前档位 | 继承通用 `temperature=1` |
| 没有任何默认 | 返回 undefined |
| 当前档配置为空对象 | 返回 `{}` |
| 嵌套对象覆盖 | `{extra:{a:1,b:2}}` 被 `{extra:{b:3}}` 替换 |
| 所有档位 null | 兜底 off；应用 off 配置 |
| 不修改源顶层 map | 通过；未覆写的嵌套对象仍共享引用 |

三项两阶段 probe 的结果为：显式 low→high 的两次 bucket 都是 high；省略 reasoning 且 off=null 的两次 bucket 都是 minimal，但 adapter 的 `reasoningEffort` 输入为 undefined；非推理模型 high 的两次 bucket 都是 off。没有观察到当前三个支持 simple adapter 的两次解析选择不同 bucket。

复现命令如下。先下载固定 commit 源码，再直接提取函数；不会安装 monorepo 依赖，也不会新增仓库测试文件。

```bash
pi_research_dir=$(mktemp -d)
curl -fsSL https://codeload.github.com/earendil-works/pi/tar.gz/abe508e1b89912adde45528136c3221eb69acdd7 -o "$pi_research_dir/pi.tar.gz"
tar -xzf "$pi_research_dir/pi.tar.gz" -C "$pi_research_dir"
PI_SOURCE_ROOT="$pi_research_dir/pi-abe508e1b89912adde45528136c3221eb69acdd7" node --input-type=module <<'JS'
import fs from 'node:fs/promises';
import assert from 'node:assert/strict';
import { pathToFileURL } from 'node:url';
const root = process.env.PI_SOURCE_ROOT + '/packages/ai/src';
const m = await fs.readFile(root + '/models.ts', 'utf8');
const s = await fs.readFile(root + '/api/simple-options.ts', 'utf8');
const models = m.slice(m.indexOf('const EXTENDED_THINKING_LEVELS:'), m.indexOf('\n/**\n * Check if two models are equal'));
const sampling = s.slice(s.indexOf('export function resolveSamplingParams('), s.indexOf('\nexport function buildBaseOptions('));
const preamble = `type Api=string; type ModelThinkingLevel='off'|'minimal'|'low'|'medium'|'high'|'xhigh'|'max'; type SamplingParams=Record<string,unknown>; type Model<T>={reasoning:boolean;thinkingLevelMap?:Partial<Record<ModelThinkingLevel,string|null>>;samplingParams?:SamplingParams;samplingParamsByThinkingLevel?:Partial<Record<ModelThinkingLevel,SamplingParams>>};\n`;
const file = process.env.PI_SOURCE_ROOT + '/probe-functions.mts';
await fs.writeFile(file, preamble + models + sampling);
const {clampThinkingLevel:clamp, resolveSamplingParams:resolve} = await import(pathToFileURL(file).href);
const rows = [];
const R = {reasoning:true};
const cases = [
 ['clamp-up',{...R,thinkingLevelMap:{low:null,medium:null},samplingParams:{temperature:1,top_p:.95},samplingParamsByThinkingLevel:{high:{temperature:.8,top_k:64}}},'low',undefined,'high',{temperature:.8,top_p:.95,top_k:64}],
 ['xhigh-missing',{...R,samplingParamsByThinkingLevel:{high:{top_k:64}}},'xhigh',undefined,'high',{top_k:64}],
 ['canonical-key',{...R,thinkingLevelMap:{minimal:'low'},samplingParamsByThinkingLevel:{minimal:{temperature:.1},low:{temperature:.2}}},'minimal',undefined,'minimal',{temperature:.1}],
 ['off-none',{...R,thinkingLevelMap:{off:'none'},samplingParamsByThinkingLevel:{off:{temperature:.7}}},'off',undefined,'off',{temperature:.7}],
 ['off-null',{...R,thinkingLevelMap:{off:null},samplingParamsByThinkingLevel:{minimal:{temperature:.8}}},'off',undefined,'minimal',{temperature:.8}],
 ['nonreasoning',{reasoning:false,samplingParamsByThinkingLevel:{off:{temperature:.7}}},'high',undefined,'off',{temperature:.7}],
 ['null-zero',{...R,samplingParams:{top_p:.95,top_k:40,temperature:1},samplingParamsByThinkingLevel:{low:{temperature:.6}}},'low',{top_p:null,top_k:0,temperature:0},'low',{top_p:null,top_k:0,temperature:0}],
 ['missing-bucket',{...R,samplingParams:{temperature:1},samplingParamsByThinkingLevel:{high:{temperature:.8}}},'medium',undefined,'medium',{temperature:1}],
 ['absent',R,'low',undefined,'low',undefined],
 ['empty',{...R,samplingParamsByThinkingLevel:{low:{}}},'low',undefined,'low',{}],
 ['shallow',{...R,samplingParams:{extra:{a:1,b:2}},samplingParamsByThinkingLevel:{low:{extra:{b:3}}}},'low',undefined,'low',{extra:{b:3}}],
 ['all-null',{...R,thinkingLevelMap:Object.fromEntries(['off','minimal','low','medium','high','xhigh','max'].map(k=>[k,null])),samplingParamsByThinkingLevel:{off:{temperature:.7}}},'high',undefined,'off',{temperature:.7}]
];
for (const [name,model,requested,request,level,want] of cases) {
 assert.equal(clamp(model,requested),level); assert.deepEqual(resolve(model,requested,request),want);
 rows.push({name,effective:level,params:want??'<undefined>'});
}
const source={...R,samplingParams:{extra:{a:1}},samplingParamsByThinkingLevel:{low:{top_k:64}}};
const before=structuredClone(source); const got=resolve(source,'low',{top_p:.5});
assert.deepEqual(source,before); assert.equal(got.extra,source.samplingParams.extra);
rows.push({name:'source-unchanged-nested-shared'});
for (const [name,model,requested] of [
 ['simple-clamp-high',cases[0][1],'low'],
 ['simple-omitted-off-null',cases[4][1],undefined],
 ['simple-nonreasoning-high',cases[5][1],'high']
]) {
 const first=resolve(model,requested??'off',{top_p:.5});
 const clamped=requested?clamp(model,requested):undefined;
 const effort=clamped==='off'?undefined:clamped;
 const second=resolve(model,effort??'off',first);
 assert.equal(clamp(model,requested??'off'),clamp(model,effort??'off'));
 assert.deepEqual(second,first); rows.push({name,effective:clamp(model,effort??'off'),effort:effort??'<undefined>',params:second});
}
console.log(JSON.stringify({node:process.version,passed:rows.length,cases:rows},null,2));
JS
```

## 9. barness-ai 接入映射与建议（尚未实现）

以下以 barness 当前 `39ab6a5` 代码为证据，区分于上述 Pi 已有行为。

| 边界 | 当前情况与建议 | 本地证据 |
|---|---|---|
| 模型目录 | 新增模型级 `SamplingParamsByThinkingLevel`，目录 key 支持 `off` 和六个思考档位；不扩大 `SimpleOptions.Reasoning` 输入枚举，仍用省略表示未请求思考；不新增没有上游依据的调用级分档字段 | [catalog_models.go](/Users/cyber/RestoX/barness/ai/catalog_models.go:17)、[reasoning.go](/Users/cyber/RestoX/barness/ai/reasoning.go:5)、[options.go](/Users/cyber/RestoX/barness/ai/options.go:73) |
| 配置隔离 | 每档验证 JSON 与保留键，快照深拷贝每档 map 及 RawMessage bytes；不能照搬 Pi helper 的嵌套引用共享 | [catalog_validate.go](/Users/cyber/RestoX/barness/ai/catalog_validate.go:38)、[catalog_snapshot.go](/Users/cyber/RestoX/barness/ai/catalog_snapshot.go:27) |
| 请求组装 | OpenAI / DeepSeek × Responses / Chat 复用同一采样解析语义：通用→夹取档位→调用→回调；保持 canonical key 与 wire value 分离 | [responses_options.go](/Users/cyber/RestoX/barness/ai/responses_options.go:173)、[chat_options.go](/Users/cyber/RestoX/barness/ai/chat_options.go:145) |
| full/simple 边界 | 当前 simple 已夹取，full 不统一夹取。若严格复刻 Pi，应保留并测试分离；若统一 full 行为，需要明确契约决策并登记差异，不能把它暗含为本特性修复 | [reasoning.go](/Users/cyber/RestoX/barness/ai/reasoning.go:97)、[responses_options.go](/Users/cyber/RestoX/barness/ai/responses_options.go:105) |
| summary/default | Responses 的 summary-only 从 medium 起选；省略 reasoning 与 off=null 必须按同一 helper 规则处理，不用最终 wire 字符串反查 | [responses_options.go](/Users/cyber/RestoX/barness/ai/responses_options.go:173) |
| Provider 范围 | 原生 Anthropic / Google 继续忽略通用与分档 map；Managed Effort、Gemini budget/level 不因这项兼容需求新增采样联动 | [规范 §采样选项](/Users/cyber/RestoX/barness/.scratch/barness-ai/spec.md:194)、[anthropic_effort.go](/Users/cyber/RestoX/barness/ai/anthropic_effort.go:26)、[gemini_options.go](/Users/cyber/RestoX/barness/ai/gemini_options.go:207) |
| 兼容基线与目录版本 | 当前规范和 ADR 冻结 pi-ai 1.0.0；接入新语义时先决定基线，更新 spec、ADR、oracle 发布件等价核验/哈希/provenance 与差分账本；目录内容变更更新 Catalog.Version | [spec.md](/Users/cyber/RestoX/barness/.scratch/barness-ai/spec.md:24)、[ADR-0004](/Users/cyber/RestoX/barness/docs/adr/0004-barness-ai-frozen-pi-oracle.md) |

建议实现前先列公共入口 E2E 失败场景：请求档位与夹取/映射不同、off/null/省略、非推理模型、summary-only、DeepSeek disabled/默认、分档与调用的 null/0 覆盖、full/simple 与回调覆写 reasoning、保留键/非法 JSON、多层快照共享。输出捕获请求和断言证据包，按项目已有差分分类账判定；不把 DeepSeek Responses 的 barness 扩展路由称作 Pi 自身已覆盖。相关约束来自 [ADR-0004](/Users/cyber/RestoX/barness/docs/adr/0004-barness-ai-frozen-pi-oracle.md) 和 [本地规范](/Users/cyber/RestoX/barness/.scratch/barness-ai/spec.md)。

## 第一方来源索引

[tag110]: https://api.github.com/repos/earendil-works/pi/git/ref/tags/v1.1.0
[tag102]: https://api.github.com/repos/earendil-works/pi/git/ref/tags/v1.0.2
[tag100]: https://api.github.com/repos/earendil-works/pi/git/ref/tags/v1.0.0
[pr]: https://github.com/earendil-works/pi/pull/9776
[pr-api]: https://api.github.com/repos/earendil-works/pi/pulls/9776
[pr-files]: https://github.com/earendil-works/pi/pull/9776/files
[feature-commit]: https://github.com/earendil-works/pi/commit/76dfb88f63ce51ff2e3fe2ead4fcf1f65f71f121
[changelog]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/CHANGELOG.md#L55-L59
[types]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/types.ts#L89-L93
[model-types]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/types.ts#L1132-L1152
[stream-types]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/types.ts#L193-L213
[helper]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/simple-options.ts#L24-L64
[clamp]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/models.ts#L1222-L1254
[chat-simple]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-completions.ts#L736-L755
[chat-thinking]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-completions.ts#L880-L977
[chat-named]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-completions.ts#L842-L853
[chat-apply]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-completions.ts#L1003-L1009
[chat-callback]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-completions.ts#L358-L376
[responses-build]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-responses.ts#L363-L388
[azure-build]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/azure-openai-responses.ts#L224-L249
[docs]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/docs/models.md#L101-L125
[config-schema]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/src/core/model-config.ts#L55-L75
[config-load]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/src/core/model-config.ts#L278-L345
[composer]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/src/core/provider-composer.ts#L164-L247
[composer-layers]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/src/core/provider-composer.ts#L553-L574
[sampling-tests]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/test/sampling-options.test.ts#L96-L234
[registry-tests]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/test/model-registry.test.ts#L757-L805
[deepseek-provider]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/providers/deepseek.ts#L6-L14
[deepseek-models]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/providers/deepseek.models.ts#L1-L8
[openrouter-provider]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/providers/openrouter.ts#L10-L31
[openai-provider]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/providers/openai.ts#L9-L33
[azure-provider]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/providers/azure.ts#L44-L54
[xai-provider]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/providers/xai.ts#L7-L20
[codex-build]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/src/api/openai-codex-responses.ts#L527-L618
[old-helper]: https://github.com/earendil-works/pi/blob/a13d35a742c6ef8462812a28fbe1d8c8b7431c32/packages/ai/src/api/simple-options.ts#L21-L50
[old-chat]: https://github.com/earendil-works/pi/blob/a13d35a742c6ef8462812a28fbe1d8c8b7431c32/packages/ai/src/api/openai-completions.ts#L998-L1002
[old-responses]: https://github.com/earendil-works/pi/blob/a13d35a742c6ef8462812a28fbe1d8c8b7431c32/packages/ai/src/api/openai-responses.ts#L363-L385
