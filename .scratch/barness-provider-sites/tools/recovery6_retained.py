"""Eight exact old08 inputs require a newly pinned84-wire/header/ledger proof."""
import hashlib
import subprocess
from pathlib import Path
import recovery6_history as h
import recovery5_retained as old
from recovery_flow_source import SourceFreeze
r=h.r;c=r.c
COMPAT=r.BASE/'recovery6-preparation/compatibility-eight.json'
COMPAT_SHA256='PENDING_FINAL_ROOT_84_REPROOF'
REPAIR_FILES={'ai/live/regional_harness_test.go','ai/internal/testkit/supportmatrix/regional_contract.go'}
BASELINE_COMMIT='8c8342606390d37dff041f11de7ff3b8d82d12d5'

def reviewed_sources(proof,current):
    baseline=h.doc(r.BASE/'recovery5/source-at-start.json')
    if baseline.get('sourceCommit')!=BASELINE_COMMIT:raise RuntimeError('exact frozen fifth source archive required')
    changes={name for name in set(baseline['files'])|set(current) if baseline['files'].get(name)!=current.get(name) and (name.startswith('ai/') or name in ('go.mod','go.sum'))}
    reviewed=proof.get('reviewedSourceDifferences')
    if not isinstance(reviewed,dict) or set(reviewed)!=changes or not REPAIR_FILES<=changes:raise RuntimeError('new exact reviewed two-file/native6 source differences required')
    commit=proof.get('currentSourceCommit','')
    if len(commit)!=40 or any(x not in '0123456789abcdef' for x in commit):raise RuntimeError('actual reproof source commit required')
    for name,sha in reviewed.items():
        if Path(name).is_absolute() or '..' in Path(name).parts or '\n' in name:raise RuntimeError('noncanonical reviewed source path')
        if current.get(name)!=sha:raise RuntimeError('reviewed source blob changed: '+name)
        raw=subprocess.check_output(['git','show',commit+':'+name],cwd=c.ROOT,env=c.ENV,stderr=subprocess.DEVNULL,timeout=30)
        if hashlib.sha256(raw).hexdigest()!=sha:raise RuntimeError('actual reproof source Git blob differs: '+name)
    return reviewed

def compatibility(stop,original,current):
    if len(COMPAT_SHA256)!=64 or h.pin(COMPAT)!=COMPAT_SHA256:raise RuntimeError('new actual84 final-wire/header/owned-ledger proof is pending or changed')
    proof=h.doc(COMPAT)
    h.exact(proof,dict(schema=1,kind='PUBLIC_CLIENT_FINAL_WIRE_GUARD_COMPATIBILITY',state='PASS',count=84,routes=8,requestBodyAndOwnedLedgerBytesUnchanged=True),'new actual84 public compatibility proof required')
    if proof.get('baselineSourceCommit')!=stop['sourceCommit'] or proof.get('baselineSpendingGuardSourceSHA256')!={n:original[n] for n in r.GUARD_FILES} or proof.get('currentSpendingGuardSourceSHA256')!=r.policy()['spendingGuardSourceSHA256']:raise RuntimeError('new84 exact original/current guard bytes required')
    counts={row['name']:row['actualCalls'] for row in stop['processes'] if row['name'] in r.RETAINED_GENERIC}
    if proof.get('routeCounts')!=counts or len(proof.get('cases',[]))!=84 or len(proof.get('originalSourceFiles',{}))!=84:raise RuntimeError('exact8 routes/84 original captures required')
    if h.pin(h.expand(proof['originalSourceIndex']))!=proof.get('originalSourceIndexSHA256'):raise RuntimeError('original84 source index changed')
    covered={name:0 for name in counts}
    for path,sha in proof['originalSourceFiles'].items():
        owners=[row['name'] for row in stop['processes'] if row['name'] in counts and path.startswith(row['nativeBundle']+'/')]
        if len(owners)!=1:raise RuntimeError('compatibility source is outside the eight retained raw trees')
        covered[owners[0]]+=1
        if stop['files'].get(path)!=sha or h.pin(h.expand(path))!=sha:raise RuntimeError('compatibility original capture outside closed4 pins')
    if covered!=counts:raise RuntimeError('exact native capture count per retained route required')
    seen=set()
    for row in proof['cases']:
        source=row.get('source','').rsplit(':',1)[0]
        if source in seen:raise RuntimeError('duplicate compatibility original case')
        seen.add(source)
        if source not in proof['originalSourceFiles']:raise RuntimeError('compatibility case lacks original capture')
        for key in ('ledgerBytesEqual','methodEqual','urlEqual','requestBodyBytesEqual','requestHeadersBytesEqual'):
            if row.get(key) is not True:raise RuntimeError('new84 actual wire/header/ledger differs')
        values={}
        for phase in ('baseline','current'):
            for kind in ('Wire','Ledger'):
                path=h.expand(row[phase+kind+'File'])
                if h.pin(path)!=row[phase+kind+'FileSHA256']:raise RuntimeError('new84 actual replay artifact changed')
                values[phase+kind]=SourceFreeze(c).read(path,2<<20)
        if values['baselineLedger']!=values['currentLedger']:raise RuntimeError('owned USD ledger bytes changed')
        left=h.strict(values['baselineWire']);right=h.strict(values['currentWire'])
        if not isinstance(left,list) or not isinstance(right,list) or len(left)!=1 or len(right)!=1:raise RuntimeError('one own final wire per compatibility case required')
        for key in ('method','url','requestBody','requestHeaders'):
            if left[0].get(key)!=right[0].get(key):raise RuntimeError('actual final request/header differs')
        if not isinstance(left[0].get('requestHeaders'),dict):raise RuntimeError('actual normalized final request headers required')
        sha=hashlib.sha256(left[0]['requestBody'].encode()).hexdigest()
        if sha!=row.get('baselineRequestSHA256') or sha!=row.get('currentRequestSHA256'):raise RuntimeError('native final body pin differs')
    if seen!=set(proof['originalSourceFiles']):raise RuntimeError('new84 proof does not cover each original capture')
    return reviewed_sources(proof,current)

