"""One finite provider-sites collection; argv is constructed here, never imported."""
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / '.scratch/barness-ai-minimax/tools'))
import paid_core as c
import paid_plan as old

ROUND = 'provider-sites-final-acceptance-1'
PHASES = ('candidate', 'glm', 'remaining')
ROUTES = ('chat', 'responses', 'messages')
REGIONAL = tuple(p + '-' + r for p in ('zai', 'minimax-cn') for r in ROUTES)
OLD = tuple(n for n, _, _, _ in old.GENERIC) + ('protocol114',) + tuple('glm-' + r for r in ROUTES) + tuple('mm-global-formal-m3-' + r for r in ROUTES)
EFFECTS = ('text-jev', 'text-luna', 'visual-jev', 'visual-decisions')
MAX_CALLS, MAX_ATTEMPTS = 726, 762


def jobs(out):
    original = old.jobs(c.ROOT, out, c.common_root())
    result = {n: original[n] for n in (*OLD, *EFFECTS)}
    for name, job in result.items():
        job.update(baseName=name, round=ROUND, phase='glm' if name.startswith('glm-') else 'remaining', networkRoute='direct')
        if name.startswith('glm-'):
            job.update(sourceVariable='ZHIPU_KEY', glmRoute=name[4:])
        if name.startswith('mm-'):
            job['environment'].update(BARNESS_MINIMAX_FORMAL_RENEWAL=ROUND,
                BARNESS_MINIMAX_LEDGER=str(c.common_root() / '.evidence/minimax-renewals' / ROUND / 'formal.jsonl'))
    for stage in ('candidate', 'formal'):
        for combo in REGIONAL:
            provider = 'zai' if combo.startswith('zai-') else 'minimax-cn'
            name = stage + '-' + combo
            result[name] = dict(name=name, baseName=name, round=ROUND,
                phase='candidate' if stage == 'candidate' else 'remaining', provider=provider,
                sourceVariable='ZAI_KEY' if provider == 'zai' else 'MINIMAX_CN_KEY',
                injectedCredential='BARNESS_REGIONAL_KEY', networkRoute='direct',
                argv=[str(out / 'live.test'), '-test.run=^TestRegional' + stage.title() + '$', '-test.timeout=1800s'],
                cwd=str(c.ROOT / 'ai/live'), timeoutSeconds=1830, maxCalls=24, maxAttempts=24,
                combo=combo, stage=stage, evidenceRoot=str(out / name),
                environment=dict(BARNESS_AI_LIVE='1', BARNESS_REGIONAL_ROUTE=combo,
                    BARNESS_REGIONAL_SERVICE=provider, BARNESS_REGIONAL_STAGE=stage,
                    BARNESS_REGIONAL_GRANT=str(out / 'grants' / (name + '.json')),
                    BARNESS_AI_EVIDENCE_DIR=str(out / name)))
    for job in result.values():
        # These are explicit missing pre-send contracts, not estimates silently
        # promoted to invoice bounds. Unknown pricing cannot reserve a dollar.
        job['costGuard'] = dict(state='FINAL_WIRE_RESERVATION',
            basis='Encoded bytes+1024, explicit image reservation, final max_tokens or sourced image-output estimate; exact model maximum catalog tier. Unknown usage keeps all reserved USD; this operational estimate is not an invoice guarantee.')
    return result


def inventory(out):
    return dict(schema=1, round=ROUND, maxCalls=MAX_CALLS, maxAttempts=MAX_ATTEMPTS,
        diagnosticMaxCalls=4, diagnosticMaxAttempts=4, jobs=jobs(out),
        credentialsRead=False, automaticProcessReruns=False,
        dollarPolicy='Reserve the entire verified process ceiling before claim/key; never release an unknown or failed process reservation.')


def live_identities():
    return tuple(old.LIVE_IDENTITIES) + REGIONAL
