# Ticket14 gap2 离线修复证据

代码提交 `1a0ce3458d5c499b95c03e3994f3e7be35e88ea2` 修复两个原生 fixture 的 SHA inventory，并让无配置的可选人工诊断在注册 evidence case 前跳过。消费者、必需 trace、断言、selector、付费预算保持原契约。付费预检现在实际执行六条固定无 Key 命令，包含完整 release package 与既有公开契约/诊断进程测试。

原有完整 release package 首先 RED1；加强既有 fixture 后的15个语义分支先全部在缺 inventory 处 RED，再通过真实 inventory 到达各自语义检查。公开诊断进程四场景先3PASS/1FAIL，再 ordinary/race 各4PASS。公开预检 CLI 先3FAIL，再3PASS，完整工具集61PASS，release package ordinary/race 与两条 vet 均退出0。所有原始日志和逐阶段源码 SHA 保留；早期 Go slice 尚未改 Python 预检，不能把它们说成最终全部源码的运行。

第一次 package race 的360秒外层限时不足，随后既有 stop helper 遇到 killpg PermissionError，原始 STARTING/exit1/日志保留，**不是通过**。另一次相同 Go argv 在明确900秒外层限时下358.550秒退出0，源码不变；不是付费重试。注释记录旧 PID 已消失，复验不执行任何保存的命令。

提交 `572570bbe4a9322fb02ed740bd2a10c9b8ce792f` 合并父分支文档/实际17行矩阵后，真实六条预检全部0。其源集合1048文件与代码提交相同。主无 Key live bundle 为398场景：336PASS/62NOT_RUN，后者是未启用的真实厂商采集；另两个 replay 子入口各1PASS，visual proxy16PASS，契约/诊断10PASS，主 audit 均0。这里不声称没有真实 Key 时62场景已经通过，也不声称修复后的完整普通/race全仓门禁已执行。

授权仍为 PENDING；既有39个闭合收据、当前 release-renewal 已知545/545消费保持不变。此次只检查，未创建新付费轮次/意图、未读取厂商 Key、未访问厂商。更早的282次独立 GLM 任务历史不并入或重置。

复验：

```sh
python3 -B .scratch/barness-ai-minimax/evidence/ticket-14/gap2/verify.py --common-root /Users/cyber/RestoX/barness
python3 -B .scratch/barness-ai-minimax/evidence/ticket-14/gap2/check-boundaries.py --common-root /Users/cyber/RestoX/barness
```

原始归档在 `.evidence/minimax-retained/ticket-14-offline-gap2/gap2-raw.tar.gz`，28,917普通文件、226,583,975 payload字节，压缩41,164,640字节，SHA `1c36af4d252552e9d42792c7e9f05296b86e75fb2a6bcaa6f9ec0cb35b8574ab`。前后逐文件 SHA 相同。一个故意制造的负例 symlink 仅投影为普通 JSON 元数据，保留原始路径/类型/目标字节与 SHA；归档不包含链接，也不跟随或重建链接。三个未执行的付费二进制只保留 hash/大小引用及原始 build/manifest 收据，避免复制无用 blob。

读取器使用已明确建立的 ticket14 固定包络：192MiB压缩、1GiB解码tar、90,000文件/100,000条目、64MiB单成员、24MiB解码inventory/8MiB压缩inventory、8MiB单观察元数据/16MiB观察总量、900MiB payload；实际解码tar278,374,400字节。全部15,761个已观察原生 manifest artifact 引用逐项核对。它不会解压到磁盘、执行保存的代码或自动重跑采集；旧 implementation 的更小包络没有被扩大。
