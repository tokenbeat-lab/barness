# 04：Decisions 正式真实入口拒绝未配置代理的直连

Status: resolved

## Failure modes before implementation

- 未设置可选 BARNESS_AI_LIVE_PROXY 时，被当作非法代理，在任何 HTTP 前拒绝正式和候选入口。
- 只修 allow_unknown 分支，导致带 Host 权限记录的直连仍失败。
- 空代理回退到 ambient HTTP(S)_PROXY，破坏显式 Host 传输边界。
- 修复误放行非 loopback、坏端口、带身份或非 HTTP 的显式代理。
- 受控 PASS 冒充真实支持；修复源码后继续使用旧当前源码证据；覆盖首轮真实失败或低报预算。

## Acceptance / seam

- 使用已经验收的独立公开进程 E2E seam，先扩展 candidate 与 formal process 的直连场景并记录 RED，后修实现。
- 候选带 Host 记录、候选 allow_unknown、正式 allow_unknown 均可无显式代理完成同一有界受控流水线。
- 显式 loopback 及已有非法代理拒绝回归保持通过，账户权限 unknown 不变为 verified。
- 无凭据 live harness 完整回归、实际 Decisions 重采集与最终完整 normal/race 门禁通过后才关闭。
- Go 内容变化后按本轮授权唯一恢复轮最多302 Calls/Attempts 重采集11支持及4效果；初轮失败原包保留。

## Diagnosis

2026-10-17 首轮 `.evidence/ticket14-remaining-20261017073529Z`：
`openai-decisions` CLI退出1，9场景 config FAIL，Calls/Attempts均0，audit零发现。
`configureDecisionsProxy` 无条件调用只接受非空显式 loopback 的 `smokeProxy`。
其他两个调用方先判断非空；已有候选/正式进程夹具总设置合法 loopback，掩盖了直连分支。
这不是服务端权限拒绝，不需要新增厂商授权；固定 endpoint、TLS 和型号不变。

2026-10-17：三个直连进程场景RED后，空代理只走直连，非空仍原loopback校验。
同一进程seam的8项直连、loopback、unknown及remote-proxy回归GREEN，完整源码提交
`9a8ea45`。收费恢复前完整live harness为206 PASS、62 NOT_RUN、1 FAIL：
仅`UC13-fifo-report`在2秒子进程上限触发deadline，原包保持FAIL。
未改变源码、时限或验收，独立同例PASS；新CLI先`-help`完成约0.524s，再对真实FIFO
以约0.0114s返回`nonregular or oversized report`、退出1。FIFO已按测试清理，模式与
argv/时长/输出留存。未捕获首个超时进程的内部栈，不把启动调度延迟认作已证明根因。
故执行一次同源码完整无凭据复核，使用新intent/目录，原失败不覆盖；只有完整复核通过
才能启用收费恢复。所有材料在`.evidence/ticket14-decisions-direct-20261017/`及恢复根。


## Answer

2026-10-17：`9a8ea45`仅在非空代理时运行既有loopback校验，空代理明确直连且不继承ambient。
三个公开进程E2E先RED，8项直连/loopback/非法代理回归GREEN，完整同源码harness复核
207 PASS、62 NOT_RUN、0 FAIL；旧FIFO时限FAIL保持原结论，未改时限或宣称已解释根因。
新真实正式Decisions9场景全PASS、9 Calls/Attempts，账户/区域/权限仍unknown而非伪造verified。
当前整版`fc8f02b`实际CLI0、pass=true、10门禁全部PASS，normal/race各9098 PASS、0 FAIL，
各861 Pi/33 UC14，17个输入审计零发现。当前真实/效果及旧失败都进入独立完整字节索引。
验收材料见`../../jev-1.14-multimodal-box/14-remaining-real/verification.md`。
