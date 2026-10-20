#!/usr/bin/env python3
"""One fixed current-source eleven renewal; old authorities remain closed.

Approval/phase logic stays separate from closed earlier authorities: this round
requires three closed known GLM results, a new M3 grant, and seven-row
publication only after all own GLM routes PASS. Stable execution,
source freezes, native settlement and preflights use the existing paid modules.
"""
import argparse,json,sys
from pathlib import Path
sys.dont_write_bytecode=True
import paid_core as c
import paid_plan as p
import paid
import routed_evidence as e
import eleven_contract as r
import eleven_inputs as inputs
import paid_checks
import eleven_grant as grant

def review_pins(out):
 return dict(sourceCommit=c.doc(out/'source-at-start.json')['sourceCommit'],sourceManifestSHA256=c.digest(out/'source-at-start.json'),binariesManifestSHA256=c.digest(out/'binaries.json'),planSHA256=c.digest(out/'plan.json'),**{field:c.digest(out/leaf) for leaf,field in r.EXTRA_PINS.items()})

def prepare(out,phase):
 history,reuse,scope=inputs.proofs(phase)
 frozen=c.committed_sources();out.mkdir(parents=True,exist_ok=False,mode=0o700)
 c.dump(out/'source-at-start.json',frozen)
 for leaf,value in [('history.json',history),('generic.json',reuse),('transition.json',scope)]:c.dump(out/leaf,value)
 c.dump(out/'plan.json',c.redact(dict(schema=1,round=r.ROUND,phase=phase,budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,history=history,generic=reuse,transition=scope,allJobs=r.jobs(out),collection=e.location(out),routes=r.jobs(out,phase),automaticProcessReruns=False,noSceneSplice=True,proxy=r.PROXY,approvalProposal=r.PROPOSAL,approvalProposalSHA256=r.PROPOSAL_SHA256)))
 auth=p.authorizations(out,c.doc(c.ROOT/'ai/release/catalog-snapshot.json')['hash']) if phase=='remaining' else {}
 for name,value in auth.items():
  field='authorizationReference' if name=='protocol-authorization.json' else 'reference';value[field]='barness-ai-minimax-ticket19';c.dump(out/name,value)
 builds=[('live.test',['go','test','-tags','live','-c','-o',out/'live.test','./ai/live'])]
 if phase=='remaining':builds.extend([('visualeval',['go','build','-o',out/'visualeval','./ai/examples/visualeval']),('chineseeval',['go','build','-o',out/'chineseeval','./ai/examples/chineseeval/cmd/chineseeval']),('minimaxverify',['go','build','-o',out/'minimaxverify','./ai/live/cmd/minimaxverify'])])
 for name,argv in builds:
  if c.execute(out,'build-'+name,argv)!=0:raise RuntimeError('fixed keyless build failed')
 c.dump(out/'binaries.json',{n:c.digest(out/n) for n,_ in builds});grant.capture(out,phase)
 auth['native-renewal-facts.json']=c.doc(out/'native-renewal-facts.json')
 c.dump(out/'prepared.json',dict(state='PREPARED',credentialsRead=False,planSHA256=c.digest(out/'plan.json'),authorizationFiles={n:c.digest(out/n) for n in auth}))
 pins=review_pins(out);c.dump(out/'prerelease-source-review.template.json',dict(decision='PENDING',reviewedBy='',**pins,authorizedRoutes=[],credentialScopes={}))
 c.dump(out/'execution-authorization.template.json',dict(decision='PENDING',authority='',requestReference='',requestText='',approvalProposalSHA256=r.PROPOSAL_SHA256,approvedAdditionalMaxCalls=216,approvedAdditionalMaxAttempts=252,round=r.ROUND,phase=phase,budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,**pins,checksSHA256='',sourceReviewSHA256=''))
 guard(out);return dict(state='PREPARED',round=r.ROUND,phase=phase,authorization='PENDING',credentialsRead=False,**pins)

