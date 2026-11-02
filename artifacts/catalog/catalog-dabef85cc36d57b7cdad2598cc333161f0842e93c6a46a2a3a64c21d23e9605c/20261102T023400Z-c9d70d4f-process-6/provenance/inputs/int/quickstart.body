> ## Documentation Index
> Fetch the complete documentation index at: https://platform.kimi.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Kimi K3.1 Pro

> Get started with the Kimi K3.1 Pro API: reasoning effort, streaming, vision input, tool calling, and a 1M-token context window.

## Overview

We're introducing Kimi K3.1 Pro, the first model in our new K3.1 series and our most capable 3T-class open-weight model to date, with a 1-million-token context window. It delivers substantial improvements across coding, 3D creation, financial research, and other agentic capabilities.

We are also releasing Kimi K3.1 Pro Highspeed, which delivers 2-3× the speed of K3.1 Pro for more responsive interactions. The full model weights will be released by Nov 4.

### Advancing the Frontier of Open Intelligence

Kimi K3.1 Pro is a post-trained version of Kimi K3, with substantially larger post-training compute. To enable large-scale joint RL across hundreds of downstream domains, we optimized every layer of the stack — the RL framework, long-context inference, training efficiency, sandbox infrastructure, rollout scheduling, and a shared platform for environment quality checks — so that each increment of RL compute delivers stable, predictable returns. As a result, K3.1 Pro achieves consistent gains in every domain.

