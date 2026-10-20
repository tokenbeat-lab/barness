"""Exactly23 own current inputs and four effects through the public consumers."""
import os
import tempfile
from pathlib import Path
import final_contract as r
import final_state as s
import final_execute as e
import renewal_common as u
c=r.c


def plan(out, selected):
    prepared=s.guard(out)
    if prepared['phase']!='remaining':
        raise RuntimeError('final23 input plan requires the final remaining-source freeze')
    value=u.json_doc(s.path(selected))
    if set(value)!= {'live','effects'} or set(value['live'])!=set(r.live_identities()) or set(value['effects'])!=set(r.EFFECTS):
        raise RuntimeError('exactly23 own identities and four effects required; old17-row finalizer is closed')
    live={};seen=set();effects={}
    for name in r.live_identities():
        base=u.bundle(value['live'][name],'live-report.json')
        if base in seen or u.json_doc(base/'live-report.json').get('combo')!=name:
            raise RuntimeError('one distinct own report per exact full identity required')
        seen.add(base);live[name]=str(base)
    for name in r.EFFECTS:
        leaf='evaluation-report.json' if name.startswith('text-') else 'report.json'
        base=u.bundle(value['effects'][name],leaf)
        if base in seen:
            raise RuntimeError('effects must retain independent own native reports')
        seen.add(base);effects[name]=str(base)
    destination=out/'finalization';matrix=destination/'matrix-candidate.json'
    merge=[str(out/'supportmatrix'),'-matrix',str(matrix),*live.values()]
    checks=[]
    for name,base in effects.items():
        checks.append([str(out/('chineseeval' if name.startswith('text-') else 'visualeval')),
            *(['-route',name[5:],'-verify',base] if name.startswith('text-') else ['-verify',base,'-require-live'])])
    checks.append([str(out/'jevformal'),'-verify',live['typesafe-classifier-jev-1.14.0'],'-require-live'])
    for combo in r.REGIONAL:
        checks.append([str(out/'regionalverify'),'-formal',live[combo]])
    for route in r.ROUTES:
        checks.append([str(out/'minimaxverify'),'-formal',live['minimax-'+route]])
    gate=[str(out/'releasegate'),'-matrix',str(matrix)]
    for flag,name in [('evaluation-jev','text-jev'),('evaluation-luna','text-luna'),('visual-evaluation-jev','visual-jev'),('visual-evaluation-decisions','visual-decisions')]:
        gate+=['-'+flag,effects[name]]
    for base in live.values():gate+=['-live',base]
    return dict(round=r.ROUND,credentialsRead=False,releaseCertified=False,liveInputs=live,effects=effects,verify=checks,
        merge=merge,preflight=[*gate,'-preflight','-out',str(destination/'preflight')],
        ordinary=[*gate,'-stage','check','-out',str(destination/'ordinary')],
        manualRace=[str(out/'releasegate'),'-stage','race','-from',str(destination/'ordinary'),'-out',str(destination/'race')])


def finalize(out, selected):
    result=plan(out,selected)
    summary=e.summarize()
    if len(summary['routes'])!=33 or any(v['state']!='COMPLETED' or v.get('budget',{}).get('status')!='PASS' for v in summary['routes']):
        raise RuntimeError('all33 own full-process facts must PASS; no capture or scenario splicing')
    import final_publication
    final_publication.validate(u.json_doc(c.ROOT/'ai/live/testdata/glm-formal/live-captures.json'))
    destination=out/'finalization';destination.mkdir(mode=0o700)
    c.dump(destination/'plan.json',c.redact(result))
    before={base:{n:u.sha(raw) for n,raw in u.tree(Path(base)).items()} for base in [*result['liveInputs'].values(),*result['effects'].values()]}
    c.dump(destination/'input-files.json',c.redact(before))
    for i,argv in enumerate(result['verify']):
        if s.run(destination,'verify-'+str(i),argv,timeout=180)!=0:
            raise RuntimeError('own final native consumer failed; private candidate is not published')
        s.guard(out)
    matrix=destination/'matrix-candidate.json'
    target=c.ROOT/'ai/live/support-matrix.json'
    original_bytes=u.file_bytes(target)
    original=u.json_doc(target)
    if len(original['rows'])!=23 or {row['combo'] for row in original['rows']}!=set(r.live_identities()):
        raise RuntimeError('formal base matrix must already contain exactly23 independent identities')
    matrix.write_bytes(u.file_bytes(c.ROOT/'ai/live/support-matrix.json'))
    if s.run(destination,'merge-all23',result['merge'],timeout=240)!=0:
        raise RuntimeError('public all23 importer refused; production matrix remains unchanged')
    updated=u.json_doc(matrix)
    rows=updated['rows']
    if len(rows)!=23 or {row['combo'] for row in rows}!=set(r.live_identities()) or any(not row.get('allPassedAt') for row in rows):
        raise RuntimeError('all23 own current full passes required')
    after={base:{n:u.sha(raw) for n,raw in u.tree(Path(base)).items()} for base in before}
    if before!=after:
        raise RuntimeError('selected immutable input bytes changed during native consumption')
    s.guard(out)
    if u.file_bytes(target)!=original_bytes:
        raise RuntimeError('tracked matrix changed during all23 public import')
    (destination/'matrix-before.json').write_bytes(original_bytes)
    fd,name=tempfile.mkstemp(prefix='.provider-sites-matrix-',dir=target.parent)
    try:
        with os.fdopen(fd,'wb') as stream:
            stream.write(u.file_bytes(matrix));stream.flush();os.fsync(stream.fileno())
        os.chmod(name,target.stat().st_mode&0o777)
        os.replace(name,target)
    finally:
        if os.path.exists(name):os.unlink(name)
    result.update(privateMatrixSHA256=c.digest(matrix),inputFilesSHA256=c.digest(destination/'input-files.json'),
        trackedMatrixSHA256=c.digest(target),automaticRace=False,
        next='Commit derived matrix/selectors, rebuild final CLI and freeze complete source/binary/input set before fresh check/ordinary, READY_FOR_RACE and the one explicitly selected formal race.')
    c.dump(destination/'result.json',c.redact(result))
    return result
