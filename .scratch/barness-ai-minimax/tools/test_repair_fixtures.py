"""Synthetic immutable native history in an isolated local Git clone; no credentials."""
import hashlib,json,os,shutil,subprocess,sys,tempfile
from pathlib import Path
TOOLS=Path(__file__).resolve().parent
BASE='8fbd17006677774368d4d77d924a8672fd205c51'
COUNTS=dict(zip(['openai-responses','openai-chat','anthropic-messages','google-gemini','deepseek-responses','deepseek-chat','openai-images','google-interactions-image','openai-decisions','typesafe-classifier','protocol114','text-jev','text-luna','visual-jev','visual-decisions'],[14,14,20,8,10,13,3,2,9,3,12,24,24,12,12]))

def seed():
 import paid_core as c,paid_plan as p,routed_contract as r,routed_evidence as e,repair_inputs as inputs,repair_anchors as anchors
 original=Path(sys.argv[2]);expected=anchors.original_receipts();certified,m3=anchors.inventories()
 def copy(reader,relative,expected_hash=None):
  e.parts(relative);raw=reader.read(relative,16<<20)
  if expected_hash is not None and hashlib.sha256(raw).hexdigest()!=expected_hash:raise ValueError('original historical fixture bytes changed')
  target=c.ROOT/relative;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(raw)
 with e.Root(original) as reader:
  for name,h in expected.items():copy(reader,'.evidence/minimax-launches/'+name+'.json',h)
  for name,j in r.jobs(c.ROOT/'.evidence/unused').items():
   record=json.loads((c.ROOT/'.evidence/minimax-launches'/(name+'.json')).read_text());relative=record['collectionRoot'].removeprefix('$REPO/');base=j['baseName']
   for leaf in ('source-at-start.json','binaries.json','plan.json','prerelease-source-review.json','execution-authorization.json','checks.json',base+'-command.json'):copy(reader,relative+'/'+leaf)
   directory=relative+('/live/' if base in {v[0] for v in p.GENERIC} else '/')+base
   leaf='formal-glm-report.json' if base.startswith('glm-') else 'evaluation-report.json' if base.startswith('text-') else 'multimodal-report.json' if base=='protocol114' else 'report.json' if base.startswith('visual-') else 'live-report.json'
   path=directory if base.startswith('visual-') else inputs.bundle(reader,directory,leaf)
   copy(reader,path+'/'+leaf,record['reportedFileSha256'])
   if base in {v[0] for v in p.GENERIC}|{'text-jev','text-luna'}:
    for inner,pin in certified[base].items():copy(reader,path+'/'+inner,pin['sha256'])
  for row in m3['files'].values():
   source=row.get('original','');relative='.evidence/'+source.partition('/.evidence/')[2]
   if '/.evidence/minimax-global-formal-glm/mm-global-formal-m3-' in source or '/.evidence/minimax-launches/mm-global-formal-m3-' in source:copy(reader,relative,row['sha256'])

