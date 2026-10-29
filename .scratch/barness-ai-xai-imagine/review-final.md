# xAI图片正式资格与有限负载审查

本轮范围见工单07与固定计划；实际阶段事实只维护于spec唯一当前状态。各轴独立审查，初审问题及修复复核保持。压力初审Standards1项、Spec2项及其零用量残留由同一实现者修复；正式消费者两轴同一资源P2去重1项修复；历史原件适配两轴0项。最终全部0遗留项。真实阶段不由静态审查替代。

## 压力 Standards 初审

已完成：固定差异的 Standards 静态审查。发现 1 项硬要求缺口，未发现需单列的代码异味。

范围：`990d00aa...c6f0c42ce302623feec164e24d68a93faaccf531`。提交：`c6f0c42c test(ai): add bounded Imagine image pressure scenarios behind explicit opt-in`。

标准：仓库 `AGENTS.md`、`docs/agents/verification.md`、ADR-0055/0056。祖先 `/Users/cyber/.codex/AGENTS.md` 为空。未发现其他适用 AGENTS、CONTRIBUTING 或 CODING_STANDARDS。

**硬要求问题**

- [P2] 准入场景缺少独立的租户身份依据。位置：`ai/e2e/xai_imagine_pressure_queue_test.go:101–117`。Observer 只与结果元数据互相比对。HTTP 期望租户又来自实际 `KeyAlias`。路由也由同一凭据生成。两租户的回复字节相同。因此，B 等待调用与 A 恢复调用交换凭据时，这些断言仍能通过。依据：AGENTS 原则 4、9；测试 SOP §1 要求隔离验证。必须给每个调用固定身份，并按唯一 Prompt/RequestID 核对预期租户、账户、密钥和 Observer。

**启发式**

未发现需报告的 baseline smell。有限分支及现有 helper 复用符合 AGENTS 原则 1、7、11；不建议新增负载控制器。

**验证**

已验证：只读检查确认八项均先经过 `pressureCase`。loopback transport 拒绝外网。负载在生成脚本后建立 heap baseline。取消与释放在等待 goroutine 前执行。

已验证：门控回执 `20261029T025425.075012Z-02xhsy9o` 为 exit0、2 PASS、8 NOT_RUN、audit 空。三个新增文件与回执 source 哈希一致。

未验证：实际压力、race 和运行时资源峰值。本审查未运行测试或网络请求。正式资格及 race 通过后，才能执行选定压力。


## 压力 Spec 初审

## Spec

审查发现 2 项压力验收缺口。范围为指定基线后的三个新增 `xai_imagine_pressure*_test.go`。本次只读文件，没有运行测试、压力、race 或厂商请求。

1. **[P2] 字节边界原件缺少负载哈希。** 固定负载报告要求「记录实际边界与 hash，不复制无界原件」。`ai/e2e/xai_imagine_pressure_boundaries_test.go:80–110` 只记录空结果摘要与 actualBytes/limitBytes；失败时 Content 为空，因此摘要的 imageHashes 也为空。五项没有记录实际输入图片、待发送请求或响应脚本的 SHA256。必须保存相应有界输入/请求/响应的实际长度和哈希。`input-total` 的 actualBytes 直接取 `limit+1`（:89–90），必须独立记录构造请求的长度/哈希，便于核对一字节边界。

2. **[P2] Observer 验收未覆盖完整归属和失败路径的内容隔离。** 固定报告要求「租户/凭据/Observer 归属」及队列「最终 Observer 无内容泄漏」。`ai/e2e/xai_imagine_pressure_test.go:163–176` 只核对租户/操作与部分消费字段，没有核对 account/binding/credential/catalog 归属。`xai_imagine_pressure_queue_test.go:101–107` 对取消/恢复只核对租户；overflow 在 :87 仅等待终态，未检查内容隔离。必须比较各阶段已解析的归属与结果，覆盖全部六调用的终态和允许元数据。保留成功、取消及拒绝的实际 Observer 原件。

其余固定范围与代码一致：plain/hooked 各两租户四并发；五张 256 KiB 输入、十张 512 KiB 候选输出；EOF 四许可/正文屏障；顺序哈希、独立金额与缺失 token；512 MiB heap-growth 上限及 TotalAlloc/headroom；六调用的有界队列、取消/恢复与资源归零。Builtin 未扩大。

