"""Exact-source freezes and fixed one-round approval pins; never acquires keys."""
import hashlib
import json
from pathlib import Path
import subprocess
import shutil
import sys
import final_contract as r
c = r.c


def path(value):
    from renewal_common import repo_path
    p = repo_path(value)
    parts = p.relative_to(c.ROOT).parts
    if len(parts) < 3 or parts[:2] != ('.evidence', 'provider-sites'):
        raise RuntimeError('new .evidence/provider-sites collection required')
    return p


def sources():
    files = c.sources()
    for file in Path(__file__).parent.glob('*'):
        if file.is_file() and file.suffix in ('.py', '.md'):
            files[file.relative_to(c.ROOT).as_posix()] = c.digest(file)
    return dict(sorted(files.items()))


def freeze():
    commit = c.git('rev-parse', 'HEAD')
    files = sources()
    for name, sha in files.items():
        raw = subprocess.check_output(['git', 'show', commit + ':' + name], cwd=c.ROOT, env=c.ENV, stderr=subprocess.DEVNULL)
        if hashlib.sha256(raw).hexdigest() != sha:
            raise RuntimeError('all operator/runtime source must be committed before freeze')
    return dict(sourceCommit=commit, files=files)


def run(out, name, argv, **kwargs):
    code = c.execute(out, name, argv, **kwargs)
    command = out / (name + '-command.json')
    value = c.doc(command)
    import datetime
    value['finishedAt'] = datetime.datetime.now(datetime.timezone.utc).isoformat()
    c.dump(command, value)
    return code


def prepare(out, phase):
    frozen = freeze()
    out.mkdir(parents=True, exist_ok=False, mode=0o700)
    c.dump(out / 'source-at-start.json', frozen)
    import final_predecessor
    predecessor=final_predecessor.disclosure()
    c.dump(out/'predecessor-disclosure.json',predecessor)
    if run(out,'go-version',['go','version'])!=0:
        raise RuntimeError('build toolchain identity unavailable')
    toolchain=dict(goVersion=(out/'go-version.log').read_text().strip(),pythonVersion=sys.version,
        executables={str(Path(shutil.which('go')).resolve()):c.digest(Path(shutil.which('go')).resolve()),
                     str(Path(sys.executable).resolve()):c.digest(Path(sys.executable).resolve())})
    c.dump(out/'toolchain.json',c.redact(toolchain))
    if run(out, 'regional-source', ['go', 'run', './ai/live/cmd/regionalverify', '-source']) != 0:
        raise RuntimeError('keyless native regional source failed')
    facts = json.loads((out / 'regional-source.log').read_text())
    source_hash = facts.get('sourceSha256', '')
    if len(source_hash) != 64 or any(v not in '0123456789abcdef' for v in source_hash):
        raise RuntimeError('native regional build source required')
    for name, sha in facts['files'].items():
        raw = subprocess.check_output(['git', 'show', frozen['sourceCommit'] + ':' + name], cwd=c.ROOT, env=c.ENV, stderr=subprocess.DEVNULL)
        if hashlib.sha256(raw).hexdigest() != sha:
            raise RuntimeError('complete native execution source must be committed before freeze')
    c.dump(out / 'regional-source.json', facts)
    builds = [('live.test', ['go', 'test', '-c', '-tags', 'live', '-ldflags',
        '-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource=' + source_hash,
        '-o', out / 'live.test', './ai/live'])]
    packages = dict(regionalpromote='./ai/live/cmd/regionalpromote', regionalverify='./ai/live/cmd/regionalverify', minimaxverify='./ai/live/cmd/minimaxverify',
        glmformal='./ai/live/cmd/glmformal', jevformal='./ai/live/cmd/jevformal', supportmatrix='./ai/live/cmd/supportmatrix',
        chineseeval='./ai/examples/chineseeval/cmd/chineseeval', visualeval='./ai/examples/visualeval', releasegate='./ai/release/cmd/releasegate')
    builds += [(n, ['go', 'build', '-o', out / n, package]) for n, package in packages.items()]
    for name, argv in builds:
        if run(out, 'build-' + name, argv, timeout=600) != 0:
            raise RuntimeError('credential-free reviewed binary build failed')
    c.dump(out / 'binaries.json', {name: c.digest(out / name) for name, _ in builds})
    import final_grants
    component_templates = final_grants.templates(out,phase)
    c.dump(out/'component-templates.json',component_templates)
    all_jobs = r.jobs(out)
    plan = r.inventory(out) | dict(phase=phase, routes={n: j for n, j in all_jobs.items() if j['phase'] == phase})
    c.dump(out / 'plan.json', c.redact(plan))
    prepared = dict(schema=1, round=r.ROUND, phase=phase, credentialsRead=False,
        nativeSourceManifestSHA256=c.digest(out/'regional-source.json'), predecessorDisclosureSHA256=c.digest(out/'predecessor-disclosure.json'), componentTemplatesSHA256=c.digest(out/'component-templates.json'), toolchainSHA256=c.digest(out/'toolchain.json'), sourceCommit=frozen['sourceCommit'], sourceManifestSHA256=c.digest(out / 'source-at-start.json'),
        binariesManifestSHA256=c.digest(out / 'binaries.json'), planSHA256=c.digest(out / 'plan.json'))
    c.dump(out / 'prepared.json', prepared)
    c.dump(out / 'prerelease-source-review.template.json', prepared | dict(decision='PENDING', reviewedBy='',
        standardsReview=dict(path='', sha256=''), specReview=dict(path='', sha256=''), authorizedRoutes=[]))
    c.dump(out / 'execution-authorization.template.json', prepared | dict(decision='PENDING', requestText='',
        requestReference='', eligibilityAuthorizationSHA256=predecessor['authorizationSHA256'],raceAuthorizationSHA256=predecessor['authorizationSHA256'], maxCalls=r.MAX_CALLS, maxAttempts=r.MAX_ATTEMPTS, maxUSD=None,
        standardCredentialScopes={}, networkRoutes={}, processReservationsMicrousd={}, componentAuthorizationSHA256={},
        predecessor=c.doc(out/'predecessor-disclosure.json'),newBudgetMicrousd=r.NEW_RESERVED,campaignMaxCalls=r.CAMPAIGN_CALLS,campaignMaxAttempts=r.CAMPAIGN_ATTEMPTS,automaticProcessReruns=False))
    guard(out)
    return prepared


