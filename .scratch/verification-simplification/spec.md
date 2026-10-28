# 验收资格与历史治理精简

Status: ready-for-agent
Type: feature

依据[已批准方案](../provider-integration-audit-20261027/simplification-plan.md)和[ADR-0055](../../docs/adr/0055-route-scoped-evidence-applicability.md)实施。先修改规范和资格失效，再摘除历史控制器与重复包装。所有历史证据、授权、claim、账本及暂停状态保留；不执行付费 live、race、压力或暂停轮次。

公共验证入口为 Client E2E、已编译 releasegate/原生 CLI 的受控子进程，以及现有 Python 采集器进程。失败方式先覆盖：无关目录变更令旧路线过期、工具变化牵动厂商资格、阶段输入不同拒绝、旧源码或原件缺失被伪报适用、相关产品源码变化被遗漏、缺场景/缺终态/篡改证据、旧权限 fallback、重复领取、取消后持续写入。优先复用现有场景，不添加实现镜像测试。

唯一当前状态见[status.md](status.md)。活动内部格式直接切换，不保留旧轮次 shim；固定历史消费者与源版本单独保留用于只读追溯。