![Kimi K3.1 Pro benchmark overview](https://mintcdn.com/moonshotai/jDcKnASQd0dgNqpK/assets/pics/k3-1-pro/benchmark-grid.png?fit=max\&auto=format\&n=jDcKnASQd0dgNqpK\&q=85\&s=0860c92750580d1d145809b4d867a054)

Across coding, agentic, and knowledge benchmarks, K3.1 Pro matches or approaches the strongest proprietary models. On Terminal-Bench 4.0, which tests \*\*\*, it scores \[xx.x%]. On Terminal-Bench-Science 0.1, which includes 70 tasks across the life, physical, Earth, mathematical, and engineering sciences, it scores \[xx.x%], above K3's \[xx.x%]. And on \[our internal kcb evaluation], which \[tests xxx], it scores \[xx.x%], the best of any open-weight model we have tested.

![Terminal-Bench 4.0 score vs. average cost per task](https://mintcdn.com/moonshotai/8qRizDfnIUj5a7Y2/assets/pics/k3-1-pro/terminal-bench-pareto.png?fit=max\&auto=format\&n=8qRizDfnIUj5a7Y2\&q=85\&s=1ff75f4facbf540a6bce98d5cd1ac780)

[Get Started](https://platform.kimi.ai/docs/guide/kimi-k3-1-pro-quickstart.md#get-started)

### Coding

K3.1 Pro covers the full range of coding work, from quick fixes in real codebases to full-project builds. It also brings a more considered visual style to games and web experiences, and knows how to build them with the right tools.

K3.1 Pro works effectively across game engines such as Unity, Unreal Engine, and Godot, combining them with web search and 3D generation to create complete experiences. Its front-end pages, interactive Three.js scenes, and 3D assets show stronger composition, typography, animation, and polish. The examples below were created end to end by K3.1 Pro:

<video autoPlay muted loop playsInline controls>
  <source src="https://mintcdn.com/moonshotai/Q56nkgCpT-8A4uaO/assets/videos/k3-1-pro/coding-celestial-palace.mp4?fit=max&auto=format&n=Q56nkgCpT-8A4uaO&q=85&s=07fe3acf5074d970c1860fc757f8d3a3" type="video/mp4" />
</video>

About this video

Kimi K3.1 Pro built a mythical Eastern celestial palace in Unreal Engine 5.8, featuring a 36-bay colonnade with a radius of 1,750 meters and a colossal dragon procedurally modeled in Blender. The main character was created using MetaHuman. The scene contains 36 million source triangles and uses Nanite for rendering and Lumen for lighting. External assets used: Cloud Forge VDB Clouds, Poly Haven/ambientCG 8K materials, and NASA LRO/LOLA moon data.

<video autoPlay muted loop playsInline controls>
  <source src="https://mintcdn.com/moonshotai/Q56nkgCpT-8A4uaO/assets/videos/k3-1-pro/coding-eastern-rpg.mp4?fit=max&auto=format&n=Q56nkgCpT-8A4uaO&q=85&s=902b28cc925a74698015ae51d46d1e22" type="video/mp4" />
</video>

About this video

Kimi K3.1 Pro created an Eastern ancient-themed 3D RPG preview video in Unreal Engine 5.8, featuring a swordsman flying through a sea of clouds, past floating islands, and into a classical city above the clouds. Concept design and part of the asset inspiration come from Rednote creator @-Duduu-. External assets used: the Gandhara Buddha and Water-Moon Guanyin 3D scans (Minneapolis Institute of Art, CC0).

<video autoPlay muted loop playsInline controls>
  <source src="https://mintcdn.com/moonshotai/Q56nkgCpT-8A4uaO/assets/videos/k3-1-pro/coding-godot-adventure.mp4?fit=max&auto=format&n=Q56nkgCpT-8A4uaO&q=85&s=1168d5409c52f3726a01086399ea495b" type="video/mp4" />
</video>

About this video

Kimi K3.1 Pro built a stylized 3D action-adventure game using Godot 4.7.2. The project features five explorable regions with procedural terrain, roads, water, vegetation, and village layouts, alongside third-person combat, a two-phase Boss fight, gathering, a day–night cycle, and campfire fast travel. External assets used: Quaternius Stylized Nature MegaKit, Quaternius Medieval Village MegaKit, Kenney Survival Kit, Quaternius Ultimate Monsters, KayKit Character Pack, KayKit Character Animations 1.1, Quaternius Universal Animation Library, CDmir Goblin, rubberduck Vampire Bat, and the Fredoka, ZCOOL KuaiLe, and LXGW WenKai fonts.

<video autoPlay muted loop playsInline controls>
  <source src="https://mintcdn.com/moonshotai/Q56nkgCpT-8A4uaO/assets/videos/k3-1-pro/coding-robotic-arm.mp4?fit=max&auto=format&n=Q56nkgCpT-8A4uaO&q=85&s=d933f2eff8a6a48213e1ed8bce6f183c" type="video/mp4" />
</video>

About this video

Kimi K3.1 built a browser-based 3D robotic-arm reveal with exploded views, transparent shells and cross-sections. It generates 2,317 parts and 268 dynamic wire routes in code, verified by 55 passing browser checks.

K3.1 Pro's long-horizon coding capabilities are also strengthened through large-scale reinforcement learning. In one internal test, it built a human walking dynamics simulator with integrated training and evaluation pipelines on top of human-atlas, an open-source anatomical model of the human body comprising 296 skeletal components, including cartilage, and 402 muscle components. Across 167 missions and 420 independent reviews, it analyzed experimental results, diagnosed problems, and iteratively refined the walking controller, enabling the simulated human to go from collapsing within seconds to walking continuously with alternating strides for 11.2 seconds. When reward shaping failed to improve arm motion, K3.1 Pro did not simply keep adjusting the reward: it ran open-loop probes of the simulator's dynamics, identified and corrected 36 errors in shoulder-muscle moment-arm calculations, and then retrained the controller.

### 3D Modeling and Design

K3.1 Pro is strong at 3D creation and spatial reasoning. From a text brief, sketch, or reference image, it can infer a subject's shape, proportions, and spatial relationships, then reconstruct it as a usable 3D model. It also has a stronger understanding of existing asset libraries and model collections.

![Kimi K3.1 Pro BenchCAD results](https://mintcdn.com/moonshotai/8qRizDfnIUj5a7Y2/assets/pics/k3-1-pro/benchcad.png?fit=max\&auto=format\&n=8qRizDfnIUj5a7Y2\&q=85\&s=b3f5ccff4348ccf9e7e0ef68aee2da17)

K3.1 Pro is strong at parametric modeling, such as mechanical parts, architectural models, products with multiple configurable parameters, and custom furniture. Without relying on MCP connectors, K3.1 Pro can directly write and run scripts for professional modeling tools such as Blender and CAD software to construct and refine assets, modify individual components, and adjust geometry, materials, and scene layout.

Below are examples of 3D assets K3.1 Pro created for brand films, games, and interactive web experiences:

<video autoPlay muted loop playsInline controls>
  <source src="https://mintcdn.com/moonshotai/Q56nkgCpT-8A4uaO/assets/videos/k3-1-pro/3d-airport-blender.mp4?fit=max&auto=format&n=Q56nkgCpT-8A4uaO&q=85&s=46c558f1e1b3d26c019ddafc454406a6" type="video/mp4" />
</video>

About this video

Kimi K3.1 Pro built an airport's indoor rainforest valley in Blender, producing a 17.5-second, 1920×1080, 60 fps modeling showcase video.
Tools used: Blender, Python, Geometry Nodes, Cycles, Pillow, NumPy, and FFmpeg. External assets used: "Kloofendal 48d Partly Cloudy (Pure Sky)" 4K HDRI from Poly Haven, by Greg Zaal and Jarod Guest, CC0 licensed, used for sky lighting and environment reflections.

<video autoPlay muted loop playsInline controls>
  <source src="https://mintcdn.com/moonshotai/Q56nkgCpT-8A4uaO/assets/videos/k3-1-pro/3d-turbofan-freecad.mp4?fit=max&auto=format&n=Q56nkgCpT-8A4uaO&q=85&s=bc19571439352f369c6a0265eaef581b" type="video/mp4" />
</video>

About this video

Kimi K3.1 Pro used Python to build a detailed turbofan model in FreeCAD, covering the fan, reduction gearbox, compressors, combustor, turbines, and external piping across 2,100 solids. The project supports step-by-step assembly, an exploded internal view, and rotational animation demonstrating the fan, low-pressure spool, and high-pressure spool at 2,780, 8,590, and 12,400 RPM respectively.

<video autoPlay muted loop playsInline controls>
  <source src="https://mintcdn.com/moonshotai/Q56nkgCpT-8A4uaO/assets/videos/k3-1-pro/3d-modular-lamp.mp4?fit=max&auto=format&n=Q56nkgCpT-8A4uaO&q=85&s=9090d50e91183b430f77cd2b08568b8c" type="video/mp4" />
</video>

About this video

Using trimesh, Kimi K3.1 Pro created a five-part, 3D-printable table lamp inspired by Fujian tulou, reimagining its distinctive circular architecture as an intricate openwork lattice.

<video autoPlay muted loop playsInline controls>
  <source src="https://mintcdn.com/moonshotai/Q56nkgCpT-8A4uaO/assets/videos/k3-1-pro/3d-cathedral.mp4?fit=max&auto=format&n=Q56nkgCpT-8A4uaO&q=85&s=69dda71df0754cf16f8d9f199fb9cf84" type="video/mp4" />
</video>

About this video

Kimi K3.1 Pro built an interactive, browser-based 3D experience in Blender, creating a cathedral's interior with roughly 3.3 million triangles and animating 8,368 components, with ribbed vaults, stone sculptures, and vivid stained-glass lighting. External assets used: stone textures from Poly Haven (CC0) and Mylius's stained-glass photograph (CC BY-SA 3.0).

[Get Started](https://platform.kimi.ai/docs/guide/kimi-k3-1-pro-quickstart.md#get-started)

### Finance

With stronger reasoning and agentic capability, K3.1 Pro  is capable of end-to-end delivery of complete financial workflows. K3.1 Pro achieves state-of-the-art results on OfficeQA Pro v2 and SpreadsheetBench v2. OfficeQA Pro v2 tests complex fiscal data analysis over real financial documents through retrieval, cross-document verification, and multi-step numerical reasoning. SpreadsheetBench v2 evaluates end-to-end spreadsheet workflows including financial modeling, cross-sheet formula dependencies, and error diagnosis. K3.1 Pro also improves over K3 on Frontier Finance, which measures support for complete investment workflows from screening to profitability analysis.

![Kimi K3.1 Pro finance benchmark results](https://mintcdn.com/moonshotai/8qRizDfnIUj5a7Y2/assets/pics/k3-1-pro/finance-benchmark.png?fit=max\&auto=format\&n=8qRizDfnIUj5a7Y2\&q=85\&s=bfad09fd344db7a21f066dfaa2046ef4)

Financial analysis and investment research demand a high degree of factual accuracy. To address this, we made targeted improvements to K3.1 Pro's factual accuracy on real-world tasks and challenging prompts. In an evaluation using de-identified conversations where users had flagged factual errors in earlier models's responses, its claim-level hallucination rate fell from 23.35% for K3 to 6.85%, while the share of responses containing at least one factual error was just 46.4% of its predecessor's rate. (This evaluation is a stress test for factual accuracy, with examples deliberately selected to probe situations where models are prone to errors. The reported rates do not represent overall error rates in daily use.)

![Kimi K3.1 Pro hallucination benchmark results](https://mintcdn.com/moonshotai/8qRizDfnIUj5a7Y2/assets/pics/k3-1-pro/finance-hallucination.png?fit=max\&auto=format\&n=8qRizDfnIUj5a7Y2\&q=85\&s=f7e2ab71daeb2cad1dc1df7651263d4d)

## Access requirements

Kimi K3.1 Pro is a flagship model: it is unlocked after a successful top-up (minimum \$1). Your cumulative top-up amount also determines your account tier and rate limits (concurrency, RPM, TPM, TPD) — see [Recharge and Rate Limits](https://platform.kimi.ai/docs/pricing/limits.md).

## Get started

* [Playground](https://platform.kimi.ai/playground)
* [Get your API key](https://platform.kimi.ai/console/api-keys)

The examples require Python 3.9+ and the OpenAI SDK. Install the SDK and initialize the client once; later Python examples reuse `client`.

```bash theme={null}
python3 -m pip install --upgrade 'openai>=1.0'
```

```python theme={null}
import os

from openai import OpenAI

client = OpenAI(
    api_key=os.environ["MOONSHOT_API_KEY"],
    base_url="https://api.moonshot.ai/v1",
)
```

## Basic call

> [!NOTE]
>
> K3.1 Pro also has a high-speed variant, `kimi-k3.1-pro-highspeed`. See [Model Pricing](https://platform.kimi.ai/docs/pricing/chat.md) for model IDs and prices.

> [tab] Python
>
> ```python theme={null}
> completion = client.chat.completions.create(
>     model="kimi-k3.1-pro",
>     messages=[{"role": "user", "content": "Introduce Kimi K3.1 Pro in one sentence."}],
> )
>
> print(completion.choices[0].message.content)
> ```

> [tab] cURL
>
> ```bash theme={null}
> curl https://api.moonshot.ai/v1/chat/completions \
>   --header "Authorization: Bearer $MOONSHOT_API_KEY" \
>   --header "Content-Type: application/json" \
>   --data '{
>     "model": "kimi-k3.1-pro",
>     "messages": [{"role": "user", "content": "Introduce Kimi K3.1 Pro in one sentence."}]
>   }'
> ```

## Reasoning effort

K3.1 Pro always has thinking mode enabled and supports configuring its reasoning effort with the top-level `reasoning_effort` request field. Reasoning effort supports `low`, `high`, and `max` (default `max`). See [Reasoning Effort](https://platform.kimi.ai/docs/guide/use-reasoning-effort.md) for usage.

```python theme={null}
completion = client.chat.completions.create(
    model="kimi-k3.1-pro",
    reasoning_effort="max",
    messages=[{"role": "user", "content": "Prove that the square root of 2 is irrational."}],
)

print(completion.choices[0].message.content)
```

For multi-turn conversations and tool calls, add the complete assistant message returned by the API to the next request. Do not keep only `content`.

## Streaming

Streaming responses provide separate `reasoning_content` and final-answer `content` deltas. See [Streaming Output](https://platform.kimi.ai/docs/guide/utilize-the-streaming-output-feature-of-kimi-api.md) for details.

```python theme={null}
stream = client.chat.completions.create(
    model="kimi-k3.1-pro",
    messages=[{"role": "user", "content": "Explain why the sky is blue."}],
    stream=True,
)

for chunk in stream:
    delta = chunk.choices[0].delta
    reasoning = getattr(delta, "reasoning_content", None)
    if reasoning:
        print(reasoning, end="", flush=True)
    if delta.content:
        print(delta.content, end="", flush=True)
```

## Vision input

For vision messages, `content` must be an array of objects, not a serialized string. See [Vision Input](https://platform.kimi.ai/docs/guide/use-kimi-vision-model.md) for formats and limits.

> [tab] Local image
>
> ```python theme={null}
> import base64
> from pathlib import Path
>
> image_data: str = base64.b64encode(Path("image.png").read_bytes()).decode()
> completion = client.chat.completions.create(
>     model="kimi-k3.1-pro",
>     messages=[
>         {
>             "role": "user",
>             "content": [
>                 {
>                     "type": "image_url",
>                     "image_url": {"url": f"data:image/png;base64,{image_data}"},
>                 },
>                 {"type": "text", "text": "Describe this image."},
>             ],
>         }
>     ],
> )
>
> print(completion.choices[0].message.content)
> ```

> [tab] Video file
>
> ```python theme={null}
> from pathlib import Path
>
> video = client.files.create(file=Path("video.mp4"), purpose="video")
> try:
>     completion = client.chat.completions.create(
>         model="kimi-k3.1-pro",
>         messages=[
>             {
>                 "role": "user",
>                 "content": [
>                     {
>                         "type": "video_url",
>                         "video_url": {"url": f"ms://{video.id}"},
>                     },
>                     {"type": "text", "text": "Summarize this video."},
>                 ],
>             }
>         ],
>     )
>     print(completion.choices[0].message.content)
> finally:
>     client.files.delete(video.id)
> ```

## Structured output

Use `json_schema` with `strict: true` to constrain the final `message.content`. Parse only that field, not `reasoning_content`.

> [accordion] Name and age schema
>
> ```python theme={null}
> import json
>
> completion = client.chat.completions.create(
>     model="kimi-k3.1-pro",
>     messages=[
>         {"role": "user", "content": "Lin is 28 years old. Extract the name and age."}
>     ],
>     response_format={
>         "type": "json_schema",
>         "json_schema": {
>             "name": "person",
>             "strict": True,
>             "schema": {
>                 "type": "object",
>                 "properties": {
>                     "name": {"type": "string"},
>                     "age": {"type": "integer"},
>                 },
>                 "required": ["name", "age"],
>                 "additionalProperties": False,
>             },
>         },
>     },
> )
>
> person: dict[str, object] = json.loads(
>     completion.choices[0].message.content or "{}"
> )
> print(person)
> ```

See [Structured Output](https://platform.kimi.ai/docs/guide/response_format.md).

## Partial Mode

Add an assistant message with `partial=True` at the end of `messages` to continue from a text prefix. Prepend that prefix when displaying the final result.

```python theme={null}
prefix: str = "Conclusion: "
completion = client.chat.completions.create(
    model="kimi-k3.1-pro",
    messages=[
        {"role": "user", "content": "In one sentence, explain why API compatibility matters."},
        {"role": "assistant", "content": prefix, "partial": True},
    ],
)

print(prefix + (completion.choices[0].message.content or ""))
```

See [Partial Mode](https://platform.kimi.ai/docs/guide/use-partial-mode-feature-of-kimi-api.md).

## Custom tools and `tool_choice`

Use `tool_choice="required"` on the first turn to require at least one tool call. After executing every call, return the complete assistant message and append one tool result with the matching `tool_call_id` for each call.

> [accordion] Minimal weather agent loop
>
> ```python theme={null}
> import json
> from typing import Any
>
> tools: list[dict[str, Any]] = [
>     {
>         "type": "function",
>         "function": {
>             "name": "get_weather",
>             "description": "Get the weather for a city",
>             "parameters": {
>                 "type": "object",
>                 "properties": {"city": {"type": "string"}},
>                 "required": ["city"],
>             },
>         },
>     }
> ]
> messages: list[Any] = [
>     {"role": "user", "content": "What is the weather in San Francisco today?"}
> ]
>
> first = client.chat.completions.create(
>     model="kimi-k3.1-pro",
>     messages=messages,
>     tools=tools,
>     tool_choice="required",
> )
> assistant_message = first.choices[0].message
> messages.append(assistant_message)
>
> for tool_call in assistant_message.tool_calls or []:
>     arguments: dict[str, str] = json.loads(tool_call.function.arguments)
>     result: str = json.dumps(
>         {"city": arguments["city"], "weather": "sunny", "temperature_c": 24}
>     )
>     messages.append(
>         {"role": "tool", "tool_call_id": tool_call.id, "content": result}
>     )
>
> final = client.chat.completions.create(
>     model="kimi-k3.1-pro",
>     messages=messages,
>     tools=tools,
> )
> print(final.choices[0].message.content)
> ```

See [Tool Choice](https://platform.kimi.ai/docs/guide/use-tool-choice.md).

## Dynamic tool loading

Place a complete tool definition in a `system` message without `content`. The tool becomes available from that message onward.

> [accordion] Load a calculator dynamically
>
> ```python theme={null}
> from typing import Any
>
> dynamic_messages: list[dict[str, Any]] = [
>     {"role": "user", "content": "Calculate 23 times 47."},
>     {
>         "role": "system",
>         "tools": [
>             {
>                 "type": "function",
>                 "function": {
>                     "name": "calculate",
>                     "description": "Evaluate an arithmetic expression",
>                     "parameters": {
>                         "type": "object",
>                         "properties": {
>                             "expression": {
>                                 "type": "string",
>                                 "description": "The arithmetic expression to evaluate",
>                             }
>                         },
>                         "required": ["expression"],
>                     },
>                 },
>             }
>         ],
>     },
> ]
> completion = client.chat.completions.create(
>     model="kimi-k3.1-pro",
>     messages=dynamic_messages,
> )
>
> print(completion.choices[0].message.tool_calls)
> ```

* Include the complete `name`, `description`, and `parameters` definition.
* The declaration takes effect at its position in `messages`.
* Keep this message in later request history; the server does not retain it.

See [Dynamic Tool Loading](https://platform.kimi.ai/docs/guide/use-dynamic-tool-loading.md).

## 1M context and TTL caching

Context caching is automatic for regular model requests — no manual cache creation: once an eligible prefix is written to the cache, hits are billed at the cached-input price, and each write incurs a one-time charge. TTL supports two values, `5m` (default) and `1h`; choose one based on your request interval — the default `5m` is enough for rapid back-and-forth, while workloads with gaps over 5 minutes should pass `prompt_cache_options.ttl` to select the `1h` tier and avoid repeated writes after expiry. The cache is stored in blocks; any partial block that does not fill a whole block cannot be written to the cache and is billed as a cache miss.

```python theme={null}
from pathlib import Path

knowledge: str = Path("knowledge-base.md").read_text(encoding="utf-8")

for question in ["Summarize the key conclusions.", "List three implementation risks."]:
    completion = client.chat.completions.create(
        model="kimi-k3.1-pro",
        messages=[
            {"role": "system", "content": knowledge},
            {"role": "user", "content": question},
        ],
    )
    print(completion.choices[0].message.content)
```

For help choosing a TTL, improving hit rates, and verifying usage, see [Context Caching](https://platform.kimi.ai/docs/guide/context-caching.md).

## Official tools

Official tools are integrated through Formula:

1. Fetch tool definitions from the Formula `/tools` endpoint.
2. Add those definitions to the Chat Completions `tools` field.
3. When the model returns `tool_calls`, submit each function name and arguments to the Formula `/fibers` endpoint.
4. Add the complete assistant message and Fiber output as the corresponding tool message.
5. Call Chat Completions again until the model returns a final answer.

See [Official Tools](https://platform.kimi.ai/docs/guide/use-official-tools.md) for the complete client and API contract. Web search is being updated and is not recommended for use in the near term.

## Important limits

* Reasoning effort is configured with the top-level `reasoning_effort` request field and supports `low`, `high`, and `max` (default `max`); K3.1 Pro always has thinking mode enabled.
* `max_completion_tokens` defaults to 131072 and can be set up to 1048576.
* `temperature=1.0`, `top_p=0.95`, `n=1`, `presence_penalty=0`, and `frequency_penalty=0` are fixed; omit them from requests.
* Return the complete assistant message unchanged in multi-turn conversations and tool calls.
* Vision input does not support public image URLs. Use base64 or `ms://<file-id>`, and make `content` an array of objects.
* Web search is being updated and is not recommended for production workflows in the near term.

## FAQ

> [accordion] How is Kimi K3.1 Pro billed?
>
> Kimi K3.1 Pro offers a 1M-token context and uses flat pay-as-you-go pricing — there is no tiering by context length. Billable items include cache writes (with separate TTL 5min and TTL 1h tiers), input (with separate rates for cache hits and misses), and output, all billed at uniform per-token prices. See [Kimi K3.1 Pro pricing](https://platform.kimi.ai/docs/pricing/chat.md).

> [accordion] How do I turn off Kimi K3.1 Pro's chain-of-thought?
>
> You can't — K3.1 Pro always thinks. If the reasoning takes too long, set `reasoning_effort` to `low` to reduce the reasoning effort. See [Reasoning Effort](https://platform.kimi.ai/docs/guide/use-reasoning-effort.md).

## Related docs

> [card] [Kimi K3.1 Pro Pricing](https://platform.kimi.ai/docs/pricing/chat.md)
>
> Review input and output prices.

> [card] [Reasoning Effort](https://platform.kimi.ai/docs/guide/use-reasoning-effort.md)
>
> Configure reasoning\_effort.

> [card] [Vision Input](https://platform.kimi.ai/docs/guide/use-kimi-vision-model.md)
>
> Send images and videos.

> [card] [Structured Output](https://platform.kimi.ai/docs/guide/response_format.md)
>
> Use strict JSON Schema.

> [card] [Partial Mode](https://platform.kimi.ai/docs/guide/use-partial-mode-feature-of-kimi-api.md)
>
> Continue from a prefix.

> [card] [Tool Choice](https://platform.kimi.ai/docs/guide/use-tool-choice.md)
>
> Control whether the model calls tools.

> [card] [Dynamic Tool Loading](https://platform.kimi.ai/docs/guide/use-dynamic-tool-loading.md)
>
> Inject tool definitions on demand.

> [card] [Tool Calling Best Practices](https://platform.kimi.ai/docs/guide/kimi-k3-tool-calling-best-practice.md)
>
> Combine tool-calling features.

> [card] [Official Tools](https://platform.kimi.ai/docs/guide/use-official-tools.md)
>
> Integrate Formula tools.
