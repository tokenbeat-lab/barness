#!/usr/bin/env python3
"""Read-only bounded custody of the HOME diagnostic regression; no provider calls."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import stat
import types

BASE=Path(__file__).resolve().parent
SOURCE='a25737aeb605b4d5ede11aafeb60916e5c5c701e'
HELPER='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f'
ARCHIVE='.evidence/minimax-retained/ticket-19-native-home-redaction/raw.tar.gz'


def helper():
    path=BASE.parents[1]/'ticket-13/implementation/verify.py'
    fd=os.open(path,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK)
    with os.fdopen(fd,'rb') as stream:
        info=os.fstat(stream.fileno());raw=stream.read((64<<10)+1)
    if not stat.S_ISREG(info.st_mode) or len(raw)>64<<10 or hashlib.sha256(raw).hexdigest()!=HELPER:
        raise ValueError('fixed bounded reader differs')
    b=types.ModuleType('bounded');b.__file__=str(path)
    exec(compile(raw,str(path),'exec'),b.__dict__);b.BASE=BASE;return b


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=p.parse_args();b=helper()
    index=b.doc(b.read(BASE,'index.json',1<<20));cap=index['capsule']
    if type(index.get('schema')) is not int or index['schema']!=1 or index['sourceCommit']!=SOURCE or cap['sourceCommit']!=SOURCE or cap['archive']['file']!=ARCHIVE:
        raise ValueError('fixed HOME regression identity differs')
    summary,items=b.verify_archive(a.common_root,cap)
    codes={'red':1,'red-assertions':1,'green':0,'race':0,'vet':0,'vet-live':0}
    for phase,code in codes.items():
        v=b.doc(items['raw/'+phase+'-command.json'])
        expected_source='source-before.go' if phase=='red' else 'source-assertions-first.go' if phase=='red-assertions' else 'source-after.go'
        if type(v.get('exitCode')) is not int or v['exitCode']!=code or v.get('credentialsRead') is not False or type(v.get('paidCalls')) is not int or v['paidCalls']!=0 or v.get('evidenceUnderHOME') is not True or v['sourceSHA256']!=b.sha(items['raw/'+expected_source]):
            raise ValueError('actual source/numeric HOME command differs')
    manifests=[];reports=0
    for phase,want in [('red','PASS'),('red-assertions','FAIL'),('green','PASS'),('race','PASS')]:
        prefix='raw/'+phase+'/'
        names=[n for n in items if n.startswith(prefix) and n.endswith('/manifest.json')]
        if len(names)!=1:raise ValueError('unique public root manifest required')
        name=names[0];root=name.removesuffix('manifest.json');m=b.doc(items[name]);audit=b.doc(items[root+'audit.json'])
        if len(m['cases'])!=1 or m['cases'][0]['id']!='MM19-renewal-controlled-refusal' or m['cases'][0]['status']!=want:
            raise ValueError('original assertion-case result differs')
        findings=audit['findings']
        if phase.startswith('red'):
            counts=[int(re.fullmatch('the home directory appears ([0-9]+) time\\(s\\)',f['detail']).group(1)) for f in findings]
            if sorted(counts)!=[1,1,1,6] or any(f['rule']!='environment' for f in findings):raise ValueError('original HOME9 audit refusal differs')
        elif findings!=[]:raise ValueError('unchanged final environment audit failed')
        for mode in ['controlled-candidate','controlled-formal','unknown-renewal']:
            command=b.doc(items[root+'MM19-renewal-controlled-refusal/'+mode+'-command.json'])
            if command['failed'] is not True or command['paid'] is not False:raise ValueError('controlled refusal result changed')
            if not phase.startswith('red'):
                entry='^TestMiniMaxFormal$' if mode=='controlled-formal' else '^TestMiniMaxCandidate$'
                if command['argv']!=['$MINIMAX_RENEWAL_TEST_BINARY','-test.run='+entry,'-test.timeout=30s'] or '$MINIMAX_RENEWAL_ROOT/renewal-'+mode+'/' not in command['output']:
                    raise ValueError('diagnostic context or exact argv lost')
        manifests.append(dict(phase=phase,caseStatus=want,auditFindings=len(findings)))
    for name,raw in items.items():
        if not name.endswith('/minimax-report.json'):continue
        r=b.doc(raw);budget=r['budget']
        if r['complete'] is not False or r.get('calls') or any(type(budget[n]) is not int or budget[n]!=0 for n in ['calls','reservedAttempts','attempts']):
            raise ValueError('controlled child gained calls/completeness')
        reports+=1
    if reports!=12:raise ValueError('all twelve raw child reports required')
    print(json.dumps(dict(verified=True,custody=summary,commands=codes,publicManifests=manifests,
        rawControlledChildReports=reports,paidCalls=0,currentFormalQualification=False),indent=2))


if __name__=='__main__':main()