压力门控位于每项负载入口之前。普通 2 PASS + 8 NOT_RUN 只证明普通门控，不能声明压力通过。

Spec：2 项，最高 P2；均为约定验收和证据缺口，没有据此认定产品运行失败。


## 压力 Standards 复核

已完成：Standards 静态复核。原 1 项 P2 问题已完成修复。未发现新硬要求问题。

范围：`c6f0c42c..9f0364a7b2d804b1078798a9ccf771a500acad16` 的三个压力测试文件。标准沿用仓库 AGENTS 原则 4、5、7、9、11、测试 SOP 及 ADR-0055/0056。

**原问题**

已验证：只读检查确认独立身份依据成立。

- `ai/e2e/xai_imagine_pressure_queue_test.go:36–68` 固定六个调用。每个调用有唯一 RequestID 和 Prompt。两租户使用不同响应字节及预期 hash。
- `ai/e2e/xai_imagine_pressure_queue_test.go:130–175` 按固定调用核对尝试数、输出 hash、密钥及路径。HTTP 期望值来自计划调用，不再来自收到的 KeyAlias。
- `ai/e2e/xai_imagine_pressure_test.go:168–238` 独立构造完整归属。Observer 核对未解析入口、完整生命周期、尝试归属和终态事实。

原有凭据交换盲点已消除。B 等待调用与 A 恢复调用交换凭据时，HTTP 及输出 hash 断言会失败。该结论来自源码推演，不是运行验证。

**新问题**

硬要求问题：0 项。

启发式问题：0 项。新增 helper 封装已有的第二个实际用途。有限场景未增加负载控制器。

`ai/e2e/xai_imagine_pressure_boundaries_test.go:73–88,150–156` 区分构造请求与实际捕获。边界证据保留字节数、顺序和 hash。

**验证**

已验证：读取普通回执 `20261029T032112.543409Z-6on615r1`。结果为 exit0、2 PASS、8 NOT_RUN、audit 空。manifest、audit 及 case artifact 哈希匹配。三个文件与回执 source 哈希一致。回执的 sourceCommit 为 `c6f0c42c`，包含提交前的修复字节。

未验证：实际压力、race、运行时峰值及新断言的运行效果。本复核未运行测试或网络请求，未修改仓库或执行 Git 操作。普通门控回执不证明压力通过。正式资格及 race 通过后，才能执行选定压力。


## 压力 Spec 首次复核

## Spec 复核

部分完成：原 2 项问题的主体已修复。仍有 1 项 P2 验收缺口。范围为 `c6f0c42c..9f0364a7b2d804b1078798a9ccf771a500acad16` 的三个压力测试文件。本次只读代码和既有回执，没有运行测试、压力、race 或厂商请求，没有执行 Git。

1. **已修复：字节边界缺少原件哈希。** `xai_imagine_pressure_boundaries_test.go:15–38,73–105,126–156` 独立记录输入/输出图片的解码长度和 SHA256、构造请求、响应脚本及实际 HTTP 请求摘要。`input-total` 从构造字节取得 actual，再将策略设为 actual−1。输入拒绝时明确没有已发送请求。原有五项一字节边界未扩大。

2. **已修复：Observer 归属及内容隔离覆盖不全。** `xai_imagine_pressure_test.go:166–238` 从预设调用身份构造期望值，区分入口未解析和终态已解析阶段。它核对 RequestID、Tenant/account/binding/credential/catalog、完整生命周期、终态副本及禁止内容。队列在 `xai_imagine_pressure_queue_test.go:130–155` 覆盖全部六个独立调用；边界在 `xai_imagine_pressure_boundaries_test.go:155` 也复用检查。

3. **[P2] 零尝试失败的 token 仍只有同源比较。** 固定负载要求独立用量及失败终态隔离。队列 `:136–147` 把零用量检查放在 attempts 循环内，因此 waiter 取消和 overflow 没有执行该检查。输入边界 `:140–141` 也未检查 result.Usage。公共 Observer 检查 `xai_imagine_pressure_test.go:219–220` 仅比较 Observer.Usage 与 result.Usage；二者若同时带入非零 token，仍可通过。请对这些零尝试失败独立断言 `result.Usage == (ai.Usage{})`，再保留现有 Observer 一致性检查。此项是验收缺口，不据此认定产品故障。

已核对既有普通回执 `032112.543409Z-6on615r1`：PASS、exit 0、2 PASS 与 8 NOT_RUN，audit 无发现，三个当前文件哈希一致。该回执只证明普通门控。压力和 race 结果仍未验证。


