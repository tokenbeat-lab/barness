"""Shared private fixed3/fixed4 state mechanics; authority values come only from compiled contexts."""
import json
from pathlib import Path
import shutil
import subprocess
import sys
import final_state as shared
from recovery_flow_source import SourceFreeze
PINNED = {'source-at-start.json': 'sourceManifestSHA256', 'regional-source.json': 'nativeSourceManifestSHA256', 'predecessor.json': 'predecessorSHA256', 'preserved-glm.json': 'preservedGLMSHA256', 'toolchain.json': 'toolchainSHA256', 'binaries.json': 'binariesManifestSHA256', 'component-templates.json': 'componentTemplatesSHA256', 'plan.json': 'planSHA256'}

class State:

    def __init__(self, flow):
        self.flow = flow
    run = staticmethod(shared.run)

    def path(self, value):
        if not value:
            raise RuntimeError('owned recovery output required')
        p = Path(value)
        p = p if p.is_absolute() else self.flow.r.c.ROOT / p
        p = self.flow.h.expand(str(p))
        parts = p.relative_to(self.flow.r.c.common_root()).parts
        if len(parts) < 3 or parts[:2] != ('.evidence', 'provider-sites'):
            raise RuntimeError('new owned .evidence/provider-sites output required')
        return p

    def doc(self, p):
        return self.flow.h.doc(p)

    def pin(self, p):
        return self.flow.h.pin(p)

    def prepare(self, out):
        predecessor = self.flow.h.facts()
        preserved = self.flow.h.preserved()
        frozen = SourceFreeze(self.flow.r.c).freeze()
        out.mkdir(parents=True, exist_ok=False, mode=448)
        for leaf, value in [('predecessor.json', predecessor), ('preserved-glm.json', preserved), ('source-at-start.json', frozen)]:
            self.flow.r.c.dump(out / leaf, value)
        if self.run(out, 'go-version', ['go', 'version']) != 0:
            raise RuntimeError('build tool identity unavailable')
        toolchain = dict(goVersion=(out / 'go-version.log').read_text().strip(), pythonVersion=sys.version, executables={str(Path(shutil.which('go')).resolve()): SourceFreeze(self.flow.r.c).digest(Path(shutil.which('go')).resolve()), str(Path(sys.executable).resolve()): SourceFreeze(self.flow.r.c).digest(Path(sys.executable).resolve())})
        self.flow.r.c.dump(out / 'toolchain.json', self.flow.r.c.redact(toolchain))
        if self.run(out, 'regional-source', ['go', 'run', './ai/live/cmd/regionalverify', '-source']) != 0:
            raise RuntimeError('native current source unavailable')
        source = json.loads((out / 'regional-source.log').read_text())
        self.flow.r.c.dump(out / 'regional-source.json', source)
        builds = [('live.test', ['go', 'test', '-c', '-tags', 'live', '-ldflags', '-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource=' + source['sourceSha256'], '-o', out / 'live.test', './ai/live'])]
        packages = dict(regionalpromote='./ai/live/cmd/regionalpromote', regionalverify='./ai/live/cmd/regionalverify', minimaxverify='./ai/live/cmd/minimaxverify', glmformal='./ai/live/cmd/glmformal', jevformal='./ai/live/cmd/jevformal', supportmatrix='./ai/live/cmd/supportmatrix', chineseeval='./ai/examples/chineseeval/cmd/chineseeval', visualeval='./ai/examples/visualeval', releasegate='./ai/release/cmd/releasegate')
        builds += [(n, ['go', 'build', '-o', out / n, p]) for n, p in packages.items()]
        for name, argv in builds:
            if self.run(out, 'build-' + name, argv, timeout=600) != 0:
                raise RuntimeError('keyless current reviewed build failed')
        self.flow.r.c.dump(out / 'binaries.json', {name: SourceFreeze(self.flow.r.c).digest(out / name) for name, _ in builds})
        templates = self.flow.n.templates(out)
        self.flow.r.c.dump(out / 'component-templates.json', templates)
        self.flow.r.c.dump(out / 'plan.json', self.flow.r.c.redact(self.flow.r.inventory(out)))
        prepared = dict(schema=1, round=self.flow.r.ROUND, validationRound=self.flow.r.ROUND, nativeRenewalId=self.flow.r.ROUND, credentialsRead=False, sourceCommit=frozen['sourceCommit'], **{self.flow.r.PREDECESSOR_SEAL_FIELD: predecessor[self.flow.r.PREDECESSOR_SEAL_FIELD]}, **self.flow.h.prepare_metadata(predecessor), originalCampaignApprovalSHA256=predecessor['authorizationSHA256'], **self.flow.r.policy())
        for leaf, field in PINNED.items():
            prepared[field] = self.pin(out / leaf)
        self.flow.r.c.dump(out / 'prepared.json', prepared)
        self.flow.r.c.dump(out / 'prerelease-source-review.template.json', prepared | dict(decision='PENDING', reviewedBy='', standardsReview=dict(path='', sha256=''), specReview=dict(path='', sha256=''), authorizedRoutes=[]))
        auth = prepared | dict(decision='PENDING', requestText='', requestReference='', maxCalls=self.flow.r.MAX_CALLS, maxAttempts=self.flow.r.MAX_ATTEMPTS, campaignMaxCalls=self.flow.r.CAMPAIGN_CALLS, campaignMaxAttempts=self.flow.r.CAMPAIGN_ATTEMPTS, newReservedMicrousd=self.flow.r.NEW_USD, retainedReservedMicrousd=self.flow.r.RETAINED_USD, maxUSD=30, automaticProcessReruns=False, predecessor=predecessor, eligibilityAuthorizationSHA256=predecessor['authorizationSHA256'], raceAuthorizationSHA256=predecessor['authorizationSHA256'], processReservationsMicrousd=self.flow.r.RESERVES, usdLedgers={n: self.flow.r.c.redact(j['usdLedger']) for n, j in self.flow.r.jobs(out).items()}, networkRoutes={n: j['networkRoute'] for n, j in self.flow.r.jobs(out).items()}, standardCredentialScopes=predecessor['credentialScopes'], componentAuthorizationSHA256={}, actualHumanApprovalSHA256='')
        auth.update(self.flow.h.scope_metadata(prepared))
        self.flow.r.c.dump(out / 'execution-authorization.template.json', auth)
        self.flow.r.c.dump(out / self.flow.r.ACTUAL_HUMAN_LEAF.replace('.json', '.template.json'), {k: v for k, v in auth.items() if k not in ('standardCredentialScopes', 'componentAuthorizationSHA256', 'networkRoutes', 'predecessor')})
        self.guard(out)
        return prepared

    def guard(self, out):
        prepared = self.doc(out / 'prepared.json')
        if any((prepared.get(k) != v for k, v in dict(round=self.flow.r.ROUND, validationRound=self.flow.r.ROUND, nativeRenewalId=self.flow.r.ROUND, **self.flow.r.policy()).items())):
            raise RuntimeError('fixed recovery authority and unchanged2 spending policy required')
        for leaf, field in PINNED.items():
            if self.pin(out / leaf) != prepared[field]:
                raise RuntimeError('frozen recovery input changed: ' + leaf)
        if self.doc(out / 'predecessor.json') != self.flow.h.facts() or self.doc(out / 'preserved-glm.json') != self.flow.h.preserved():
            raise RuntimeError('immutable prior facts changed')
        frozen = self.doc(out / 'source-at-start.json')
        if frozen['files'] != SourceFreeze(self.flow.r.c).files() or frozen['sourceCommit'] != self.flow.r.c.git('rev-parse', 'HEAD') or prepared['sourceCommit'] != frozen['sourceCommit']:
            raise RuntimeError('complete source commit/set changed')
        for name, sha in self.doc(out / 'binaries.json').items():
            if self.pin(out / name) != sha:
                raise RuntimeError('current frozen binary changed')
        for executable, sha in self.doc(out / 'toolchain.json')['executables'].items():
            value = executable
            for label, p in (('$REPO', self.flow.r.c.ROOT), ('$COMMON', self.flow.r.c.common_root()), ('$HOME', Path.home())):
                value = value.replace(label, str(p))
            if SourceFreeze(self.flow.r.c).digest(Path(value)) != sha:
                raise RuntimeError('frozen build executable changed')
        for leaf, sha in self.doc(out / 'component-templates.json').items():
            if self.pin(out / leaf) != sha:
                raise RuntimeError('component template changed')
        current = json.loads(subprocess.check_output([out / 'regionalverify', '-source'], cwd=self.flow.r.c.ROOT, env=self.flow.r.c.ENV, timeout=30))
        if current != self.doc(out / 'regional-source.json') or self.doc(out / 'plan.json') != self.flow.r.c.redact(self.flow.r.inventory(out)):
            raise RuntimeError('native source or reconstructed fixed24 plan changed')
        return prepared

    def approved(self, out, job):
        prepared = self.guard(out)
        review = self.doc(out / 'prerelease-source-review.json')
        auth = self.doc(out / 'execution-authorization.json')
        human = self.flow.h.actual_approval(prepared)
        for value in (review, auth):
            if value.get('decision') != 'APPROVED' or any((value.get(k) != v for k, v in prepared.items())):
                raise RuntimeError('new exact independent review and human source approval required')
        if not review.get('reviewedBy') or set(review.get('authorizedRoutes', [])) != set(self.flow.r.jobs(out)):
            raise RuntimeError('all24 exact independently reviewed routes required')
        for field in ('standardsReview', 'specReview'):
            proof = review[field]
            if self.pin(self.path(proof['path'])) != proof['sha256']:
                raise RuntimeError('retained independent review changed')
        prior = self.doc(out / 'predecessor.json')
        for key, value in dict(**self.flow.h.scope_metadata(prepared), maxCalls=self.flow.r.MAX_CALLS, maxAttempts=self.flow.r.MAX_ATTEMPTS, campaignMaxCalls=self.flow.r.CAMPAIGN_CALLS, campaignMaxAttempts=self.flow.r.CAMPAIGN_ATTEMPTS, newReservedMicrousd=self.flow.r.NEW_USD, retainedReservedMicrousd=self.flow.r.RETAINED_USD, maxUSD=30, automaticProcessReruns=False, predecessor=prior, eligibilityAuthorizationSHA256=prior['authorizationSHA256'], raceAuthorizationSHA256=prior['authorizationSHA256'], processReservationsMicrousd=self.flow.r.RESERVES, usdLedgers={n: self.flow.r.c.redact(j['usdLedger']) for n, j in self.flow.r.jobs(out).items()}, networkRoutes={n: j['networkRoute'] for n, j in self.flow.r.jobs(out).items()}, standardCredentialScopes=prior['credentialScopes'], actualHumanApprovalSHA256=self.pin(self.flow.h.HUMAN), requestText=human['requestText'], requestReference=human['requestReference']).items():
            if auth.get(key) != value:
                raise RuntimeError('fixed exact new human scope/budget/proxy/prefix required: ' + key)
        scope = auth['standardCredentialScopes'][job['sourceVariable']]
        if scope.get('provider') != job['provider'] or scope.get('purpose') != 'standard' or scope.get('confirmed') is not True:
            raise RuntimeError('original own standard service qualification required')
        return (prepared, auth, scope, job['reservedMicrousd'])

    def checks(self, out):
        self.guard(out)
        codes = {}
        pins = {}
        for name, argv, timeout in [('vet', ['go', 'vet', './...'], 600), ('vet-live', ['go', 'vet', '-tags', 'live', './...'], 600), ('no-key-live', ['go', 'test', '-tags', 'live', '-count=1', '-timeout=20m', './ai/live'], 1260)]:
            codes[name] = self.run(out, name, argv, timeout=timeout, env={'BARNESS_AI_EVIDENCE_DIR': str(out / 'no-key-live')})
            pins[name] = self.pin(out / (name + '-command.json'))
            if codes[name] != 0:
                break
        result = dict(passed=len(codes) == 3 and all((v == 0 for v in codes.values())), commands=codes, commandReceipts=pins, credentialsRead=False)
        self.flow.r.c.dump(out / 'checks.json', result)
        self.guard(out)
        if not result['passed']:
            raise RuntimeError('keyless checks failed')
        return result
