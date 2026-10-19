"""Only current own eleven collections may atomically replace seven matrix rows."""
import os,tempfile
import paid_core as c
import paid_plan as p
import paid
import renewal_common as u
import eleven_contract as r
import eleven_inputs as inputs

def run(out):
 from eleven_validation import guard,summarize
 guard(out)
 if c.doc(out/'plan.json')['phase']!='remaining':raise RuntimeError('finalization requires remaining phase')
 actual=summarize(out)
 if len(actual['routes'])!=11 or any(v['state']!='REPORTED' or v['status']!='PASS' for v in actual['routes']):raise RuntimeError('all eleven own complete PASS reports required; failed scenes cannot be spliced')
 glm,_=inputs.glm_collection()
 if c.doc(out/'transition.json')['publication'].get('threeOwnPASS') is not True:raise RuntimeError('seven-row promotion requires strict same-source GLM publication')
 target=u.repo_path('ai/live/support-matrix.json');before=u.file_bytes(target);original=c.doc(target)
 if u.sha(before)!=c.doc(out/'transition.json')['matrixSHA256']:raise RuntimeError('matrix changed before atomic seven-row merge')
 intent=c.common_root()/'.evidence/minimax-renewals'/r.ROUND/'finalize-intent.json'
 c.dump(intent,dict(state='STARTING',credentialsRead=False,sourceCommit=c.git('rev-parse','HEAD'),matrixBeforeSHA256=u.sha(before)),exclusive=True)
 destination=out/'finalization';destination.mkdir(mode=0o700)
 candidate=destination/'matrix-candidate.json';candidate.write_bytes(before)
 support=destination/'jev-support';protocol=p.unique(out/'protocol114','multimodal-report.json')
 u.public(destination,'derive-jev-support',['go','run','./ai/live/cmd/jevformal','-protocol',protocol,'-effects',out/'visual-jev','-out',support,'-require-live'])
 guard(out)
 commands=[]
 for route in ('chat','responses','messages'):
  bundle=p.unique(out/('mm-global-formal-m3-'+route),'minimax-report.json')
  commands.append(('verify-m3-'+route,[out/'minimaxverify','-formal',bundle]))
 for route in ('jev','luna'):commands.append(('verify-text-'+route,[out/'chineseeval','-route',route,'-verify',p.unique(out/('text-'+route),'evaluation-report.json')]))
 for route in ('jev','decisions'):commands.append(('verify-visual-'+route,[out/'visualeval','-verify',out/('visual-'+route),'-require-live']))
 commands.append(('verify-derived-jev',['go','run','./ai/live/cmd/jevformal','-verify',support,'-require-live']))
 for name,argv in commands:u.public(destination,name,argv);guard(out)
 generic=c.doc(out/'generic.json')['inputs'];root=c.common_root()
 bundles=[root/generic[name]['bundle'] for name,_,_,_ in p.GENERIC]
 bundles+=[p.unique(glm/('glm-'+route),'live-report.json') for route in ('chat','responses','messages')]
 bundles+=[p.unique(out/('mm-global-formal-m3-'+route),'live-report.json') for route in ('chat','responses','messages')]
 bundles.append(support)
 u.public(destination,'validate-seventeen-current-inputs',['go','run','./ai/live/cmd/supportmatrix','-matrix',candidate,*bundles])
 guard(out);updated=c.doc(candidate);oldrows={v['combo']:v for v in original['rows']};newrows={v['combo']:v for v in updated['rows']};protected={n for n,_,_,_ in p.GENERIC}
 if (len(oldrows)!=17 or len(newrows)!=17 or len(updated['rows'])!=17 or set(oldrows)!=set(newrows) or
     {k:v for k,v in original.items() if k!='rows'}!={k:v for k,v in updated.items() if k!='rows'} or
     any(newrows[n]!=oldrows[n] for n in protected) or any(not newrows[n].get('allPassedAt') for n in set(newrows)-protected)):
  raise RuntimeError('atomic publication must preserve ten generic rows and contain seven own successful rows')
 raw=u.file_bytes(candidate)
 if u.file_bytes(target)!=before:raise RuntimeError('matrix changed during public candidate validation')
 fd,name=tempfile.mkstemp(prefix='.eleven-matrix-',dir=target.parent)
 try:
  with os.fdopen(fd,'wb') as stream:stream.write(raw);stream.flush();os.fsync(stream.fileno())
  os.replace(name,target)
 finally:
  if os.path.exists(name):os.unlink(name)
 result=dict(state='COMPLETED',credentialsRead=False,strictElevenCurrent=True,releaseCertified=False,updatedRows=7,preservedGenericRows=10,verifiedSupportInputs=17,sourceCommit=c.git('rev-parse','HEAD'),matrixBeforeSHA256=u.sha(before),matrixAfterSHA256=u.sha(raw),current=actual['current'],total=actual['total'],noAutomaticReruns=True)
 c.dump(destination/'result.json',result);c.dump(intent,result)
 return result