## 压力 Spec 最终复核

## Spec 最终复核

已完成：压力残留 P2 的只读复核。行动问题为 0 项。

范围：`9f0364a7..9f8fcf60ac8eea653e76f13115370e2c1754337f`。差异仅为 `ai/e2e/xai_imagine_pressure_test.go` 的 8 行新增。

原要求是独立核对零尝试失败的用量。`:198–202` 对零尝试结果独立比较 `ai.Usage{}`。`:226–229` 对终态 Observer 独立比较同一零值，并核对没有 Attempt。检查覆盖队列取消、overflow 与输入越界。它没有用 Observer 和结果相等替代独立期望。

零值只是没有消费事实的载体。实现没有给无尝试调用补造 UsageReporting。已有真实 Attempt 的 Unreported 检查保持。范围、负载、目录及正式资格合同保持。

已验证：静态差异符合本项要求。未验证：新增压力断言的运行结果。本次没有运行测试、压力、race 或厂商请求。


## 正式消费者 Standards 初审

已完成：正式消费者的 Standards 静态审查。发现 1 项硬要求问题。未发现需单列的代码异味。

范围：`git diff 990d00aa...b24e81ba`，19 个文件。唯一提交：`b24e81ba feat(ai): qualify retained xAI Images captures offline`。标准：AGENTS 原则 3、4、7、9、11，测试 SOP §3–4，ADR-0055/0056。

**硬要求问题**

- [P2] 内层图片资源缺少严格字段校验。位置：`ai/internal/testkit/supportmatrix/xai_images_capture.go:149–157,167–199`。消费者将 `image/images` 解码为只有 `type/url` 的结构。Go 解码会忽略额外字段，并接受大小写别名。外层键集合校验不覆盖该层。因此，保留原 URL 和图片字节，同时加入 `file_id` 或将键改为 `Type/URL`，仍满足当前语义检查。同步更新 manifest/producer 哈希也不会改变图片 hash。消费者会接受偏离固定原生契约的请求。依据：AGENTS 原则 3、4；正式缺口报告「必须独立核验的合同」第 6 条。必须在转换前核对每个资源的精确 `type/url` 键集合。必须增加对应公共 CLI 拒绝场景。

该结论来自源码推演。未执行篡改验证。

**启发式**

0 项。图片原生规则独立封装。通用来源、目录及 ledger 规则未放宽。新增 `ClaimEvidenceInputs` 复用可信 Git 存储及双 key 核验。

静态检查确认：通用 Merge 调用图片原生验证。releasegate 的 generic 路径也经图片 manifest 验证。原包只读；投影保留原测量身份与时间。

**验证**

已验证：读取普通回执 `20261029T032448.975974Z-woi6rmcn`。exit0，88.763 秒。主 bundle 为 147 PASS，audit 空。回执来源为 `b24e81ba`；当前来源文件哈希无差异。该回执覆盖 25 项消费者场景及相邻回归。

未验证：上述内层资源负例和真实原件正式资格。本审查未运行测试、网络、race 或压力，未修改仓库或执行 Git 变更操作。普通受控结果不表示真实正式资格通过。


## 正式消费者 Spec 初审

## Spec

部分完成：正式消费者只读审查发现 1 项 P2。范围为 `990d00aa...b24e81ba9ee000aa2a23e89aff01b0303ba79ebf`，19 个文件。本次没有运行测试、厂商请求、正式资格、race 或压力，没有合并正式消费者代码。

1. **[P2] 嵌入请求的资源对象仍接受额外字段和字段别名。** Spec §3 要求「不允许……file_id……其他资源生命周期」，并要求最终字节「严格解析，重做 schema、重复字段、null」。`ai/internal/testkit/supportmatrix/xai_images_capture.go:149–158` 把 requestBody 解为仅含 Type/URL 的资源 DTO，但 regionalDecode 最终使用普通 json.Unmarshal。资源对象的未知字段被忽略，TYPE/URL 别名也能填入同一字段。`:174–199` 只核对解码后的 Type/URL。因 requestBody 是外层 JSON 字符串，外层 xaiReadJSON 和全包 null 检查没有深入它。保留正确内联图片，再在 image 或 images 成员加入 `file_id`（包括 null）或非规范 TYPE/URL，当前代码仍能接受。请在解码前逐资源校验精确 type/url 字段集合、规范拼写与非 null，并增加公共消费者拒绝场景。此项是静态确认的校验缺口，未运行篡改场景。

