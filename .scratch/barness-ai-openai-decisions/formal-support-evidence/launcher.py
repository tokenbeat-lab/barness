from pathlib import Path
import datetime
import hashlib
import json
import os
import subprocess

root = Path.cwd()
keys = []
for line in (root / '.env').read_text().splitlines():
    name, sep, value = line.partition('=')
    if sep and name.strip() == 'OPENAI_KEY':
        value = value.strip()
        if len(value) >= 2 and value[0] == value[-1] and value[0] in ['"', "'"]:
            value = value[1:-1]
        keys.append(value)
if len(keys) != 1 or not keys[0]:
    raise SystemExit('OPENAI_KEY missing or ambiguous')
key = keys[0]
base = root / '.evidence/formal-decisions-live-authorized'
base.mkdir(parents=True, exist_ok=True)
out = root / '.scratch/barness-ai-openai-decisions/formal-support-evidence'
source_dir = out / 'execution-source'
source_dir.mkdir(exist_ok=True)
files = list((root / 'ai').glob('*.go')) + list((root / 'ai/live').glob('*.go'))
files += list((root / 'ai/internal/testkit/supportmatrix').glob('*.go'))
files += [root / 'go.mod', root / 'go.sum', root / 'ai/release/catalog-snapshot.json']
source_hashes = {}
for file in files:
    if file.parent != root / 'ai/live' and file.name.endswith('_test.go'):
        continue
    relative = file.relative_to(root)
    dest = source_dir / relative
    dest.parent.mkdir(parents=True, exist_ok=True)
    raw = file.read_bytes()
    dest.write_bytes(raw)
    source_hashes[str(relative)] = hashlib.sha256(raw).hexdigest()
(out / 'execution-source-snapshot.json').write_text(json.dumps(source_hashes, indent=2) + '\n')
env = {k: os.environ[k] for k in ['PATH', 'HOME', 'TMPDIR'] if k in os.environ}
env.update({'BARNESS_AI_LIVE': '1', 'BARNESS_AI_LIVE_COMBO': 'openai-decisions',
            'BARNESS_AI_LIVE_KEY_OPENAI_DECISIONS': key, 'BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT': '1',
            'BARNESS_AI_LIVE_PROXY': 'http://127.0.0.1:7890', 'BARNESS_AI_EVIDENCE_DIR': str(base)})
cmd = ['go', 'test', '-tags', 'live', '-timeout', '180s', './ai/live', '-run', '^TestFormalDecisions$', '-count=1', '-json']
start = datetime.datetime.now(datetime.timezone.utc).isoformat()
r = subprocess.run(cmd, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=210)
log = r.stdout.decode().replace(key, '[REDACTED]').replace(str(Path.home()), '<host-home>')
(out / 'live-execution.jsonl').write_text(log)
receipt = {'command': cmd, 'startedAt': start, 'finishedAt': datetime.datetime.now(datetime.timezone.utc).isoformat(),
           'exitCode': r.returncode, 'credentialSource': '.env:OPENAI_KEY', 'proxy': 'http://127.0.0.1:7890',
           'accountVerification': 'unknown', 'regionVerification': 'unknown', 'permissionVerification': 'unknown',
           'userAuthorization': 'issue 10 formal live; carries issue 09 explicit user authorization for credential, proxy and unknown provenance',
           'logSHA256': hashlib.sha256(log.encode()).hexdigest(),
           'sourceSnapshotSHA256': hashlib.sha256((out / 'execution-source-snapshot.json').read_bytes()).hexdigest()}
(out / 'live-receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')
print(json.dumps({'exitCode': r.returncode, 'log': str((out / 'live-execution.jsonl').relative_to(root))}))
