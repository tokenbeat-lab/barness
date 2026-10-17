#!/usr/bin/env python3
"""One preregistered current-catalog legacy refresh; no previous-task balance."""
import argparse
import datetime as dt
import json
from pathlib import Path
import subprocess
import sys
import zipfile
sys.dont_write_bytecode=True
import commands as c
ROOT=c.ROOT
ROUTES=(
 ('openai-responses','openai','OPENAI_KEY',32),('openai-chat','openai','OPENAI_KEY',32),
 ('openai-images','openai','OPENAI_KEY',4),('openai-decisions','openai','OPENAI_KEY',12),
 ('anthropic-messages','anthropic','ANTHROPIC_KEY',32),('google-gemini','google','GEMINI_KEY',32),
 ('google-interactions-image','google','GEMINI_KEY',4),('deepseek-responses','deepseek','DEEPSEEK_KEY',32),
 ('deepseek-chat','deepseek','DEEPSEEK_KEY',32),('protocol113','typesafe','TYPESAFE_KEY',6),
 ('protocol114','typesafe','TYPESAFE_KEY',12),('text-luna','openai','OPENAI_KEY',24),
 ('text-jev','typesafe','TYPESAFE_KEY',24),('visual-decisions','openai','OPENAI_KEY',12),
 ('visual-jev','typesafe','TYPESAFE_KEY',12))
CAPS={r[0]:r[3] for r in ROUTES}