其余核对符合固定范围：原六场景、七图及 ticks 使用独立冻结事实；token 保持 Unreported；来源/producer/授权及双 key claim 纳入 57 项输入。原 candidate 的 stage/mode/time 保持，消费者只返回内存投影。当前 Builtin 精确限制为 5 参考、2 输出、三 size、三 quality、两 ratio。通用 Merge 必须重新读取原生捕获；单独序列化 PASS 不能取得资格。通用来源与目录比较没有放宽。没有发现扩大其他 Provider 资格的差异。

已核对：measured-source 的 545 个哈希条目与原件全部相等，规范哈希为原 7a77…4113。既有普通 `032448.975974Z-woi6rmcn` 为 PASS/exit0、88.763 秒、147 PASS case，audit 无发现，新增产品文件哈希一致。该回执不是实际原件的正式资格结论；后者仍未验证。


## 正式消费者 Standards 复核

已完成：Standards 静态复核。原 P2 已完成修复。新硬要求问题 0 项，启发式问题 0 项。

正式修复范围：`e462e797..d9359d4864408bfac0f536af56a6ffb5fbc4e117`，3 个文件，76 行新增、1 行删除。提交：`d9359d48 fix(ai): reject noncanonical Imagine reference objects`。

标准沿用 AGENTS 原则 3、4、7、9，测试 SOP §3–4、ADR-0055/0056。

**原 P2**

已验证：只读检查 `ai/internal/testkit/supportmatrix/xai_images_capture.go:57–68`。

- 额外 `file_id`：精确两键校验拒绝该字段。其值为 null 时同样拒绝。
- `TYPE/URL` 别名：精确查找小写 `type/url`，拒绝别名。
- 重复键：`regionalDecode` 在转换前检查全部重复字段。
- null 或错误类型：必需值校验及独立 DTO 转换拒绝这些输入。

单图和多图均调用同一严格解码器。已有 URL、图片字节及顺序校验继续执行。修复没有放宽通用来源、授权、目录或 ledger 规则。

`ai/e2e/xai_images_qualification_test.go:12–23` 注册 12 项公共 CLI 负例。变更覆盖原 P2 的 8 种单图/多图输入，并保留 4 项重复 URL/null URL 守卫。

**压力增量**

已验证：静态检查 `9f0364a7..9f8fcf60` 的 8 行断言。当前 `ai/e2e/xai_imagine_pressure_test.go:198–202,226–228` 独立要求零尝试结果和终态 Observer 的 Usage 为零值。该断言没有把零值写成厂商报告。原有固定身份、完整生命周期和凭据归属校验保持。

**验证**

已验证：读取普通回执 `20261029T033326.218368Z-wsnl6hno`。exit0，117.549 秒。主 bundle 为 183 PASS，audit 空，含全部 37 项消费者断言及相邻回归。manifest/audit 哈希匹配。实际来源为 `d9359d48`；当前来源文件哈希无差异。

未验证：真实原件正式资格、race、压力及压力增量的运行效果。本复核未运行测试或网络，未修改仓库或执行 Git 变更。受控普通验证不表示真实资格通过。


## 正式消费者 Spec 复核

## Spec 复核

已完成：原资源 P2 的只读复核。行动问题为 0 项。原 P2 已修复。

范围：`e462e797..d9359d4864408bfac0f536af56a6ffb5fbc4e117`，3 文件、76 行新增。本次没有运行测试、压力、race、厂商请求或真实原件资格，没有修改代码或合并分支。

原要求：Spec §3 禁止 file_id 与其他资源生命周期。最终字节必须重验 schema、重复字段和 null。

`ai/internal/testkit/supportmatrix/xai_images_capture.go:57–68` 现在由资源自己的 UnmarshalJSON 检查原始对象。它先要求恰好两个规范键 type/url，并拒绝缺失、null 和重复字段。随后使用无递归的私有类型转换。最终请求的单 image 指针与 images 数组均进入该校验。原有 type 值、内联 URI、MIME、图片哈希与顺序检查保持。

