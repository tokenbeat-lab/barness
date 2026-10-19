"""Immutable old actuals and the sole GLM-data transition for ticket19."""
import hashlib,json,subprocess
import paid_core as c
import paid_plan as p
import paid
import routed_evidence as e
import renewal_common as u
import repair_anchors as anchors
import repair_inputs as old
import repair_contract as repair
import eleven_contract as r
BASE='751ad941a7dcda4392ee9dd73ff6089f8c846f0e'
FIXED_SIX={'glm-chat':'06e59ac3f16e315e22fa31bde2b19ce6431f6741416132aac1fb2e7bdf6860c2','glm-responses':'cb31e629693a71f54c9e80128c7990085587a85d2a7fe651e5e64a92908a8cab','glm-messages':'1904b9417eacaadfac17bcb3a7c3f5c4c6643daa6db5a84dfc99d01f257229a3','protocol114':'035058b4d47396903c0d9324833b99159217c765059f4d9ba9d3d3e7acb069be','visual-jev':'8f6f2b831e83e5b171279e84412c233b22f8de9c684791658a84928a19d31578','visual-decisions':'b812952144db1098820d2d0bf7f3016856e93b59c219b0126da94ef7b0e7cfbd'}
ONE='.evidence/glm-messages-cancel-20261019-1/'
ONE_PINS={'paid-intent.json':'85ad80d8b3bbee3b7b4a782b0ee49044e2507219f99f81039ed7165314bfec51','actual-command.json':'a23936d88decb4575959b18811a963f725c8d2ac2fe2b9d0da05c3e08c826129','actual/20261018T230218.566865000Z/diagnostic-glm-messages-cancel-report.json':'78395266585b5c632bced498789d2135d57c214860e59cba48b1c317d483267c'}
FORMAL_SHA='18da13b55406c7d4a1754ad26984c617185cae3dbef270e68dff95db9cd2e4ef'

def history():
 previous,collections=old.historical();receipts=dict(previous['receipts']);calls=attempts=545
 with e.Root(c.common_root()) as reader:
  for base,pin in FIXED_SIX.items():
   job=repair.jobs(c.ROOT/'.evidence/history-location')[repair.ROUND+'-'+base];relative='.evidence/minimax-launches/'+job['name']+'.json';raw=reader.read(relative)
   actual=e.settled(job,root=c.common_root(),contract=repair,expected_owner=anchors.OWNER)
   if u.sha(raw)!=pin or actual.get('state')!='REPORTED':raise ValueError('certified closed fixed6 actual changed or unknown')
   receipts[job['name']]=dict(receiptSHA256=pin,**actual);calls+=actual['calls'];attempts+=actual['attempts']
  for name,pin in ONE_PINS.items():
   if u.sha(reader.read(ONE+name))!=pin:raise ValueError('certified single diagnostic original changed')
  formal=reader.read('.evidence/minimax-stages/formal.jsonl')
  if u.sha(formal)!=FORMAL_SHA:raise ValueError('original M3 formal ledger changed; cannot reset or rehash')
 if len(receipts)!=45 or (calls+1,attempts+1)!=(648,648):raise ValueError('closed prior648/648 required')
 return dict(receipts=receipts,diagnostic=dict(root=ONE,pins=ONE_PINS,calls=1,attempts=1),known=dict(calls=648,attempts=648),separateM3Formal=dict(calls=39,attempts=36),originalFormalLedgerSHA256=FORMAL_SHA),collections

def generic_source(name):
 if name in ('go.mod','go.sum','ai/release/catalog-snapshot.json'):return True
 if not name.endswith('.go') or name.endswith('_test.go'):return False
 return name.startswith('ai/') and not name.startswith(('ai/live/','ai/e2e/','ai/release/','ai/internal/testkit/','ai/examples/'))

