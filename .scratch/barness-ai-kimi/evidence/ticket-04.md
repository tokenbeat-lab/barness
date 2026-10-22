# Ticket 04 Messages ordinary verification

公共 Client + 受控 Provider seam；无厂商调用。所有原件已按原 attempt 名称复制并逐文件校验到主工作区 `.evidence/barness-ai-kimi-messages/e2e/`，保留全部失败和构建失败。

最终受测提交：`66eab996b860801d188507d071e8e572e355b658`。最终命令：

```sh
python3 .scratch/barness-ai-test-process-review/tools/run.py --feature barness-ai-kimi-messages --stage e2e -- go test -json ./ai/e2e ./ai -run 'TestKimiMessages|TestAnthropic|TestGLMMessages|TestMiniMaxMessages|TestProtocol(Text|History|Usage|Options|Failures)' -count=1 -timeout=10m
```

最终退出 0：665 个叶测试 PASS，其中 Kimi Messages 196 个；主 E2E bundle 和 audit 通过。`./ai` 在该表达式下仅编译/包执行成功，无匹配叶测试，不当作其完整行为回归。

最终 bundle：`../../../.evidence/barness-ai-kimi-messages/e2e/20261022T141834.119268Z-_tmhbk4_/bundles/20261022T141838.988654000Z/manifest.json`。

实现覆盖：稳定资源/Bearer/version、独立 effort、顶层缓存、命名/最终字段拒绝、四入口工具往返、实际 signed/unsigned thinking 与来源降级、严格终态/工具 JSON、收费计数与实际 TTL/成本区间、HTTP 分类与有限重试、字节/队列/工具限制、取消/超时、Header guards、回调/请求/Observer/Result 快照。无推断 Kimi request-ID 头。

live / race / 压力均 NOT_RUN：本工单仅授权普通受控验证；候选目录和受控回放不产生真实资格。完整集成回归及发布判定由唯一 spec 当前状态记录，release.pass=false。

| Attempt | Status | Exit | Receipt SHA-256 |
| --- | --- | --- | --- |
| `20261022T140446.610647Z-eqna0z43` | FAIL | 1 | `c541d3995e92bea4025750ca3c68fe08017f4a852a017b8af711061ef32f79c4` |
| `20261022T140553.551677Z-33bl8_30` | FAIL | 1 | `5ea1796b454fc67df9520b7cc24ef1c8938689d1f468802b40f7acf195a078f2` |
| `20261022T140634.694097Z-nqi54x90` | PASS | 0 | `a8468442c3987be6efcfaf67a9b8c5e5c38b4f9a16cf3ea703a738b9313e5865` |
| `20261022T140731.021557Z-fkg5vc6r` | FAIL | 1 | `31d7324bce0a827f9f66590242d95beda72241ba38096df2b0ab678257af8713` |
| `20261022T140818.082710Z-zy4stan1` | PASS | 0 | `de0ee8b5aa9a6ce59e2af2ab2d780777f3305b3fa6346c75ab9f77f312762f4c` |
| `20261022T140923.153082Z-fqre9itq` | FAIL | 1 | `18ba1bc82e4bbbeab4d8c0be9fab3a7112b2a20953c105d3059b6c80d5de2291` |
| `20261022T140943.799527Z-9ebhfr9o` | FAIL | 1 | `31d266f13dd251b3ff506959bcc4161dd0b4f859d1b5ffcda9e4ebb86db1a992` |
| `20261022T140955.374730Z-vzurb8ys` | FAIL | 1 | `5d20d91cc25ae52ed35de0b4f3c6a0bb1fe9f382bd60ac70ca9180960f722ba1` |
| `20261022T141027.363830Z-akt_8e_e` | PASS | 0 | `10ab2478efa0c8afab3d1a2a27125ca741a95d5114205f830ff0bcd8beb75de1` |
| `20261022T141059.786760Z-p1n8wo4t` | FAIL | 1 | `c474c8b0fbbdbe5d56846f160595d949c907a932137b4682ba710f6893658407` |
| `20261022T141117.118213Z-fuy0dhfy` | FAIL | 1 | `3fd0e4d07908200ffe2212a44a3c84e1456acdcc11d51219acb76b0c15dd65fc` |
| `20261022T141219.445540Z-sjizawdg` | FAIL | 1 | `17a437fec93a17abb1eb7beea29a103f51196b43c09241bd03222dca1afb9ab5` |
| `20261022T141246.500110Z-mwdttoim` | PASS | 0 | `1d1f19b80ec5203be911ab1b3f06f405ebeb7b53ceaefa72a757d83c3b6f7abf` |
| `20261022T141435.613535Z-kupyj_m4` | FAIL | 1 | `1425f7f369ac58a53d10cf6587f743232ec0c4e93daaafc46157f558fb12d5b7` |
| `20261022T141548.262973Z-k7fapl_s` | FAIL | 1 | `a70a75529851958ee601960571df7ce7d4e95f33850cbc26634b74d97f18a4d6` |
| `20261022T141608.482985Z-i1eo3qsq` | PASS | 0 | `abdbbe2b01270aa0e1a89dee74645ae9f36af9826c1b78c17ef4d27e8a4e31fb` |
| `20261022T141730.491918Z-oz4onw2i` | FAIL | 1 | `b9a494ca2c3da7dd29cf1b0132ab5b32d2206aae2d97d7960e5f68a176bba916` |
| `20261022T141739.226312Z-6qrvmfxa` | PASS | 0 | `a69e0767551dc7381d47f963a6c2a56ff656028876c73abe2cb2b1adecdafdcc` |
| `20261022T141801.199719Z-4aihei4x` | FAIL | 1 | `1b4f73bebfa6a67424072f9317fe09a3f941bd857c1c2d9726fe905818940240` |
| `20261022T141817.416147Z-ehjalrqy` | FAIL | 1 | `182a1ca21a3e289e1c4671642ead2de1269c0b2c43123870ead847667dd1aecc` |
| `20261022T141834.119268Z-_tmhbk4_` | PASS | 0 | `3b11daa6ed36e29e5e8b7df961a995adcbcbf5deb7752270abf08533f9cc4cb3` |
