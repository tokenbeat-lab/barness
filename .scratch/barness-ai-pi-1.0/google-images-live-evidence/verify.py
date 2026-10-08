#!/usr/bin/env python3
"""Verify checksums, actual report outcomes, raster headers and own matrix row."""
import hashlib
import json
from pathlib import Path

folder = Path(__file__).resolve().parent
root = folder.parents[2]


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


for name, expected in read(folder / 'sha256.json').items():
    assert sha(folder / name) == expected, f'artifact changed: {name}'
for name, expected in read(folder / 'source-hashes.json').items():
    assert sha(root / name) == expected, f'source changed: {name}'
report = read(folder / 'live-pass/live-report.json')
assert report['operation'] == 'image' and report['provider'] == 'google' and report['api'] == 'google-interactions'
assert report['expected'] == ['generation', 'reference-edit']
assert [s['outcome'] for s in report['scenarios']] == ['PASS', 'PASS']
assert report['budget']['callsUsed'] == report['budget']['imagesUsed'] == report['budget']['httpAttempts'] == 2
assert report['budget']['maxImages'] == report['budget']['maxCalls'] == 4
assert report['budget']['environmentRetries'] == 0
for scenario in report['scenarios']:
    a = scenario['imageCalls'][-1]
    assert a['path'] == '/v1beta/interactions' and a['status'] == 200 and a['verified'] == 1
    assert a['width'] == a['height'] == 1024 and a['outputFormat'] == 'jpeg'
    assert a['usageReporting'] == 'partial'
    g = a['google']
    assert not g.get('responseId') and g['responseModel'] == report['model']
    assert g['status'] == 'completed' and g['inlineImages'] == 1 and g['fixedRequest']
    assert g['thoughtInModalities'] is False and not g['uri'] and not g['continuation']
    assert g['usage']['output_tokens_by_modality'] == [{'modality': 'image', 'tokens': 1120}]
    path = folder / 'live-pass' / scenario['caseId']
    prefix = 'attempt-1-' + scenario['id']
    check = read(path / (prefix + '-image-check.json'))['validation']
    raster = path / (prefix + '-output.jpeg')
    assert sha(raster) == check['sha256'] and raster.stat().st_size == check['bytes']
    assert raster.read_bytes()[:2] == b'\xff\xd8'
    exchange = read(path / (prefix + '-exchanges.json'))[-1]
    request = json.loads(exchange['requestBody'])
    assert request['store'] is False and 'delivery' not in request['response_format']
    assert request['response_format'] == {'type': 'image', 'aspect_ratio': '1:1', 'image_size': '1K'}
    assert exchange['requestHeaders']['x-goog-api-key'] == '[REDACTED]'
    assert not exchange.get('truncated')
assert [s['outcome'] for s in read(folder / 'delivery-refused/live-report.json')['scenarios']] == ['FAIL', 'FAIL']
assert [s['outcome'] for s in read(folder / 'not-run/live-report.json')['scenarios']] == ['NOT_RUN', 'NOT_RUN']
row = next(r for r in read(folder / 'matrix.json')['rows'] if r['combo'] == report['combo'])
assert row['allPassedAt'] == report['finishedAt']
assert [c['outcome'] for c in row['capabilities']] == ['PASS', 'PASS']
assert not read(folder / 'live-pass/audit.json')['findings']
assert not read(folder / 'audit.json')['findings']
if (folder / 'offline-summary.json').exists():
    assert read(folder / 'offline-summary.json')['passing_cases'] > 4000
    assert all(c['status'] == 'PASS' for c in read(folder / 'offline-manifest.json')['cases'])
if (folder / 'live-harness-manifest.json').exists():
    harness = read(folder / 'live-harness-manifest.json')
    assert harness['audit_findings'] == 0
    assert all(c['status'] == 'PASS' for c in harness['cases'])
    assert {c['id'] for c in harness['cases']} >= {
        'P09-live-capture-' + name for name in
        ['request-prefix', 'response-prefix', 'malformed', 'read-interrupted']}
if (folder / 'verification.json').exists():
    verification = read(folder / 'verification.json')
    assert verification['status'] == 'PASS' and verification['failed_cases'] == 0
    assert verification['run_audit_findings'] == 0 and verification['all_recorded_gauges_zero']
    assert verification['review']['axes']['Standards']['remaining'] == 0
    assert verification['review']['axes']['Spec']['remaining'] == 0
    assert all(c['exit_code'] == 0 for c in verification['commands'])
    trace = read(folder / 'trace.json')
    assert trace['catalog_snapshot_current'] and trace['differential_pass']
    assert trace['own_live_row']['status'] == 'PASS'
    assert len(trace['items']) == 3 and all(i['status'] == 'PASS' for i in trace['items'])
print('PASS: issue 14 source/artifact hashes, own live evidence, raster checks and matrix')