def generic(collections):
 certified,_=anchors.inventories();inputs={};files={}
 with e.Root(c.common_root()) as reader:
  for name,_,_,_ in p.GENERIC:
   directory=collections[name]['relativeRoot']+'/live/'+name;path=old.bundle(reader,directory,'live-report.json');native=old.tree(reader,path);anchors.require(name,path,native,certified);files.update(native)
   inputs[name]=dict(bundle=path,sourceCommit=collections[name]['receipt']['sourceCommit'])
 selected={n:h for n,h in c.sources().items() if generic_source(n)}
 raw=subprocess.check_output(['git','ls-tree','-r',BASE],cwd=c.ROOT,env=c.ENV).decode().splitlines()
 baseline={line.partition('\t')[2]:line.partition('\t')[0].split()[2] for line in raw if generic_source(line.partition('\t')[2])}
 if set(selected)!=set(baseline):raise ValueError('generic relevant producer set changed since reviewed current runtime')
 for name in selected:
  data=u.file_bytes(c.ROOT/name)
  if hashlib.sha1(b'blob '+str(len(data)).encode()+b'\0'+data).hexdigest()!=baseline[name]:raise ValueError('generic runtime/SDK/catalog changed; no affected-combination exemption')
 return dict(inputs=inputs,files=files,producerBaseline=BASE,producerFiles=selected,policy='existing affected-combination policy; native original source/timestamps remain; full current public consumers required')

def glm_collection():
 receipts={};roots=set();source=set()
 for name,job in r.jobs(c.ROOT/'.evidence/unused','glm').items():
  actual=e.settled(job,contract=r)
  if actual.get('state')!='REPORTED' or actual.get('status') not in ('PASS','FAIL'):raise ValueError('three own current-round closed known GLM processes required; unknown blocks remaining')
  if actual['status']=='PASS' and (actual['calls'],actual['attempts'])!=(22,22):raise ValueError('own PASS requires all22 independent GLM scenarios')
  with e.Root(c.common_root()) as reader:reader_raw=reader.read('.evidence/minimax-launches/'+name+'.json');receipt=e.document(reader_raw)
  roots.add(receipt['collectionRoot']);source.add(receipt['sourceCommit']);receipts[name]=dict(receipt=receipt,actual=actual,receiptSHA256=u.sha(reader_raw))
 if len(roots)!=1 or len(source)!=1:raise ValueError('one own GLM collection/source required')
 return c.ROOT/next(iter(roots)).removeprefix('$REPO/'),receipts

def transition(phase):
 if phase=='glm':return dict(phase=phase,selectorSHA256={n:c.digest(c.ROOT/n) for n in sorted(r.GLM_DATA)},matrixSHA256=old.matrix_scope())
 out,receipts=glm_collection();frozen=c.doc(out/'source-at-start.json');current=c.sources()
 passed=all(v['actual']['status']=='PASS' for v in receipts.values())
 if passed:
  from renewal_glm import validate_eleven_publication
  published=validate_eleven_publication(c.doc(c.ROOT/'ai/live/testdata/glm-formal/live-captures.json'))|dict(published=True)
  if {n:h for n,h in frozen['files'].items() if n not in r.GLM_DATA}!={n:h for n,h in current.items() if n not in r.GLM_DATA}:raise ValueError('only four GLM publication selectors may change between phase freezes')
  u.committed(r.GLM_DATA)
 else:
  if frozen['files']!=current:raise ValueError('known GLM FAIL continuation requires every execution source/data byte unchanged; no publication')
  published=dict(threeOwnPASS=False,published=False,sourceCommit=frozen['sourceCommit'])
 return dict(phase=phase,glmRoot=out.relative_to(c.ROOT).as_posix(),glmSourceCommit=frozen['sourceCommit'],glmSourceManifestSHA256=c.digest(out/'source-at-start.json'),publication=published,selectorSHA256={n:c.digest(c.ROOT/n) for n in sorted(r.GLM_DATA)},matrixSHA256=old.matrix_scope(),canonicalReceipts={n:v['receiptSHA256'] for n,v in receipts.items()})

def proofs(phase):
 value,collections=history();return value,generic(collections),transition(phase)
