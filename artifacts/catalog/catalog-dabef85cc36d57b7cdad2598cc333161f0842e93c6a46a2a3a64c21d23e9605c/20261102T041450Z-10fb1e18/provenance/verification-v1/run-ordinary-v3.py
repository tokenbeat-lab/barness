#!/usr/bin/env python3
"""Run complete ordinary packages using a frozen input inventory."""
import hashlib,json,os,sys
from pathlib import Path
if len(sys.argv)!=2:raise SystemExit('usage: run-ordinary-v3.py <input-inventory.json>')
manifest=Path(sys.argv[1]);inputs=json.loads(manifest.read_text())
if not isinstance(inputs,dict) or not inputs:raise SystemExit('nonempty input inventory required')
for name,digest in inputs.items():
 p=Path(name)
 if p.is_absolute() or '..' in p.parts or p.is_symlink() or not p.is_file() or hashlib.sha256(p.read_bytes()).hexdigest()!=digest:raise SystemExit('selected input mismatch: '+name)
script=Path('artifacts/catalog/work/20261102T041450Z-10fb1e18/verification/run-ordinary-v3.py')
argv=['python3','tools/verification/run.py','--feature','catalog-model-id-policy','--stage','e2e','--timeout','2700']
for name in sorted(set(inputs)|{manifest.as_posix(),script.as_posix()}):argv.extend(['--input',name])
argv.extend(['--','go','test','-json','./ai/e2e','./ai','./ai/examples/chineseeval/...','./catalog/...','./cmd/catalogctl','./cmd/cataloggen','-count=1','-timeout=40m'])
os.execvp(argv[0],argv)
