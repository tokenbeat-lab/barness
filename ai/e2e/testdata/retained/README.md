# Retained ordinary-test inputs

These tracked fixtures make the public evidence-consumer regressions repeatable
from Git. They contain synthetic/redacted historical captures, not credentials
or current live support. Original scratch archives and verdicts stay unchanged.

- `jev-baseline`: exact original ZIP and anchored manifest. Its historical
  release remains FAIL; the tracked historical verifier reads it via --bundle.
- `decisions-text-choice`: exact original ZIP and anchored manifest, including
  red runs and NOT_RUN live. Its verifier never invokes a Provider.
- `decisions-runtime/bundle`: the same six P10-OD58 revocation cases previously
  selected from the first runs/full manifest in runtime.zip. Every original
  selected case file and audit byte is retained. Only the derived manifest's
  case selection/serialization changes; it is not a newly accepted native run.

Each provenance.json pins the source archive/manifest and retained file bytes.
Do not regenerate pinned archives, rewrite historical verdicts, or weaken the
verifier anchors. Missing/corrupt inputs must fail the existing public tests.
Future input changes follow [the testing SOP](../../../../docs/agents/verification.md)
and [the remediation issue](../../../../.scratch/scratch-artifact-remediation/issues/01-retained-ordinary-inputs.md).

The runtime selection and exact-archive fixtures intentionally remain distinct:
their existing consumers require different provenance and identity contracts.
