#!/usr/bin/env python3
"""One reviewed additive finite trace revision; never change prior IDs/cases."""
import copy,gzip,hashlib,json,re
from pathlib import Path
BASE=Path(__file__).resolve().parent;ROOT=BASE.parents[3];TRACE=ROOT/'ai/release/traceability.json';sha=lambda v:hashlib.sha256(v).hexdigest();OLD='f1de4ce6b8a81f08b0ce27c56bc01f6c8fae9e8a8ca51c114c44d5a55acd21b6'

def normalized(v):
 rows=[]
 for r in sorted(v['items'],key=lambda x:x['id']):
  row=dict(id=r['id'],scenarios=sorted(r['scenarios']),cases=r.get('cases',''))
  if r.get('requiredCases'):row['requiredCases']=sorted(r['requiredCases'])
  rows.append(row)
 raw=json.dumps(dict(scenarios=dict(sorted(v['scenarios'].items())),items=rows),ensure_ascii=False,separators=(',',':')).replace('<','\\u003c').replace('>','\\u003e').replace('&','\\u0026').encode();return raw

def main():
 if (BASE/'trace-expansion.json').exists():raise ValueError('exclusive finite revision proof required')
 before=TRACE.read_bytes();old=json.loads(before)
 if sha(normalized(old))!=OLD:raise ValueError('original approved normalized mapping differs')
 manifest=next((ROOT/'.evidence/minimax-ticket17/green-final-native-shapes/bundles').glob('*/manifest.json'));cases=json.loads(manifest.read_bytes())['cases'];ids=sorted(c['id'] for c in cases)
 if len(ids)!=44 or len(set(ids))!=44 or any(c['status']!='PASS' for c in cases):raise ValueError('actual complete44 public GREEN identities required')
 new=copy.deepcopy(old);changed={};wanted={'MM-R20','MM-A17','MM-US55'}
 for row in new['items']:
  if row['id'] not in wanted:continue
  previous=list(row['requiredCases']);pattern=row['cases']
  if not pattern.endswith(')$') or set(ids)&set(previous):raise ValueError('unexpected prior finite mapping')
  row['requiredCases']=previous+ids;row['cases']=pattern[:-2]+'|'+'|'.join(re.escape(n) for n in ids)+')$';changed[row['id']]=dict(beforeRequired=len(previous),afterRequired=len(row['requiredCases']),added=ids,removed=[],oldRegexRetainedAsPrefix=True)
 if set(changed)!=wanted:raise ValueError('exact three related requirement groups required')
 oldrows={r['id']:r for r in old['items']};newrows={r['id']:r for r in new['items']}
 if any(newrows[k]!=r for k,r in oldrows.items() if k not in wanted) or any(old[k]!=new[k] for k in old if k!='items'):raise ValueError('unrelated release contracts changed')
 lines=before.decode().splitlines()
 for i,line in enumerate(lines):
  for name in wanted:
   if line.lstrip().startswith('{"id":"'+name+'",'):
    lines[i]='    '+json.dumps(newrows[name],ensure_ascii=False,separators=(',',':'))+(',' if line.rstrip().endswith(',') else '')
 after=('\n'.join(lines)+'\n').encode()
 if json.loads(after)!=new:raise ValueError('minimal three-line trace rewrite differs')
 TRACE.write_bytes(after);(BASE/'trace-before.json.gz').write_bytes(gzip.compress(before,mtime=0));(BASE/'trace-after.json.gz').write_bytes(gzip.compress(after,mtime=0));proof=dict(beforeSHA256=sha(before),afterSHA256=sha(after),oldNormalizedSHA256=OLD,newNormalizedSHA256=sha(normalized(new)),unchangedItemIDs=sorted(set(oldrows)-wanted),changedItems=changed,traceItems=len(oldrows),newPublicCases=44,priorRequiredCasesRemoved=0,liveCombosUnchanged=True,scenarioPatternsUnchanged=True,source='Actual green-final-native-shapes public manifest; original current approval hash verified first',credentialsRead=False,paidCalls=0);(BASE/'trace-expansion.json').write_text(json.dumps(proof,separators=(',',':'))+'\n');print(json.dumps({k:proof[k] for k in ['oldNormalizedSHA256','newNormalizedSHA256','traceItems','newPublicCases','priorRequiredCasesRemoved']},indent=2))
if __name__=='__main__':main()
