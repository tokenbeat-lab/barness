#!/usr/bin/env python3
"""Complete metadata for immutable bytes after observation selection was refused."""
import gzip,hashlib,json,os
from pathlib import Path
import freeze as f
BASE=f.BASE;sha=lambda raw:hashlib.sha256(raw).hexdigest()

def recovered(label,observe,small=False):
 leaf='inventory.json.gz' if small else label+'-inventory.json.gz';encoded=(BASE/leaf).read_bytes()
 with gzip.GzipFile(fileobj=__import__('io').BytesIO(encoded)) as stream:raw=stream.read((8<<20 if small else 40<<20)+1)
 limit=8<<20 if small else 40<<20
 if len(raw)>limit:raise ValueError('existing inventory bound')
 pins=f.b.doc(raw);source_encoded=(BASE/'source-at-start.json.gz').read_bytes();source=gzip.decompress(source_encoded);archive=f.EXT/('glm-raw.tar.gz' if small else 'glm-'+label+'-raw.tar.gz');digest=hashlib.sha256()
 with f.b.opened(f.COMMON,archive.relative_to(f.COMMON).as_posix()) as stream:
  for chunk in iter(lambda:stream.read(65536),b''):digest.update(chunk)
 cap=dict(sourceCommit=f.SOURCE,archive=dict(file=archive.relative_to(f.COMMON).as_posix(),bytes=archive.stat().st_size,sha256=digest.hexdigest(),prefix='repair-glm-'+('actual' if small else label)),inventory=dict(file=leaf,compressedSHA256=sha(encoded),decodedBytes=len(raw),decodedSHA256=sha(raw),files=len(pins)),sourceSnapshot=dict(file='source-at-start.json.gz',compressedSHA256=sha(source_encoded),decodedBytes=len(source),decodedSHA256=sha(source),archiveMember='frozen/source-at-start.json'),observe=sorted(set(observe)),originalsUnchangedAfterSnapshot=True)
 if small:cap['id']='actualRepairGLM7931';f.b.verify_archive(f.COMMON,cap)
 else:cap['limits']=f.large.LIMITS;f.large.verify(f.b,BASE,f.COMMON,cap)
 archive.chmod(0o444);return cap,pins

def main():
 if (BASE/'index.json').exists():raise ValueError('new index is exclusive; do not repeat completed retention')
 paid=f.b.doc(gzip.decompress((BASE/'inventory.json.gz').read_bytes()))
 observed=[n for n in paid if n.startswith(('frozen/','canonical/','selectors/','references/','controllers/')) or n.startswith('actual/') and n.endswith(('/formal-glm-report.json','/glm-integrity.json','/glm-execution-source.json','/live-report.json','/audit.json','/assertions.json','/attempt-1-result.json','/attempt-1-exchanges.json'))]
 cap,pins=recovered('actual',observed,True);preflight=f.b.doc(gzip.decompress((BASE/'preflight-inventory.json.gz').read_bytes()));pobs=['frozen/source-at-start.json','frozen/actual-keyless-case-summary.json','projection/known-intentional-link.json']
 for item in f.b.doc(f.raw(f.G/'actual-keyless-case-summary.json',1<<20))['manifests']:
  name='preflight/'+item['manifest'].removeprefix('.evidence/minimax-release-repair-glm-1/');pobs.append(name);audit=name.removesuffix('manifest.json')+'audit.json'
  if audit in preflight:pobs.append(audit)
 pcap,_=recovered('preflight',pobs)
 binary={'frozen/'+n:f.G/n for n in ('live.test','source-at-start.json','binaries.json','build-live.test-command.json')};bcap,_=f.snapshot('binary',binary,{},False,[n for n in binary if n!='frozen/live.test'])
 roots={route:next(n.removesuffix('/formal-glm-report.json') for n in pins if n.startswith('actual/'+route+'/') and n.endswith('/formal-glm-report.json')) for route in ('chat','responses','messages')}
 projection=dict(originalRelativePath=f.KNOWN_LINK,targetSHA256=f.LINK_SHA,archiveMember='projection/known-intentional-link.json',linksFollowed=False)
 if sha(os.readlink(os.fsencode(f.G/f.KNOWN_LINK)))!=f.LINK_SHA:raise ValueError('original link metadata changed')
 index=dict(schema=1,kind='actual-failed-repair-glm-three',actualSource=f.SOURCE,round=f.ROUND,boundedHelperSHA256=f.readers.SMALL,fullReaderSHA256=f.readers.FULL,intentionalLinkProjection=projection,routes=roots,capsule=cap,preflight=pcap,binary=bcap,realCredentialsReadByRetention=False,retentionPaidCalls=0)
 f.dump(BASE/'index.json',index)
 f.dump(BASE/'snapshot-authoring-history.json',dict(initialPublicVerifier='missing entrypoint RED before implementation',initialScan='unexpected intentional fixture symlink refused; reproduced numeric1 before exact metadata-only projection',firstSnapshot='paid archive closed; preflight regular bytes/inventory/post-hashes complete, observation count refused before index publication',completion='existing paid/preflight archives verified without rewriting; choose only four declared actual manifests within original256 observation cap; binary archived once',rawArchivesRewritten=False))
 print(json.dumps(dict(completed=True,paid=cap['archive'],preflight=pcap['archive'],binary=bcap['archive']),indent=2))
if __name__=='__main__':main()
