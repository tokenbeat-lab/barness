#!/usr/bin/env python3
"""Run adoption and ChineseEvaluation public acceptance without vendor calls."""
import os
from pathlib import Path
script=Path('artifacts/catalog/work/20261102T023400Z-c9d70d4f/verification/run-builtin-green-v1.py')
inputs=[script.as_posix()]
for directory in ('catalog','cmd/catalogctl','cmd/cataloggen'):
    inputs.extend(p.as_posix() for p in Path(directory).rglob('*') if p.is_file())
argv=['python3','tools/verification/run.py','--feature','catalog-kimi-k31-pro','--stage','e2e','--timeout','1200']
for path in sorted(set(inputs)):
    argv.extend(['--input',path])
argv.extend(['--','go','test','-json','./ai/e2e','-run','^(TestKimiK31BuiltinAdoption|TestTerraAdoptedBuiltinSource|TestTerraAdoptedSeedExport|TestChineseEvaluation)','-count=1','-timeout=15m'])
os.execvp(argv[0],argv)
