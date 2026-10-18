#!/usr/bin/env python3
"""One pending finite six-process release repair; original authorities stay closed."""
import argparse,json,sys
from pathlib import Path
sys.dont_write_bytecode=True
import paid_core as c
import paid_plan as p
import paid
import routed_evidence as e
import repair_contract as r
import repair_inputs as inputs
import paid_checks

def review_pins(out):
 return dict(sourceCommit=c.doc(out/'source-at-start.json')['sourceCommit'],sourceManifestSHA256=c.digest(out/'source-at-start.json'),binariesManifestSHA256=c.digest(out/'binaries.json'),planSHA256=c.digest(out/'plan.json'),**{field:c.digest(out/leaf) for leaf,field in r.EXTRA_PINS.items()})

def prepare(out,phase):
 history,reuse,scope=inputs.proofs(phase)
 if scope['matrixSHA256']!=inputs.MATRIX_SHA:raise RuntimeError('paid preparation requires original authenticated17 matrix; finalizer alone may merge four after collection')
 frozen=c.committed_sources();out.mkdir(parents=True,exist_ok=False,mode=0o700)
 c.dump(out/'source-at-start.json',frozen)
 for leaf,value in [('history.json',history),('reuse.json',reuse),('scope.json',scope)]:c.dump(out/leaf,value)
 c.dump(out/'plan.json',c.redact(dict(schema=1,round=r.ROUND,phase=phase,budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,history=history,reuse=reuse,scope=scope,collection=e.location(out),routes=r.jobs(out,phase),automaticProcessReruns=False,noSceneSplice=True,proxy=r.PROXY,approvalProposal=r.PROPOSAL,approvalProposalSHA256=r.PROPOSAL_SHA256)))
 auth=p.authorizations(out,c.doc(c.ROOT/'ai/release/catalog-snapshot.json')['hash']) if phase=='effects' else {}
 for name,value in auth.items():
  field='authorizationReference' if name=='protocol-authorization.json' else 'reference';value[field]='barness-ai-minimax-ticket15';c.dump(out/name,value)
 builds=[('live.test',['go','test','-tags','live','-c','-o',out/'live.test','./ai/live'])]
 if phase=='effects':builds.append(('visualeval',['go','build','-o',out/'visualeval','./ai/examples/visualeval']))
 for name,argv in builds:
  if c.execute(out,'build-'+name,argv)!=0:raise RuntimeError('fixed keyless build failed')
 c.dump(out/'binaries.json',{n:c.digest(out/n) for n,_ in builds});c.dump(out/'prepared.json',dict(state='PREPARED',credentialsRead=False,planSHA256=c.digest(out/'plan.json'),authorizationFiles={n:c.digest(out/n) for n in auth}))
 pins=review_pins(out);c.dump(out/'prerelease-source-review.template.json',dict(decision='PENDING',reviewedBy='',**pins,authorizedRoutes=[],credentialScopes={}))
 c.dump(out/'execution-authorization.template.json',dict(decision='PENDING',authority='',requestReference='',requestText='',approvalProposalSHA256=r.PROPOSAL_SHA256,approvedAdditionalMaxCalls=108,approvedAdditionalMaxAttempts=132,round=r.ROUND,phase=phase,budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,**pins,checksSHA256='',sourceReviewSHA256=''))
 guard(out);return dict(state='PREPARED',round=r.ROUND,phase=phase,authorization='PENDING',credentialsRead=False,**pins)

def guard(out):
 frozen=c.guard(out);plan=c.doc(out/'plan.json');phase=plan.get('phase')
 if phase not in ('glm','effects'):raise RuntimeError('fixed glm/effects phase required')
 history,reuse,scope=inputs.proofs(phase)
 if (plan.get('round')!=r.ROUND or plan.get('budget')!=r.BUDGET or plan.get('accountingScope')!=r.ACCOUNTING_SCOPE or plan.get('collection')!=e.location(out) or
     plan.get('routes')!=c.redact(r.jobs(out,phase)) or plan.get('proxy')!=r.PROXY or plan.get('automaticProcessReruns') is not False or
     plan.get('approvalProposalSHA256')!=r.PROPOSAL_SHA256 or plan.get('approvalProposal')!=r.PROPOSAL):raise RuntimeError('fixed six-process contract changed')
 for leaf,value in [('history.json',history),('reuse.json',reuse),('scope.json',scope)]:
  if c.doc(out/leaf)!=value or plan.get(leaf.removesuffix('.json'))!=value:raise RuntimeError('historical native input/source proof changed')
 return frozen

