"""Index sealed current release evidence and record owned temporary cleanup."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

W = Path('/tmp/barness-glm-13')
R = Path('/Users/cyber/RestoX/barness')
E = R / '.evidence/barness-ai-glm/ticket-13'
D = W / '.scratch/barness-ai-glm/evidence/ticket-13'


def digest(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        while chunk := stream.read(1 << 20):
            h.update(chunk)
    return h.hexdigest()


def doc(path):
    return json.loads(path.read_bytes())


def dump(path, value):
    path.write_text(json.dumps(value, indent=2) + '\n')


seal = doc(E / 'current-release-seal.json')
assert seal['passVerdict']
freeze = doc(R / '.evidence/glm13-corrective-current-r1/source-at-start.json')
tracked = subprocess.check_output(['git', 'ls-files', 'ai'], cwd=W, text=True).splitlines()
paths = {p for p in tracked if p.endswith(('.go', '.json')) and p != 'ai/live/support-matrix.json'}
paths.update(['go.mod', 'go.sum'])
paths.update(str((D / n).relative_to(W)) for n in ('commands.py', 'run.py', 'glm_launcher.py', 'rollback.py', 'rollback-probe.txt', 'corrective.py'))
assert paths == set(freeze['files']), 'complete frozen source/config additions or deletions'
assert all(digest(W / p) == sha for p, sha in freeze['files'].items())
assert freeze['gitCommit'] == seal['sourceCommit']
before_path = Path('/tmp/barness-glm-user-untracked-before.json')
before = doc(before_path)
after = {p: dict(sha256=digest(R / p), bytes=(R / p).stat().st_size) for p in before}
assert before == after
shutil.copyfile(before_path, E / 'user-untracked-before.json')
dump(E / 'user-untracked-after.json', after)
oracle = doc(W / '.evidence/glm13-legacy-current-r3/oracle-environment.json')
link = W / oracle['link']
assert link.is_symlink() and os.readlink(link) == oracle['target']
assert all(digest(link.parent / name) == sha for name, sha in oracle['lockedSourcePins'].items())
assert all(digest(R / link.parent.relative_to(W) / name) == sha for name, sha in oracle['lockedSourcePins'].items())
link.unlink()
cache = D / '__pycache__'
if cache.exists():
    shutil.rmtree(cache)
oracle.update(removedAfterSeal=True, rootDependencyDirectoryRetained=Path(oracle['target']).is_dir(),
    rootAndWorkerLockedSourcePinsUnchanged=True)
dump(E / 'owned-cleanup.json', dict(oracle=oracle, ownedPythonCacheRemoved=not cache.exists(),
    userUntrackedFiles=len(after), userUntrackedBytesUnchanged=True,
    rootHEAD=subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=R, text=True).strip(),
    rootBranch=subprocess.check_output(['git', 'branch', '--show-current'], cwd=R, text=True).strip(),
    rootTrackedStatus=subprocess.check_output(['git', 'status', '--porcelain', '--untracked-files=no'], cwd=R, text=True),
    rollbackTemporaryRestoration='removed by public rollback receipt; helper worktree not retained',
    implementerWorktree='parent merger owns final worktree cleanup'))

# The three actual execution binaries are checked against both their immutable
# prepay hashes and durable canonical copies; no build runs during sealing.
binaries = doc(R / '.evidence/glm13-corrective-current-r1/binaries.json')
assert all(digest(W / '.evidence/glm13-corrective-current-r1' / n) == sha for n, sha in binaries.items())
assert all(digest(R / '.evidence/glm13-corrective-current-r1' / n) == sha for n, sha in binaries.items())
dump(E / 'final-source-applicability.json', dict(finalExecutionCommit=seal['sourceCommit'],
    completeFrozenFiles=len(freeze['files']), allFrozenBytesPreserved=True, sourceFreezeSHA256=digest(R / '.evidence/glm13-corrective-current-r1/source-at-start.json'),
    measuredGLMCommit='34853063396ad58476e77dfe52463b3a9154ee77',
    glmSupportConsumerCommit='adc694d84e54d47fb120cbc13031ffa5d5074f87',
    finalConsumer='complete consumed Go source unchanged; production current-source validator passed normal/race and gate',
    originalLegacyMeasuredCommit='dbc6c4e048e8578997839134af2203f3d102ddeb',
    correctiveMeasuredCommit=freeze['gitCommit'], actualFullGateCommit=seal['sourceCommit'],
    reusedReports='10 generic +3 GLM own support +2 Chinese effects retain original source IDs; exact selected-source proof and public current consumers pass',
    correctivePlan='36 max/36 actual; Jev own support12 and visual12+12, no balance/retry',
    finalDocumentationAndArtifactTooling='outside measured execution source set; source/config/paid launchers unchanged',
    binarySHA256=binaries, binaryExactCopiesVerified=True,
    prepaySourceApplicability=dict(path=str(R / '.evidence/glm13-corrective-current-r1/source-applicability.json'), sha256=digest(R / '.evidence/glm13-corrective-current-r1/source-applicability.json')),
    originalFailedRelease='historical-r3-preservation.json; original dbc full CLI exit1, immutable',
    historicalTicket12Selectors='.scratch/barness-ai-glm/evidence/ticket-13/history/',
    historicalBaseline='rollback-history/reconciliation.json; original source/catalog only'))
shutil.copyfile(Path(__file__), D / 'index.py')
shutil.copyfile(D / 'seal.py', E / 'seal.py')
shutil.copyfile(D / 'reconcile-rollback-history.py', E / 'reconcile-rollback-history.py')
shutil.copyfile(D / 'index.py', E / 'index.py')
shutil.copyfile(D / 'preserve-history.py', E / 'preserve-history.py')
shutil.copyfile(D / 'preserve-default-history.py', E / 'preserve-default-history.py')

pins = {}
for parent, dirs, files in os.walk(E):
    for name in files:
        path = Path(parent) / name
        if path.name in ('durable-evidence-files.json', 'final-index.json'):
            continue
        assert not path.is_symlink(), str(path)
        pins[str(path.relative_to(E))] = digest(path)
dump(E / 'durable-evidence-files.json', pins)
record = dict(schema=1, passVerdict=True, sourceCommit=seal['sourceCommit'],
    durableRoot=str(E),
    catalogVersion='2026-10-17.2', catalogHash='sha256:81d6d9f3b69e9231985967e1ec8e13640aa7f4fdfd9e7ccef72f7d05320a6942',
    actualFullRelease=dict(report='release-report.json', sha256=digest(E / 'release-report.json'),
        seal='current-release-seal.json', sealSHA256=digest(E / 'current-release-seal.json'),
        normalCases=seal['bundles']['offline']['cases'], raceCases=seal['bundles']['race']['cases'],
        gates=len(seal['gates']), audits=seal['audits'], liveIdentities=14, protocolCombinations=13,
        effects=4, commandsPassed=True, suppliedOfflineBundle=False),
    trace=dict(requirements=22, acceptanceGroups=15, stories=62, items=99, finiteCaseIDs=643,
        allRequiredCasesNormalAndRacePass=True, receipt='complete-trace-receipts.json',
        sha256=digest(E / 'complete-trace-receipts.json')),
    sourceApplicability=dict(path='final-source-applicability.json', sha256=digest(E / 'final-source-applicability.json')),
    paid=seal['paid'],
    historicalGLM=dict(originalsPreserved=True, glmOnlyLogicalCalls=217, glmOnlyHTTPAttempts=215,
        scope='exploration19/17 + candidate66/66 + formal12 66/66 + formal13 66/66; no total invoice claim'),
    rollback=dict(current='rollback-index.json', historicalReconciliation='rollback-history/reconciliation.json',
        restoredBaseline='f5cf5362610a8bdb8f3e170f408f3858c149e540', completeTrackedFiles=6682,
        exactSelectedHistoricalPayloads=201887, historicalMeasuredGoModuleFiles=600),
    proofIndexes={name:digest(E / name) for name in [
        'formal-current/capture-index.json', 'formal-current/independent-reconciliation.json',
        'formal-current/independent-parent-usage.json', 'legacy-paid-index.json',
        'legacy-paid-root-verification.json', 'normal-root-verification.json', 'race-root-verification.json',
        'effects-root-verification.json', 'effect-summary.json', 'current-effect-summary.json',
        'current-normal-root-verification.json', 'current-race-root-verification.json',
        'current-case-regression-scope.json',
        'final-whole-gate-root-verification.json',
        'prereview/corrective-3d3cd0f/barness-glm-corrective-paid-root-verification.json',
        'prereview/corrective-3d3cd0f/barness-glm-corrective-source-root-verification.json',
        'prereview/corrective-3d3cd0f/barness-glm-corrective-reuse-root-verification.json',
        'prereview/corrective-3d3cd0f/prerelease-source-review.json',
        'prereview/corrective-3d3cd0f/prepaid-checks.json',
        'prereview/corrective-3d3cd0f/matrix-root-verification.json',
        'historical-r3-preservation.json', 'historical-r3-files.json',
        'remaining-history-preservation.json', 'default-e2e-history-preservation.json',
        'failed-full-gate-r1-index.json', 'corrective-paid-index.json',
        'consumer-diagnostic-r2-index.json',
        'prereview/ten-axis-review-reports.json', 'rollback-index.json',
        'rollback-history/reconciliation.json', 'owned-cleanup.json']},
    verificationTools={name:dict(source=str((D / name).relative_to(W)), sha256=digest(D / name))
        for name in ('seal.py', 'reconcile-rollback-history.py', 'preserve-history.py', 'preserve-default-history.py', 'index.py')},
    canonicalTrees=seal['trees'],
    durableInventory=dict(path='durable-evidence-files.json', sha256=digest(E / 'durable-evidence-files.json'), files=len(pins)),
    scopes=dict(prepaidReviews='all original and corrective exact-source two-axis reports retained; all findings closed before respective paid collection; final parent review remains separate',
        intermediateReceipts='legacy-paid-index and independent original normal/race/effect receipts are preserved checkpoints of the closed original round; historical-r3 retains actual failed CLI; current release-report/current-release-seal provide the new complete final verdict',
        independentManifestReceipts='current-normal/current-race independently verify cases and payloads only; actualFullRelease/current-release-seal supplies the whole CLI verdict',
        originalWorkerPaths='preserved provenance; durableRoot/canonicalTrees contain usable current copies; remaining-history-preservation.json accounts for all otherwise-unsealed historical trees; default-e2e-history-preservation.json retains34 original focused runs',
        deployment='not performed', accountRegionPermissionsBilling='UNKNOWN'))
dump(E / 'final-index.json', record)
dump(D / 'index.json', record)
print('Final tracked evidence index PASS;', len(pins), 'durable payload pins')