class Isolated:
 SOURCE_REF="7931b163f983a8b362374163b7949d31e68bbec4"
 def setUp(self):
  self.tmp=tempfile.TemporaryDirectory(prefix='release-repair-cli-');self.root=Path(self.tmp.name)/'repo';source=TOOLS.parents[2];self.outputs=set();self.journal=[]
  self.env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ}
  subprocess.run(['git','clone','-q','--shared','--no-hardlinks',str(source),str(self.root)],env=self.env,check=True)
  if self.SOURCE_REF:subprocess.run(['git','checkout','-q',self.SOURCE_REF],cwd=self.root,env=self.env,check=True)
  self.tools=self.root/'.scratch/barness-ai-minimax/tools';shutil.copytree(TOOLS,self.tools,dirs_exist_ok=True,ignore=shutil.ignore_patterns('__pycache__'))
  subprocess.run(['git','add','.scratch/barness-ai-minimax/tools'],cwd=self.root,env=self.env,check=True)
  subprocess.run(['git','-c','user.name=Controlled','-c','user.email=controlled@example.invalid','commit','--allow-empty','-qm','synthetic public CLI source'],cwd=self.root,env=self.env,check=True)
  self.bin=self.root/'.fake-bin';self.bin.mkdir();(self.bin/'go').write_text('#!'+sys.executable+'\nimport pathlib,sys,os,json,shutil\na=sys.argv[1:];r=pathlib.Path('+repr(str(self.root))+')\nwith (r/"go-probe.jsonl").open("a") as f:f.write(json.dumps(dict(argv=a,environmentKeys=sorted(os.environ)))+"\\n")\nif "-o" in a:\n p=pathlib.Path(a[a.index("-o")+1]);shutil.copyfile(r/".fake-bin/producer",p);p.chmod(0o755)\nname="vet-live" if a[:2]==["vet","-tags"] else "vet" if a[:1]==["vet"] else "no-key-live" if "./ai/live" in a else "no-key-release-package" if "./ai/internal/testkit/release" in a else "no-key-visual-proxy" if "^TestVisualEvaluationProxyProcess$" in a else "no-key-release-contracts"\nfail=r/"fail-check-name"\nsys.exit(7 if "-o" not in a and ((r/"fail-next-check").exists() or (fail.exists() and fail.read_text()==name)) else 0)\n');(self.bin/'go').chmod(0o755)
  (self.bin/'producer').write_text('#!'+sys.executable+'\nprint("synthetic-no-selected-job")\n')
  self.env.update(PATH=str(self.bin)+os.pathsep+self.env['PATH'],OPENAI_API_KEY='AMBIENT-SYNTHETIC',HTTP_PROXY='http://forbidden.invalid')
  common=subprocess.check_output(['git','rev-parse','--git-common-dir'],cwd=source,env=self.env,text=True).strip();common=(source/common).resolve().parent
  subprocess.run([sys.executable,'-B',str(self.tools/'test_repair_fixtures.py'),'--seed',str(common)],cwd=self.root,env=self.env,check=True)
  self.cli=[sys.executable,'-B',str(self.tools/'release_repair_validation.py')]
 def selected_producer(self):
  (self.bin/'producer').write_text('#!'+sys.executable+'\n' + "import json,os,pathlib,subprocess,sys,time\nroot=pathlib.Path.cwd()\nwhile not (root/'go.mod').is_file():root=root.parent\nargv=sys.argv[1:];route=os.environ.get('BARNESS_GLM_ROUTE');failed=False\nif route:\n base='glm-'+route;bundle=pathlib.Path(os.environ['BARNESS_AI_EVIDENCE_DIR'])/'session';leaf='formal-glm-report.json';failed=route=='responses' and (root/'.fake-bin/fail-responses').exists();n=21 if failed else 22\n value=dict(schema=1,kind='glm-formal',mode='live',route=route,outcome='FAIL' if failed else 'PASS',complete=not failed,auditPassed=True,gitCommit=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),budget=dict(callsUsed=n,httpAttempts=n,attemptsReserved=n))\nelif 'BARNESS_MM_ROUTE' in os.environ:\n bundle=pathlib.Path(os.environ['BARNESS_AI_EVIDENCE_DIR'])/'session';leaf='multimodal-report.json';value=dict(status='PASS',budget=dict(actual=dict(calls=12,attempts=12)))\nelse:\n bundle=pathlib.Path(argv[argv.index('-out')+1]);leaf='report.json';value=dict(status='COMPLETE',budget=dict(calls=12,attempts=12))\nbundle.mkdir(parents=True);(bundle/leaf).write_text(json.dumps(value))\nif route:\n for name in ('glm-integrity.json','live-report.json','glm-execution-source.json'):(bundle/name).write_text('{}')\nprint(json.dumps(dict(os.environ)));time.sleep(0.05);sys.exit(1 if failed else 0)\n")
 def approve(self,out):
  import hashlib
  base=self.root/out;plan=json.loads((base/'plan.json').read_text());review=json.loads((base/'prerelease-source-review.template.json').read_text())
  review.update(decision='APPROVED',reviewedBy='synthetic independent current-source reviewer',authorizedRoutes=list(plan['routes']),credentialScopes={j['sourceVariable']:'release-repair-validation-1/'+j['networkRoute'] for j in plan['routes'].values()})
  (base/'prerelease-source-review.json').write_text(json.dumps(review));auth=json.loads((base/'execution-authorization.template.json').read_text())
  auth.update(decision='APPROVED',authority='explicit-additional-budget-reply',requestReference='synthetic NEW108/132 question and reply',requestText='Approve additional108/132 for this fixed6 only',checksSHA256=hashlib.sha256((base/'checks.json').read_bytes()).hexdigest(),sourceReviewSHA256=hashlib.sha256((base/'prerelease-source-review.json').read_bytes()).hexdigest())
  (base/'execution-authorization.json').write_text(json.dumps(auth))
 def tearDown(self):
  try:
   target=os.environ.get('BARNESS_REPAIR_TEST_ARTIFACT_DIR')
   if target:self.retain(Path(target)/self.id())
  finally:self.tmp.cleanup()
 def retain(self,target):
  # The test controller opts into this finite artifact. Historical native inputs,
  # original keys and intentional negative links are never followed or copied.
  import stat
  target.mkdir(parents=True,exist_ok=False);inventory={};total=0
  def copy(path):
   nonlocal total
   if not stat.S_ISREG(path.lstat().st_mode):raise ValueError('regular controlled test artifact required')
   if path.suffix not in ('.json','.jsonl','.log'):return
   raw=path.read_bytes();total+=len(raw)
   if len(raw)>16<<20 or total>96<<20 or len(inventory)>=10000:raise ValueError('fixed test artifact bound')
   name=path.relative_to(self.root).as_posix();dest=target/name;dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes(raw);inventory[name]=dict(bytes=len(raw),sha256=hashlib.sha256(raw).hexdigest())
  for name in sorted(self.outputs):
   base=self.root/name
   if not base.is_dir():continue
   for directory,dirs,files in os.walk(base,followlinks=False):
    if any((Path(directory)/n).is_symlink() for n in dirs):raise ValueError('test output directory link forbidden')
    for leaf in files:copy(Path(directory)/leaf)
  for path in (self.root/'.evidence/minimax-launches').glob(getattr(self,'INTENT_GLOB','release-repair-validation-1-*.json')):copy(path)
  for relative in getattr(self,'EXTRA_ARTIFACTS',()):
   for path in (self.root/relative).glob('*.json'):copy(path)
  for name in ('go-probe.jsonl',):
   if (self.root/name).exists():copy(self.root/name)
  (target/'public-cli-journal.json').write_text(json.dumps(self.journal,indent=2)+'\n')
  (target/'inventory.json').write_text(json.dumps(dict(files=inventory,bytes=total,credentials='synthetic test values only; no .env copied',paidVendorCalls=0),indent=2)+'\n')
 def invoke(self,*args,ok=True):
  self.outputs.add(args[1]);p=subprocess.run([*self.cli,*args],cwd=self.root,env=self.env,text=True,capture_output=True,timeout=30)
  self.journal.append(dict(args=args,exitCode=p.returncode,stdout=p.stdout.replace(str(self.root),'$ISOLATED_REPO'),stderr=p.stderr.replace(str(self.root),'$ISOLATED_REPO')))
  self.assertEqual(p.returncode==0,ok,p.stderr[-1000:]+p.stdout[:200]);return p
if __name__=='__main__' and sys.argv[1]=='--seed':seed()
