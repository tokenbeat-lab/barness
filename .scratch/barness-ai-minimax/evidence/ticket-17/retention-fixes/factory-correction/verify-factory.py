#!/usr/bin/env python3
"""Read the bounded raw repair capsule without executing any captured program."""
import argparse,json,sys
from pathlib import Path
BASE=Path(__file__).resolve().parent;FRONT=BASE.parents[1];sys.path.insert(0,str(FRONT));import readers
CASES={'index-size':'artifact size limit','inventory-compressed':'artifact size limit','inventory-decoded':'decoded metadata pin or limit','seal-metadata':'artifact size limit','output-grow':'output archive size bound'}
def main(common):
 b=readers.small();full=readers.full();index=b.doc(b.read(BASE,'capsule-index.json',1<<20))
 if index['kind']!='resource-factory-correction-history' or index['sourceCommit']!='e1f3b8405f8d4dddd30f2e035597db5f15fd89b0' or index['archive']['file']!='.evidence/minimax-retained/ticket-17-final-payload/resource-factory-correction-raw.tar.gz' or index['credentialsRead'] is not False or type(index['paidCalls']) is not int or index['paidCalls']!=0 or index['releaseCertified'] is not False:raise ValueError('fixed factory repair custody differs')
 pins,values,archive=full.verify(b,BASE,common,index);source=b.doc(values['source-proof.json']);before=b.sha(values['sources/red-resource-probe.py.txt']);after=b.sha(values['sources/green-resource-probe.py.txt'])
 if source['beforeSourceSHA256']!=before or source['files']['resource-probe.py']!=after or source['currentBytesEqualActualGREEN'] is not True or b.sha(b.read(FRONT/'retention-fixes','resource-probe.py',64<<10))!=after:raise ValueError('actual measured old/current constructor bytes differ')
 for phase in ('red','green'):
  for case,reason in CASES.items():
   row=b.doc(values['tdd/'+phase+'-'+case+'/receipt.json'])
   if type(row['exitCode']) is not int or row['exitCode']!=1 or row['injected'] is not True or row['originalsUnchanged'] is not True or row['originalBefore']!=row['originalAfter'] or row['sourceSHA256']!=(before if phase=='red' else after):raise ValueError('actual numeric injected constructor boundary differs')
   events=[json.loads(v) for v in values['tdd/'+phase+'-'+case+'/events.jsonl'].splitlines()]
   if sum(v['event']=='injected' for v in events)!=1:raise ValueError('one genuine OS-open replacement required')
   if phase=='red':
    if row['constructionPresent'] is not True or row['verificationInvoked'] is not True or row['sealRewritten'] is not True:raise ValueError('historical RED must remain a late refusal')
   elif row['verificationInvoked'] is not False or row['sealRewritten'] is not False or row['constructionPresent'] is not (case=='seal-metadata') or reason not in row['diagnostic']:raise ValueError('current constructor must refuse before verification/seal publication')
 partial=b.doc(values['collection-first-refusal.json'])
 if partial['archiveCreated'] is not False or any(pins.get(name)!=pin for name,pin in partial['partialFiles'].items()):raise ValueError('initial nofollow refusal partial source custody differs')
 positive=b.doc(values['legitimate/receipt.json']);verification=b.doc(values['legitimate/verification.json']);construction=b.doc(values['legitimate/construction.json'])
 if type(positive['exitCode']) is not int or positive['exitCode']!=0 or positive['sourceUnchanged'] is not True or positive['sourceBeforeSHA256']!=after or positive['sourceAfterSHA256']!=after or type(verification['exitCode']) is not int or verification['exitCode']!=1 or construction['archive']['sha256']!='80befd9d76176b28c60ad7b7de1868ed3fde6902dd7a9ced832820194d784b34' or construction['archive']['bytes']!=90847871:raise ValueError('normal private construction/public counterfeit rejection differs')
 print(json.dumps(dict(verified=True,archive=archive,publicFactoryBoundaryPairs=5,lateREDConstructorExits=[1]*5,earlyGREENConstructorExits=[1]*5,legitimateFactoryExit=0,legitimateCounterfeitVerifierExit=1,originalsUnchanged=True,credentialsRead=False,paidCalls=0,releaseCertified=False),indent=2))
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--common-root',type=Path,required=True);main(p.parse_args().common_root)
