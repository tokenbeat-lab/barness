# Historical catalog fixture

`pre-terra-builtin.json` is the exact `ai/catalog_builtin.json` bytes from
commit `37a283ee4ef743f93d3a12132d1bc2c3211fb1e5`. Its SHA-256 is
`288da66d30a54eef6376fe9afff3827b76e27645578d96ed968b0097be6f28f6`.

It exists to replay the Chinese evaluation v3 configuration against its
original catalog. It is not the current default or a vendor qualification.

`pre-kimi-k31-builtin.json` copies `ai/catalog_builtin.json` at commit
`1827f0956804c37bef8c32594903b895099ee4e5` (SHA-256
`04107af45ba03e3a4e7660aa5255997303b8c1442b74fc6d905b37fba0ba80ca`).
It binds the 119 previous identities and Chinese evaluation v4 historical pin.

`kimi-k31-approved.json` retains the six-candidate CLI's approved final snapshot
(SHA-256 `6fb2842be21797a6aa448a92244d12b4c528c212d74eff8e16220a883e89818a`).
It independently checks which exact bytes were adopted. Neither fixture grants
vendor qualification.
