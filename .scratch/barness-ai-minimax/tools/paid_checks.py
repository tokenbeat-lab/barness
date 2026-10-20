"""The six actual keyless commands shared by the two fixed renewal authorities."""
import paid_core as c
import routed_evidence as e

def commands(out):
 return (
  ('vet',['go','vet','./...'],600,None),
  ('vet-live',['go','vet','-tags','live','./...'],600,None),
  ('no-key-live',['go','test','-tags','live','-count=1','-timeout=20m','./ai/live'],1260,{'BARNESS_AI_EVIDENCE_DIR':str(out/'no-key-live')}),
  ('no-key-visual-proxy',['go','test','./ai/e2e','-run','^TestVisualEvaluationProxyProcess$','-count=1','-timeout=5m'],360,None),
  ('no-key-release-contracts',['go','test','./ai/e2e','-count=1','-run','^(TestMiniMaxAdditiveReleaseContract|TestGLMReleaseTraceContract|TestReleaseRecordedBundleDiagnosticProcess)$'],360,{'BARNESS_AI_EVIDENCE_DIR':str(out/'no-key-release-contracts')}),
  ('no-key-release-package',['go','test','./ai/internal/testkit/release','-count=1'],360,None))

def run(out,guard,pin_receipts=False):
 codes={}
 for name,argv,timeout,env in commands(out):
  guard(out);codes[name]=c.execute(out,name,argv,env=env,timeout=timeout);guard(out)
 result=dict(passed=all(code==0 for code in codes.values()),commands=codes,credentialsRead=False)
 if pin_receipts:result['commandReceipts']={name:c.digest(out/(name+'-command.json')) for name in codes}
 c.dump(out/'checks.json',result);return result

def verify(out):
 """A true boolean alone is insufficient authority for selected credential access."""
 relative=out.relative_to(c.ROOT).as_posix()
 with e.Root(c.ROOT) as reader:
  checks=e.document(reader.read(relative+'/checks.json'));verify_at(reader,relative,checks)
 return checks

def verify_at(reader,relative,checks):
 out=c.ROOT/relative
 expected={name:0 for name,_,_,_ in commands(out)}
 if checks.get('passed') is not True or checks.get('credentialsRead') is not False or checks.get('commands')!=expected or set(checks.get('commandReceipts',{}))!=set(expected):raise RuntimeError('six actual zero keyless checks and receipt pins required')
 for name,argv,_,_ in commands(out):
  raw=reader.read(relative+'/'+name+'-command.json');v=e.document(raw)
  if c.redact(argv)!=v.get('argv') or v.get('cwd')!='$REPO' or v.get('state')!='COMPLETED' or type(v.get('exitCode')) is not int or v['exitCode']!=0 or v.get('credentialsRead') is not False or v.get('credentialsInherited') is not False or v.get('sourceVariable') is not None or v.get('injectedCredential') is not None or v.get('launcherAudit')!={'passed':True,'findings':[]} or e.u.sha(raw)!=checks['commandReceipts'][name]:raise RuntimeError('original fixed keyless command receipt changed or failed')
 return checks