def guard(out):
    prepared = c.doc(out / 'prepared.json')
    frozen = c.doc(out / 'source-at-start.json')
    if prepared.get('round')!=r.ROUND:
        raise RuntimeError('only fixed active round2 is executable; original1 remains closed')
    import final_predecessor
    if c.digest(out/'predecessor-disclosure.json')!=prepared['predecessorDisclosureSHA256'] or c.doc(out/'predecessor-disclosure.json')!=final_predecessor.disclosure():
        raise RuntimeError('closed predecessor facts changed')
    if c.digest(out/'component-templates.json')!=prepared['componentTemplatesSHA256']:
        raise RuntimeError('frozen native component template inventory changed')
    for leaf,sha in c.doc(out/'component-templates.json').items():
        if c.digest(out/leaf)!=sha:
            raise RuntimeError('frozen native component template bytes changed')
    if c.digest(out/'toolchain.json')!=prepared['toolchainSHA256']:
        raise RuntimeError('frozen build-toolchain manifest changed')
    for executable,sha in c.doc(out/'toolchain.json')['executables'].items():
        for label,value in [('$REPO',str(c.ROOT)),('$COMMON',str(c.common_root())),('$HOME',str(Path.home()))]:
            executable=executable.replace(label,value)
        if c.digest(Path(executable))!=sha:
            raise RuntimeError('reviewed build tool executable changed')
    if frozen['files'] != sources() or frozen['sourceCommit'] != c.git('rev-parse', 'HEAD'):
        raise RuntimeError('frozen complete source commit/set changed')
    for leaf, field in [('source-at-start.json', 'sourceManifestSHA256'), ('binaries.json', 'binariesManifestSHA256'), ('plan.json', 'planSHA256')]:
        if c.digest(out / leaf) != prepared[field]:
            raise RuntimeError('frozen source/binary/plan manifest changed')
    for name, sha in c.doc(out / 'binaries.json').items():
        if c.digest(out / name) != sha:
            raise RuntimeError('frozen binary changed')
    if c.digest(out/'regional-source.json') != prepared['nativeSourceManifestSHA256']:
        raise RuntimeError('frozen native execution source manifest changed')
    # Use the actual frozen native selector as well as operator sources: it owns
    # compressed captures and other non-Go/JSON dependencies and detects additions.
    current = json.loads(subprocess.check_output([str(out/'regionalverify'), '-source'], cwd=c.ROOT, env=c.ENV, timeout=30))
    if current != c.doc(out/'regional-source.json'):
        raise RuntimeError('complete native execution source set changed before key')
    expected = c.redact(r.inventory(out) | dict(phase=prepared['phase'], routes={n: j for n, j in r.jobs(out).items() if j['phase'] == prepared['phase']}))
    if c.doc(out / 'plan.json') != expected:
        raise RuntimeError('fixed reconstructed job plan changed')
    return prepared


