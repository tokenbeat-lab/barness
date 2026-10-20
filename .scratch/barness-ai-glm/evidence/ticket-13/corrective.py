#!/usr/bin/env python3
"""One separately reserved 36-call source-corrective round; no prior balance."""
import argparse
import json
from pathlib import Path
import subprocess
import sys
sys.dont_write_bytecode=True
import commands as c
from run import Round, ROUTES

NAMES=('protocol114','visual-decisions','visual-jev')
COMPONENTS=tuple(row for row in ROUTES if row[0] in NAMES)
assert tuple(row[0] for row in COMPONENTS)==NAMES
assert sum(row[3] for row in COMPONENTS)==36
ORIGINAL=c.ROOT/'.evidence/glm13-legacy-current-r3'


def source_files():
 files=c.sources()
 files[str(Path(__file__).resolve().relative_to(c.ROOT))]=c.digest(Path(__file__))
 return files


def plan():
 return dict(reference='barness-ai-glm-ticket13-corrective-r1',
  source='authorized R21 necessary completion after retained actual CLI failure',
  maxCalls=36,maxAttempts=36,providerCaps={'typesafe':24,'openai':12},
  previousTaskBalanceCarried=False,automaticReruns=False,
  closedPriorActualCalls=246,closedPriorActualHTTP=246,
  cumulativePlannedMaxCalls=410,cumulativePlannedMaxAttempts=434,
  accountRegionPermissions='UNVERIFIED',dollarCap=None,proxy=None,
  components=[dict(name=n,provider=p,sourceVariable=k,maxCalls=cap,maxAttempts=cap)
   for n,p,k,cap in COMPONENTS],
  componentContracts='original fixed support/effect questions, images, outputs, retries and deadlines')


