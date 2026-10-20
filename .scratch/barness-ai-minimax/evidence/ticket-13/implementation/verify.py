#!/usr/bin/env python3
"""Read-only, bounded verification of the two fixed keyless history capsules."""
import argparse
import gzip
import hashlib
import json
import os
import re
from pathlib import Path, PurePosixPath
import stat
import sys
import tarfile

BASE=Path(__file__).resolve().parent
CAP=32<<20
DECODED=192<<20


def relative(value):
    if not isinstance(value,str) or not value or PurePosixPath(value).is_absolute() or any(p in ('','.','..') for p in value.split('/')):
        raise ValueError('invalid owned artifact path')
    return value.split('/')


def opened(root,name):
    fd=os.open(root,os.O_RDONLY|os.O_DIRECTORY|os.O_NOFOLLOW)
    try:
        names=relative(name)
        for i,part in enumerate(names):
            flags=os.O_RDONLY|os.O_NOFOLLOW|(os.O_DIRECTORY if i<len(names)-1 else os.O_NONBLOCK)
            child=os.open(part,flags,dir_fd=fd);os.close(fd);fd=child
        info=os.fstat(fd)
        if not stat.S_ISREG(info.st_mode):raise ValueError('regular artifact required')
        stream=os.fdopen(fd,'rb');fd=None;return stream
    finally:
        if fd is not None:os.close(fd)


def read(root,name,limit):
    with opened(root,name) as stream:
        if os.fstat(stream.fileno()).st_size>limit:raise ValueError('artifact size limit')
        raw=stream.read(limit+1)
        if len(raw)>limit:raise ValueError('artifact byte limit')
        return raw


def sha(raw):return hashlib.sha256(raw).hexdigest()


def doc(raw):
    def pairs(items):
        value={}
        for key,item in items:
            if key in value:raise ValueError('duplicate JSON field')
            value[key]=item
        return value
    value=json.loads(raw,object_pairs_hook=pairs)
    if not isinstance(value,dict):raise ValueError('JSON object required')
    return value


class Limited:
    def __init__(self,stream,limit):self.stream=stream;self.limit=limit;self.count=0
    def read(self,size=-1):
        remaining=self.limit-self.count
        raw=self.stream.read(min(size if size>=0 else remaining+1,remaining+1))
        self.count+=len(raw)
        if self.count>self.limit:raise ValueError('decoded byte limit')
        return raw


def decoded_gzip(info):
    encoded=read(BASE,info['file'],8<<20)
    if sha(encoded)!=info['compressedSHA256']:raise ValueError('inventory compressed pin')
    import io
    with gzip.GzipFile(fileobj=io.BytesIO(encoded)) as stream:raw=stream.read((8<<20)+1)
    if len(raw)>8<<20 or len(raw)!=info['decodedBytes'] or sha(raw)!=info['decodedSHA256']:
        raise ValueError('inventory decoded pin or limit')
    return raw


def inventory(capsule):
    info=capsule['inventory'];values=doc(decoded_gzip(info))
    if len(values)!=info['files'] or len(values)>32768:raise ValueError('inventory file count')
    if capsule.get('innerInventory'):values['inventory.json']=capsule['innerInventory']
    for name,value in values.items():
        relative(name)
        if (not isinstance(value,dict) or type(value.get('bytes')) is not int or not 0<=value['bytes']<=64<<20 or
            not isinstance(value.get('sha256'),str) or re.fullmatch('[a-f0-9]{64}',value['sha256']) is None):raise ValueError('invalid inventory member')
    source=capsule['sourceSnapshot'];raw=decoded_gzip(source)
    if doc(raw).get('sourceCommit')!=capsule['sourceCommit'] or values[source['archiveMember']]['sha256']!=sha(raw):
        raise ValueError('source snapshot does not match original capsule')
    return values


