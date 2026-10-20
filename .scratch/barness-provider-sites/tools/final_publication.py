"""Only this finite round's new canonical full22 GLM originals may publish."""
import final_contract as r
import final_native
c=r.c


def require(route):
    return final_native.publication_receipt(route)


def validate(index, views=None):
    import renewal_common as u
    import renewal_glm as g
    from repair_publication import zipped, FIELDS
    if (index.get('collectionRound')!=r.ROUND or set(index.get('routes',{}))!=set(r.ROUTES) or
        set(index.get('launcherReceiptSHA256',{}))!=set(r.ROUTES)):
        raise RuntimeError('new own three-route GLM publication index required')
    commits=set();usage={};originals={};pins={}
    for route in r.ROUTES:
        receipt,sha=require(route)
        if index['launcherReceiptSHA256'][route]!=sha:
            raise RuntimeError('new canonical GLM receipt bytes changed')
        root=receipt['collectionRoot'].replace('$REPO/','')
        base=u.bundle(root+'/'+receipt['baseName'],'formal-glm-report.json');files=u.tree(base)
        captured={'schema':1,**{key:u.sha(files[leaf]) for key,leaf in FIELDS.items()}}
        if receipt.get('capturePins')!=captured:
            raise RuntimeError('new original capture changed after actual launcher')
        row=index['routes'][route];prefix=base.name+'/'
        selected,members=zipped(row,[prefix+n for n in files])
        if (members!={prefix+n for n in files} or any(selected[prefix+n]!=raw for n,raw in files.items()) or
            row.get('reportPath')!=prefix+'formal-glm-report.json' or row.get('reportSHA256')!=captured['formalReportSha256']):
            raise RuntimeError('new immutable GLM archive must preserve all canonical bytes')
        report=u.json_doc(base/'formal-glm-report.json')
        if index.get('sourceManifestSHA256')!=report.get('sourceManifestHash'):
            raise RuntimeError('new own model-source identity changed')
        usage[route]={k:row[k] for k in ('archiveSHA256','reportSHA256')}|g.known_usage(files,report)
        commits.add(receipt['sourceCommit']);originals[route]=files;pins[route]=captured
    if len(commits)!=1 or index.get('originalLiveSourceCommit')!=next(iter(commits)):
        raise RuntimeError('new all-three GLM actual source commit required')
    expected=dict(schema=1,sourceCommit=next(iter(commits)),basis=g.USAGE_BASIS,routes=usage)
    if u.json_doc(c.ROOT/g.BASE/'known-usage.json')!=expected:
        raise RuntimeError('new independently reconciled known usage changed')
    if u.file_bytes(c.ROOT/g.BASE/'live-captures.sha256')!=(c.digest(c.ROOT/g.BASE/'live-captures.json')+'\n').encode():
        raise RuntimeError('new raw selector companion changed')
    if views is None:
        views=u.json_doc(c.ROOT/g.BASE/'support-views.json')
    if views.get('kind')!='no-paid-derived-glm-support-views' or set(views.get('routes',{}))!=set(r.ROUTES):
        raise RuntimeError('new three own public support views required')
    for route in r.ROUTES:
        row=views['routes'][route];files=originals[route]
        required=['original/'+n for n in files]+['glm-support-derivation.json']
        selected,members=zipped(row,required)
        import json
        derived=json.loads(selected['glm-support-derivation.json'])
        if (row.get('reportPath')!='live-report.json' or row.get('files')!=len(members) or
            {n for n in members if n.startswith('original/')}!={'original/'+n for n in files} or
            any(selected['original/'+n]!=raw for n,raw in files.items())):
            raise RuntimeError('new own support view must retain every original canonical byte')
        fields=dict(schema=1,kind='glm-derived-live-support',route=route,
            rawCaptureIndexSHA256=c.digest(c.ROOT/g.BASE/'live-captures.json'),
            rawArchiveSHA256=index['routes'][route]['archiveSHA256'],originalFormalReportSHA256=pins[route]['formalReportSha256'],
            originalSupportReportSHA256=pins[route]['supportReportSha256'],originalGitCommit=next(iter(commits)))
        if any(derived.get(k)!=v for k,v in fields.items()):
            raise RuntimeError('new support derivation provenance changed')
    return dict(round=r.ROUND,threeOwnPASS=True,credentialsRead=False,paidCalls=0)
