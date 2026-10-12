# 工单 05：分类内联图片离线证据

独立图片单选已通过公共 Classify / HookedClient.Classify 完成闭环。Images 使用真 PNG/JPEG/WebP
资产，按照输入顺序发为一个 user message 的 text-first data URLs。State 仍必填；空对象允许
图片判断。无图保持 input string，旧 Jev wire 与图像生成回归通过。

三个条件分别开启图片资源、型号能力和租户权限：Classifier.Images 非 nil、型号声明 image
及正的 MaxInputImages/MIME 子集、Binding.AllowClassifierImages=true。默认关闭；GenerateImages
政策独立。入口按已知长度先检查数量、单图/总 decoded 和请求下界，再校验严格编码/header、复制；
Hook 后检查所有 message/content/图片/问题和完整 JSON，权限来自固定 Binding。

入口政策超限为 resource_limit/scope，型号/API 超限为 invalid_request/capability，Binding
拒绝为 tenant_denied/capability；以上均零凭据/Attempt/HTTP。Hook 越权、坏 schema、超资源分别为
tenant_denied/request、callback_failed/request、resource_limit/request。不分配 raster、不修补、不拉取。

| PNG/JPEG/WebP 示例 | decoded 字节 | base64 字节 | data URL 字节 |
| --- | ---: | ---: | ---: |
| red.png | 77 | 104 | 126 |
| blue.jpg | 613 | 820 | 843 |
| gopher.webp | 442 | 592 | 615 |

这次三图调用 decoded 总量 1132、base64 总量 1516、State 与图片已知长度 1597、完整公共 JSON 1769、
最终 native JSON 1952 字节。二进制原资产、MIME/hash、读数和最终捕获均保存，不以 arbitrary base64
冒充资产。WebP 来源与 BSD license 随源码 fixture 保存。

起点 `ffe70054`，实现 `0d823b9`，审查与交叉核验修复至 `926ed8a`，分支 main。
完整失败清单在 [image-failures.md](../image-failures.md)。两轴审查已解决 raw text 比例复制、合法
message discriminator、跨窗口 surrogate 误计、typed 整请求预检和 byte-slice encoder 语义问题，
最终 Standards 0 / Spec 0，见 [REVIEW](REVIEW.md)。

最终生产源码 `926ed8a`：131 个图片公共用例全部通过；专项 race + 原 Classifier/Jev 对照共
431 PASS。最后完整开启 Pi 的全仓测试退出 0，7158 E2E case：7149 PASS、9 既有 opt-in NOT_RUN；
815 项现有 Pi 差分全部 PASS。两种 vet 均退出 0。
全仓库 race 在初始实现 `0d823b9` 退出 0（6312 PASS、826 NOT_RUN，未开启 Pi）；审查改动由最终
431-case race 验证。validation.json 分别记录每次执行的实际编译版本，不将中途版本冒称最终版本。

images.zip 保存九次红色历史、最终专项 431 个完整 case 的 fixture/capture/result/Attempt/Observer/
assertions/gauges、源码/spec 快照、完整测试日志与全套 manifest/audit 汇总。全套其他 case 的完整
原始文件仍在本地运行生成的 .evidence；提交包对这些 case 明确只保留汇总，不伪造缺失捕获。
每个 ZIP 成员的 SHA-256 在 manifest.json，外层 SHA256SUMS 固定压缩包、manifest、validation、README、REVIEW。
整包使用现有 auditbundle 审计，audit.json 无发现。红色记录保留原 verdict，重放不覆盖历史。

本票只声明合成宿主目录和 loopback HTTP 的离线能力。未新增内置型号，未执行厂商 live、
新 Pi Decisions 差分、图片设计负载或发布门禁；它们由 07/08/09/10/13 等后续工单提供独立证据。

核验固定包（仓库根目录）：

```sh
(cd .scratch/barness-ai-openai-decisions/image-evidence && shasum -a 256 -c SHA256SUMS)
unzip -t .scratch/barness-ai-openai-decisions/image-evidence/images.zip
```

在临时目录做整包审计：

```sh
image_bundle_dir="$(mktemp -d)"
unzip -q .scratch/barness-ai-openai-decisions/image-evidence/images.zip -d "$image_bundle_dir"
go run ./ai/internal/testkit/audit/cmd/auditbundle "$image_bundle_dir"
```

准确重放当前源码（产生新包，不改写本包；逐例命令另见 assertions/manifest）：

```sh
go test ./ai/e2e -run '^TestDecisionsImages' -count=1
BARNESS_AI_PIDIFF=1 go test -race ./ai/e2e -run '^Test(DecisionsImages|JevRegression|Classifier)' -count=1 -json
go vet ./...
go vet -tags live ./...
BARNESS_AI_PIDIFF=1 go test ./... -count=1 -json
```
