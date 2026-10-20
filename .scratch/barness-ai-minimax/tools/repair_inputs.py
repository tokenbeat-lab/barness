"""Reopen immutable native history and prove the precise affected-source boundary."""
import os,re,json
from pathlib import Path
import subprocess
import paid_core as c
import paid_plan as p
import paid
import routed_contract as old
import routed_evidence as e
import renewal_common as u
import repair_anchors as anchors
BASE='8fbd17006677774368d4d77d924a8672fd205c51'
OFFLINE={'ai/internal/testkit/release/plan.go','ai/internal/testkit/release/artifacts_test.go','ai/internal/testkit/release/decisions_record_test.go','ai/e2e/release_recorded_bundle_test.go','ai/e2e/release_recorded_bundle_process_test.go','ai/live/glm_assembly_test.go'}
APPROVED={'ai/internal/testkit/release/plan.go':'0026bdf6d8a93c6765a5e0455d3d37a0271821b771900bb77d52c53d7fc7a97b','ai/internal/testkit/release/artifacts_test.go':'7e70f5ffa9c53d0ea785bf48487758d213c6a14e6b22b169a0ef6c4ceba93efd','ai/internal/testkit/release/decisions_record_test.go':'c58f068e08b23cd13b2ca534b9d18d930f1d367c0f248fa1e949331f928225fb','ai/e2e/release_recorded_bundle_test.go':'cb0e26849683cc5bf9fef8bb3b0722fbce32b1baae16ce19cad0864638667abd','ai/e2e/release_recorded_bundle_process_test.go':'1b39ba3dba57263469671cf02924c1e134e41ca6a1ceccb788bd99c6ae9dd662','ai/live/glm_assembly_test.go':'2fd107f18e3d34b6793fd256a69322bd335e9be42f50d232fbca963fe25afaf9'}
POLICY={'.scratch/barness-ai-minimax/evidence/ticket-13/postgate-'+n+'.json':h for n,h in [('source-policy-review','ce53bc70c0900fdf48b1f3818e4393f975d24b0c9e698d921b9f97880855097a'),('minimal-scope-spec-review','f4ab925705043a9269c4c02fd787874db7d10a42e22d1a7cf214382964c894d5')]}

def tree(reader,relative):
 result={};total=entries_seen=0
 def walk(name,depth=0):
  nonlocal total,entries_seen
  if depth>20:raise ValueError('bounded native tree depth')
  fd=reader.open(name,directory=True)
  try:
   with os.scandir(fd) as entries:
    for entry in entries:
     entries_seen+=1
     if entries_seen>32768:raise ValueError('bounded entire native tree shape, including empty directories')
     path=name+'/'+entry.name;e.parts(path)
     if entry.is_symlink():raise ValueError('native evidence links forbidden')
     if entry.is_dir(follow_symlinks=False):walk(path,depth+1)
     else:
      raw=reader.read(path,16<<20);total+=len(raw)
      if len(result)>=32768 or total>512<<20:raise ValueError('bounded native tree inventory')
      result[path]=dict(bytes=len(raw),sha256=u.sha(raw))
  finally:os.close(fd)
 walk(relative);return result

def bundle(reader,directory,leaf):
 fd=reader.open(directory,directory=True);found=[]
 try:
  with os.scandir(fd) as entries:
   for count,entry in enumerate(entries):
    if count>=1024 or entry.is_symlink():raise ValueError('bounded unambiguous native sessions required')
    if entry.is_dir(follow_symlinks=False):
     name=directory+'/'+entry.name
     try:reader.read(name+'/'+leaf,16<<20)
     except FileNotFoundError:continue
     found.append(name)
 finally:os.close(fd)
 if len(found)!=1:raise ValueError('one original native bundle required')
 return found[0]