def guard(out):
 frozen=c.guard(out);plan=c.doc(out/'plan.json');phase=plan.get('phase')
 if phase not in ('glm','remaining'):raise RuntimeError('fixed glm/remaining phase required')
 history,reuse,scope=inputs.proofs(phase)
 if (plan.get('round')!=r.ROUND or plan.get('budget')!=r.BUDGET or plan.get('accountingScope')!=r.ACCOUNTING_SCOPE or plan.get('collection')!=e.location(out) or
     plan.get('routes')!=c.redact(r.jobs(out,phase)) or plan.get('allJobs')!=c.redact(r.jobs(out)) or plan.get('proxy')!=r.PROXY or plan.get('automaticProcessReruns') is not False or
     plan.get('approvalProposalSHA256')!=r.PROPOSAL_SHA256 or plan.get('approvalProposal')!=r.PROPOSAL):raise RuntimeError('fixed eleven-process contract changed')
 if phase=='remaining':grant.validate_facts(out)
 for leaf,value in [('history.json',history),('generic.json',reuse),('transition.json',scope)]:
  if c.doc(out/leaf)!=value or plan.get(leaf.removesuffix('.json'))!=value:raise RuntimeError('historical native input/source proof changed')
 return frozen

def approved(out,job,require_human=True):
 frozen=guard(out);paid_checks.verify(out)
 review=c.doc(out/'prerelease-source-review.json');pins=review_pins(out);value=c.doc(out/'execution-authorization.json')
 if review.get('decision')!='APPROVED' or not review.get('reviewedBy') or job['name'] not in review.get('authorizedRoutes',[]) or any(review.get(k)!=v for k,v in pins.items()) or review.get('credentialScopes',{}).get(job['sourceVariable'])!=r.ROUND+'/'+job['networkRoute'] or value!=r.expected_authorization(value,c.doc(out/'plan.json')['phase'],pins,c.digest(out/'checks.json'),c.digest(out/'prerelease-source-review.json')):raise RuntimeError('new exact-source/history/reuse/scope reviews and explicit216/252 approval required')
 if require_human:human(out)
 return frozen

def human(out,create=False):
 value=c.doc(out/'execution-authorization.json')
 expected=dict(schema=1,decision='APPROVED',renewalId=r.ROUND,maxCalls=216,maxAttempts=252,requestReference=value['requestReference'],requestText=value['requestText'],approvalProposalSHA256=r.PROPOSAL_SHA256,firstPhaseSourceCommit=c.doc(out/'source-at-start.json')['sourceCommit'])
 path=c.common_root()/'.evidence/minimax-renewals'/r.ROUND/'human-approval.json'
 if create:
  if c.doc(out/'plan.json')['phase']!='glm':raise RuntimeError('human authority starts only at initial GLM phase')
  c.dump(path,expected,exclusive=True)
 else:
  with e.Root(c.common_root()) as reader:original=e.document(reader.read(path.relative_to(c.common_root()).as_posix()))
  if c.doc(out/'plan.json')['phase']=='remaining':expected['firstPhaseSourceCommit']=c.doc(out/'transition.json')['glmSourceCommit']
  if any(type(original.get(n)) is not int for n in ('schema','maxCalls','maxAttempts')):raise RuntimeError('exact integer human budget required')
  if original!=expected:raise RuntimeError('one immutable new human approval must cover both phases')
 return dict(credentialsRead=False,humanApprovalSHA256=c.digest(path),authorization='APPROVED fixed216/252 only')

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
 current=dict(calls=sum(v['calls'] or 0 for v in rows),attempts=sum(v['attempts'] or 0 for v in rows),conservativeCalls=sum(v['conservativeCalls'] for v in rows),conservativeAttempts=sum(v['conservativeAttempts'] for v in rows),maxCalls=216,maxAttempts=252)
 if current['conservativeCalls']>216 or current['conservativeAttempts']>252:raise RuntimeError('fixed additional216/252 exceeded')
 return dict(round=r.ROUND,budget=r.BUDGET,accountingScope=r.ACCOUNTING_SCOPE,routes=rows,current=current,total=dict(calls=648+current['calls'],attempts=648+current['attempts'],conservativeCalls=648+current['conservativeCalls'],conservativeAttempts=648+current['conservativeAttempts'],maxCalls=864,maxAttempts=900),automaticProcessReruns=False,credentialsRead=False)

