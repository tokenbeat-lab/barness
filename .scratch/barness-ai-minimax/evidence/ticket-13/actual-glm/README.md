# 票13中国区直连：实际 GLM 三路

实际来源 `51c3433`，固定三路各22场景，零自动重试。Chat22 PASS / 22 calls / 22 Attempts；Messages同样22/22/22。Responses19 PASS、3 FAIL，22场景实际只有21 calls / 21 Attempts：GLM5.3 `tool-call` 与 `control-max` 各发生一次Attempt，约90s未收到HTTP响应头，受 ResponseHeaderTimeout 终止；缺少工具前置结果的 `tool-result` 保留 FAIL，未生成调用或Attempt。GLM5.2/5.3 native-history 均 PASS。

本轮 GLM 合计65/65，原封保留旧300/300后，当前 MiniMax 发布的 GLM/legacy 续期作用域累计365/365。MiniMax自身candidate/formal预算与早先独立GLM任务历史另存，不属于此数值。未报告 usage 不表示免费。两次超时没有 HTTP status，不能据此定位 DNS/TCP/TLS、厂商内部原因或认定账户权限问题。

三份原始包均有完整性清单和 audit0；它们不授予三路 PASS。Responses为 `outcome: FAIL`、`complete: false`，没有成功 launcher capturePins；保留现有清单，不补造成功 anchor。公开 renewal CLI 实际拒绝新三路发布，四份 raw/usage/support selector 前后逐字节相同；其间生成的 Chat ZIP 是失败流水线的中间文件。未发布新GLM支持视图，也未发布Jev支持包。

`index.json` 固定外部原始归档和清单 pins；`summary.json` 与 `scenarios.jsonl` 来自独立 Python 对原始场景、最终Attempt、wire和预算求和。`verify.py` 只读校验所有归档成员、原 seal、审计、实际 launcher/command pins、独立消费及发布拒绝。复用并锁定[实现归档校验器](../implementation/verify.py)的字节；不导入付费工具，不读取 `.env`、不调用 `summarize`、不访问网络、不运行任何真实请求。

```sh
python3 -B .scratch/barness-ai-minimax/evidence/ticket-13/actual-glm/verify.py
# 只在临时文件上复核7种拒绝，前后均校验原始包，不重放实际请求。
python3 -B .scratch/barness-ai-minimax/evidence/ticket-13/actual-glm/check-boundaries.py
```

在主工作树运行；其他工作树增加 `--common-root <主工作树绝对路径>`。原始归档位于 common root 的 `.evidence/minimax-retained/ticket-13-actual-glm/`，这里不重复提交。约220MB预检与36MB binary只保存原位置、有限元数据和hash引用，不复制进此归档。83来源预检已在修复前退役，无当时APPROVED或付费intent；单列历史，不续借为51实际验收。

独立校验的失败边界：归档或成员变更、路径越界、符号链接、重复/遗漏成员、超限解压、source/command/report pins不符、消费或场景数不符、FAIL被改成PASS、缺失capturePins被补造、selector变化均拒绝。完整发布门禁和其余十五项由父任务另行保存；这份有限归档不证明整份 spec 已通过，也不授权再运行付费场景。