def historical():
 prior={};receipts={};collections={};calls=attempts=300
 common=c.common_root();jobs=old.jobs(c.ROOT/'.evidence/unused-history-location');expected=anchors.original_receipts()
 with e.Root(common) as reader:
  for name,digest in expected.items():
   if u.sha(reader.read('.evidence/minimax-launches/'+name+'.json'))!=digest:raise ValueError('certified original receipt changed; cannot rehash as history')
  for name,count in old.PRIOR.items():
   raw=reader.read('.evidence/minimax-launches/'+name+'.json');v=e.document(raw);budget=v.get('budget',{})
   if (v.get('state')!='COMPLETED' or v.get('route')!=name or v.get('credentialsRead') is not True or v.get('launcherAudit',{}).get('passed') is not True or budget.get('state')!='REPORTED' or budget.get('status') not in ('PASS','FAIL') or type(budget.get('calls')) is not int or budget['calls']!=count or type(budget.get('attempts')) is not int or budget['attempts']!=count or re.fullmatch('[a-f0-9]{40}',v.get('sourceCommit','')) is None or re.fullmatch('[a-f0-9]{64}',v.get('reportedFileSha256','')) is None):raise ValueError('original21 validated300 history required')
   receipts[name]=dict(receiptSHA256=u.sha(raw),sourceCommit=v['sourceCommit'],reportSHA256=v['reportedFileSha256'],calls=count,attempts=count,status=budget['status'])
  for name,job in jobs.items():
   actual=e.settled(job,root=common,expected_owner=anchors.OWNER)
   if actual['state']!='REPORTED':raise ValueError('all original39 must settle; unknown preserves full caps and denies preparation')
   raw=reader.read('.evidence/minimax-launches/'+name+'.json');receipt=e.document(raw);relative=receipt['collectionRoot'].removeprefix('$REPO/');e.parts(relative)
   if not relative.startswith('.evidence/'):raise ValueError('original history collection root required')
   metadata={}
   for leaf in ('source-at-start.json','binaries.json','plan.json','prerelease-source-review.json','execution-authorization.json','checks.json',job['baseName']+'-command.json'):
    metadata[relative+'/'+leaf]=u.sha(reader.read(relative+'/'+leaf))
   receipts[name]=dict(receiptSHA256=u.sha(raw),sourceCommit=receipt['sourceCommit'],reportSHA256=receipt['reportedFileSha256'],calls=actual['calls'],attempts=actual['attempts'],status=actual['status'],metadata=metadata)
   collections[job['baseName']]=dict(relativeRoot=relative,receipt=receipt)
   calls+=actual['calls'];attempts+=actual['attempts']
 if len(receipts)!=39 or (calls,attempts)!=(545,545):raise ValueError('immutable known39 total545/545 required; no reset')
 return dict(receipts=receipts,known=dict(calls=calls,attempts=attempts)),collections

def evaluation_source(name):
 # Same complete selection as the unchanged public evaluation_source.go; this
 # standalone keyless preparer cannot import Go. Current public verifiers remain
 # mandatory, and the source-scope guard rejects edits to that selector itself.
 if name in ('go.mod','go.sum','ai/examples/chineseeval/testdata/recorded-jev-20261008.json'):return True
 if not name.endswith('.go') or name.endswith('_test.go'):return False
 if name.startswith(('ai/release/','ai/e2e/','ai/live/','ai/internal/testkit/release/')):return False
 if name.startswith('ai/examples/'):return name.startswith(('ai/examples/chineseeval/','ai/examples/requestid/'))
 return name.startswith('ai/')


def reusable(collections):
 common=c.common_root();inputs={};files={};source_maps={};certified,m3_index=anchors.inventories()
 with e.Root(common) as reader:
  for name in [*(j[0] for j in p.GENERIC),'text-jev','text-luna']:
   relative=collections[name]['relativeRoot'];directory=relative+('/live/' if name in {j[0] for j in p.GENERIC} else '/')+name
   leaf='evaluation-report.json' if name.startswith('text-') else 'live-report.json';path=bundle(reader,directory,leaf)
   inputs[name]=dict(bundle=path,report=path+'/'+leaf,sourceCommit=collections[name]['receipt']['sourceCommit']);native_files=tree(reader,path);anchors.require(name,path,native_files,certified);files.update(native_files)
   if name.startswith('text-'):
    native=e.document(reader.read(path+'/code-snapshot.json'));source_maps[name]=native;selected={n:h for n,h in native.items() if evaluation_source(n)}
    current={n for n in c.sources() if evaluation_source(n)}
    if len(native)!=860 or len(selected)!=261 or set(selected)!=current or any(c.digest(c.ROOT/n)!=h.removeprefix('sha256:') for n,h in selected.items()):raise ValueError('unchanged complete261 selected text source required; preserve full860 snapshot')
  for route in ('chat','responses','messages'):
   name='minimax-'+route;directory='.evidence/minimax-global-formal-glm/mm-global-formal-m3-'+route;path=bundle(reader,directory,'minimax-report.json')
   receipt_raw=reader.read('.evidence/minimax-launches/mm-global-formal-m3-'+route+'.json');receipt=e.document(receipt_raw);report=e.document(reader.read(path+'/minimax-report.json'))
   certified_receipts=[v['sha256'] for v in m3_index['files'].values() if v.get('original','').endswith('/.evidence/minimax-launches/mm-global-formal-m3-'+route+'.json')]
   if certified_receipts!=[u.sha(receipt_raw)]:raise ValueError('certified original M3 canonical receipt changed')
   if (receipt.get('state')!='COMPLETED' or receipt.get('credentialsRead') is not True or receipt.get('origin',{}).get('reportSha256')!=u.sha(reader.read(path+'/minimax-report.json')) or report.get('outcome')!='PASS' or report.get('complete') is not True or report.get('identity')!=p.identity(dict(route=route,model='MiniMax-M3',purpose='standard'))):raise ValueError('original own successful formal M3 origin required')
   native=report['executionSource'];source_maps[name]=native
   if len(native)!=263 or any(c.digest(c.ROOT/n)!=h for n,h in native.items()):raise ValueError('unchanged native263 M3 source required')
   inputs[name]=dict(bundle=path,report=path+'/live-report.json',sourceCommit=report['gitCommit'],launcherReceiptSHA256=u.sha(receipt_raw));native_files=tree(reader,path);anchors.require(name,path,native_files,certified);files.update(native_files)
 if len(files)>32768 or sum(v['bytes'] for v in files.values())>512<<20:raise ValueError('bounded reusable native inventory')
 return dict(ownerSHA256=u.sha(str(common).encode()),inputs=inputs,files=files,sourceMaps=source_maps,certifiedInventoryPins=anchors.PINS,policy='original commits/timestamps and payload remain; public current verifiers required separately')

