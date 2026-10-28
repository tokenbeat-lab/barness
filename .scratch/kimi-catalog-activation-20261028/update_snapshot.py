"""Export public discovery, requiring an exact six-Kimi-only catalog delta."""
from pathlib import Path
import json,sys
R=Path(__file__).resolve().parents[2];B=R/'.evidence/kimi-catalog-activation-20261028';S=R/'.scratch/kimi-catalog-activation-20261028'
sys.path.insert(0,str(R/'tools/verification'))
from receipt_files import save,sha,source_snapshot
from run_process import environment,isolate_go_home,capture
p=R/'ai/release/catalog-snapshot.json';before=p.read_bytes();(B/'before/ai__release__catalog-snapshot.json').write_bytes(before)
out=B/'catalog-export';out.mkdir();env,_=environment('e2e');isolate_go_home(env,'e2e',out);env.update(GOPROXY='off',GOSUMDB='off')
argv=['go','run','./.scratch/kimi-catalog-activation-20261028/export_catalog.go'];r=capture(argv,R,env,out,180,[])
save(out/'receipt.json',dict(argv=argv,source=source_snapshot(R),files={x.name:sha(x) for x in out.iterdir() if x.is_file()},**r))
assert r['status']=='PASS' and r['exitCode']==0,(out/'stderr.log').read_text()
a=json.loads(before);b=json.loads((out/'stdout.jsonl').read_text());old=a['catalog'];new=b['catalog']
added=[m for m in new['models'] if m['provider'] in ['moonshot','moonshot-cn']]
without={**new,'models':[m for m in new['models'] if m['provider'] not in ['moonshot','moonshot-cn']]}
assert without==old,'public catalog changed beyond Kimi additions'
active=json.loads((R/'ai/catalog_kimi.json').read_text());assert added==active['models'] and len(added)==6
p.write_bytes((out/'stdout.jsonl').read_bytes())
save(B/'catalog-snapshot-update.json',dict(pass_=True,publicExport={'path':str((out/'receipt.json').relative_to(R)),'sha256':sha(out/'receipt.json')},beforeSHA256=sha(B/'before/ai__release__catalog-snapshot.json'),afterSHA256=sha(p),oldCatalogHash=a['hash'],newCatalogHash=b['hash'],onlyAddedKimiModels=6,allOtherCatalogFactsIdentical=True,vendorCalls=0))
print(json.dumps({'snapshotUpdated':True,'onlyAddedKimiModels':6,'otherCatalogFacts':'unchanged'}))
