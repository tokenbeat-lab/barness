#!/usr/bin/env python3
"""Capture targeted public identity-policy acceptance."""
from pathlib import Path
import os
script=Path('artifacts/catalog/work/20261102T041450Z-10fb1e18/verification/run-green-v4.py')
inputs=[script.as_posix()]
for directory in ('catalog','cmd/catalogctl','cmd/cataloggen'):
    inputs.extend(p.as_posix() for p in Path(directory).rglob('*') if p.is_file())
argv=['python3','tools/verification/run.py','--feature','catalog-model-id-policy','--stage','e2e','--timeout','1200']
for path in sorted(set(inputs)):
    argv.extend(['--input',path])
argv.extend(['--','go','test','-json','./ai/e2e','-run','^Test(CatalogFutureModel|KimiK31CatalogRejectsExpandedCapabilities|XAICatalogBoundary|LocalCatalogClassifierFailureTaxonomy|CatalogMaintenanceCLIKnownDefinitionCodeGap|JevBoxCatalog)','-count=1','-timeout=15m'])
os.execvp(argv[0],argv)
