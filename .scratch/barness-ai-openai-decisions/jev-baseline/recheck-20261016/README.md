# 工单 01：2026-10-16 当前工作树复核

工单在本次开始前已经 resolved，公共 E2E、fixture 和固定包均已存在。本次仅重放并补充当前证据，没有修改生产代码、测试或原始固定包；不重新实现已经交付的切片。

受测 HEAD 为 `291d2e52066bc846170d561626d87e0ecc429dcc`，当前分支为 `main`。开始时存在其他工作的文档及未跟踪文件；逐文件 Hash 单独记录并在完成时核对，不把工作树描述为 clean。本次仅使用受控本地 Provider 和冻结 Pi，没有真实厂商调用。

## 基线与当前版本

原包仍固定 `8ea61b8d70de6c84fe27be34dc7d21c5c40841d9`，原 655 个 case 的状态仍为 652 PASS、2 NOT_RUN、1 个预期先红 FAIL。`verify.py --audit` 本次执行成功，历史 release FAIL 继续保留。

当前目录从基线的 `2026-10-11.1` 演进到 `2026-10-13.1`，Hash 为 `sha256:faae843ac61fea6e3ae59bb75165dc97fcf6aa409dabfe0a2f7c9dccbba12029`。与包内原始输入清单相比，64 个既有 `ai/` 输入改变或删除，另新增 104 个 Go 文件。完整偏移清单和当前 Go 源码 Hash 在本次归档中；偏移来自后续 Decisions、图片能力、宿主评估及压力流完整性工作，不能把当前执行称作原 HEAD 的重建。

`ai/e2e/jev_{regression,rejection_order,baseline_evidence}_test.go` 和 `testdata/typesafe/jev-regression.json` 自原实现提交 `625b50d` 后没有改动。Pi 的 package/lockfile、provenance 与 43 文件模型数据清单也与设计基线相同：唯一运行版本仍为 1.1.0，commit `abe508e1b89912adde45528136c3221eb69acdd7`，model-data SHA-256 为 `8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8`。运行器后续新增 source pins、独立 Decisions 分派和网络边界校验；当前 9 个 Node 检查全部通过。

## 当前执行

| 本次执行 | 结果 |
| --- | --- |
| 已有 Jev、目录、回调、unary、官方 wire replay | 568 PASS、180 NOT_RUN；退出 0 |
| 显式 Pi Jev classify/catalog 差分 | 12 PASS、pending 0；退出 0 |
| plain/hooked 公共行为与拒绝顺序 | 98 PASS；退出 0 |
| 固定包反篡改 | 4 PASS；退出 0 |
| Oracle 启动及协议边界 | 9 PASS；退出 0 |
| 相关 race | 416 PASS、178 NOT_RUN；退出 0 |
| 普通及 live build tags 的 vet | 均退出 0 |
| 开启 Pi 的完整测试套件 | 7,436 PASS、187 NOT_RUN；861 个差分全部 PASS、pending 0；退出 0 |

准确命令、退出码、每次执行的版本、状态计数和产物 Hash 见 [receipt.json](receipt.json)。原始捕获、结果、Attempt、Observer、assertions、用例重放命令、日志及输入清单保存在本地 `recheck.zip`；该生成归档按仓库 `.gitignore` 不进入 Git。原始 `baseline.zip`、`manifest.json`、`validation.json` 和 `validation.zip` 没有改写。

初次打包时缺少集合根 `manifest.json`，既有审计器拒绝把该目录视为证据包。已保留退出 1 的诊断日志与原因，补充引用实际已完成运行的集合清单后再执行原审计器；没有改变审计规则或删除失败记录。打包诊断在归档的 `packaging-diagnostics.json` 和 `whole-recheck-audit-initial.log`。

`existing` 的 180 NOT_RUN 中，2 个目录差分在该阶段未开启 Pi，另 178 个是后续新增的图片压力场景和导入检查。独立 `jev-differential` 阶段显式开启 Pi，12 个 case 全部 PASS；相关 race 的 178 NOT_RUN 同样属于图片压力场景和导入检查。完整套件的 187 NOT_RUN 为上述 178 个加上原有 mixed/policy/tool-argument 压力与规模检查 9 个。缺少压力执行不视作通过。本次不作全仓库 race、设计压力或完整 releasegate 的结论。

从仓库根目录重放（输出目录必须不存在）：

```sh
python3 .scratch/barness-ai-openai-decisions/jev-baseline/verify.py --audit
python3 .scratch/barness-ai-openai-decisions/jev-baseline/replay.py \
  --out .evidence/jev-baseline-next
env -u BARNESS_AI_LIVE -u BARNESS_AI_PRESSURE \
  BARNESS_AI_PIDIFF=1 \
  BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/jev-baseline-next/full" \
  go test ./... -count=1 -json
```

如未安装冻结 oracle，先执行原基线说明中的 `npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node`。本次使用已有安装副本，pins 和安装数据由实际差分及启动拒绝检查核验，没有重新下载源码或重跑历史等价重建。

核对本次本地归档及其每个成员（从仓库根目录执行）：

```sh
python3 - <<'PY'
import hashlib, json, pathlib, zipfile
folder = pathlib.Path('.scratch/barness-ai-openai-decisions/jev-baseline/recheck-20261016')
receipt = json.loads((folder / 'receipt.json').read_text())
sha = lambda raw: hashlib.sha256(raw).hexdigest()
archive = folder / 'recheck.zip'
assert sha(archive.read_bytes()) == receipt['archive_sha256']
with zipfile.ZipFile(archive) as bundle:
    raw = bundle.read('files.json')
    assert sha(raw) == receipt['files_manifest_sha256']
    files = json.loads(raw)
    names = bundle.namelist()
    assert len(names) == len(set(names))
    assert set(names) == set(files) | {'files.json'}
    for name, expected in files.items():
        assert sha(bundle.read(name)) == expected, name
print('PASS: current recheck archive and all artifacts verified')
PY
```

本次两轴审查见 [REVIEW.md](REVIEW.md)。这些结果用于确认 Jev 的离线比较仍可重放，不据此新增 Decisions 或发布支持声明。
