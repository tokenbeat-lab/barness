#!/usr/bin/env python3
"""Read-only bounded custody of ticket19 native implementation; no vendor certification."""
import argparse
import collections
import hashlib
import json
import os
from pathlib import Path
import stat
import types

BASE=Path(__file__).resolve().parent
HELPER='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f'
SOURCE='f599c6cee2a8c7a666cfbcccb5ad44ccd94b575e'
ARCHIVE='.evidence/minimax-retained/ticket-19-native-renewal/raw.tar.gz'
CODES={1:1,2:0,3:1,4:1,5:1,6:0,7:1,8:1,9:1,10:0,11:0,12:0,13:0,14:1,15:0,16:0,17:0,18:0,19:0,20:0,21:0,22:0,23:2}


def helper():
    path=BASE.parents[1]/'ticket-13/implementation/verify.py'
    fd=os.open(path,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK)
    with os.fdopen(fd,'rb') as stream:
        info=os.fstat(stream.fileno());raw=stream.read((64<<10)+1)
    if not stat.S_ISREG(info.st_mode) or len(raw)>64<<10 or hashlib.sha256(raw).hexdigest()!=HELPER:
        raise ValueError('fixed bounded helper differs')
    b=types.ModuleType('bounded');b.__file__=str(path)
    exec(compile(raw,str(path),'exec'),b.__dict__);b.BASE=BASE;return b


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=parser.parse_args();b=helper()
    index=b.doc(b.read(BASE,'index.json',1<<20));cap=index['capsule']
    if (type(index.get('schema')) is not int or index['schema']!=1 or index.get('kind')!='keyless-native-formal-renewal-implementation-only' or
        index.get('sourceCommit')!=SOURCE or cap['sourceCommit']!=SOURCE or cap['archive']['file']!=ARCHIVE):
        raise ValueError('fixed native implementation identity differs')
    summary,items=b.verify_archive(a.common_root,cap)
    results={}
    for name,raw in items.items():
        leaf=name.removeprefix('raw/')
        if '/' in leaf or not leaf.endswith('.json') or not leaf[:2].isdigit():continue
        number=int(leaf[:2]);v=b.doc(raw)
        if number not in CODES or number in results or type(v.get('exitCode')) is not int or v['exitCode']!=CODES[number] or v.get('paid') is not False:
            raise ValueError('original numeric command history differs')
        results[number]=v['exitCode']
    if set(results)!=set(CODES):raise ValueError('missing original command receipt')
    proof=b.doc(items['raw/committed-byte-equality.json'])
    if proof.get('allSourceBytesEqual') is not True or type(proof.get('files')) is not int or proof['files']!=872 or proof.get('measuredHEAD')!=index['sourceMeasuredHEAD']:
        raise ValueError('WIP to committed source byte proof differs')
    harness=[]
    for name,raw in items.items():
        if name.startswith(('raw/15-harness-green/','raw/18-final-harness-race/')) and name.endswith('/manifest.json'):
            v=b.doc(raw);statuses=collections.Counter(c['status'] for c in v['cases'])
            if not v['cases'] or set(statuses)!={'PASS'}:raise ValueError('final controlled harness outcomes differ')
            audit=b.doc(items[name.removesuffix('manifest.json')+'audit.json'])
            if audit.get('findings')!=[]:raise ValueError('final controlled audit differs')
            harness.append(dict(member=name,cases=len(v['cases']),statuses=dict(statuses),auditFindings=0))
    if len(harness)!=2:raise ValueError('two independent controlled harness bundles required')
    counts=[]
    for suffix,expected in [('minimax-stages/formal.jsonl',{'reserve':39,'attempt':36}),
                            ('minimax-renewals/current-source-eleven-renewal-1/formal.jsonl',{'reserve':40,'attempt':37})]:
        key='raw/positive-api-fixture/.evidence/'+suffix
        rows=[b.doc(line) for line in items[key].splitlines()];actual=dict(collections.Counter(r['kind'] for r in rows))
        if actual!=expected:raise ValueError('synthetic durable ledger fixture count differs')
        counts.append(dict(member=key,counts=actual,realProviderCalls=0))
    print(json.dumps(dict(verified=True,custody=summary,numericCommands=results,controlledHarness=harness,
        syntheticLedgerMechanism=counts,sourceMeasuredAs='751ad94 plus WIP; exact872 compiler bytes equal f599',
        paidCalls=0,formalVendorQualification=False),ensure_ascii=False,indent=2))


if __name__=='__main__':main()
