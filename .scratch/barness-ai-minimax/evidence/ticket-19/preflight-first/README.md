# 首次最终前置检查：5通过、1失败

来源 `0bbd73d`，1074份完整执行来源。六项真实无Key命令全部关闭；[checks.json](checks.json)记录普通vet、live vet、视觉代理、发布契约与release包均exit0，完整live包exit1。所有命令不读凭据、不继承凭据、launcher audit通过；0厂商请求、0付费intent、0真实grant。

失败来自新`TestMiniMaxFormalRenewalControlledRefusal`的证据记录：子进程输出包含证据目录，根工作区位于HOME下，`MM19-renewal-controlled-refusal`原样记录输出与断言诊断，最终脱敏审计拒绝9处本机HOME路径。`/tmp`工作树下此前定向运行未覆盖该路径布局；原结果不改写为通过。见[原完整live输出](no-key-live.log)与[数字退出凭据](no-key-live-command.json)。

仅修正受控测试证据中路径的记录形式，保持命令执行、拒绝断言、审计策略、Provider及授权消费者不变；修复需先复现该失败，再在新提交下重新冻结前置检查。原冻结来源、二进制、计划、六命令及日志按[index.json](index.json)保留，不在此目录覆盖后续测量。
