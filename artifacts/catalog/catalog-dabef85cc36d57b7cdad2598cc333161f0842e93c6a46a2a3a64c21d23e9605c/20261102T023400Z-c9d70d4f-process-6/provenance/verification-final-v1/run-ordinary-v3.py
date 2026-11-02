#!/usr/bin/env python3
"""Run complete ordinary suite against an explicitly frozen input inventory."""
import hashlib,json,os,sys
from pathlib import Path
if len(sys.argv) != 2:
    raise SystemExit('usage: run-ordinary-v3.py <frozen-input-inventory.json>')
manifest = Path(sys.argv[1])
inventory = json.loads(manifest.read_text())
if not isinstance(inventory, dict) or not inventory:
    raise SystemExit('nonempty path to SHA-256 map required')
for name, expected in inventory.items():
    p=Path(name)
    if p.is_absolute() or '..' in p.parts or p.is_symlink() or not p.is_file() or hashlib.sha256(p.read_bytes()).hexdigest()!=expected:
        raise SystemExit('selected input mismatch: '+name)
script=Path('artifacts/catalog/work/20261102T023400Z-c9d70d4f/verification/run-ordinary-v3.py')
argv=['python3','tools/verification/run.py','--feature','catalog-kimi-k31-pro','--stage','e2e','--timeout','2700']
for name in sorted(set(inventory)|{manifest.as_posix(),script.as_posix()}):
    argv.extend(['--input',name])
argv.extend(['--','go','test','-json','./ai/e2e','./ai','./ai/examples/chineseeval/...','./catalog/...','./cmd/catalogctl','./cmd/cataloggen','-count=1','-timeout=40m'])
os.execvp(argv[0],argv)
