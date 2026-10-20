"""Select exactly33 roles from41 actual records without erasing a failure."""
from pathlib import Path
import recovery_contract as r
import recovery_history as h
import recovery_state as s
import recovery_execute as e
c=r.c

def roles(out):
    s.guard(out);summary=e.summarize()
    if summary['actualProcessCount']!=41 or len(summary['routes'])!=24 or any(v['state']!='COMPLETED' or v['budget']['status']!='PASS' for v in summary['routes']):raise RuntimeError('all24 new full processes must pass;41 old/new records stay retained')
    selected={}
    for combo in r.REGIONAL:
        name='candidate-'+combo;p=c.common_root()/'.evidence/minimax-launches'/(r.old.ROUND+'-'+name+'.json');value=s.doc(p)
        if value.get('state')!='COMPLETED' or value.get('exitCode')!=0 or value.get('budget',{}).get('status')!='PASS' or not value.get('nativeConsumerReceiptSha256'):raise RuntimeError('six genuinely promoted historical candidate roles required')
        selected[name]=dict(round=r.old.ROUND,receiptSHA256=s.pin(p),historical=True)
    import final_publication
    h.preserved()
    for route in r.ROUTES:
        receipt,sha=final_publication.require(route);selected['glm-'+route]=dict(round=r.old.ROUND,receiptSHA256=sha,historical=True,measuredCommit=receipt['sourceCommit'])
    for name,job in r.jobs(out).items():
        receipt=s.doc(e.marker(name));command=out/(name+'-command.json')
        if (receipt.get('sourceCommit')!=s.doc(out/'prepared.json')['sourceCommit'] or receipt.get('collectionRoot')!=c.redact(str(out)) or receipt.get('exitCode')!=0 or receipt.get('commandReceiptSha256')!=s.pin(command) or s.doc(command).get('launcherAudit')!=dict(passed=True,findings=[])):raise RuntimeError('all24 own current3 native process receipts required')
        import recovery_native
        recovery_native.require_consumed(out,job,receipt)
        selected[name]=dict(round=r.ROUND,receiptSHA256=s.pin(e.marker(name)),historical=False)
    if len(selected)!=33:raise RuntimeError('exact33 effective full-process roles required')
    return selected

def own_base(out,name):
    import final_execute
    job=r.jobs(out)[name];receipt=s.doc(e.marker(name));base=final_execute.report_path(out,job).parent
    if receipt.get('reportedFileSha256')!=s.pin(final_execute.report_path(out,job)):raise RuntimeError('own completed report changed')
    return base

def bind_inputs(out,result):
    # Only native derived GLM/Jev views may differ from direct measured roots.
    import final_publication
    import renewal_common as u
    selected=roles(out)
    live=result['liveInputs'];effects=result['effects']
    for name,base in live.items():
        if name.startswith('zhipu-') or name=='typesafe-classifier-jev-1.14.0':continue
        jobname='formal-'+name if name in r.REGIONAL else 'mm-global-formal-m3-'+name.removeprefix('minimax-') if name.startswith('minimax-') else name
        if Path(base)!=own_base(out,jobname):raise RuntimeError('final input must be its new3 canonical own capture')
    for name,base in effects.items():
        if Path(base)!=own_base(out,name):raise RuntimeError('effect must be its current3 canonical own capture')
    protocol=own_base(out,'protocol114');visual=own_base(out,'visual-jev');jev=Path(live['typesafe-classifier-jev-1.14.0'])
    for target,source in [(jev/'protocol',protocol),(jev/'effects',visual)]:
        if u.tree(target)!=u.tree(source):raise RuntimeError('Jev derivation must retain exact current3 protocol and visual trees')
    # Current public importer verifies each immutable raw/derived GLM archive;
    # additionally require its exact indexed original data, not another pass.
    index=u.json_doc(c.ROOT/'ai/live/testdata/glm-formal/live-captures.json');views=u.json_doc(c.ROOT/'ai/live/testdata/glm-formal/support-views.json');final_publication.validate(index,views)
    from repair_publication import zipped
    for route in r.ROUTES:
        base=Path(live['zhipu-'+route]);actual=u.tree(base);archived,members=zipped(views['routes'][route],list(actual))
        if members!=set(actual) or actual!=archived:raise RuntimeError('preserved GLM support input must equal indexed original derivation')
    return selected
