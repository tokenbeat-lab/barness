#!/usr/bin/env python3
"""Read-only exact negative fixture custody; never a release qualification."""
import argparse,json,sys
from pathlib import Path
BASE=Path(__file__).resolve().parent.parent;sys.path.insert(0,str(BASE));import readers
p=argparse.ArgumentParser();p.add_argument('--common-root',type=Path,required=True);a=p.parse_args();b=readers.small();full=readers.full();original=b.doc((BASE/'index.json').read_bytes());op,om,_=full.verify(b,BASE,a.common_root,original);rows=[]
for kind in ('boolean','float'):
 cap=b.doc((BASE/'field-repair'/(kind+'-negative-index.json')).read_bytes())
 if cap['kind']!='non-authoritative-resealed-negative-fixture' or cap['archive']['file']!='.evidence/minimax-retained/ticket-15-new-actual/incomplete-ordinary-'+kind+'-negative-raw.tar.gz' or cap['sourceSnapshot']!=original['sourceSnapshot']:raise ValueError('fixed non-authoritative negative identity required')
 pins,m,archive=full.verify(b,BASE,a.common_root,cap);different=[n for n in set(op)|set(pins) if op.get(n)!=pins.get(n)];bad=b.doc(m['receipts/ordinary.json']);good=b.doc(om['receipts/ordinary.json'])
 if different!=['receipts/ordinary.json'] or type(bad['exitCode']) is not (bool if kind=='boolean' else float) or bad['exitCode']!=1 or type(good['exitCode']) is not int or good['exitCode']!=1:raise ValueError('only exact synthetic exit-code type delta allowed')
 del bad['exitCode'];del good['exitCode']
 if bad!=good:raise ValueError('other command field altered')
 rows.append(dict(kind=kind,archive=archive,syntheticNativeField='exitCode',releaseEvidence=False))
print(json.dumps(dict(verified=True,negativeFixtures=rows,originalActualExitType='int',originalActualExit=1,credentialsRead=False,paidCalls=0,releaseCertified=False),indent=2))
