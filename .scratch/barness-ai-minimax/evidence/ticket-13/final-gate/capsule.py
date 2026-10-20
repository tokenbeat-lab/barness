"""Finite streaming reader for each fixed complete ordinary/race suite.

This is a distinct retention protocol: the earlier implementation capsule has
32MiB/192MiB/32k limits and remains unchanged. Here 133k complete-suite files
require explicit fixed larger limits; no captured command is executed.
"""
import gzip
import hashlib
import io
import os
import re
import tarfile

LIMITS=dict(compressedBytes=256<<20,decodedTarBytes=2<<30,files=160000,entries=200000,
            memberBytes=64<<20,observedMetadataBytes=16<<20,observedTotalBytes=64<<20,
            inventoryDecodedBytes=40<<20,inventoryCompressedBytes=8<<20,rawMemberBytes=1600<<20)


def decoded(b,base,info,limit):
    encoded=b.read(base,info['file'],LIMITS['inventoryCompressedBytes'])
    if b.sha(encoded)!=info['compressedSHA256']:raise ValueError('compressed metadata pin differs')
    with gzip.GzipFile(fileobj=io.BytesIO(encoded)) as stream:raw=stream.read(limit+1)
    if len(raw)>limit or len(raw)!=info['decodedBytes'] or b.sha(raw)!=info['decodedSHA256']:raise ValueError('decoded metadata pin or limit')
    return raw


def verify(b,base,common,index):
    if index['limits']!=LIMITS:raise ValueError('fixed streaming limits differ')
    pins=b.doc(decoded(b,base,index['inventory'],LIMITS['inventoryDecodedBytes']))
    if len(pins)!=index['inventory']['files'] or len(pins)>LIMITS['files']:raise ValueError('fixed file count limit')
    for name,pin in pins.items():
        b.relative(name)
        if (type(pin.get('bytes')) is not int or not 0<=pin['bytes']<=LIMITS['memberBytes'] or
            not isinstance(pin.get('sha256'),str) or re.fullmatch('[a-f0-9]{64}',pin['sha256']) is None):raise ValueError('invalid regular member pin')
    if sum(v['bytes'] for v in pins.values())>LIMITS['rawMemberBytes']:raise ValueError('fixed raw member byte limit')
    observed=index['observe']
    if len(observed)>256 or len(observed)!=len(set(observed)) or not set(observed)<=set(pins):raise ValueError('fixed observed metadata inventory')
    if sum(pins[n]['bytes'] for n in observed)>LIMITS['observedTotalBytes']:raise ValueError('observed total byte limit')
    expected_source=decoded(b,base,index['sourceSnapshot'],1<<20)
    if b.doc(expected_source)['sourceCommit']!=index['sourceCommit'] or pins[index['sourceSnapshot']['archiveMember']]['sha256']!=b.sha(expected_source):raise ValueError('original source snapshot pin differs')
    archive=index['archive'];selected={};seen=set();entries=0
    with b.opened(common,archive['file']) as stream:
        before=os.fstat(stream.fileno())
        if before.st_size!=archive['bytes'] or before.st_size>LIMITS['compressedBytes']:raise ValueError('compressed archive size or limit')
        digest=hashlib.sha256();count=0
        for raw in iter(lambda:stream.read(1<<20),b''):
            count+=len(raw)
            if count>LIMITS['compressedBytes']:raise ValueError('compressed archive byte limit')
            digest.update(raw)
        if count!=archive['bytes'] or digest.hexdigest()!=archive['sha256']:raise ValueError('compressed archive pin differs')
        stream.seek(0)
        with gzip.GzipFile(fileobj=stream) as zipped:
            limited=b.Limited(zipped,LIMITS['decodedTarBytes'])
            with tarfile.open(fileobj=limited,mode='r|') as tar:
                for member in tar:
                    entries+=1
                    if entries>LIMITS['entries']:raise ValueError('fixed tar entry limit')
                    parts=b.relative(member.name)
                    if len(parts)>40 or parts[0]!=archive['prefix'] or len(parts)<2 or not member.isfile():raise ValueError('owned regular tar member required')
                    name='/'.join(parts[1:]);pin=pins.get(name)
                    if name in seen or pin is None or member.size!=pin['bytes']:raise ValueError('duplicate, foreign or resized tar member')
                    if name in observed and member.size>LIMITS['observedMetadataBytes']:raise ValueError('observed member byte limit')
                    digest=hashlib.sha256();size=0;chunks=[]
                    with tar.extractfile(member) as original:
                        for raw in iter(lambda:original.read(65536),b''):
                            size+=len(raw)
                            if size>pin['bytes']:raise ValueError('regular member byte limit')
                            digest.update(raw)
                            if name in observed:chunks.append(raw)
                    if size!=pin['bytes'] or digest.hexdigest()!=pin['sha256']:raise ValueError('original member pin differs')
                    if name in observed:selected[name]=b''.join(chunks)
                    seen.add(name)
            while limited.read(65536):pass
            expanded=limited.count
        after=os.fstat(stream.fileno())
        if (before.st_size,before.st_mtime_ns,before.st_ctime_ns)!=(after.st_size,after.st_mtime_ns,after.st_ctime_ns):raise ValueError('archive changed while verifying')
    if seen!=set(pins):raise ValueError('missing original archive member')
    return pins,selected,dict(files=len(seen),entries=entries,rawBytes=sum(v['bytes'] for v in pins.values()),decodedTarBytes=expanded,
                              archiveBytes=archive['bytes'],archiveSHA256=archive['sha256'])