class Corrective(Round):
 # This is a distinct experiment, not a retry branch of the closed legacy
 # round. Reuse its unchanged component argv and selected-key subprocess seam.
 def guard(self):
  if source_files()!=c.doc(self.out/'source-at-start.json')['files']:
   raise RuntimeError('frozen complete source/config changed')
  for name,sha in c.doc(self.out/'binaries.json').items():
   if c.digest(self.out/name)!=sha:
    raise RuntimeError('frozen binary changed')

 def prepare(self):
  self.out.mkdir(parents=True,exist_ok=False,mode=0o700)
  commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=c.ROOT,env=c.ENV,text=True).strip()
  files=source_files()
  import hashlib
  for name,sha in files.items():
   raw=subprocess.check_output(['git','show',commit+':'+name],cwd=c.ROOT,env=c.ENV)
   if hashlib.sha256(raw).hexdigest()!=sha:
    raise RuntimeError('execution source must be committed before freeze')
  c.dump(self.out/'source-at-start.json',dict(files=files,gitCommit=commit))
  c.dump(self.out/'authorization-scope.json',plan())
  catalog=c.doc(c.ROOT/'ai/release/catalog-snapshot.json')
  for route,model,provider,service in (
   ('decisions','gpt-6-luna','openai','https://api.openai.com/v1/decisions'),
   ('jev','jev-1.14.0','typesafe','https://api.typesafe.ai/v1/systemone')):
   proof=dict(route=route,model=model,service=service,catalogHash=catalog['hash'],
    accountAlias=provider+'-local-unverified',regionAlias='unverified',
    reference=plan()['reference'],dedicatedCredential=True,maxCalls=12,maxAttempts=12)
   c.dump(self.out/('visual-'+route+'-authorization.json'),proof)
   if route=='jev':
    proof.pop('catalogHash');proof['authorizationReference']=proof.pop('reference')
    c.dump(self.out/'protocol-authorization.json',proof)
  for name,argv in (
   ('build-live',['go','test','-tags','live','-c','-o',self.out/'live.test','./ai/live']),
   ('build-text',['go','build','-o',self.out/'chineseeval','./ai/examples/chineseeval/cmd/chineseeval']),
   ('build-visual',['go','build','-o',self.out/'visualeval','./ai/examples/visualeval'])):
   if self.execute(name,argv,timeout=240)!=0:
    raise RuntimeError('credential-free build failed')
  c.dump(self.out/'binaries.json',{n:c.digest(self.out/n) for n in ('live.test','chineseeval','visualeval')})
  self.guard();c.dump(self.out/'prepared.json',dict(state='PREPARED',credentialsRead=False))

 def jobs(self):
  for name,argv,options in super().jobs():
   if name in NAMES:
    yield name,argv,options

 def summarize(self):
  rows=[]
  for name,provider,_,cap in COMPONENTS:
   receipt=self.out/(name+'-command.json')
   row=dict(name=name,provider=provider,maxCalls=cap,state='UNSTARTED',calls=0,
    attempts=0,conservativeCalls=0,conservativeAttempts=0)
   if receipt.is_file():
    candidates=([self.out/name/'report.json'] if name.startswith('visual-') else
     list((self.out/name).glob('*/multimodal-report.json')))
    if len(candidates)!=1 or not candidates[0].is_file():
     row.update(state='UNKNOWN',calls=None,attempts=None,conservativeCalls=cap,conservativeAttempts=cap)
    else:
     report=c.doc(candidates[0]);budget=report['budget']
     if name=='protocol114':budget=budget['actual']
     calls,attempts=budget['calls'],budget['attempts']
     if not 0<=calls<=cap or not 0<=attempts<=cap:
      raise RuntimeError('corrective component budget exceeded')
     row.update(state='REPORTED',calls=calls,attempts=attempts,
      conservativeCalls=calls,conservativeAttempts=attempts,status=report.get('status',report.get('outcome')))
   rows.append(row)
  totals={k:sum(r.get(k) or 0 for r in rows) for k in ('calls','attempts','conservativeCalls','conservativeAttempts')}
  for provider,cap in plan()['providerCaps'].items():
   if any(sum(r[k] for r in rows if r['provider']==provider)>cap for k in ('conservativeCalls','conservativeAttempts')):
    raise RuntimeError('corrective provider budget exceeded')
  if max(totals['conservativeCalls'],totals['conservativeAttempts'])>36:
   raise RuntimeError('corrective task budget exceeded')
  totals.update(maxCalls=36,maxAttempts=36)
  c.dump(self.out/'consumption.json',dict(components=rows,total=totals));return totals

 def live(self):
  self.guard()
  if c.doc(self.out/'live-harness-accepted.json')['exitCode']!=0:
   raise RuntimeError('no-key harness must pass')
  review=c.doc(self.out/'prerelease-source-review.json')
  if review.get('decision')!='APPROVED' or review.get('sourceCommit')!=c.doc(self.out/'source-at-start.json')['gitCommit']:
   raise RuntimeError('prerelease source review identity missing')
  with (self.out/'paid-corrective-round-started.json').open('x') as stream:
   json.dump(dict(maxCalls=36,maxAttempts=36,automaticReruns=False),stream)
  codes={}
  for name,argv,options in self.jobs():
   self.guard();codes[name]=self.execute(name,argv,**options);self.summarize()
  self.guard()
  if not all(code==0 for code in codes.values()) or tuple(codes)!=NAMES:
   c.dump(self.out/'live-summary.json',dict(commands=codes,total=self.summarize(),allCommandsSucceeded=False))
   raise SystemExit(1)
  for route in ('luna','jev'):
   codes['verify-text-'+route]=self.execute('verify-text-'+route,[self.out/'chineseeval',
    '-route',route,'-verify',c.bundle(ORIGINAL/('text-'+route))])
  for route in ('decisions','jev'):
   codes['verify-visual-'+route]=self.execute('verify-visual-'+route,[self.out/'visualeval',
    '-verify',self.out/('visual-'+route),'-require-live'])
  codes['support']=self.execute('support',['go','run','./ai/live/cmd/jevformal',
   '-protocol',c.bundle(self.out/'protocol114'),'-effects',self.out/'visual-jev',
   '-out',self.out/'support','-require-live'])
  if codes['support']==0:
   codes['verify-support']=self.execute('verify-support',['go','run','./ai/live/cmd/jevformal',
    '-verify',self.out/'support','-require-live'])
  self.guard();passed=all(code==0 for code in codes.values())
  c.dump(self.out/'live-summary.json',dict(commands=codes,total=self.summarize(),allCommandsSucceeded=passed))
  if not passed:raise SystemExit(1)

 def live_inputs(self):
  views=c.doc(ORIGINAL/'glm-support-inputs.json')
  return [Path(views[r]) if r.startswith('zhipu-') else self.out/'support'
   if r=='typesafe-classifier-jev-1.14.0' else c.bundle(ORIGINAL/'protocol113')
   if r=='typesafe-classifier' else c.bundle(ORIGINAL/'live'/r)
   for r in c.doc(c.ROOT/'ai/release/traceability.json')['liveCombos']]

 def gate(self):
  # Both original text effects retain their source IDs and satisfy the exact
  # unchanged execution-source selector; only the two visual effects are new.
  super().gate(text_root=ORIGINAL)


def main():
 parser=argparse.ArgumentParser(description=__doc__)
 parser.add_argument('mode',choices=('plan','prepare','check','live','merge','gate','summarize'))
 parser.add_argument('out');args=parser.parse_args()
 if Path(args.out).is_absolute() or '..' in Path(args.out).parts:
  parser.error('repository-relative owned evidence root required')
 if args.mode=='plan':print(json.dumps(plan(),indent=2));return
 getattr(Corrective(args.out),args.mode)()
if __name__=='__main__':main()
