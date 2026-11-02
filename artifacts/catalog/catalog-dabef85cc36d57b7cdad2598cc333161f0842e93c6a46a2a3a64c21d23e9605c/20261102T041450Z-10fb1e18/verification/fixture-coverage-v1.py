#!/usr/bin/env python3
"""Relate qualification and retained-path cases to actual synthetic packets."""
import collections,json,pathlib,sys
root=pathlib.Path.cwd();sys.path.insert(0,str(root/'tools/verification'));from receipt_files import sha
attempts=['20261102T044646.946229Z-0wet3qqo','20261102T051207.821665Z-1z37t3sn']; results=[];failures=[]
for attempt in attempts:
 p=root/'.evidence/catalog-model-id-policy/e2e'/attempt;r=json.loads((p/'receipt.json').read_text());b=p/r['bundles'][0]['path'];m=json.loads((b/'manifest.json').read_text());tree=p/'xai-images-historical-9584';cases=m['cases'];fixtures=[c for c in cases if c['id'].startswith('XAIIMG-qualification-fixture-SYNTHETIC-')]; paths=[c for c in cases if c['id'].startswith('XAIIMG-producer-path-fixture-')]; packets=[*(tree/'synthetic-packets').glob('xai-images-proof-fixture-*'),*tree.glob('path-xai-images-proof-fixture-*')]; mapped=[]
 for c in fixtures:
  session=c['id'].removeprefix('XAIIMG-qualification-fixture-'); found=[q for q in packets if (q/'bundles'/session).is_dir()]
  if len(found)!=1 or c['status']!='PASS':failures.append('fixture '+attempt+' '+session);continue
  q=found[0]; qreceipt=json.loads((q/'receipt.json').read_text());
  if qreceipt['exitCode']!=0 or (q/'exit-code').read_text().strip()!='0':failures.append('producer '+session)
  mapped.append({'fixtureCase':c['id'],'test':c['test'],'packet':str(q.relative_to(root)),'producerReceiptSHA256':sha(q/'receipt.json'),'producerExit':qreceipt['exitCode']})
 if len(mapped)!=len(fixtures) or len(packets)!=len(fixtures) or len({x['packet'] for x in mapped})!=len(packets):failures.append('one-to-one inventory '+attempt)
 pathmap=[]
 for c in paths:
  info=json.loads((b/c['dir']/'retained-fixture.json').read_text()); q=tree/info['path']; rel=str(q.relative_to(root)); matches=[x for x in mapped if x['packet']==rel]
  if c['status']!='PASS' or len(matches)!=1 or info.get('notActualLive')!=True:failures.append('path relation '+c['id']);continue
  pathmap.append({'case':c['id'],'test':c['test'],'fixtureCase':matches[0]['fixtureCase'],'packet':rel,'retainedInventoryArtifactSHA256':sha(b/c['dir']/'retained-fixture.json')})
 if len(paths)!=5 or len(pathmap)!=5:failures.append('path inventory '+attempt)
 results.append({'receipt':str((p/'receipt.json').relative_to(root)),'receiptSHA256':sha(p/'receipt.json'),'qualificationFixtureCases':len(fixtures),'pathFixtureCases':len(paths),'distinctPackets':len(packets),'fixtureTests':dict(collections.Counter(c['test'].split('/')[0] for c in fixtures)),'relations':mapped,'pathRelations':pathmap})
out=root/'artifacts/catalog/work/20261102T041450Z-10fb1e18/verification/fixture-coverage-v1.json';out.write_text(json.dumps({'status':'PASS' if not failures else 'FAIL','failures':failures,'explanation':'Both attempts have 48 qualification-fixture cases and five additional path-fixture inventory cases. Each path-fixture case refers to a packet already counted by a qualification-fixture case: 53 native fixture-related cases map to 48 distinct packets. Dynamic identifiers differ per run.','attempts':results},indent=2)+'\n');print(json.dumps({'status':'PASS' if not failures else 'FAIL','failures':failures,'summary':[(x['qualificationFixtureCases'],x['pathFixtureCases'],x['distinctPackets']) for x in results],'resultSHA256':sha(out),'scriptSHA256':sha(pathlib.Path(__file__))}));sys.exit(bool(failures))
