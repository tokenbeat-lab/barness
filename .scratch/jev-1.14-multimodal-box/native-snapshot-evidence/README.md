# 01：TypeSafe native 快照的公开行为证据

本票将 TypeSafe 最终请求从“回解为公共 ClassifierRequest”改为独立 native 快照。
Host 的公共请求、最终协议模型及返回答案在各自边界转换；Hook 最终改名、换 kind、
重排后的题集是唯一答案依据。删除公共 decoder 的 native/noul 模式，不保留内部兼容分支。
题目仍为 choice/score/bool 文本；图片/Box 未在本票启用。

起点为 `6b893c4c7e26b7d84092644ae8f4fe7aaeb84d3b`，生产实现提交为 `d918842`，
最终验证在 `29b300e` 执行。分支为 `main`；开始时存在其他文档与规划改动，本票没有将其
纳入实现提交。新增 E2E 在修改生产代码前跑绿，完整[失败方式清单](../native-snapshot-failures.md)
与[两轴审查](REVIEW.md)独立保存。

## 验证与范围

完整结果和所有准确命令见 `receipt.json`。Baseline 的公开 focused 回归为 702 PASS / 180 NOT_RUN，
显式 Jev Pi 差分为 12 PASS。补充的 32 个公开 case 在旧实现及新实现均通过，
覆盖保留/替换 Hook、最终三种题型、精确图例、已持有 RawMessage 的改写、有限重试固定字节、
逐题损坏原子失败、封闭 native schema/重复键、能力限制及 canonical HTML 转义的精确字节边界。
补充接缝仍是 HookedClient.Classify；既有 baseline 同时覆盖 Client.Classify。

| 最终验证 | 结果 |
| --- | --- |
| 同一公开 focused 回归 | 702 PASS、180 NOT_RUN |
| 显式 Jev/Pi classify 与目录差分 | 12 PASS，pending 0 |
| 普通 / live build tags 的 vet | 均退出 0 |
| 相关 race | 448 PASS、178 NOT_RUN |
| 开启冻结 Pi 的完整套件（只执行一次） | 7,468 PASS、187 NOT_RUN；861 个差分全部 PASS |
| Oracle 启动与固定历史包核验 | 9 项 Node 检查及原验证器均退出 0 |

NOT_RUN 项未被计作通过，主要是未启用的图片压力/导入检查；focused 的两个目录差分在显式
Jev 差分阶段另行执行。完整套件包含冻结 Pi 差分，但没有真实服务调用或设计压力运行。
Live 为 NOT_RUN；本票通过只证明受控 HTTP 文本行为等价，不声明真实服务、图片能力或整版发布 PASS。

目录版本仍为 `2026-10-13.1`，Hash 为
`sha256:faae843ac61fea6e3ae59bb75165dc97fcf6aa409dabfe0a2f7c9dccbba12029`。
TypeSafe fixtures、目录源码、支持矩阵、Pi pins/模型数据清单和旧固定 Jev 包在 baseline/final
逐文件 Hash 相同。冻结 Pi 仍为 1.1.0；实际差分及 9 项 Node 检查验证其安装与协议边界。

## 证据内容

生成的 `evidence.zip` 在本目录，按仓库约定不提交 Git；`receipt.json` 固定归档与清单 Hash。
归档包含 baseline 与 final 原始记录，每个 case 保留 fixture Hash、输入、最终 HTTP 请求/响应脚本、
返回结果或错误、Attempt/Observer、Host 读取次数、资源断言和精确重放命令。
本票没有图片输入，图片 Hash 项明确不适用；文本 fixtures 的原始 Hash 与脱敏后 artifact Hash 分开保存。

每个 run 的 `manifest.json` 固定所有 case artifacts；集合的 `files.json` 固定包括 manifest、日志、
审计在内的每个文件。`inputs.json` 包含受测生产/测试源与稳定输入的 Hash；final 的
`configuration.json` 保存有限资源策略、合成目录及补充测试覆盖的配置源码和 Hash，
这些配置源码与 baseline 的 `inputs.json` 相同。所有标准审计均由既有 auditbundle 执行。

完整测试退出 0 后，首轮集合汇总递归读到了两个测试内部的示例 `manifest.json`，
因它们不含 cases 而退出 1。`packaging-diagnostics.json` 保留该失败；修复 `90157fa`
把范围限定为 command/timestamp 清单，使用 `--finish-only` 校验全部命令收据与源码 Hash 后
继续审计和封存原产物，未再次执行完整测试。该修复也经两轴独立审查，均 0 项发现。

## 重放

从仓库根目录执行，输出目录必须不存在，两个阶段都只使用受控 HTTP：

```sh
python3 .scratch/jev-1.14-multimodal-box/native-snapshot-evidence/replay.py \
  --phase baseline --out .evidence/native-snapshot-next-baseline
python3 .scratch/jev-1.14-multimodal-box/native-snapshot-evidence/replay.py \
  --phase final --out .evidence/native-snapshot-next-final \
  --baseline .evidence/native-snapshot-next-baseline
```

重放当前工作树不会重建历史起点；历史 baseline 的原始 Hash 与 captures 保存在本次固定包。
若 oracle 安装缺失，使用原 lockfile：

```sh
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
```

核验本次已保存归档（不解压、不发网络请求）：

```sh
python3 - <<'PY'
import hashlib, json, pathlib, zipfile
folder = pathlib.Path('.scratch/jev-1.14-multimodal-box/native-snapshot-evidence')
receipt = json.loads((folder / 'receipt.json').read_text())
sha = lambda raw: hashlib.sha256(raw).hexdigest()
assert sha((folder / 'evidence.zip').read_bytes()) == receipt['archive_sha256']
with zipfile.ZipFile(folder / 'evidence.zip') as bundle:
    names = bundle.namelist()
    assert len(names) == len(set(names))
    for phase in ('baseline', 'final'):
        prefix = phase + '/'
        raw = bundle.read(prefix + 'files.json')
        assert sha(raw) == receipt[phase]['files_manifest_sha256']
        files = json.loads(raw)
        assert {n for n in names if n.startswith(prefix)} == {prefix + n for n in files} | {prefix + 'files.json'}
        for name, expected in files.items():
            assert sha(bundle.read(prefix + name)) == expected
print('PASS: archive and every baseline/final artifact verified')
PY
```

回滚没有数据迁移：直接恢复本票前代码，或按原规则禁用 classifier Binding / 型号授权以停止新调用。
已经 pin 的调用继续按原快照完成，已报告的用量保留；后续图片工单仍需自己的验收与证据。
