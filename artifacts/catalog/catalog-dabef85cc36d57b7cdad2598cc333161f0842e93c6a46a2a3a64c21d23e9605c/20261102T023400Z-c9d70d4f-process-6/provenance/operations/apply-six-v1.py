#!/usr/bin/env python3
import base64,datetime,hashlib,json,pathlib,subprocess,sys
W=pathlib.Path('artifacts/catalog/work/20261102T023400Z-c9d70d4f'); STORE='artifacts/catalog/work/store'; BIN=str(W/'bin/catalogctl-frozen'); H=str(W/'operate-v1.py'); idx=[]
def run(label,args,timeout=360):
 p=subprocess.run(['python3',H,'--label',label,'--timeout',str(timeout),'--',BIN,*args],check=False)
 if p.returncode: raise RuntimeError('command failed: '+label)
 return json.loads((W/'commands'/label/'stdout.jsonl').read_text())
rows=[('int','moonshot','openai-completions','chat'),('int','moonshot','openai-responses','responses'),('int','moonshot','anthropic-messages','messages'),('cn','moonshot-cn','openai-completions','chat'),('cn','moonshot-cn','openai-responses','responses'),('cn','moonshot-cn','anthropic-messages','messages')]
for i,(reg,provider,api,kind) in enumerate(rows,1):
 tag=f'{i:02d}-{reg}-{kind}'
 if i==1:
  frozen=json.loads((W/'commands/10-official-int-chat-freeze/stdout.jsonl').read_text())['candidate']
 else:
  meta=json.loads((W/'inputs'/reg/'pricing.json').read_text());url=meta.get('requestedURL',meta.get('url'));stamp=meta.get('retrievedAtUTC',meta.get('fetchedAt'))
  base=['add','--store',STORE,'--operation','chat','--provider',provider,'--api',api,'--model','kimi-k3.1-pro','--source-kind','official','--source',str(W/'inputs'/reg/'pricing.body'),'--source-url',url,'--source-time',stamp,'--definition',str(W/'definitions-v1'/f'{reg}-{kind}-definition.json'),'--decisions',str(W/'definitions-v1'/f'{reg}-{kind}-decisions.json'),'--reason','Official K3.1 Pro text implementation subset with reviewed standard rates']
  proposal=run('20-'+tag+'-preview',base+['--dry-run'])
  if not proposal.get('candidate'): raise RuntimeError('no candidate '+tag)
  frozen=run('21-'+tag+'-freeze',base+['--confirm'])['candidate']
  if frozen['id']!=proposal['candidate']['id']: raise RuntimeError('frozen differs from preview')
 cid=frozen['id']; snap=base64.b64decode(frozen['snapshot']); val=json.loads(snap); ver=val['catalog']['version']; digest=hashlib.sha256(snap).hexdigest()
 if not ver.startswith('catalog-') or '/' in ver or ver in ('.','..'): raise RuntimeError('unsafe version')
 common=['check','--store',STORE,'--id',cid,'--repo','.','--evidence-feature','catalog-kimi-k31-pro','--timeout','5m']
 pre=run('22-'+tag+'-check-preview',common+['--dry-run'])
 checked=run('23-'+tag+'-check',common+['--confirm'])
 if checked.get('status')!='applicable': raise RuntimeError('check not applicable '+tag)
 ref=checked['reference']
 query=run('24-'+tag+'-candidate-applicable',['candidate','--store',STORE,'--id',cid,'--repo','.','--check',ref])
 if query.get('status')!='applicable': raise RuntimeError('candidate query not applicable '+tag+': '+str(query.get('status')))
 parent=pathlib.Path('artifacts/catalog')/ver;parent.mkdir(parents=True,exist_ok=True)
 directory=parent/('20261102T023400Z-c9d70d4f-process-'+str(i));directory.mkdir(exist_ok=False)
 dest=directory/(ver+'.json')
 apply=['apply','--store',STORE,'--id',cid,'--repo','.','--check',ref,'--out',str(dest)]
 run('25-'+tag+'-apply-preview',apply+['--dry-run'])
 result=run('26-'+tag+'-apply',apply+['--confirm'])
 if dest.read_bytes()!=snap: raise RuntimeError('applied bytes differ '+tag)
 (directory/(ver+'.json.sha256')).write_text(digest+'  '+ver+'.json\n')
 (directory/'PROCESS.md').write_text('Process snapshot; not a completed delivery.\n' if i<6 else 'Final snapshot; delivery assembly in progress, no completion marker.\n')
 idx.append({'identity':[provider,api,'kimi-k3.1-pro'],'candidate':cid,'baseRevision':frozen['baseDigest'],'snapshotSHA256':digest,'catalogVersion':ver,'catalogHash':val['hash'],'directory':str(directory),'checkReference':ref,'checkStatus':checked['status'],'apply':result})
 (W/'applied-index-v1.json').write_text(json.dumps(idx,indent=2,ensure_ascii=False)+'\n')
 print('APPLIED',tag,ver,flush=True)
final=pathlib.Path(idx[-1]['directory']); review=final/'review'
run('30-final-review-preview',['export','--store',STORE,'--repo','.','--review-dir',str(review),'--dry-run'])
run('31-final-review-export',['export','--store',STORE,'--repo','.','--review-dir',str(review),'--confirm'])
print('REVIEW',review,flush=True)
