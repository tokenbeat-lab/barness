"""Fresh unchanged GLM raw archives, independent known usage and support views."""
import json
from pathlib import Path
import re
import tempfile
import paid_core as c
import renewal_common as u

BASE='ai/live/testdata/glm-formal/'
RAW_NAMES=[BASE+n for n in ('live-captures.json','live-captures.sha256','known-usage.json')]


def known_usage(files,report):
    total_input=total_output=unknown=attempts=0
    scenarios=report.get('scenarios',[])
    if len(scenarios)!=22 or len({s.get('id') for s in scenarios})!=22:
        raise RuntimeError('complete22 own model-qualified scenarios required')
    for scenario in scenarios:
        path='GLM11-'+scenario['id'].replace('/','-')+'/attempt-1-result.json'
        value=json.loads(files[path]);records=value['result']['metadata']['attempts']
        if len(records)!=1:raise RuntimeError('one actual final Attempt per fixed GLM scenario required')
        for attempt in records:
            attempts+=1;usage=attempt.get('usage',{})
            values=[usage.get(k,0) for k in ('input','cacheRead','cacheWrite','output')]
            if any(type(v) is not int or v<0 for v in values):raise RuntimeError('nonnegative exact token counters required')
            total_input+=sum(values[:3]);total_output+=values[3]
            if attempt.get('usageReporting')!='complete':unknown+=1
    return dict(inputSide=total_input,output=total_output,unknownUsageAttempts=unknown,actualAttempts=attempts)


def capture(out,values):
    snapshot=u.json_doc(c.ROOT/'ai/release/catalog-snapshot.json')
    source=u.sha(u.file_bytes(c.ROOT/'ai/live/testdata/glm-candidate/sources.json'))
    code=u.file_bytes(c.ROOT/'ai/internal/testkit/supportmatrix/glm.go').decode()
    pin=re.search(r'const GLMSourcePin = "([a-f0-9]{64})"',code)
    if pin is None or pin.group(1)!=source:raise RuntimeError('current independently pinned GLM source manifest required')
    entries={};known={};commits=set();paths=[]
    with tempfile.TemporaryDirectory(prefix='raw-validation-',dir=out) as temporary:
        for route in u.ROUTES:
            base=u.bundle(values[route],'formal-glm-report.json');files=u.tree(base)
            report=json.loads(files['formal-glm-report.json'])
            if (report.get('route')!=route or report.get('schema')!=1 or report.get('mode')!='live' or
                report.get('kind')!='glm-formal' or report.get('outcome')!='PASS' or report.get('complete') is not True or
                report.get('auditPassed') is not True or report.get('sourceManifestHash')!=source or
                report.get('catalogHash')!=snapshot['hash'] or report.get('catalogVersion')!=snapshot['catalog']['version'] or
                re.fullmatch(r'[a-f0-9]{40}',report.get('gitCommit','')) is None):
                raise RuntimeError('fresh own current GLM formal identity/source/catalog required')
            receipt_path=c.common_root()/'.evidence/minimax-launches'/('glm-'+route+'.json')
            receipt=u.json_doc(receipt_path)
            if (receipt.get('state')!='COMPLETED' or receipt.get('exitCode')!=0 or receipt.get('credentialsRead') is not True or
                receipt.get('launcherAudit',{}).get('passed') is not True or receipt.get('route')!='glm-'+route or
                receipt.get('sourceCommit')!=report['gitCommit']):
                raise RuntimeError('fresh successful canonical GLM launcher receipt required')
            # A completed process is not authority for subsequently edited
            # captures. These hashes are written only by the successful fixed
            # launcher, and the sealed inventory binds all remaining payloads.
            pins=receipt.get('capturePins',{})
            expected={'schema':1}
            for field,leaf in (('formalReportSha256','formal-glm-report.json'),
                ('integritySha256','glm-integrity.json'),('supportReportSha256','live-report.json'),
                ('executionSourceSha256','glm-execution-source.json')):
                if leaf not in files:raise RuntimeError('original launcher-pinned GLM capture missing')
                expected[field]=u.sha(files[leaf])
            if not isinstance(pins,dict) or type(pins.get('schema')) is not int or pins!=expected:
                raise RuntimeError('GLM original capture bytes differ from canonical launcher pins')
            commit=report['gitCommit'];commits.add(commit)
            copied=Path(temporary)/route/base.name
            for name,raw in files.items():
                dest=copied/name;dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes(raw)
            paths.append(copied)
            archive=out/('glm-'+route+'-'+commit[:12]+'.zip')
            entry=dict(archive=str(archive.relative_to(c.ROOT)),archiveSHA256=u.archive(archive,files,base.name+'/'),
                       reportPath=base.name+'/formal-glm-report.json',reportSHA256=u.sha(files['formal-glm-report.json']))
            entries[route]=entry
            known[route]={k:entry[k] for k in ('archiveSHA256','reportSHA256')}|known_usage(files,report)
        if len(commits)!=1:raise RuntimeError('all three own GLM routes must share actual source commit')
        u.validate_glm(out,paths,'validate-glm-raw')
    commit=next(iter(commits))
    index=dict(schema=1,originalLiveSourceCommit=commit,sourceManifestSHA256=source,routes=entries)
    usage=dict(schema=1,sourceCommit=commit,
        basis='Independent Python final Attempt sums: input+cacheRead+cacheWrite once; output once; reasoning already output. Partial/cancelled/unreported remains unknown; known lower bound, not invoice.',routes=known)
    # Keep recoverable originals in the new task directory; historical archives
    # are immutable. Index bytes and their SHA pin publish only after validation.
    before={n:u.file_bytes(c.ROOT/n) for n in RAW_NAMES}
    c.dump(out/'raw-selector-before.json',{n:raw.decode() for n,raw in before.items()})
    try:
        c.dump(c.ROOT/RAW_NAMES[0],index)
        (c.ROOT/RAW_NAMES[1]).write_text(c.digest(c.ROOT/RAW_NAMES[0])+'\n')
        c.dump(c.ROOT/RAW_NAMES[2],usage)
    except BaseException:
        for name,raw in before.items():(c.ROOT/name).write_bytes(raw)
        raise
    checkpoint=dict(schema=1,sourceCommit=commit,selectorSHA256={n:c.digest(c.ROOT/n) for n in RAW_NAMES},
        catalogHash=snapshot['hash'],sourceManifestSHA256=source,credentialsRead=False,
        next='commit raw selectors/known usage and all importer source, then glm-views; commit views before legacy/Jev freeze')
    c.dump(out/'raw-published.json',checkpoint)
    return checkpoint


