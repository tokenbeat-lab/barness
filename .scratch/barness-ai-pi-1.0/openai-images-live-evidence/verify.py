#!/usr/bin/env python3
"""Repeatable offline checks of own live evidence, catalog and atomic merge."""
from pathlib import Path
import base64
import hashlib
import json
import os
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
base = Path(__file__).resolve().parent
bundle = base / 'live' / '20261008T071013.136836000Z'
read = lambda p: json.loads(p.read_text())
report = read(bundle / 'live-report.json')
assert report['combo'] == 'openai-images'
assert (report['operation'], report['provider'], report['api']) == ('image', 'openai', 'openai-images')
assert report['model'] == 'gpt-image-2.5-sunburst-2026-09-08'
assert report['endpoint'] == 'https://api.openai.com/v1'
assert report['expected'] == ['json-edit', 'generation', 'mask-edit']
assert report['budget']['callsUsed'] == report['budget']['imagesUsed'] == report['budget']['httpAttempts'] == 3
assert report['budget']['environmentRetries'] == 0
assert read(bundle / 'audit.json')['findings'] == []
cost = 0.0
for case in read(bundle / "manifest.json")["cases"]:
    assert "-run '^TestLive$/^openai-images$'" in case["replay"]
for sc in report['scenarios']:
    assert sc['outcome'] == 'PASS'
    directory = bundle / sc['caseId']
    exchange = read(next(directory.glob('*-exchanges.json')))[0]
    result = read(next(directory.glob('*-result.json')))['result']
    validation = read(next(directory.glob('*-image-check.json')))['validation']
    raster = next(directory.glob('*-output.*')).read_bytes()
    assert hashlib.sha256(raster).hexdigest() == validation['sha256']
    assert len(raster) == validation['bytes']
    assert base64.b64decode(result['content'][0]['data'], validate=True) == raster
    assert validation['width'] == validation['height'] == 1024
    assert exchange['url'] == 'https://api.openai.com' + sc['imageCalls'][0]['path']
    request = json.loads(exchange['requestBody'])
    assert request['model'] == report['model'] and request['size'] == '1024x1024' and request['n'] == 1
    assert 'input_fidelity' not in request
    assert exchange['status'] == 200 and exchange['requestHeaders']['authorization'] == '[REDACTED]'
    assert exchange['responseHeaders']['x-request-id'] == sc['providerRequestIds'][0]
    assert result['metadata']['attempts'][0]['usageReporting'] == 'complete'
    modality = result['usage']['modalities']
    estimate = (modality['inputText'] * 5 + modality['inputImage'] * 8 + modality['outputImage'] * 30) / 1_000_000
    assert abs(estimate - result['usage']['cost']['total']) < 1e-12
    cost += estimate
assert abs(cost - 0.029431) < 1e-12

# Exercise the actual CLI with a valid report followed by an invalid report.
# The failed batch must not change the file, even though the first report passed.
original = json.loads(subprocess.check_output(['git', 'show', 'c0b860c80e2175676984d93dbfa32a063be0cb71:ai/live/support-matrix.json'], cwd=root))
initial = json.loads(json.dumps(original))
initial['rows'].append(dict(combo='openai-images', operation='image', provider='openai', api='openai-images'))
with tempfile.TemporaryDirectory() as temporary:
    folder = Path(temporary)
    matrix = folder / 'matrix.json'
    matrix.write_text(json.dumps(initial, indent=2) + '\n')
    before = matrix.read_bytes()
    bad = dict(report, provider='deepseek')
    bad['finishedAt'] = '2026-10-09T00:00:00Z'
    bad_path = folder / 'bad.json'
    bad_path.write_text(json.dumps(bad))
    result = subprocess.run(['go', 'run', './ai/live/cmd/supportmatrix', '-matrix', str(matrix), str(bundle), str(bad_path)], cwd=root, capture_output=True, text=True)
    assert result.returncode != 0 and matrix.read_bytes() == before
    result = subprocess.run(['go', 'run', './ai/live/cmd/supportmatrix', '-matrix', str(matrix), str(bundle)], cwd=root, capture_output=True, text=True)
    assert result.returncode == 0
    merged = read(matrix)
    published = read(root / 'ai/live/support-matrix.json')
    assert merged == published
    assert merged['rows'][:-1] == original['rows']
    assert merged['rows'][-1]['allPassedAt'] == report['finishedAt']

# Rebuild the current price/catalog snapshot to a temporary destination.
with tempfile.TemporaryDirectory() as temporary:
    snapshot = Path(temporary) / 'snapshot.json'
    subprocess.run(['go', 'run', './ai/release/cmd/releasegate', '-write-snapshot', '-snapshot', str(snapshot)], cwd=root, check=True, capture_output=True)
    assert snapshot.read_bytes() == (root / 'ai/release/catalog-snapshot.json').read_bytes()
summary = dict(status='PASS', liveCases=3, calls=3, images=3, retries=0, estimatedUSD=round(cost, 6), rasterChecksums=True, atomicMerge=True, otherRowsUnchanged=True, catalogSnapshotCurrent=True)
(base / 'verification.json').write_text(json.dumps(summary, indent=2) + '\n')
print(json.dumps(summary))