class Round:
 def __init__(self,out):
  self.commands=c.Commands(out);self.out=self.commands.out;self.execute=self.commands.execute

 def guard(self):
  if c.sources()!=c.doc(self.out/'source-at-start.json')['files']:
   raise RuntimeError('frozen complete source/config changed; collection stopped')
  for name,sha in c.doc(self.out/'binaries.json').items():
   if c.digest(self.out/name)!=sha:
    raise RuntimeError('frozen binary changed')

 def prepare(self):
  self.out.mkdir(parents=True,exist_ok=False,mode=0o700)
  commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,env=c.ENV,text=True).strip()
  files=c.sources()
  for name,sha in files.items():
   raw=subprocess.check_output(['git','show',commit+':'+name],cwd=ROOT,env=c.ENV)
   import hashlib
   if hashlib.sha256(raw).hexdigest()!=sha:
    raise RuntimeError('execution source must be committed before freeze')
  c.dump(self.out/'source-at-start.json',dict(files=files,gitCommit=commit))
  c.dump(self.out/'authorization-scope.json',dict(reference='barness-ai-glm-ticket13-r1',
   source='current implement-spec R21 and issue13; required current-hash release renewal',
   maxCalls=302,maxAttempts=302,previousTaskBalanceCarried=False,automaticReruns=False,
   providerCaps={'openai':116,'anthropic':32,'google':36,'deepseek':64,'typesafe':54},
   accountRegionPermissions='UNVERIFIED',dollarCap=None,proxy=None,
   routes=[dict(route=r,provider=p,sourceVariable=k,maxCalls=n,maxAttempts=n) for r,p,k,n in ROUTES],
   componentContracts='unchanged checked-in live/evaluation budgets, tokens, images, questions, retries and timeouts'))
  cat=c.doc(ROOT/'ai/release/catalog-snapshot.json')
  for route,model,provider,service in (
   ('decisions','gpt-6-luna','openai','https://api.openai.com/v1/decisions'),
   ('jev','jev-1.14.0','typesafe','https://api.typesafe.ai/v1/systemone')):
   proof=dict(route=route,model=model,service=service,catalogHash=cat['hash'],
    accountAlias=provider+'-local-unverified',regionAlias='unverified',reference='barness-ai-glm-ticket13-r1',
    dedicatedCredential=True,maxCalls=12,maxAttempts=12)
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

 def check(self):
  self.guard()
  code=self.execute('live-harness',['go','test','-tags','live','-count=1','-timeout=20m','./ai/live'],
   extra={'BARNESS_AI_EVIDENCE_DIR':str(self.out/'live-harness')},timeout=1260)
  self.guard();c.dump(self.out/'live-harness-accepted.json',dict(exitCode=code))
  if code!=0:raise SystemExit(code)

 def jobs(self):
  for route,provider,key,cap in ROUTES[:9]:
   extra={'BARNESS_AI_LIVE':'1','BARNESS_AI_LIVE_COMBO':route,
    'BARNESS_AI_LIVE_ACCOUNT_ALIAS':provider+'-local-unverified@region-unverified',
    'BARNESS_AI_EVIDENCE_DIR':str(self.out/'live'/route)}
   formal=route=='openai-decisions'
   if formal:extra['BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT']='1'
   selector='^TestFormalDecisions$' if formal else '^TestLive$/^'+route+'$'
   limit,timeout=('180s',210) if formal else (('10m',660) if cap==4 else ('30m',2100))
   yield route,[self.out/'live.test','-test.run='+selector,'-test.timeout='+limit],dict(extra=extra,
    timeout=timeout,cwd=ROOT/'ai/live',source=key,injected='BARNESS_AI_LIVE_KEY_'+route.upper().replace('-','_'))
  yield 'protocol113',[self.out/'live.test','-test.run=^TestLive$/^typesafe-classifier$','-test.timeout=10m'],dict(
   extra={'BARNESS_AI_LIVE':'1','BARNESS_AI_LIVE_COMBO':'typesafe-classifier',
    'BARNESS_AI_LIVE_ACCOUNT_ALIAS':'typesafe-local-unverified@region-unverified',
    'BARNESS_AI_EVIDENCE_DIR':str(self.out/'protocol113')},cwd=ROOT/'ai/live',timeout=660,
   source='TYPESAFE_KEY',injected='BARNESS_AI_LIVE_KEY_TYPESAFE_CLASSIFIER')
  yield 'protocol114',[self.out/'live.test','-test.run=^TestMultimodalFormal$','-test.timeout=180s'],dict(
   extra={'BARNESS_MM_ROUTE':'jev','BARNESS_MM_FORMAL':'1','BARNESS_MM_LIVE':'1',
    'BARNESS_MM_AUTHORIZATION':str(self.out/'protocol-authorization.json'),
    'BARNESS_AI_EVIDENCE_DIR':str(self.out/'protocol114')},cwd=ROOT/'ai/live',timeout=210,
   source='TYPESAFE_KEY',injected='BARNESS_MM_KEY')
  for route,key,injected in (('luna','OPENAI_KEY','BARNESS_AI_CHINESE_EVAL_OPENAI_KEY'),
                             ('jev','TYPESAFE_KEY','BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY')):
   yield 'text-'+route,[self.out/'chineseeval','-route',route,'-live','-out',self.out/('text-'+route)],dict(
    extra={'BARNESS_AI_CHINESE_EVAL':'1','BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS':
     ('openai' if route=='luna' else 'typesafe')+'-local-unverified@region-unverified'},timeout=630,
    source=key,injected=injected)
  for route,key in (('decisions','OPENAI_KEY'),('jev','TYPESAFE_KEY')):
   yield 'visual-'+route,[self.out/'visualeval','-formal','-mode','live','-route',route,'-entry','plain',
    '-authorization',self.out/('visual-'+route+'-authorization.json'),'-out',self.out/('visual-'+route)],dict(
    timeout=270,source=key,injected='BARNESS_VISUAL_KEY')

 def live(self):
  self.guard()
  if c.doc(self.out/'live-harness-accepted.json')['exitCode']!=0:
   raise RuntimeError('no-key harness must pass')
  # Parent supplies this exact-source internal review receipt before collection.
  review=c.doc(self.out/'prerelease-source-review.json')
  if review.get('decision')!='APPROVED' or review.get('sourceCommit')!=c.doc(self.out/'source-at-start.json')['gitCommit']:
   raise RuntimeError('prerelease source review identity missing')
  with (self.out/'paid-round-started.json').open('x') as stream:
   json.dump(dict(maxCalls=302,maxAttempts=302,automaticReruns=False),stream)
  codes={}
  for name,argv,options in self.jobs():
   self.guard();codes[name]=self.execute(name,argv,**options);self.summarize()
  self.guard()
  for route in ('luna','jev'):
   found=list((self.out/('text-'+route)).glob('*/manifest.json'))
   codes['verify-text-'+route]=self.execute('verify-text-'+route,[self.out/'chineseeval','-route',route,
    '-verify',found[0].parent]) if len(found)==1 else None
  for route in ('decisions','jev'):
   codes['verify-visual-'+route]=self.execute('verify-visual-'+route,[self.out/'visualeval','-verify',
    self.out/('visual-'+route),'-require-live'])
  found=list((self.out/'protocol114').glob('*/manifest.json'))
  codes['support']=self.execute('support',['go','run','./ai/live/cmd/jevformal','-protocol',found[0].parent,
   '-effects',self.out/'visual-jev','-out',self.out/'support','-require-live']) if len(found)==1 else None
  if codes['support']==0:
   codes['verify-support']=self.execute('verify-support',['go','run','./ai/live/cmd/jevformal',
    '-verify',self.out/'support','-require-live'])
  self.guard();c.dump(self.out/'live-summary.json',dict(commands=codes,total=self.summarize(),
   allCommandsSucceeded=all(v==0 for v in codes.values())))
  if not all(v==0 for v in codes.values()):raise SystemExit(1)

 def summarize(self):
  rows=[]
  for name,provider,_,cap in ROUTES:
   path=self.out/name/'report.json' if name.startswith('visual-') else None
   if path is None:
    pattern=(name+'/*/evaluation-report.json' if name.startswith('text-') else
     name+'/*/'+('multimodal-report.json' if name=='protocol114' else 'live-report.json')
     if name.startswith('protocol') else 'live/'+name+'/*/live-report.json')
    found=list(self.out.glob(pattern));path=found[0] if len(found)==1 else None
   receipt=self.out/(name+'-command.json')
   if not receipt.is_file():
    rows.append(dict(name=name,provider=provider,maxCalls=cap,state='UNSTARTED',calls=0,attempts=0,
     conservativeCalls=0,conservativeAttempts=0));continue
   if path is None or not path.is_file():
    rows.append(dict(name=name,provider=provider,maxCalls=cap,state='UNKNOWN',
     conservativeCalls=cap,conservativeAttempts=cap));continue
   report=c.doc(path);b=report['budget'];b=b['actual'] if name=='protocol114' else b
   calls,attempts=(b['calls'],b['attempts']) if name.startswith(('text-','visual-')) or name=='protocol114' else (b['callsUsed'],b['httpAttempts'])
   if not 0<=calls<=cap or not 0<=attempts<=cap:raise RuntimeError('component budget exceeded')
   rows.append(dict(name=name,provider=provider,maxCalls=cap,state='REPORTED',calls=calls,attempts=attempts,
    conservativeCalls=calls,conservativeAttempts=attempts,status=report.get('status',report.get('outcome'))))
  total={k:sum(row.get(k,0) for row in rows) for k in ('calls','attempts','conservativeCalls','conservativeAttempts')}
  total.update(maxCalls=302,maxAttempts=302)
  limits={'openai':116,'anthropic':32,'google':36,'deepseek':64,'typesafe':54}
  for p,cap in limits.items():
   if max(sum(r[k] for r in rows if r['provider']==p) for k in ('conservativeCalls','conservativeAttempts'))>cap:
    raise RuntimeError('provider budget exceeded')
  if max(total['conservativeCalls'],total['conservativeAttempts'])>302:raise RuntimeError('task budget exceeded')
  c.dump(self.out/'consumption.json',dict(routes=rows,total=total));return total

 def live_inputs(self):
  views=c.doc(self.out/'glm-support-inputs.json')
  return [Path(views[r]) if r.startswith('zhipu-') else self.out/'support' if r=='typesafe-classifier-jev-1.14.0'
   else c.bundle(self.out/'protocol113') if r=='typesafe-classifier' else c.bundle(self.out/'live'/r)
   for r in c.doc(ROOT/'ai/release/traceability.json')['liveCombos']]

 def merge(self):
  self.guard()
  if not c.doc(self.out/'live-summary.json')['allCommandsSucceeded']:
   raise RuntimeError('strict producer verdicts must pass before atomic matrix merge')
  matrix=ROOT/'ai/live/support-matrix.json';(self.out/'matrix-before.json').write_bytes(matrix.read_bytes())
  code=self.execute('matrix-merge',['go','run','./ai/live/cmd/supportmatrix',*self.live_inputs()],timeout=300)
  (self.out/'matrix-after.json').write_bytes(matrix.read_bytes())
  if code!=0:raise SystemExit(code)

 def gate(self):
  self.guard()
  args=['go','run','./ai/release/cmd/releasegate','-out',self.out/'gate',
   '-evaluation-jev',c.bundle(self.out/'text-jev'),'-evaluation-luna',c.bundle(self.out/'text-luna'),
   '-visual-evaluation-jev',self.out/'visual-jev','-visual-evaluation-decisions',self.out/'visual-decisions']
  for path in self.live_inputs():args+=['-live',path]
  code=self.execute('releasegate',args,timeout=6600)
  report=self.out/'gate/release-report.json'
  available=report.is_file()
  passed=code==0 and available and c.doc(report).get('pass') is True
  self.guard();c.dump(self.out/'gate-summary.json',dict(exitCode=code,
   passVerdict=passed,reportAvailable=available))
  if code!=0:raise SystemExit(code)
  if not passed:raise SystemExit(1)


def main():
 parser=argparse.ArgumentParser(description=__doc__)
 parser.add_argument('mode',choices=('prepare','check','live','merge','gate','summarize'))
 parser.add_argument('out');args=parser.parse_args()
 if Path(args.out).is_absolute() or '..' in Path(args.out).parts:parser.error('new repository-relative evidence root required')
 getattr(Round(args.out),args.mode)()
if __name__=='__main__':main()