def views(out):
    state=u.json_doc(out/'raw-published.json')
    if any(c.digest(c.ROOT/n)!=sha for n,sha in state['selectorSHA256'].items()):
        raise RuntimeError('fresh raw selector checkpoint changed')
    u.committed(RAW_NAMES)
    index=u.json_doc(c.ROOT/RAW_NAMES[0]);entries={};paths=[]
    parent=out/'support-views';parent.mkdir(exist_ok=False,mode=0o700)
    for route in u.ROUTES:
        raw=index['routes'][route]
        if c.digest(u.repo_path(raw['archive']))!=raw['archiveSHA256']:raise RuntimeError('fresh raw immutable archive changed')
        path=parent/route
        u.public(out,'derive-glm-'+route,['go','run','./ai/live/cmd/glmformal','-route',route,'-out',path])
        files=u.tree(path);archive=out/('support-'+route+'.zip')
        entries[route]=dict(archive=str(archive.relative_to(c.ROOT)),archiveSHA256=u.archive(archive,files),
                            reportPath='live-report.json',files=len(files))
        paths.append(path)
    u.validate_glm(out,paths,'validate-glm-support-views')
    target=c.ROOT/BASE/'support-views.json'
    (out/'support-views-before.json').write_bytes(u.file_bytes(target))
    c.dump(target,dict(schema=1,kind='no-paid-derived-glm-support-views',routes=entries))
    result=dict(credentialsRead=False,routes=entries,next='commit support-view selectors before fresh complete legacy/Jev source freeze')
    c.dump(out/'views-published.json',result);return result
