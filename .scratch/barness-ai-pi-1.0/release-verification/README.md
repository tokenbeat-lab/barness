# 工单 02：pi-ai 1.0.0 发布件等价核验

2026-10-08 核验通过：冻结源码加入发布件模型数据后，ai 的 **811/811** 个
`dist/` 文件与 telemetry 的 **24/24** 个 `dist/` 文件逐字节一致，合计 **835** 个文件，
差异 **0**。核验范围包括隐藏文件、JS、声明、source map 与模型数据，
不比较 npm 包的 README 或 package.json 排版，不执行上游测试。
另一个全新临时环境已再次安装和重建通过，来源、模型、两包逐文件哈希及失败场景结果
与提交证据一致；可按下列命令重放。

这份材料是发布件等价证明；barness 在 1.0.0 下的差分为 **NOT_RUN**。
本工单未改动 `ai/internal/testkit/pioracle/node`，唯一运行中的 oracle 仍为 **0.87.1**。
工单 03 应依据本证明替换其 lockfile、provenance 和账本，然后完成新基线差分。
这里没有模型调用入口或第二套 runner，核验安装与源码仅存在于运行后删除的临时目录。

## 固定来源与完整性

| 项目 | 固定值 |
| --- | --- |
| 源码 | [pi tag v1.0.0](https://github.com/earendil-works/pi/tree/v1.0.0)，[commit a13d35a742c6ef8462812a28fbe1d8c8b7431c32](https://github.com/earendil-works/pi/tree/a13d35a742c6ef8462812a28fbe1d8c8b7431c32) |
| ai npm gitHead | `a13d35a742c6ef8462812a28fbe1d8c8b7431c32` |
| telemetry npm gitHead | `a13d35a742c6ef8462812a28fbe1d8c8b7431c32` |
| ai integrity | `sha512-3/W1vdDaVtpeMd23ElvJC12HLA5yS/BGqqcXF+0SK082dN7cbgNcCwguTBRBC258Ke8SzSvUW1B75iAf8w8IxA==` |
| telemetry integrity | `sha512-WjNBj5TYIiPZFQEz2WlULcDwPLaKwIlmsKjVeYM+LJSbnSp38kWsHUJysIDGnu23IcLbKoPtywsvYfUJCeZePA==` |
| 发布件安装 lock SHA-256 | `d7cf8bf81f0292b7980d4b101eccf190d3513e9214e281824196fa2c7eca413a` |
| 冻结源码 lock SHA-256 | `d5ca36c6b5af93e5d5ef79deedae8efff575adf6ae080bd2093ff29bb7c5664b` |
| 模型数据完整 SHA-256 | `8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e` |
| 模型数据 | 43 文件，含 `.manifest.json`；schema 6，generatedAt `2026-10-01T18:57:11.882Z` |
| structureHash | `235f2f320916ab6b0d7193e0bf66ec7983e9bc05abeddd7264923fb1e7eaf76e` |
| 本次工具链 | Node.js `v24.15.0`、npm `12.0.2`、TypeScript `7.0.2`、Go `go1.26.2`，darwin/arm64 |

tag、目标 commit 与两个实际 npm gitHead **完全相同**，无需像 0.87.1 那次
解释 CHANGELOG 差异。核验同时拒绝源码工作树的未提交修改。
元数据取自官方 [ai registry](https://registry.npmjs.org/@earendil-works/pi-ai/1.0.0)
与 [telemetry registry](https://registry.npmjs.org/@earendil-works/pi-telemetry/1.0.0)。
下载的两个 tarball 都重新计算 SHA-512；`npm ci` 另按已固定 lockfile 校验安装包。

发布件安装锁定 85 个外部包；ai 与 telemetry 都精确固定 `1.0.0`，
telemetry 的上游 `^1.0.0` 由 override 限定。
源码复用目标 commit 中原有 lockfile，包含 385 项外部包记录；不生成或改写它。
逐项核对安装版本；仅允许 lock 明确标为 optional / devOptional / extraneous 的记录未安装。
平台选择与跳过脚本可能影响 optional 包的安装，实际情况逐项写入清单。
源码 lock 原有 4 项 extraneous 旧嵌套依赖不参与安装或构建，保留它们的精确版本和完整性记录，
不修改上游锁。它们的影响限于清单条目，源码与产物的判断未放宽。

模型数据仅从发布包的 `dist/providers/data/` 复制到冻结源码的
`packages/ai/src/providers/data/`；没有调用 `generate-models` 或 `hydrate:model-data`。
完整模型哈希沿用 ADR-0004 定义：文件名排序，对每个文件形成
`<sha256(file)>  <name>\n`，含 `.manifest.json`，再对全部行计算 SHA-256。
[逐文件输入清单](model-data.sha256) 固定全部 43 个文件；运行时对名称集合和字节哈希均核对。

## 重放

需要 Git、Node.js ≥22.19、npm、项目要求的 Go 工具链，以及 GitHub / npm registry 网络。
从本仓库的干净 checkout 根目录运行，输出目录须尚不存在：

```sh
node .scratch/barness-ai-pi-1.0/release-verification/run.mjs \
  .evidence/barness-ai-pi-1.0/02-clean-replay
```

脚本从已提交的 [inputs.json](inputs.json)、[package.json](package.json) 与
[package-lock.json](package-lock.json) 读取固定输入。
安装 / 构建子进程采用环境白名单、空 npm / Git 用户配置和新的 npm 缓存。
先验证 tarball，再执行两次 `npm ci --ignore-scripts --no-audit --no-fund`，
并核对 tag commit、lock 哈希与实际安装版本。
按 ADR-0004 顺序构建 chord、tui、telemetry，再执行 ai 的 `build:offline`
（包含上游 `check:model-data` 与 TypeScript 编译）。npm 在构建时设为 offline；
构建没有模型目录网络刷新步骤。

随后 [check.mjs](check.mjs) 比较完整文件集合与文件字节，
[failures.mjs](failures.mjs) 在真实安装 / 重建材料上逐一篡改并恢复输入，
验证错版本、gitHead 不一致、缺数据、数据变化、两包构建差异、未锁依赖、tarball 损坏及审计秘密。
这组失败清单在实现前写入 [FAILURES.md](FAILURES.md)，检查器不存在时先得到失败产物缺失的红灯。
9 个场景必须全部拒绝；任一失败导致非零退出与 FAIL 报告。
另由 [faults.mjs](faults.mjs) 通过 CLI 对网络与文件系统施加故障，
验证下载超过 20 MB 时在读取中取消响应，以及证据写入失败时仍清理临时目录。
这两项来自代码审查，先补充失败清单与验收场景，再修正下载和清理边界。

脚本调用仓库已有审计规则的
`go run ./ai/internal/testkit/audit/cmd/auditbundle <bundle>`，
审计实际输出中的凭据、主目录、主机名、JSON 和请求捕获。
失败报告只记录阶段，不回显原始错误或秘密。全部日志脱敏后写入，成功 / 失败都会审计，
临时工作树与安装副本随后移除。已有输出目录拒绝覆盖，旧 PASS 不能冒充新运行的结果。

## 已提交证据与工单 03 交接

| 材料 | 内容 |
| --- | --- |
| [report.json](evidence/report.json) | 来源、完整模型哈希、两包范围 / 文件数 / 差异数，1.0 差分 NOT_RUN |
| [packages.json](evidence/packages.json) | 实际 registry 版本、gitHead、tarball integrity |
| [dependencies.json](evidence/dependencies.json) | 两个 lock 哈希、精确版本 / URL / integrity 与安装情况 |
| [dist-ai.sha256](evidence/dist-ai.sha256)、[dist-telemetry.sha256](evidence/dist-telemetry.sha256) | 每行分别为发布文件哈希、重建文件哈希、相对文件名 |
| [failure-checks.json](evidence/failure-checks.json) | 9 个篡改拒绝场景 PASS |
| [boundary-checks.json](evidence/boundary-checks.json) | 下载超限与证据写入失败的 2 个边界场景 PASS |
| [commands.json](evidence/commands.json)、[commands.log](evidence/commands.log) | 实际重放步骤、退出状态与脱敏构建输出 |
| [audit.json](evidence/audit.json) | 脱敏审计 `findings: []` |
| [inputs.sha256](evidence/inputs.sha256) | 相对仓库根的脚本、配置与审计命令源文件完整性 |
| [SHA256SUMS](evidence/SHA256SUMS) | 相对证据目录的全部交付材料完整性（不自包含） |

核对脚本输入（从仓库根）与证据（从证据目录）：

```sh
shasum -a 256 -c .scratch/barness-ai-pi-1.0/release-verification/evidence/inputs.sha256
cd .scratch/barness-ai-pi-1.0/release-verification/evidence
shasum -a 256 -c SHA256SUMS
```

工单 03 可取 `inputs.json` 中的版本、commit、两个 npm gitHead、integrity
与模型数据完整哈希来更新唯一 oracle 的 provenance，并用本工单的锁定依赖作为迁移输入。
这些值尚未启用为运行时 provenance；切换仍须完成聊天修正、目录比对与全量差分门禁。
完整控制验证及双轴审查结论见 [verification.md](verification.md)。