def approved(out,job):
 frozen=guard(out);c.reviewed(out,job);paid_checks.verify(out)
 review=c.doc(out/'prerelease-source-review.json');pins=review_pins(out);value=c.doc(out/'execution-authorization.json')
 if any(review.get(k)!=v for k,v in pins.items()) or review.get('credentialScopes',{}).get(job['sourceVariable'])!=r.ROUND+'/'+job['networkRoute'] or value!=r.expected_authorization(value,c.doc(out/'plan.json')['phase'],pins,c.digest(out/'checks.json'),c.digest(out/'prerelease-source-review.json')):raise RuntimeError('new exact-source/history/reuse/scope reviews and explicit108/132 approval required')
 return frozen

def counters(out,job):
 unknown=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None)
 try:
  relative=out.relative_to(c.ROOT).as_posix()
  with e.Root(c.ROOT) as reader:
   raw=reader.read(relative+'/'+job['baseName']+'-command.json');command=e.document(raw)
   with e.Root(c.common_root()) as common:intent=e.document(common.read('.evidence/minimax-launches/'+job['name']+'.json'))
   if command.get('state')!='COMPLETED' or intent.get('state')!='COMPLETED' or e.u.sha(raw)!=intent.get('commandReceiptSha256'):return unknown
   return e.native_counters(e.document(e.report_bytes(reader,relative,job['baseName'])),job)
 except (OSError,RuntimeError,ValueError,KeyError,TypeError):return unknown

def summarize(out):
 guard(out);rows=[]
 for name,job in r.jobs(out).items():
  marker=paid.marker(name);actual=dict(state='NOT_RUN',status='NOT_RUN',calls=0,attempts=0)
  if marker.exists() or marker.is_symlink():actual=e.settled(job,contract=r)
  known=actual.get('state') in ('REPORTED','NOT_RUN') and type(actual.get('calls')) is int and type(actual.get('attempts')) is int
  if not known:actual=dict(state='UNKNOWN',status='UNKNOWN',calls=None,attempts=None)
  rows.append(dict(route=name,baseName=job['baseName'],networkRoute=job['networkRoute'],maxCalls=job['maxCalls'],maxAttempts=job['maxAttempts'],**{k:actual[k] for k in ('state','status','calls','attempts')},conservativeCalls=actual['calls'] if known else job['maxCalls'],conservativeAttempts=actual['attempts'] if known else job['maxAttempts']))
 current=dict(calls=sum(v['calls'] or 0 for v in rows),attempts=sum(v['attempts'] or 0 for v in rows),conservativeCalls=sum(v['conservativeCalls'] for v in rows),conservativeAttempts=sum(v['conservativeAttempts'] for v in rows),maxCalls=108,maxAttempts=132)
 if current['conservativeCalls']>108 or current['conservativeAttempts']>132:raise RuntimeError('fixed additional108/132 exceeded')
 return dict(round=r.ROUND,budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,routes=rows,current=current,total=dict(calls=545+current['calls'],attempts=545+current['attempts'],conservativeCalls=545+current['conservativeCalls'],conservativeAttempts=545+current['conservativeAttempts'],maxCalls=653,maxAttempts=677),automaticProcessReruns=False,credentialsRead=False)