def verify_archive(common,capsule):
    expected=inventory(capsule);archive=capsule['archive'];seen=set();captured={};entries=0
    with opened(common,archive['file']) as stream:
        info=os.fstat(stream.fileno())
        if info.st_size!=archive['bytes'] or info.st_size>CAP:raise ValueError('compressed archive size')
        digest=hashlib.sha256();compressed=0
        for chunk in iter(lambda:stream.read(65536),b''):
            compressed+=len(chunk)
            if compressed>CAP:raise ValueError('compressed archive byte limit')
            digest.update(chunk)
        if compressed!=archive['bytes']:raise ValueError('compressed archive size changed')
        if digest.hexdigest()!=archive['sha256']:raise ValueError('compressed archive pin')
        stream.seek(0)
        with gzip.GzipFile(fileobj=stream) as zipped:
            limited=Limited(zipped,DECODED)
            with tarfile.open(fileobj=limited,mode='r|') as bundle:
                names=set()
                for member in bundle:
                    entries+=1
                    if entries>32768:raise ValueError('archive entry limit')
                    parts=relative(member.name)
                    if len(parts)>32 or parts[0]!=archive['prefix'] or member.name in names:raise ValueError('archive path or duplicate')
                    names.add(member.name)
                    if member.isdir():continue
                    if not member.isfile() or len(parts)<2:raise ValueError('regular archive member required')
                    name='/'.join(parts[1:]);pin=expected.get(name)
                    if pin is None or member.size!=pin['bytes']:raise ValueError('archive member set or size')
                    if name in capsule.get('observe',[]) and member.size>1<<20:raise ValueError('observed metadata limit')
                    original=bundle.extractfile(member);digest=hashlib.sha256();count=0;chunks=[]
                    with original:
                        for chunk in iter(lambda:original.read(65536),b''):
                            count+=len(chunk)
                            if count>pin['bytes']:raise ValueError('member payload limit')
                            digest.update(chunk)
                            if name in capsule.get('observe',[]):chunks.append(chunk)
                    if count!=pin['bytes'] or digest.hexdigest()!=pin['sha256']:raise ValueError('archive member pin')
                    if name in capsule.get('observe',[]):captured[name]=b''.join(chunks)
                    seen.add(name)
            while limited.read(65536):pass
            decoded=limited.count
        after=os.fstat(stream.fileno())
        if (after.st_size,after.st_mtime_ns,after.st_ctime_ns)!=(info.st_size,info.st_mtime_ns,info.st_ctime_ns):
            raise ValueError('original archive changed during verification')
    if seen!=set(expected):raise ValueError('missing archive member')
    return dict(id=capsule['id'],archiveSHA256=archive['sha256'],files=len(seen),entries=entries,decodedTarBytes=decoded),captured


def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--common-root',type=Path,default=BASE.parents[4])
    a=parser.parse_args()
    for line in read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
        expected,name=line.split('  ',1)
        if sha(read(BASE,name,8<<20))!=expected:raise ValueError('local metadata pin')
    index=doc(read(BASE,'index.json',1<<20))
    if index.get('schema')!=1 or {v.get('id') for v in index['capsules']}!={'implementation83','reviewfix51'} or len(index['capsules'])!=2:
        raise ValueError('fixed two-capsule index required')
    outputs=[];all_observed={}
    for capsule in index['capsules']:
        result,observed=verify_archive(a.common_root,capsule);outputs.append(result);all_observed[capsule['id']]=observed
    expected=doc(read(BASE,'observations.json',1<<20))
    for item in expected['manifests']:
        observed=all_observed['implementation83'];manifest=doc(observed[item['manifest']]);audit=doc(observed[item['audit']])
        statuses={status:sum(c['status']==status for c in manifest['cases']) for status in ('PASS','FAIL')}
        if (len(manifest['cases'])!=item['cases'] or statuses!=item['statuses'] or
            len(audit['findings'])!=item['auditFindings'] or manifest['git_commit']!=item['measuredGitCommit']):
            raise ValueError('historical observation differs')
    for check in expected['checks']:
        raw=all_observed[check['capsule']][check['member']]
        if 'tests' in check:
            if re.search(rb'Ran '+str(check['tests']).encode()+rb' tests in [0-9.]+s\s+OK\s*$',raw) is None:
                raise ValueError('archived test result differs')
        elif raw!=b'':raise ValueError('historical vet log differs')
    print(json.dumps(dict(schema=1,verified=True,credentialsRead=False,paidCalls=0,capsules=outputs,
        scope='immutable keyless implementation history; no current provider or release certification'),indent=2))


if __name__=='__main__':
    try:main()
    except Exception as error:
        print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