MATRIX_SHA='7fbe72c37324a5700b0cff9416bf2a05543167ca1eed88186741669e39c6190f'

def matrix_scope():
 raw=subprocess.check_output(['git','show','73ef82888b7a9c623d3ff4c33f331057f0275af4:ai/live/support-matrix.json'],cwd=c.ROOT,env=c.ENV)
 if u.sha(raw)!=MATRIX_SHA:raise ValueError('original authenticated17 matrix pin differs')
 original=e.document(raw);current=c.doc(c.ROOT/'ai/live/support-matrix.json');fresh={'zhipu-chat','zhipu-responses','zhipu-messages','typesafe-classifier-jev-1.14.0'}
 before={r['combo']:r for r in original['rows']};after={r['combo']:r for r in current['rows']}
 if {k:v for k,v in current.items() if k!='rows'}!={k:v for k,v in original.items() if k!='rows'} or len(after)!=17 or len(current['rows'])!=17 or set(after)!=set(before) or any(after[n]!=v for n,v in before.items() if n not in fresh):raise ValueError('matrix must preserve original13 rows/histories exactly')
 return c.digest(c.ROOT/'ai/live/support-matrix.json')

def scope(phase):
 changed=c.git('diff','--name-only',BASE,'HEAD').splitlines();selected=set(c.sources());disallowed=[];data=[]
 if any(c.digest(c.ROOT/n)!=h for n,h in POLICY.items()):raise ValueError('independently certified source policy review changed')
 if any(c.digest(c.ROOT/n)!=h for n,h in APPROVED.items()):raise ValueError('named offline/GLM delta must match independently reviewed exact source')
 allowed_data={'ai/live/testdata/glm-formal/'+n for n in ('live-captures.json','live-captures.sha256','known-usage.json','support-views.json')}
 for name in changed:
  if name in ('ai/live/support-matrix.json','ai/live/testdata/minimax-formal/origins.json'):continue
  if name in allowed_data:data.append(name);continue
  if (name in selected or name.startswith('ai/') and name.endswith(('.go','.json','.md'))) and name not in OFFLINE and not name.startswith('.scratch/barness-ai-minimax/tools/'):disallowed.append(name)
 if disallowed:raise ValueError('affected scope changed outside reviewed offline/GLM-only producer exception: '+','.join(disallowed))
 if data:
  if phase!='effects':raise ValueError('GLM phase must precede new raw/view publication')
  index=c.doc(c.ROOT/'ai/live/testdata/glm-formal/live-captures.json')
  import repair_contract as r
  if index.get('collectionRound')!=r.ROUND:raise ValueError('only new canonical same-source3PASS publication may change GLM data')
  from renewal_glm import validate_repair_publication
  validate_repair_publication(index)
 return dict(baseline=BASE,matrixSHA256=matrix_scope(),matrixBaselineSHA256=MATRIX_SHA,changedSelected=sorted(n for n in changed if n in selected),permittedGLMData=sorted(data),policyReports={n:c.digest(c.ROOT/n) for n in POLICY},catalogSHA256=c.digest(c.ROOT/'ai/release/catalog-snapshot.json'),SDK={n:c.digest(c.ROOT/n) for n in ('go.mod','go.sum')},reusedProducerPolicy='unchanged runtime/SDK/model/catalog/shared live producer; explicit renewedGLM-only assembly exception')

def proofs(phase):
 history,collections=historical();return history,reusable(collections),scope(phase)