`ai/e2e/xai_images_qualification_mutations_test.go:108–155` 修改嵌入的 requestBody，并刷新外层 manifest/producer 哈希。该安排防止拒绝仅来自旧外层哈希。`xai_images_qualification_test.go:13–24` 对单资源和多资源各覆盖 file_id、null file_id、TYPE、URL、重复 url、null url，共 12 项。多资源包含第二项的字段变异，覆盖数组后续元素。

已验证：静态差异符合原要求，没有发现新增遗漏、错误或范围扩大。修复只改变本路线的只读资格校验与受控过程场景。原 candidate、Builtin、通用授权和来源规则保持。

已核对既有普通回执 `033326.218368Z-wsnl6hno`：PASS/exit0、117.549 秒、183 PASS case，audit 无发现。12 项资源拒绝断言全部 PASS。三个差异文件、manifest 和 audit 哈希均匹配。该回执证明受控普通过程；真实原件正式资格仍未验证。


## 历史原件增量 Standards

已完成：历史输入增量的 Standards 静态复核。硬要求问题 0 项，启发式问题 0 项。

范围：`285f4e752ef849894e036023ea95eba7683672ad...a857e6b997606741194fba63c3213863984d0d84`。提交：`a857e6b9 fix(ai): retain historical Imagine producer path and document evidence`。标准：AGENTS 原则 3、4、5、8、9，测试 SOP §3–4，ADR-0055/0056。

**边界**

已验证：只读检查确认范围未扩大。

- `xai_images_producer.go:49–52` 先固定 bundle 的绝对路径。既有仓库内路径、逐层 symlink 拒绝及二进制哈希校验继续执行。
- `xai_images_producer.go:143–157` 只为原执行库存中缺少 producer/selected 记录的 Markdown 查询 Git。原库存仍绑定固定 545 项及 canonical hash。已有错 hash、缺失 Go/JSON 继续拒绝。
- `xai_images_history.go:16–35` 固定原 40 位提交，校验 canonical 仓库路径和 SHA。读取 Git blob 使用 10 秒上下文及 1 MiB 上限。超时、缺失、超限、错字节均拒绝。
- 文档哈希进入 `@xai-original-source/<commit>/<path>`。消费者没有读取当前 README，也没有写回原 producer。通用 record、授权、来源及目录规则保持。

已验证：独立只读 Git blob 核对匹配。原 `kimi/README.md` 为 1078 字节，原 `regional/README.md` 为 5172 字节。两项 SHA 均与固定执行库存相等。

该增量遵循 SOP 的原版本输入核验规则。它没有恢复旧执行权限。

**验证**

已验证：读取普通回执 `20261029T035230.406738Z-hqziivbd`。exit0，139.663 秒。主 bundle 为 215 PASS，audit 空。路径 5 项、文档 5 项、37 项原生消费者及选定相邻回归均有 PASS 原件。manifest/audit 哈希匹配。2105 项来源与当前 `a857e6b9` 无差异。

未验证：本轮真实正式 stage、race、压力。本复核未运行测试或网络，未修改仓库或执行 Git 变更。此前绝对路径 inspect 的成功仅表示诊断通过。首次真实消费的失败保持，不以诊断替代正式回执。


## 历史原件增量 Spec

## Spec 增量复核

已完成：`285f4e752...a857e6b997606741194fba63c3213863984d0d84` 的只读 Spec 复核。行动问题为 0 项。没有发现合同遗漏、错误或范围扩大。

SOP §3 明确允许「live 原输入可按原测量提交核验留存字节」，并要求找不到原字节即拒绝沿用。ADR-0055 要求原测量的来源、输入与原件保持不可变。

`ai/internal/testkit/supportmatrix/xai_images_producer.go:49–52` 把 bundle 路径转为绝对路径。该修改修复后续绝对 root 与相对 attempt 的 filepath.Rel 错误。原有仓库内路径、普通文件与二进制哈希校验保持。

`:143–158` 只对原来源集合中、同时缺于 producer source/selectedInputs 的 Markdown 启用 Git 核验。已有错误哈希不会走此分支。Go、JSON 和未提交授权不获豁免。

`xai_images_history.go:16–37` 固定原测量提交，检查规范仓库路径和 SHA256。读取上限为 1 MiB，时限为 10 秒。原 blob 必须匹配原来源哈希。`xai_images_producer.go:153` 把原提交与两文档哈希加入传递输入。代码没有读取当前 README 代替旧字节，也没有改写原 grant、stage、mode、报告或 producer。

