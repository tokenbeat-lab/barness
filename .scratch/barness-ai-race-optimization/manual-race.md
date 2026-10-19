# 人工 race 待执行清单

当前事实统一见 [status.json](status.json)。本轮实际 race 启动0次，厂商调用0次；方案实施授权不授予下一次 race 或付费续期。

## 当前前提

- 实现来源：`0433459`；双vet和完整普通测试通过。普通测试的实际命令、环境与原件见 [verification.md](verification.md)。
- 当前外部证明 preflight 失败：Jev1.14自身支持、Jev视觉效果、Decisions视觉效果的来源不再适用于改动后的发布工具。其余16份支持及两份中文效果仍被原消费者接受。
- 因外部输入未就绪，本轮没有生成正式 READY_FOR_RACE 或 check-state。独立普通回归不能伪装成阶段一成功。
- 三份证明续期需另行明确范围/资格/预算授权；本轮不修改原证明、selector、历史hash或矩阵。

## 正式第二阶段接口（尚未执行）

现有输入取得当前资格后，先完成新版第一阶段，取得真实 READY_FOR_RACE 的全新目录；若输入或执行来源变化，旧普通结果不能直接续接。编译后的CLI保留退出2的待人工语义（go run会包装成自身退出1）。

```sh
go build -o .evidence/releasegate ./ai/release/cmd/releasegate
.evidence/releasegate -stage check -out <new-check-dir> <current-input-flags>
# 只有人工明确选择后运行：
.evidence/releasegate -stage race -from <new-check-dir> -out <new-race-dir>
```

固定race范围与实际子命令为 `go test -race -json -count=1 -timeout=75m ./...`，差分和压力均开启，凭据隔离；生成独立race bundle，再调用原十项门禁。输入计划、普通回执/原件、编译来源、测试、构建工具及其环境都重新核对，拒绝输入覆盖或混合来源。

参考历史5次全量race记录约37.0–48.9分钟；本版本尚未测得实际耗时。包级上限75分钟，CLI整个race子进程上限80分钟。只统计实际执行，合成fake-Go调度测试不计作race验证。

## 独立离线 race 选项

若人工只选择完整离线race而暂不续期真实证明，可在无厂商凭据环境运行同一完整Go包集合及差分/压力、JSON事件，保留独立原生bundle及实际命令回执。它只能报告已运行的离线范围；外部来源资格和完整发布认证仍开放。此次没有执行该选项。

人工选择后的失败、检测器报告或来源漂移保留为失败；不因为失败、收尾、collect-all或实施方案曾获批而自动追加race轮次。
