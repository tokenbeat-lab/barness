#!/usr/bin/env python3
"""Read-only replay of exact correction evidence; negative fixtures never certify."""
import argparse,copy,hashlib,json,sys
from pathlib import Path
BASE=Path(__file__).resolve().parent;sys.path.insert(0,str(BASE.parent));import readers
RESOURCE='stages/green-final-native-shapes/bundles/20261018T123017.584602000Z/MM04-final-block-field-text-missing/resources.json'
MANIFEST='stages/green-final-native-shapes/bundles/20261018T123017.584602000Z/manifest.json'
def main(common):
 b=readers.small();full=readers.full();proof=b.doc(b.read(BASE,'correction-evidence.json',1<<20))
 if proof['credentialsRead'] is not False or type(proof['paidCalls']) is not int or proof['paidCalls']!=0 or proof['releaseCertified'] is not False:raise ValueError('correction qualification differs')
 for name,pin in proof['files'].items():
  raw=b.read(BASE,name,8<<20)
  if len(raw)!=pin['bytes'] or b.sha(raw)!=pin['sha256']:raise ValueError('correction file pin differs')
 for pair in proof['creatorPairs']:
  red=b.doc(b.read(BASE,'tdd/'+pair['red']+'/receipt.json',1<<20));green=b.doc(b.read(BASE,'tdd/'+pair['green']+'/receipt.json',1<<20))
  if red['originalsUnchanged'] is not True or green['originalsUnchanged'] is not True or not red['injected'] or not green['injected'] or type(green['exitCode']) is not int or green['exitCode']!=1 or green['timeout'] or green['indexPublished']:raise ValueError('public producer boundary correction differs')
  if red['case']!=green['case'] or red['creator']!=green['creator'] or red['producerSHA256']==green['producerSHA256']:raise ValueError('distinct actual before/after producers required')
  if pair['kind']=='timeout':valid=red['timeout'] is True and red['exitCode'] in (-15,-9)
  elif pair['kind']=='late-refusal':valid=type(red['exitCode']) is int and red['exitCode']==1 and red['nextCommandRead'] is True
  else:valid=type(red['exitCode']) is int and red['exitCode']==0 and red['indexPublished'] is True
  if not valid:raise ValueError('honest producer RED category differs')
 original=b.doc(b.read(BASE.parent,'index.json',1<<20))
 if b.sha(b.read(BASE.parent,'index.json',1<<20))!=proof['originalIndexSHA256'] or original['archive']['sha256']!='5d9b5d5ad032fdcba150d294cf65f0944f6219068ec810f49a86caa45d71ef8d':raise ValueError('original immutable ticket17 index differs')
 original_view=dict(original,observe=[RESOURCE,MANIFEST]);pins,values,_=full.verify(b,BASE.parent,common,original_view);old=b.doc(values[RESOURCE]);keys={'bodies','calls','events','hostPermits','permits','waiters'}
 if set(old)!=keys or any(type(v) is not int or v!=0 for v in old.values()):raise ValueError('original88 zero records are not being revised')
 results=[]
 for label in ('false','float'):
  front=BASE/'negative-fixtures'/label;index=b.doc(b.read(front,'stream-index.json',1<<20));view=dict(index,observe=[RESOURCE,MANIFEST]);other,observed,archive=full.verify(b,front,common,view);changed={name for name in pins if pins[name]!=other.get(name)}
  if set(other)!=set(pins) or changed!={RESOURCE,MANIFEST}:raise ValueError('private capsule has unrelated native delta')
  row=b.doc(observed[RESOURCE]);expected=copy.deepcopy(old);expected['bodies']=False if label=='false' else 0.0
  if row!=expected or type(row['bodies']) is not (bool if label=='false' else float) or any(type(row[k]) is not int for k in keys-{'bodies'}):raise ValueError('exact one typed counterfeit counter required')
  manifest=b.doc(values[MANIFEST]);after=copy.deepcopy(manifest);case=next(v for v in after['cases'] if v['id']=='MM04-final-block-field-text-missing');case['artifacts']['resources.json']=other[RESOURCE]['sha256']
  if b.doc(observed[MANIFEST])!=after:raise ValueError('exact saved artifact digest delta required')
  for phase,expected_exit in [('red',0),('green',1)]:
   receipt=b.doc(b.read(front,phase+'.json',1<<20));stdout=b.read(front,phase+'.stdout',1<<20);stderr=b.read(front,phase+'.stderr',1<<20)
   if type(receipt['exitCode']) is not int or receipt['exitCode']!=expected_exit or b.sha(stdout)!=receipt['stdoutSHA256'] or b.sha(stderr)!=receipt['stderrSHA256'] or (phase=='green' and b'original public resource release differs' not in stderr):raise ValueError('actual public resource RED/GREEN binding differs')
  results.append(dict(value=label,unchangedOtherFiles=len(pins)-2,archive=archive,qualification='NON_AUTHORITATIVE_NEGATIVE_FIXTURE'))
 positive=b.doc(b.read(BASE,'tdd/positive-creator/receipt.json',1<<20))
 if type(positive['exitCode']) is not int or positive['exitCode']!=0 or not positive['originalsUnchanged']:raise ValueError('bounded legitimate public creator must pass')
 print(json.dumps(dict(verified=True,creatorCorrections=len(proof['creatorPairs']),resourceNegativeFixtures=results,legitimatePrivateCreatorExit=0,credentialsRead=False,paidCalls=0,releaseCertified=False),indent=2))
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--common-root',type=Path,required=True);main(p.parse_args().common_root)