def run_selected(out,job):
 frozen=approved(out,job);summary=summarize(out)
 if any(v['state']=='UNKNOWN' for v in summary['routes']):raise RuntimeError('unknown consumption retains whole cap and blocks further selected credentials')
 if r.phase_for(job['baseName'])=='remaining':inputs.glm_collection()
 evidence=Path(job['evidenceRoot']);base=job['baseName'];name=job['name'];intent=paid.marker(name)
 if evidence.exists() or evidence.is_symlink() or (out/(base+'-command.json')).exists():raise RuntimeError('fresh native output required; no automatic rerun')
 record=dict(state='STARTING',route=name,baseName=base,validationRound=r.ROUND,phase=c.doc(out/'plan.json')['phase'],sourceCommit=frozen['sourceCommit'],sourceVariable=job['sourceVariable'],credentialsRead=False,maxCalls=job['maxCalls'],maxAttempts=job['maxAttempts'],networkRoute=job['networkRoute'],collectionRoot=str(out),collectionOwnerSHA256=e.location(out)['rootSHA256'],automaticReruns=False,accountingScope=r.ACCOUNTING_SCOPE,sourceManifestSha256=c.digest(out/'source-at-start.json'),binariesManifestSha256=c.digest(out/'binaries.json'),planSha256=c.digest(out/'plan.json'),sourceReviewSha256=c.digest(out/'prerelease-source-review.json'),authorizationSha256=c.digest(out/'execution-authorization.json'),**{field:c.digest(out/leaf) for leaf,field in r.EXTRA_PINS.items()})
 if base.startswith('mm-'):record.update(renewalId=r.ROUND,renewalGrantSha256=grant.require(out))
 c.dump(intent,c.redact(record),exclusive=True)
 code=c.execute(out,base,job['argv'],env=job['environment'],timeout=job['timeoutSeconds'],cwd=job['cwd'],source=job['sourceVariable'],injected=job['injectedCredential'],receipt=intent)
 guard(out);record=c.doc(intent);record['commandReceiptSha256']=c.digest(out/(base+'-command.json'));c.dump(intent,record)
 actual=counters(out,job);record['budget']=actual
 if actual['state'] in ('REPORTED','NOT_RUN'):
  with e.Root(c.ROOT) as reader:record['reportedFileSha256']=e.u.sha(e.report_bytes(reader,out.relative_to(c.ROOT).as_posix(),base))
 c.dump(intent,record)
 if code==0 and base.startswith('glm-') and actual['status']=='PASS':record['capturePins']=paid.pin_glm_capture(out,job|dict(name=base),record);c.dump(intent,record)
 if code==0 and base.startswith('mm-') and actual['status']=='PASS':
  record['origin']=paid.pin_actual_origin(out,job|dict(name=base),record);c.dump(intent,record)
  if c.execute(out,base+'-origin',[out/'minimaxverify','-formal','-publish',r.report_path(out,job).parent],timeout=120)!=0:raise RuntimeError('own native origin verification/publication failed; no rerun')
 c.dump(out/(name+'-intent.json'),record);c.dump(out/'eleven-consumption.json',summarize(out))
 if code!=0 or actual['state']=='UNKNOWN' or actual['status']=='FAIL':raise RuntimeError('finite process failed; retain actual consumption and continue only remaining valid fixed jobs, no retry')
 return dict(route=name,baseName=base,exitCode=code,budget=actual)


def main():
 parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('mode',choices=('prepare','check','preview','run-selected','summarize','authorize','grant','finalize'));parser.add_argument('out');parser.add_argument('route',nargs='?');parser.add_argument('--phase',choices=('glm','remaining'));a=parser.parse_args();out=paid.out_path(a.out)
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
   result=inventory[a.route]|dict(credentialsRead=False,authorization='PENDING independent current-source reviews and new216/252 reply')
  elif a.mode=='summarize':result=summarize(out)
  elif a.mode=='grant':result=grant.create(out)
  elif a.mode=='finalize':
   import eleven_finalize
   result=eleven_finalize.run(out)
  elif a.mode=='authorize':
   for job in inventory.values():approved(out,job,require_human=False)
   result=human(out,create=True)
  else:
   if a.route not in inventory:raise RuntimeError('one fixed complete native job required')
   result=run_selected(out,inventory[a.route])
 print(json.dumps(c.redact(result),ensure_ascii=False,indent=2))
if __name__=='__main__':
 try:main()
 except Exception as error:print('REFUSED: '+str(c.redact(str(error))),file=sys.stderr);sys.exit(1)
