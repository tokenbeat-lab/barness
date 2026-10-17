#!/usr/bin/env python3
"""One same-source DeepSeek Chat confirmation inside unchanged task caps."""
import importlib.util
import json
from pathlib import Path
import sys

sys.dont_write_bytecode = True
HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('recovery',HERE/'recover.py')
recovery = importlib.util.module_from_spec(spec)
spec.loader.exec_module(recovery)
r = recovery.r


def run(out):
    job = recovery.Recovery(out)
    job.guard()
    summary = r.doc(job.out/'live-summary.json')
    if {k for k,v in summary['commands'].items() if v!=0}!={'deepseek-chat'}:
        raise RuntimeError('only the diagnosed DeepSeek Chat failure may be rechecked')
    # The complete SSE already demonstrates empty native reasoning, rather than
    # adapter loss. Keep the failed route and all counters; do not narrow tests.
    failed = r.doc(r.bundle(job.out/'live/deepseek-chat')/'live-report.json')
    bad = [s for s in failed['scenarios'] if s['outcome']!='PASS']
    if len(bad)!=1 or bad[0]['id']!='reasoning-history' or bad[0]['errorCategory']!='assertion':
        raise RuntimeError('different failure needs new diagnosis')
    initial = r.doc(recovery.INITIAL/'consumption.json')
    current = r.doc(job.out/'consumption-all.json')
    rows = initial['routes']+current['routes']
    deepseek = sum(row['calls'] for row in rows if row['name'].startswith('deepseek-'))
    deepseek_attempts = sum(row['attempts'] for row in rows if row['name'].startswith('deepseek-'))
    before = r.doc(job.out/'task-consumption.json')
    if (max(before['conservativeCalls'],before['conservativeAttempts'])+32>550
            or max(deepseek,deepseek_attempts)+32>128):
        raise RuntimeError('existing task/provider cap cannot reserve confirmation')
    with (job.out/'deepseek-chat-recheck-reservation.json').open('x') as marker:
        json.dump(dict(maxCalls=32,maxAttempts=32,taskMaxCalls=550,providerMaxCalls=128,
            previousConsumption=before,previousProviderCalls=deepseek,previousProviderAttempts=deepseek_attempts,
            noAutomaticResend=True,sourceUnchanged=True),marker,indent=2)
    code = job.execute('deepseek-chat-recheck',[job.out/'live.test',
        '-test.run=^TestLive$/^deepseek-chat$','-test.timeout=30m'],
        extra={'BARNESS_AI_LIVE':'1','BARNESS_AI_LIVE_COMBO':'deepseek-chat',
            'BARNESS_AI_LIVE_ACCOUNT_ALIAS':'deepseek-local-unverified@region-unverified',
            'BARNESS_AI_EVIDENCE_DIR':str(job.out/'deepseek-chat-recheck')},timeout=2100,cwd=r.ROOT/'ai/live',
        source='DEEPSEEK_KEY',injected='BARNESS_AI_LIVE_KEY_DEEPSEEK_CHAT')
    job.guard()
    report = r.doc(r.bundle(job.out/'deepseek-chat-recheck')/'live-report.json')
    calls,attempts = report['budget']['callsUsed'],report['budget']['httpAttempts']
    if not 0<=calls<=32 or not 0<=attempts<=32:
        raise RuntimeError('confirmation component exceeded')
    total = {k:before[k]+(calls if k in ('calls','conservativeCalls') else attempts)
        for k in ('calls','attempts','conservativeCalls','conservativeAttempts')}
    total.update(maxCalls=550,maxAttempts=550,recoveryRounds=1,targetedConfirmations=1)
    r.dump(job.out/'task-consumption-final.json',total)
    accepting = code==0 and all(s['outcome']=='PASS' for s in report['scenarios']) and not r.doc(
        r.bundle(job.out/'deepseek-chat-recheck')/'audit.json')['findings']
    r.dump(job.out/'accepted-deepseek-chat.json',dict(accepting=accepting,
        selectedBundle=str(r.bundle(job.out/'deepseek-chat-recheck').relative_to(r.ROOT)),
        originalFailedBundle=str(r.bundle(job.out/'live/deepseek-chat').relative_to(r.ROOT)),
        calls=calls,attempts=attempts,oldFailureRetained=True))
    if not accepting:
        raise SystemExit(1)


if __name__ == '__main__':
    if len(sys.argv)!=2 or Path(sys.argv[1]).is_absolute() or '..' in Path(sys.argv[1]).parts:
        raise SystemExit('new repository-relative completed recovery root required')
    run(sys.argv[1])
