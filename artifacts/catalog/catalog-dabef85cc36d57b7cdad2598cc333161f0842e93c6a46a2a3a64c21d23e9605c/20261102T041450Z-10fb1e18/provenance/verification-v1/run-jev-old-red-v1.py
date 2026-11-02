#!/usr/bin/env python3
"""Capture the unchanged historical Jev name-only assumptions."""
import os
argv=['python3','tools/verification/run.py','--feature','catalog-model-id-policy','--stage','e2e','--input','artifacts/catalog/work/20261102T041450Z-10fb1e18/verification/run-jev-old-red-v1.py','--','go','test','-json','./ai/e2e','-run','^TestJevBoxCatalog$','-count=1','-timeout=15m']
os.execvp(argv[0],argv)
