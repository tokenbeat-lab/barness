#!/usr/bin/env python3
"""Bounded localTLS deadline evidence replay; never executes captured tests or calls vendors."""
import argparse,collections,hashlib,json,os,stat,sys,types
from pathlib import Path,PurePosixPath
BASE=Path(__file__).resolve().parent
HELPER='b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f';CODE='dcd30f7ce6bf65f5bd1101697c8e5c98155ceb98'
SOURCE='d8e7100695a5747a801bf5f8966cef1a2d129e4c';ASSEMBLY='ai/live/glm_assembly_test.go'

def helper():
 p=BASE.parents[1]/'ticket-13/implementation/verify.py';fd=os.open(p,os.O_RDONLY|os.O_NOFOLLOW|os.O_NONBLOCK)
 with os.fdopen(fd,'rb') as f:info=os.fstat(f.fileno());raw=f.read((64<<10)+1)
 if not stat.S_ISREG(info.st_mode) or len(raw)>64<<10 or hashlib.sha256(raw).hexdigest()!=HELPER:raise ValueError('pinned reader differs')
 b=types.ModuleType('bounded_headers');b.__file__=str(p);exec(compile(raw,str(p),'exec'),b.__dict__);b.BASE=BASE;return b

def analyse(b,index,m):
 def doc(n):return b.doc(m[n])
 pins=b.inventory(index['capsule']);red=doc('tdd/source-red-headers.json');green=doc('tdd/source-green-headers.json');original=doc('tdd/source-before.json')
 if red['sourceCommit']!='06229f68de6e9dc3100724f05365bb4bd414667a' or green['sourceCommit']!=SOURCE or green['workingTreeChanged'] is not True or len(green['files'])!=1048:raise ValueError('measured HEAD/WIP identity differs')
 if [n for n in green['files'] if red['files'].get(n)!=green['files'][n]]!=[ASSEMBLY]:raise ValueError('header slice source scope differs')
 old=m['tdd/assembly-red-headers.go'];new=m['tdd/assembly-green-headers.go']
 line=b'\tpolicy.CallTimeout = 120 * time.Second\n'
 if old.count(line)!=1 or old.replace(line,line+b'\tpolicy.ResponseHeaderTimeout = 120 * time.Second\n')!=new or new!=m['source/glm_assembly_test.go']:raise ValueError('production RED/GREEN delta differs')
 if b.sha(old)!=red['files'][ASSEMBLY] or b.sha(new)!=green['files'][ASSEMBLY] or b.sha(m['tdd/assembly-committed-before.go'])!=original['files'][ASSEMBLY]:raise ValueError('assembly snapshot differs')
 note=doc('tdd/setup-copy-note.json')
 if note['initialCopySHA256']!=pins['tdd/assembly-before.go']['sha256'] or note['committedSourceSHA256']!=original['files'][ASSEMBLY] or m['tdd/assembly-before.go']==m['tdd/assembly-committed-before.go']:raise ValueError('setup WIP projection history differs')
 outcomes={}
 for stage in ('red-headers','green-headers','green-normal','green-host-cancel','race-host-cancel','vet','vet-live'):
  v=doc('tdd/'+stage+'-command.json');code=1 if stage=='red-headers' else 0
  if (v['state']!='COMPLETED' or v['exitCode']!=code or v['credentialsRead'] is not False or v['credentialsInherited'] is not False or v['sourceVariable'] is not None or v['injectedCredential'] is not None or v['launcherAudit']!={'passed':True,'findings':[]} or any(any(k2 in k for k2 in ('KEY','TOKEN','SECRET')) for k in v['environmentKeys'])):raise ValueError('keyless command differs')
  if stage!='red-headers' and doc('tdd/source-'+stage+'.json')['files']!=green['files']:raise ValueError('GREEN source changed')
  outcomes[stage]=dict(exitCode=code,seconds=v['seconds'])
 if doc('tdd/red-headers-command.json')['argv']!=doc('tdd/green-headers-command.json')['argv'] or b'ResponseHeaderTimeout (1m30s)' not in m['tdd/red-headers.log']:raise ValueError('genuine header RED differs')
 manifests=[];refs=0;late=None
 for n in sorted(m):
  if not n.endswith('/manifest.json'):continue
  v=doc(n);root=n.rsplit('/',1)[0];audit=doc(root+'/audit.json')
  if audit['findings']:raise ValueError('native audit findings')
  counts=dict(collections.Counter(c['status'] for c in v['cases']));stage=n.split('/')[1]
  want={'FAIL':1} if stage=='red-headers' else {'PASS':1 if stage=='green-headers' else 2 if stage=='green-normal' else 179}
  if counts!=want:raise ValueError('native case outcomes differ')
  for case in v['cases']:
   directory=PurePosixPath(case['dir'])
   if directory.is_absolute() or '..' in directory.parts:raise ValueError('unsafe native artifact path')
   for leaf,digest in case['artifacts'].items():
    b.relative(leaf)
    if pins.get(root+'/'+str(directory/leaf),{}).get('sha256')!=digest:raise ValueError('native artifact pin differs')
    refs+=1
   if stage=='green-headers':late=root+'/'+str(directory)
  manifests.append(dict(member=n,cases=len(v['cases']),statuses=counts,auditFindings=0))
 if len(manifests)!=5 or late is None:raise ValueError('fixed native bundles missing')
 timing=doc(late+'/timing.json');result=doc(late+'/result.json')
 if (not 90000<timing['headersAfterMs']<120000 or not timing['headersAfterMs']<=timing['callElapsedMs']<120000 or timing['transportRequests']!=1 or timing['localTLSPeer'] is not True or timing['vendorCalls']!=0 or
     len(result['metadata']['attempts'])!=1 or result['message']['stopReason']!='stop' or result['message']['content'][0]['text']!='Controlled GLM.'):raise ValueError('real controlled91s boundary differs')
 impact=doc('tdd/selected-source-impact.json')
 for route in ('chat','responses','messages'):
  report=doc('historical/m3-'+route+'.json');files=report['executionSource']
  if len(files)!=263 or any((b.sha(b.read(BASE,'M3-selected-README.md',1<<20)) if n=='ai/internal/testkit/minimax/README.md' else green['files'].get(n))!=h for n,h in files.items()):raise ValueError('M3 formal selected files changed')
  glm=doc('historical/glm-'+route+'-source.json')
  if len(glm)!=217 or [n for n,h in glm.items() if green['files'].get(n)!=h]!=[ASSEMBLY]:raise ValueError('GLM prior source expiry differs')
  mmrow=next(r for r in impact['evidence'] if r['route']==route and r['kind'].startswith('M3'));grow=next(r for r in impact['evidence'] if r['route']==route and r['kind'].startswith('GLM'))
  if mmrow['reportSHA256']!=pins['historical/m3-'+route+'.json']['sha256'] or grow['sourceMapSHA256']!=pins['historical/glm-'+route+'-source.json']['sha256'] or mmrow['changed']!=[] or grow['changed']!=[ASSEMBLY] or grow['expired'] is not True:raise ValueError('source applicability proof differs')
 for leaf in ('selected-source-impact.json','setup-copy-note.json'):
  if b.read(BASE,leaf,1<<20)!=m['tdd/'+leaf]:raise ValueError('frontdoor/raw proof differs')
 return dict(schema=1,codeCommit=CODE,measuredGREENHEAD=SOURCE,workingTreeChanged=True,sourceFiles=1048,assemblySHA256=b.sha(new),commands=outcomes,nativeManifests=manifests,artifactReferencesVerified=refs,lateHeaders=timing,syntheticUsage=result['metadata']['attempts'][0]['usageReporting'],M3FormalSelectedMapsUnchanged=3,priorGLMSelectedMapsExpired=3,priorGLMFailuresUnchanged=True,setupCopyExplicitlyWIP=True,credentialsRead=False,paidCalls=0,vendorCalls=0,releaseCertified=False,scope='Local controlled TLS proves only prospective90s→120s header-policy behavior; no vendor120s success or new paid authority')

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--common-root',type=Path,default=BASE.parents[4]);a=p.parse_args();b=helper()
 for line in b.read(BASE,'SHA256SUMS',1<<20).decode().splitlines():
  digest,name=line.split('  ',1)
  if b.sha(b.read(BASE,name,8<<20))!=digest:raise ValueError('local metadata pin differs')
 ix=b.doc(b.read(BASE,'index.json',1<<20));cap=ix['capsule']
 if ix['codeCommit']!=CODE or cap['id']!='controlledGLMHeader120' or cap['sourceCommit']!=SOURCE or cap['archive']['file']!='.evidence/minimax-retained/ticket-16-glm-header-alignment/deadline-alignment-raw.tar.gz' or cap['archive']['prefix']!='glm-header-alignment' or len(cap['observe'])>512:raise ValueError('fixed capsule identity differs')
 pins=b.inventory(cap)
 if len(set(cap['observe']))!=len(cap['observe']) or not set(cap['observe'])<=set(pins) or sum(pins[n]['bytes'] for n in cap['observe'])>16<<20:raise ValueError('bounded metadata observation differs')
 archive,m=b.verify_archive(a.common_root,cap);summary=analyse(b,ix,m)
 if summary!=b.doc(b.read(BASE,'summary.json',1<<20)):raise ValueError('independent controlled summary differs')
 print(json.dumps(dict(verified=True,codeCommit=CODE,credentialsRead=False,paidCalls=0,vendorCalls=0,releaseCertified=False,archive=archive),indent=2))

if __name__=='__main__':
 try:main()
 except Exception as error:print('REFUSED: '+type(error).__name__+': '+str(error),file=sys.stderr);sys.exit(1)