def run_selected(out,job):
 frozen=approved(out,job);summary=summarize(out)
 if any(v['state']=='UNKNOWN' for v in summary['routes']):raise RuntimeError('unknown consumption retains whole cap and blocks further selected credentials')
 if r.phase_for(job['baseName'])=='effects' and any(v['state']!='REPORTED' or v['status'] not in ('PASS','FAIL') for v in summary['routes'] if v['baseName'].startswith('glm-')):raise RuntimeError('all three own GLM processes must close first; honest known FAIL permits effects without publication')
 evidence=Path(job['evidenceRoot']);base=job['baseName'];name=job['name'];intent=paid.marker(name)
 if evidence.exists() or evidence.is_symlink() or (out/(base+'-command.json')).exists():raise RuntimeError('fresh native output required; no automatic rerun')
 record=dict(state='STARTING',route=name,baseName=base,validationRound=r.ROUND,phase=c.doc(out/'plan.json')['phase'],sourceCommit=frozen['sourceCommit'],sourceVariable=job['sourceVariable'],credentialsRead=False,maxCalls=job['maxCalls'],maxAttempts=job['maxAttempts'],networkRoute=job['networkRoute'],collectionRoot=str(out),collectionOwnerSHA256=e.location(out)['rootSHA256'],automaticReruns=False,accountingScope=r.ACCOUNTING_SCOPE,sourceManifestSha256=c.digest(out/'source-at-start.json'),binariesManifestSha256=c.digest(out/'binaries.json'),planSha256=c.digest(out/'plan.json'),sourceReviewSha256=c.digest(out/'prerelease-source-review.json'),authorizationSha256=c.digest(out/'execution-authorization.json'),**{field:c.digest(out/leaf) for leaf,field in r.EXTRA_PINS.items()})
 c.dump(intent,c.redact(record),exclusive=True)
 code=c.execute(out,base,job['argv'],env=job['environment'],timeout=job['timeoutSeconds'],cwd=job['cwd'],source=job['sourceVariable'],injected=job['injectedCredential'],receipt=intent)
 guard(out);record=c.doc(intent);record['commandReceiptSha256']=c.digest(out/(base+'-command.json'));c.dump(intent,record)
 actual=counters(out,job);record['budget']=actual
 if actual['state'] in ('REPORTED','NOT_RUN'):
  with e.Root(c.ROOT) as reader:record['reportedFileSha256']=e.u.sha(e.report_bytes(reader,out.relative_to(c.ROOT).as_posix(),base))
 c.dump(intent,record)
 if code==0 and base.startswith('glm-') and actual['status']=='PASS':record['capturePins']=paid.pin_glm_capture(out,job|dict(name=base),record);c.dump(intent,record)
 c.dump(out/(name+'-intent.json'),record);c.dump(out/'repair-consumption.json',summarize(out))
 if code!=0 or actual['state']=='UNKNOWN' or actual['status']=='FAIL':raise RuntimeError('finite process failed; retain actual consumption and continue only remaining valid fixed jobs, no retry')
 return dict(route=name,baseName=base,exitCode=code,budget=actual)


def main():
 parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('mode',choices=('prepare','check','preview','run-selected','summarize'));parser.add_argument('out');parser.add_argument('route',nargs='?');parser.add_argument('--phase',choices=('glm','effects'));a=parser.parse_args();out=paid.out_path(a.out)
 if a.mode=='prepare':
  if a.phase is None:raise RuntimeError('prepare requires fixed phase')
  result=prepare(out,a.phase)
 else:
  if a.phase:raise RuntimeError('phase was frozen by prepare')
  guard(out);inventory=r.jobs(out,c.doc(out/'plan.json')['phase'])
  if a.mode=='check':
   result=paid_checks.run(out,guard,pin_receipts=True)
   if not result['passed']:raise RuntimeError('fixed six keyless checks failed; no selected intent or key')
   template=c.doc(out/'execution-authorization.template.json')
   if template.get('decision')!='PENDING':raise RuntimeError('checks may update only a PENDING authorization template')
   template['checksSHA256']=c.digest(out/'checks.json');c.dump(out/'execution-authorization.template.json',template)
  elif a.mode=='preview':
   if a.route not in inventory:raise RuntimeError('one fixed complete native job required')
   result=inventory[a.route]|dict(credentialsRead=False,authorization='PENDING independent current-source reviews and new108/132 reply')
  elif a.mode=='summarize':result=summarize(out)
  else:
   if a.route not in inventory:raise RuntimeError('one fixed complete native job required')
   result=run_selected(out,inventory[a.route])
 print(json.dumps(c.redact(result),ensure_ascii=False,indent=2))
if __name__=='__main__':
 try:main()
 except Exception as error:print('REFUSED: '+str(c.redact(str(error))),file=sys.stderr);sys.exit(1)
