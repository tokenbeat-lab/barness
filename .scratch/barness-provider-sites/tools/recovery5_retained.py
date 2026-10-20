"""Eight unchanged generic captures, proven against all measured source and wire bytes."""
import hashlib
from pathlib import Path
import recovery5_history as h
from recovery_flow_source import SourceFreeze
r=h.r;c=r.c
# Exact four reviewed budget-test blobs; the HOME fix only aliases its owned
# classifier evidence roots. Guard/wire/ledger compatibility stays pinned below.
BUDGET_SOURCE_COMMIT='8cbaea6c689f264afcb17e3b97fd0de3d8458c25'
COMPAT_SHA256='f5d37cc92abc7d9f4db844e73ad8e00efc63d7c316824ae43afff917280907a7'
BUDGET_TESTS={'ai/live/final_budget_process_test.go','ai/live/final_budget_pipeline_test.go','ai/live/final_budget_plan_replay_test.go','ai/examples/visualeval/final_budget_test.go'}


def source_changes(original,current):
    names=set(original)|set(current);allowed=[]
    for name in sorted(names):
        if original.get(name)==current.get(name):continue
        if name in BUDGET_TESTS:
            import subprocess
            raw=subprocess.check_output(['git','show',BUDGET_SOURCE_COMMIT+':'+name],cwd=c.ROOT,env=c.ENV,stderr=subprocess.DEVNULL)
            if hashlib.sha256(raw).hexdigest()!=current.get(name):raise RuntimeError('reviewed actual budget test source changed: '+name)
        if name=='ai/internal/testkit/finalbudget/estimate.go' or name in BUDGET_TESTS or name.startswith(('ai/internal/testkit/minimaxlive/','ai/live/cmd/minimaxverify/','.scratch/barness-provider-sites/tools/')):
            allowed.append(name)
        else:raise RuntimeError('retained generic complete execution source drift: '+name)
    return allowed


def compatibility(stop,original):
    if h.pin(h.COMPAT)!=COMPAT_SHA256:raise RuntimeError('fixed actual84 compatibility proof changed')
    proof=h.doc(h.COMPAT)
    h.exact(proof,dict(schema=1,kind='PUBLIC_CLIENT_FINAL_WIRE_GUARD_COMPATIBILITY',state='PASS',count=84,routes=8,requestBodyAndOwnedLedgerBytesUnchanged=True),'actual84 public wire/owned-ledger compatibility proof required')
    if proof.get('baselineSourceCommit')!=stop['sourceCommit'] or proof.get('baselineSpendingGuardSourceSHA256')!={n:original[n] for n in r.GUARD_FILES} or proof.get('currentSpendingGuardSourceSHA256')!=r.policy()['spendingGuardSourceSHA256']:raise RuntimeError('compatibility proof must bind exact original/current guard bytes')
    counts={row['name']:row['actualCalls'] for row in stop['processes'] if row['name'] in r.RETAINED_GENERIC}
    if proof.get('routeCounts')!=counts or len(proof.get('cases',[]))!=84 or len(proof.get('originalSourceFiles',{}))!=84:raise RuntimeError('exact8 routes/84 original captures required')
    source_index=h.expand(proof['originalSourceIndex'])
    if h.pin(source_index)!=proof['originalSourceIndexSHA256']:raise RuntimeError('original84 wire source index changed')
    for path,sha in proof['originalSourceFiles'].items():
        if stop['files'].get(path)!=sha or h.pin(h.expand(path))!=sha:raise RuntimeError('compatibility original capture outside closed4 pins')
    for row in proof['cases']:
        if row.get('source','').rsplit(':',1)[0] not in proof['originalSourceFiles']:raise RuntimeError('compatibility case lacks original capture')
        for key in ('ledgerBytesEqual','methodEqual','urlEqual','requestBodyBytesEqual'):
            if row.get(key) is not True:raise RuntimeError('compatibility wire or ledger changed')
        baseline=[];current=[]
        for phase,result in [('baseline',baseline),('current',current)]:
            for kind in ('Wire','Ledger'):
                path=h.expand(row[phase+kind+'File'])
                if h.pin(path)!=row[phase+kind+'FileSHA256']:raise RuntimeError('actual compatibility artifact changed')
                result.append(SourceFreeze(c).read(path,2<<20))
        if baseline[1]!=current[1]:raise RuntimeError('owned USD ledger bytes changed')
        left=h.strict(baseline[0]);right=h.strict(current[0])
        if not isinstance(left,list) or not isinstance(right,list) or len(left)!=1 or len(right)!=1:raise RuntimeError('one final wire per compatibility case required')
        for key in ('method','url','requestBody'):
            if left[0].get(key)!=right[0].get(key):raise RuntimeError('actual final request differs')
        sha=hashlib.sha256(left[0]['requestBody'].encode()).hexdigest()
        if sha!=row.get('baselineRequestSHA256') or sha!=row.get('currentRequestSHA256'):raise RuntimeError('native final request body pin differs')
    return h.pin(h.COMPAT)


def verify():
    stop=h.doc(h.STOP);archive=h.doc(r.BASE/'recovery4/source-at-start.json');oldfiles=archive['files'];freeze=SourceFreeze(c)
    if archive['sourceCommit']!=stop['sourceCommit']:raise RuntimeError('retained source commit differs')
    # The historical archive is checked against its Git object, never current source.
    import subprocess
    commit=archive['sourceCommit']
    if len(commit)!=40 or any(v not in '0123456789abcdef' for v in commit):raise RuntimeError('historical source commit format differs')
    for name in oldfiles:
        if not name or '\n' in name or '\r' in name or Path(name).is_absolute() or '..' in Path(name).parts:raise RuntimeError('historical source path differs')
    queries=''.join(commit+':'+name+'\n' for name in oldfiles).encode()
    raw=subprocess.check_output(['git','cat-file','--batch'],input=queries,cwd=c.ROOT,env=c.ENV,stderr=subprocess.DEVNULL,timeout=30)
    if len(raw)>512<<20:raise RuntimeError('historical Git source byte bound exceeded')
    offset=0
    for name,sha in oldfiles.items():
        end=raw.find(b'\n',offset);header=raw[offset:end].split()
        if end<offset or len(header)!=3 or header[1]!=b'blob':raise RuntimeError('historical source blob unavailable')
        size=int(header[2]);start=end+1
        if size<0 or size>64<<20 or hashlib.sha256(raw[start:start+size]).hexdigest()!=sha or raw[start+size:start+size+1]!=b'\n':raise RuntimeError('original generic measured Git source differs')
        offset=start+size+1
    if offset!=len(raw):raise RuntimeError('unexpected historical Git source bytes')
    changed=source_changes(oldfiles,freeze.files());proof=compatibility(stop,oldfiles)
    roles={}
    for name in r.RETAINED_GENERIC:
        row=next(v for v in stop['processes'] if v['name']==name)
        if row['exitCode']!=0 or row['parentBudget']['status']!='PASS':raise RuntimeError('retained generic needs original full PASS')
        parent=h.f.owned('.evidence/minimax-launches/'+h.OLD+'-'+name+'.json')
        roles[name]=dict(round=h.OLD,receiptSHA256=h.pin(parent),historical=True,base=str(h.expand(row['nativeBundle'])),sourceCommit=stop['sourceCommit'],nativeReportSHA256=row['nativeReportSHA256'],nativeConsumerReceiptSHA256=row['nativeConsumerReceiptSHA256'])
    return dict(roles=roles,compatibilitySHA256=proof,sourceDifferences=changed,wholeReservedMicrousd=12400000,refundMicrousd=0)
