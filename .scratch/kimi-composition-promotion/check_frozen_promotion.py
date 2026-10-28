"""Replay the existing positive controlled CLI scenario with exact frozen bytes."""
from pathlib import Path
import hashlib,json,sys
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/kimi-composition-promotion'
sys.path.insert(0,str(ROOT/'tools/verification'));from receipt_files import source_snapshot,save,sha
from run_process import environment,capture,isolate_go_home
freeze=json.load(open(BASE/'build-v2/freeze.json'));expected=json.load(open(BASE/'full-ordinary-selection-v3.json'))['argv'];expected=expected[expected.index('--')+1:]
receipts=[]
for p in sorted((BASE/'e2e').glob('*/receipt.json')):
 r=json.load(open(p))
 if r['argv']==expected and r['status']=='PASS' and r['exitCode']==0:receipts.append((p,r))
assert len(receipts)==1
p,r=receipts[0];b=[b for b in r['bundles'] if b['package']=='./ai/e2e'];assert len(b)==1
bundle=p.parent/b[0]['path'];original=bundle/'KCP-six-paired-controlled-promotion/controlled-proposal.json';proposal=json.load(open(original))
private=bundle.parent/'kimi-private-authority'/(bundle.name+'-'+hashlib.sha256(b'TestKimiHistoryComposition').hexdigest()[:12])/'source.git'
assert private.is_dir() and private.resolve()!=ROOT/'.git'
out=BASE/'frozen-promotion-v2';out.mkdir();env,_=environment('e2e');isolate_go_home(env,'e2e',out);env.update(GIT_DIR=str(private),GIT_WORK_TREE=str(ROOT))
binary=ROOT/freeze['regionalpromote']['path'];assert sha(binary)==freeze['regionalpromote']['sha256']
argv=[str(binary),'-kimi','-controlled','-proposal',str(out/'controlled.json')]+[str(ROOT/c['composition']) for c in proposal['candidates']]
source=source_snapshot(ROOT);result=capture(argv,ROOT,env,out,180,[])
assert result['exitCode']==0 and (out/'controlled.json').read_bytes()==original.read_bytes()
assert source==source_snapshot(ROOT)==freeze['source'] and (ROOT/'ai/catalog_kimi.json').read_bytes()==(BASE/'initial/catalog-before.json').read_bytes()
value={'kind':'exact-frozen-kimi-promotion-controlled-replay','argv':argv,'source':source,'frozenBinary':freeze['regionalpromote'],'privateAuthority':str(private.relative_to(ROOT)),'controlledProofExactlyMatchesOrdinary':True,'independentMeasurements':12,'catalogUnchanged':True,'vendorCalls':0,'paidGrantsIssued':0,'realClaims':0,'files':{x.name:sha(x) for x in out.iterdir() if x.is_file()},**result};save(out/'receipt.json',value)
print({'exitCode':result['exitCode'],'sameProofBytes':True,'measurements':12,'catalogActive':False})
