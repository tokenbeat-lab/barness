"""Shared private fixed recovery finalize mechanics; authority values come only from compiled contexts."""
import os
import tempfile
from pathlib import Path
import renewal_common as u

class Finalizer:

    def __init__(self, flow):
        self.flow = flow

    def plan(self, out, selected):
        prepared = self.flow.s.guard(out)
        value = self.flow.s.doc(self.flow.s.path(selected))
        if set(value) != {'live', 'effects'} or set(value['live']) != set(self.flow.r.live_identities()) or set(value['effects']) != set(self.flow.r.EFFECTS):
            raise RuntimeError('exactly23 own identities and four effects required; old17-row finalizer is closed')
        live = {}
        seen = set()
        effects = {}
        for name in self.flow.r.live_identities():
            base = u.bundle(value['live'][name], 'live-report.json')
            if base in seen or self.flow.files.doc(base / 'live-report.json').get('combo') != name:
                raise RuntimeError('one distinct own report per exact full identity required')
            seen.add(base)
            live[name] = str(base)
        for name in self.flow.r.EFFECTS:
            leaf = 'evaluation-report.json' if name.startswith('text-') else 'report.json'
            base = u.bundle(value['effects'][name], leaf)
            if base in seen:
                raise RuntimeError('effects must retain independent own native reports')
            seen.add(base)
            effects[name] = str(base)
        destination = out / 'finalization'
        matrix = destination / 'matrix-candidate.json'
        merge = [str(out / 'supportmatrix'), '-matrix', str(matrix), *live.values()]
        checks = []
        for name, base in effects.items():
            checks.append([str(out / ('chineseeval' if name.startswith('text-') else 'visualeval')), *(['-route', name[5:], '-verify', base] if name.startswith('text-') else ['-verify', base, '-require-live'])])
        checks.append([str(out / 'jevformal'), '-verify', live['typesafe-classifier-jev-1.14.0'], '-require-live'])
        for combo in self.flow.r.REGIONAL:
            checks.append([str(out / 'regionalverify'), '-formal', live[combo]])
        for route in self.flow.r.ROUTES:
            checks.append([str(out / 'minimaxverify'), '-formal', live['minimax-' + route]])
        gate = [str(out / 'releasegate'), '-matrix', str(matrix)]
        for flag, name in [('evaluation-jev', 'text-jev'), ('evaluation-luna', 'text-luna'), ('visual-evaluation-jev', 'visual-jev'), ('visual-evaluation-decisions', 'visual-decisions')]:
            gate += ['-' + flag, effects[name]]
        for base in live.values():
            gate += ['-live', base]
        return dict(round=self.flow.r.ROUND, credentialsRead=False, releaseCertified=False, liveInputs=live, effects=effects, verify=checks, merge=merge, preflight=[*gate, '-preflight', '-out', str(destination / 'preflight')], ordinary=[*gate, '-stage', 'check', '-out', str(destination / 'ordinary')], manualRace=[str(out / 'releasegate'), '-stage', 'race', '-from', str(destination / 'ordinary'), '-out', str(destination / 'race')])

    def finalize(self, out, selected):
        result = self.plan(out, selected)
        summary = self.flow.e.summarize()
        roles = self.flow.p.bind_inputs(out, result)
        import final_publication
        if not self.flow.r.RETAINED_GENERIC:final_publication.validate(self.flow.files.doc(self.flow.r.c.ROOT / 'ai/live/testdata/glm-formal/live-captures.json'))
        destination = out / 'finalization'
        destination.mkdir(mode=448)
        self.flow.r.c.dump(destination / 'plan.json', self.flow.r.c.redact(result))
        self.flow.r.c.dump(destination / self.flow.r.FINAL_SELECTION_FILE, dict(roles=roles, accounting=summary))
        before = {base: {n: u.sha(raw) for n, raw in self.flow.files.tree(Path(base)).items()} for base in [*result['liveInputs'].values(), *result['effects'].values()]}
        self.flow.r.c.dump(destination / 'input-files.json', self.flow.r.c.redact(before))
        for i, argv in enumerate(result['verify']):
            if self.flow.s.run(destination, 'verify-' + str(i), argv, timeout=180) != 0:
                raise RuntimeError('own final native consumer failed; private candidate is not published')
            self.flow.s.guard(out)
        matrix = destination / 'matrix-candidate.json'
        target = self.flow.r.c.ROOT / 'ai/live/support-matrix.json'
        original_bytes = self.flow.files.read(target)
        original = self.flow.files.doc(target)
        if len(original['rows']) != 23 or {row['combo'] for row in original['rows']} != set(self.flow.r.live_identities()):
            raise RuntimeError('formal base matrix must already contain exactly23 independent identities')
        matrix.write_bytes(self.flow.files.read(self.flow.r.c.ROOT / 'ai/live/support-matrix.json'))
        if self.flow.s.run(destination, 'merge-all23', result['merge'], timeout=240) != 0:
            raise RuntimeError('public all23 importer refused; production matrix remains unchanged')
        updated = self.flow.files.doc(matrix)
        rows = updated['rows']
        if len(rows) != 23 or {row['combo'] for row in rows} != set(self.flow.r.live_identities()) or any((not row.get('allPassedAt') for row in rows)):
            raise RuntimeError('all23 own current full passes required')
        after = {base: {n: u.sha(raw) for n, raw in self.flow.files.tree(Path(base)).items()} for base in before}
        if before != after:
            raise RuntimeError('selected immutable input bytes changed during native consumption')
        self.flow.s.guard(out)
        if self.flow.files.read(target) != original_bytes:
            raise RuntimeError('tracked matrix changed during all23 public import')
        (destination / 'matrix-before.json').write_bytes(original_bytes)
        fd, name = tempfile.mkstemp(prefix='.provider-sites-matrix-', dir=target.parent)
        try:
            with os.fdopen(fd, 'wb') as stream:
                stream.write(self.flow.files.read(matrix))
                stream.flush()
                os.fsync(stream.fileno())
            os.chmod(name, target.stat().st_mode & 511)
            os.replace(name, target)
        finally:
            if os.path.exists(name):
                os.unlink(name)
        result.update(privateMatrixSHA256=self.flow.files.digest(matrix), inputFilesSHA256=self.flow.files.digest(destination / 'input-files.json'), trackedMatrixSHA256=self.flow.files.digest(target), automaticRace=False, next='Commit derived matrix/selectors, rebuild final CLI and freeze complete source/binary/input set before fresh check/ordinary, READY_FOR_RACE and the one explicitly selected formal race.')
        self.flow.r.c.dump(destination / 'result.json', self.flow.r.c.redact(result))
        return result