def verify():
    stop=h.doc(r.BASE/'recovery4-stopped/index.json');archive=h.doc(r.BASE/'recovery4/source-at-start.json');original=archive['files'];current=SourceFreeze(c).files()
    if archive['sourceCommit']!=stop['sourceCommit']:raise RuntimeError('retained original08 source differs')
    # Original execution archive stays tied to its own Git objects, not current6.
    for name,sha in original.items():
        if not name or Path(name).is_absolute() or '..' in Path(name).parts or '\n' in name:raise RuntimeError('unsafe historical source path')
    queries=''.join(archive['sourceCommit']+':'+n+'\n' for n in original).encode()
    raw=subprocess.check_output(['git','cat-file','--batch'],input=queries,cwd=c.ROOT,env=c.ENV,timeout=30);offset=0
    if len(raw)>512<<20:raise RuntimeError('historical source archive byte ceiling')
    for name,sha in original.items():
        end=raw.find(b'\n',offset);header=raw[offset:end].split()
        if end<offset or len(header)!=3 or header[1]!=b'blob':raise RuntimeError('historical source blob unavailable')
        size=int(header[2]);start=end+1
        if size<0 or size>64<<20 or hashlib.sha256(raw[start:start+size]).hexdigest()!=sha or raw[start+size:start+size+1]!=b'\n':raise RuntimeError('original generic measured Git source differs')
        offset=start+size+1
    if offset!=len(raw):raise RuntimeError('unexpected historical Git source bytes')
    reviewed=compatibility(stop,original,current)
    remaining=dict(current)
    for name in reviewed:
        if name in original:remaining[name]=original[name]
        else:remaining.pop(name,None)
    changes=old.source_changes(original,remaining)
    roles={}
    for name in r.RETAINED_GENERIC:
        row=next(v for v in stop['processes'] if v['name']==name)
        if row['exitCode']!=0 or row['parentBudget']['status']!='PASS':raise RuntimeError('retained generic requires original full PASS')
        parent=h.f.owned('.evidence/minimax-launches/provider-sites-final-acceptance-4-'+name+'.json')
        roles[name]=dict(round='provider-sites-final-acceptance-4',receiptSHA256=h.pin(parent),historical=True,base=str(h.expand(row['nativeBundle'])),sourceCommit=stop['sourceCommit'],nativeReportSHA256=row['nativeReportSHA256'],nativeConsumerReceiptSHA256=row['nativeConsumerReceiptSHA256'])
    return dict(roles=roles,compatibilitySHA256=h.pin(COMPAT),sourceDifferences=sorted(set(changes)|set(reviewed)),reviewedSourceDifferences=reviewed,wholeReservedMicrousd=12400000,refundMicrousd=0)