已验证：独立读取原提交的两份 Git blob。1078/5172 字节均匹配原 545 项来源中的哈希。受控场景保留绝对/相对路径正例、外部/符号链接负例、原 Git 缺失/错误及运行代码/JSON 缺失负例。

已核对既有普通 `035230.406738Z-hqziivbd`：PASS/exit0、139.663 秒、215 PASS case，audit 无发现。当前 2105 项来源与冻结文件及哈希完全相等。manifest/audit 哈希匹配。

未验证：真实原件的新正式 consumer 结论。本次没有运行测试、网络、资格、race 或压力，没有合并代码。绝对 inspect 的诊断结论不代替正式 consumer。



## 压力首轮拒绝码增量：barness-xai-images-pressure-admission-standards.md

已完成：单行压力断言的 Standards 复核。硬要求问题 0 项，启发式问题 0 项。

范围：`aa0d189f...4c89b6e475f2b431f494fc2b1c72c9af5b6fe5c4`。净差异仅为 `ai/e2e/xai_imagine_pressure_queue_test.go:115` 的错误码期望。

**结论**

已验证：只读检查确认修复符合既有合同。

`ai/admission.go:150–153` 在等待队列满时返回 `CodeAdmissionDenied`。`ai/e2e/admission_test.go:113–133` 的公共准入场景核对同一合同。压力测试现使用相同错误码。

零 HTTP 尝试、空输出、调用身份、Observer、取消和资源释放断言保持。产品代码及设计负载没有变化。依据：AGENTS 原则 5、7、9；测试 SOP §1–3、§5–6；ADR-0055。

**普通证据**

已验证：读取回执 `20261029T040732.976339Z-izmt500t`。exit0，28 PASS、8 NOT_RUN，audit 空。28 PASS 包含 26 项公共准入基线和 2 项 xAI 基线。

实际受测 sourceCommit 为 `ea2e5820`，包含提交前的修复字节。2105 项来源哈希与最终当前文件一致。manifest/audit 哈希匹配。原压力 `20261029T040433.052044Z-k9ltuj9p` 的 exit1 保持。

**后续必要范围**

此门控与公共准入基线可以作为本增量的冻结普通证据。必须同时引用此前完整普通、消费者和产品回执，并记录旧来源仅差此断言一行的适用性依据。

正式资格可以只读重消费原件。随后按既定顺序，使用原 race/pressure argv 冻结同一当前来源。保持原有限 case、负载、预算及超时。没有新增厂商事实需求，无需厂商请求或扩大负载，也无需重跑完整 E2E。

未验证：修复后的实际压力及后续正式/race 回执。本复核未运行测试、压力、race 或网络，未修改仓库或执行 Git 变更。当前普通的 8 NOT_RUN 不表示压力通过。


## 压力首轮拒绝码增量：barness-xai-images-pressure-admission-spec.md

## Spec 增量复核

已完成：`aa0d189f...4c89b6e475f2b431f494fc2b1c72c9af5b6fe5c4` 的只读 Spec 复核。行动问题为 0 项。

差异只有 `ai/e2e/xai_imagine_pressure_queue_test.go:115` 的错误码期望。CodeResourceLimit 改为 CodeAdmissionDenied。

ADR-0008:28 规定：等待数达到 MaxAdmissionWaiters 时，新尝试立即以 admission_denied 拒绝。`ai/admission.go:151–153` 实现同一公共合同。既有 `ai/e2e/admission_test.go:114–130` 验收该语义。本功能要求复用既有 Admission，因此该修复纠正测试期望，不改变产品语义。原有限负载范围中的 resource_limit 是旧测试期望错误，不能据此修改公共准入合同。

同一断言仍要求无 Attempt、无 Content。六个逻辑调用、四个 HTTP 尝试、租户/账户/凭据/目录身份、独立零用量、Observer、取消和最终资源释放检查全部保持。没有扩大负载，也没有把拒绝改成成功。

已核对既有普通回执 `040732.976339Z-izmt500t`：PASS/exit0、28 PASS、8 NOT_RUN。audit 无发现。修复文件、manifest 和 audit 哈希匹配。回执记录测量时 HEAD 为 ea2e5820，实际文件哈希对应最终修复字节。该回执只证明普通门控及相邻准入检查。

未验证：修复后的压力运行结果。本次没有运行测试、压力、race、正式资格或厂商请求，没有修改代码或合并分支。旧压力 FAIL 保持，其他 9 项 PASS 不替代整轮通过。