def approved(out, job):
    return approved_metadata(out, job, guard(out))


def approved_metadata(out, job, prepared):
    review = c.doc(out / 'prerelease-source-review.json')
    auth = c.doc(out / 'execution-authorization.json')
    for value in (review, auth):
        if value.get('decision') != 'APPROVED' or any(value.get(k) != v for k, v in prepared.items() if k != 'credentialsRead'):
            raise RuntimeError('new exact-source/binary/plan independent review and human approval required')
    if not review.get('reviewedBy') or job['name'] not in review.get('authorizedRoutes', []):
        raise RuntimeError('explicit reviewed own job required')
    for field in ('standardsReview', 'specReview'):
        proof = review.get(field, {})
        if not proof.get('path') or c.digest(path(proof['path'])) != proof.get('sha256'):
            raise RuntimeError('both retained exact independent review files required')
    from decimal import Decimal
    maximum = Decimal(str(auth.get('maxUSD')))
    if not maximum.is_finite() or maximum != 30:
        raise RuntimeError('existing30USD campaign ceiling and new bounded paid approval required')
    if (type(auth.get('maxCalls')) is not int or auth['maxCalls'] != r.MAX_CALLS or
        type(auth.get('maxAttempts')) is not int or auth['maxAttempts'] != r.MAX_ATTEMPTS or
        auth.get('automaticProcessReruns') is not False or not auth.get('requestText') or not auth.get('requestReference')):
        raise RuntimeError('new explicit722/758 fixed-round human reply required')
    prior=c.doc(out/'predecessor-disclosure.json')
    import final_predecessor
    if prior!=final_predecessor.disclosure():raise RuntimeError('fixed immutable first-cohort history required')
    if (auth.get('predecessor')!=prior or auth.get('newBudgetMicrousd')!=r.NEW_RESERVED or
        auth.get('campaignMaxCalls')!=r.CAMPAIGN_CALLS or auth.get('campaignMaxAttempts')!=r.CAMPAIGN_ATTEMPTS or
        auth.get('eligibilityAuthorizationSHA256')!=prior['authorizationSHA256'] or auth.get('raceAuthorizationSHA256')!=prior['authorizationSHA256']):
        raise RuntimeError('same closed predecessor, original eligibility/race and new bounded paid scope required')
    scope = auth.get('standardCredentialScopes', {}).get(job['sourceVariable'], {})
    if scope != prior['credentialScopes'].get(job['sourceVariable']):
        raise RuntimeError('same previously confirmed own credential service/account/purpose required')
    if (scope.get('provider') != job['provider'] or scope.get('purpose') != 'standard' or
        scope.get('confirmed') is not True or not scope.get('accountAlias') or not scope.get('regionAlias')):
        raise RuntimeError('own confirmed standard provider/account/region scope required')
    network = auth.get('networkRoutes', {}).get(job['name'])
    if network not in ('direct', 'http://127.0.0.1:7890') or job['provider'] in ('zhipu', 'minimax-cn', 'deepseek') and network != 'direct':
        raise RuntimeError('explicit direct or fixed loopback proxy route required')
    reserve = auth.get('processReservationsMicrousd', {}).get(job['name'])
    if type(reserve) is not int or reserve <= 0:
        raise RuntimeError('positive conservative process USD reservation required')
    return prepared, auth, scope, reserve
