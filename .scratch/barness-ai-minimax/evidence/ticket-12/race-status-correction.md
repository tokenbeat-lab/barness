# 5abc race diagnosis correction

The preserved 5abc1a5 race manifest has **4 PASS test cases**. There is no
confirmed child timeout. The implementer inferred a timeout from a child
approaching its synthetic limit, then interrupted the slow parent final audit
while reducing redundant fault injection. That inference was wrong.

The original command receipt is retained with exit **-15**, and its final audit
receipt is absent. Thus this run is **assertions PASS / audit incomplete / not
Go PASS**. The provisional `race-interruption.json` reason is retained as
curation history and is superseded by this note; it must not be cited as an
observed timeout or a test failure. No raw manifest, assertion or receipt is
rewritten.

The subsequent single-fault fixture at 6ff05b8 completed ordinary and race:
4 PASS each, exit0, independent artifact audit0 findings. Its atomic-once change
bounds repeated synthetic negative-probe cost; production capture/approval
source is byte-identical to independently reviewed 5abc1a5. No vendor call or
actual proof was affected.
