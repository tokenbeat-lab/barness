# 工单 02：Decisions 文本单选证据

本票增加公共 Decisions API/Options、文本能力目录、RefusalAnswer 及可实际发送和解析的文本单选路径。
宿主目录为合成 `classifier/openai/openai-decisions/host-decisions`，真实 HTTP 由 loopback 受控 Provider 接收；
厂商真实请求为 NOT_RUN，内置目录及九路线支持矩阵没有修改。后续真实调用遵循用户要求使用本地代理。

实施起点 `a8a4686d7b47a48d4514ac1c459f54b95ff2f8e0`，当前分支 main。
起点已有 GLOSSARY 与其他规划改动均保留。失败目录先写在[失败清单](../text-choice-failures.md)，
测试沿用已确认的公共 Classify / HookedClient.Classify 和 Catalog/Lookup/List/Hash/NewClient 接缝。

`text-choice.zip` 固定六次红色运行与最终 653 个全部 PASS 的公共 case；其中新 P10 195 个，
其余为 Jev、目录、回调、unary 及已开启 Pi 的既有回归。红色历史保留真实失败，包括测试装配限额错误，
不把它们改写成 PASS。新增 scope/能力拒绝同时检查 resolver/credential/Hook/Attempt/admission/HTTP 次数。
初始模型 capability 不支持的 score/bool 在凭据前拒绝；Hook 不能开启未实现的 message/image/kind。

| 验收面 | 公共证据 |
| --- | --- |
| OD01/02/03/09/10/12 | 旧目录省略/构造范围保留；四维身份、白名单/目录交集、操作/Provider/API/租户拒绝与零凭据/发送 |
| OD04/05/06 | nil/value/pointer Options，typed nil/错协议；能力/Hash/NewClient 一致、入口与三种查询出口深复制 |
| OD11/14/16 | State 原文或精确 compact，数字键确定次序，字符串 true/false，null/空说明省略，native 仅 model/input/questions |
| OD22/23/25 | Keep/Replace 最终换题、迟变引用隔离、越权/坏 schema 拒绝、初始和最终完整编码题目/JSON 限额 |
| OD27/28/29/30/31/32 | name 回配、typed 单题拒答、完整名称/候选集、精确字段、finite/range/sum/最高项并列；损坏整体清空 |
| R13/14/19 | 2xx 后不重放、单次固定凭据、回调顺序、Observer/Attempt、body/permit 归零及脱敏审计 |

目录新 Input/MinChoices/DescriptionFormats 的省略解释不改写旧 JSON 或内置 Hash。
新文本模型显式至少两个候选且 string 说明。基础已知 input/output/total 与 input-only Cost 保留，
Reporting 保守为 partial；完整 details/presence、固定 272000 阶梯由工单 04 验证交付。
score/bool、图片公共契约、混合并发/故障、Decisions 自身 Pi 差分与 live 分别由 03/05/06/07/09 交付。

压缩包包含每 case 的 Provider 脚本/捕获、结果/错误、Attempt、Observer、assertions、fixture 哈希、准确重放命令，
另附输入/限额、测试源码、变更源码 SHA-256 清单、起点身份、日志与整包审计。
外层 `manifest.json` 固定压缩包及全部 4,929 文件的 SHA-256，`verify.py` 以固定 manifest 哈希拒绝重写；
解压数量/字节/路径/成员类型均有界。审计不放宽原规则。缺包、篡改包、重写 manifest 的公共 process E2E 先红后绿。

核验固定证据（无真实 API 请求）：

```sh
python3 .scratch/barness-ai-openai-decisions/text-choice-evidence/verify.py --audit
go test ./ai/e2e -run '^TestDecisionsEvidence$' -count=1
python3 .scratch/barness-ai-openai-decisions/jev-baseline/verify.py --audit
```

从仓库根目录重放当前源码；指定新目录不会覆盖固定历史，case manifest 还包含逐例准确命令：

```sh
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
BARNESS_AI_PIDIFF=1 go test ./ai/e2e -run '^Test(Decisions|JevRegression|Classifier|Unary|OperationCatalog)' -count=1
BARNESS_AI_PIDIFF=1 go test -race ./ai/e2e -run '^Test(Decisions|JevRegression|Classifier|Unary|OperationCatalog)' -count=1
go vet ./...
go vet -tags live ./...
BARNESS_AI_PIDIFF=1 go test ./... -count=1 -json
```

完整最终测试、race、审查及源码适用性收据在本目录的 validation 文件中单独记录；
固定 Jev 基线核验 655 case，历史 release FAIL 仍原样保留，不是当前 live 状态或本票的新发布结论。
