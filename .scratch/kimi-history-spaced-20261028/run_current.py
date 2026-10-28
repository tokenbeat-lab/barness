"""One human-approved Kimi route, once; predecessor candidate pass is mandatory."""
from common import *
from datetime import timedelta
from decimal import Decimal
import re, shlex, time, uuid, traceback

route=sys.argv[1] if len(sys.argv)==2 else ''
key=''
try:
 assert route in REMAINING,'exact remaining Kimi route required'
 assert not (B/'STOP.json').exists(),'current sequence is stopped'
 pre=load(B/'preflight.json');assert pre['pass_']
 f=frozen();assert f['source']==pre['source']
 done=[]
 for r in ORDER:
  p=B/'results'/f'{r}.json'
  if p.exists():
   result=load(p);assert result['pass'] and result['route']==r
   check_ref(result['candidateReceipt']);check_ref(result['candidateResult'])
   done.append(r)
 assert done==ORDER[:len(done)] and len(done)<6 and route==ORDER[len(done)],'serial approved order required'
 grantPath=B/'grants'/f'{route}.json'
 assert not grantPath.exists(),'this route approval cannot be retried or reissued'
 sr=next(v for v in pre['machineScopes'] if Path(v['path']).stem==route)
 scopePath=check_ref(sr);scope=load(scopePath);g=scope['machineTemplate'].copy()
 assert scope['route']==route and g['schema']==2 and g['combo']==route
 assert g['mode']=='live' and g['stage']=='candidate' and g['purpose']=='xai-history-supplement'
 assert g['contract']=='kimi/'+route+'/xai-origin-history/v1'
 assert g['sourceSha256']==NATIVE and g['binarySha256']==f['kimi']['sha256']
 assert g['baselineManifestSha256']==PINS[route] and g['models']==['kimi-k3']
 assert g['scenarioIds']==['kimi-k3/xai-origin-history']
 assert g['maxCalls']==g['maxAttempts']==1 and g['maxRequestedOutputTokens']==g['maxReservedInputTokens']==4096
 assert g['maxUSD']==0.10 and g['callIntervalMs']==60000 and scope['retries']==0
 assert Decimal(scope['reservationAt4096InputOutputUSD'])<=Decimal('0.10')
 for n in ['manifest','report','seal']:check_ref(scope['baseline'][n])
 check_ref(scope['producer']);check_ref(scope['confirmedFacts'])
 for name,h in f['regionalSource']['files'].items():assert sha(R/name)==h,name
 gatePath=B/'cooldowns'/f'{route}.json'
 gate=load(gatePath)
 assert gate['route']==route and gate['status']=='ELAPSED' and gate['requiredSeconds']==COOLDOWN_SECONDS
 check_ref(gate['anchor'])
 assert datetime.now(timezone.utc)>=datetime.fromisoformat(gate['earliestGrantAt'])
 # Retrieve only the selected site's known assignment; never source .env,
 # place credentials in argv/inputs, or forward another site's credential.
 keyname='KIMI_CN_KEY' if route.startswith('kimi-cn-') else 'KIMI_KEY'
 values=[]
 with (R/'.env').open() as stream:
  for line in stream:
   m=re.match(r'^\s*(?:export\s+)?'+re.escape(keyname)+r'\s*=\s*(.*)$',line)
   if not m:continue
   raw=m[1].strip()
   if raw.startswith(('"',"'")):
    parts=shlex.split(raw,comments=True,posix=True)
    assert len(parts)==1,'selected credential assignment ambiguous'
    values.append(parts[0])
   else:values.append(raw.split(' #',1)[0].strip())
 assert len(values)==1 and values[0] and not values[0].startswith('$'),'selected credential missing or ambiguous'
 key=values[0];del values
 (B/'grants').mkdir(exist_ok=True)
 now=datetime.now(timezone.utc)
 g.update(id='kh-'+route+'-'+now.strftime('%Y%m%dT%H%M%SZ')+'-'+uuid.uuid4().hex[:8],accountAlias=scope['identity']['accountAlias'],regionAlias=scope['identity']['regionAlias'],expiresAt=(now+timedelta(minutes=30)).isoformat().replace('+00:00','Z'))
 assert len(g['id'])<=64 and g['accountAlias']=='KIMI' and g['regionAlias']==('CN' if route.startswith('kimi-cn-') else 'GLOBAL')
 save(grantPath,g);os.chmod(grantPath,0o600)
 issuance=B/'issuance';issuance.mkdir(exist_ok=True)
 save(issuance/f'{route}.json',dict(route=route,grant=ref(grantPath),approval=ref(B/'approval.json'),machineScope=sr,issuedAt=now.isoformat(),expiresAt=g['expiresAt'],maxCalls=1,maxHTTPAttempts=1,maxUSD='0.10',completedPredecessors=done,remainingRouteGrantsIssued=False,cooldown=ref(gatePath)))
 state('正在执行'+route+'唯一一次已批准单项。新grant为30分钟有效；1调用/1HTTP/USD0.10、输入/输出各4096、零重试、本路外部冷却300秒已结束，进程内首次发送仍等待60秒。已通过路线：'+('、'.join(done) or '无')+'。其他未开始路线尚未签发grant；任何失败立即停止。')
 env,_=environment('e2e')
 env.update(BARNESS_AI_LIVE='1',BARNESS_REGIONAL_ROUTE=route,BARNESS_REGIONAL_STAGE='candidate',BARNESS_REGIONAL_SERVICE=g['provider'],BARNESS_REGIONAL_GRANT=str(grantPath),BARNESS_REGIONAL_KEY=key)
 inputs=[name for name in f['regionalSource']['files'] if name not in pre['source']['files']]
 inputs += [str(p.relative_to(R)) for p in [Path(__file__).resolve(),S/'common.py',S/'spec.md',gatePath,R/gate['anchor']['path'],B/'approval.json',B/'preflight.json',scopePath,grantPath,FREEZE,R/scope['baseline']['manifest']['path'],R/scope['producer']['path'],R/scope['confirmedFacts']['path']]]
 inputs=sorted(set(inputs))
 binary=R/f['kimi']['path']
 argv=[sys.executable,'tools/verification/run.py','--feature','kimi-history-spaced-20261028','--stage','live','--route',route,'--timeout','240']
 for name in inputs:argv+=['--input',name]
 argv+=['--',str(binary),'-test.run=^TestKimiHistorySupplement$','-test.count=1','-test.timeout=240s']
 commands=B/'commands';commands.mkdir(exist_ok=True)
 save(commands/f'{route}.json',dict(argv=argv,selectedInputs={n:sha(R/n) for n in inputs},credentialVariable=keyname,credentialInjectedOnlyAs='BARNESS_REGIONAL_KEY',credentialStored=False,paidProcessLaunchLimit=1,reservationAt4096InputOutputUSD=scope['reservationAt4096InputOutputUSD']))
 wrappers=B/'wrappers';wrappers.mkdir(exist_ok=True)
 startedAt=datetime.now(timezone.utc).isoformat();started=time.monotonic()
 # Exactly one native producer launch through the frozen collector. No retry.
 proc=subprocess.run(argv,cwd=R,env=env,capture_output=True)
 stdout=proc.stdout.decode(errors='replace').replace(key,'[REDACTED]');stderr=proc.stderr.decode(errors='replace').replace(key,'[REDACTED]')
 (wrappers/f'{route}.stdout').write_text(stdout);(wrappers/f'{route}.stderr').write_text(stderr)
 env.pop('BARNESS_REGIONAL_KEY',None);key=''
 rows=[json.loads(line) for line in stdout.splitlines() if line.startswith('{')]
 assert len(rows)==1 and 'attempt' in rows[0],'capture result unknown; never retry'
 attempt=R/rows[0]['attempt'];receiptPath=attempt/'receipt.json';producer=load(receiptPath)
 indices=B/'attempt-indices';indices.mkdir(exist_ok=True)
 save(indices/f'{route}.json',dict(route=route,wrapperExitCode=proc.returncode,nativeReceipt=ref(receiptPath),nativeExitCode=producer['exitCode'],captureStatus=producer['status'],startedAt=startedAt,finishedAt=datetime.now(timezone.utc).isoformat(),elapsedSeconds=time.monotonic()-started,grant=ref(grantPath)))
 assert proc.returncode==0 and producer['status']=='PASS' and producer['exitCode']==0,'native producer/capture did not pass'
 assert len(producer['bundles'])==1,'exactly one native measurement required'
 bundle=attempt/producer['bundles'][0]['path'];report=load(bundle/'regional-report.json')
 assert report['regional']['complete'] is False and report['regional']['mode']=='live' and report['regional']['stage']=='candidate'
 assert report['expected']==['kimi-k3/xai-origin-history'] and len(report['scenarios'])==1
 assert report['scenarios'][0]['outcome']=='PASS'
 assert report['budget']['callsUsed']==report['budget']['httpAttempts']==1
 assert report['regional']['reservedUSD']<=0.10
 def evidence_ref(bundle,producerRef):
  return dict(bundle=str(bundle.relative_to(R)),manifestSHA256=sha(bundle/'manifest.json'),reportSHA256=sha(bundle/'regional-report.json'),sealSHA256=sha(bundle/'regional-integrity.json'),producer=producerRef)
 baseline=R/scope['baseline']['bundle'];baselineRef=evidence_ref(baseline,scope['producer']);supplementRef=evidence_ref(bundle,ref(receiptPath))
 compositions=B/'compositions';compositions.mkdir(exist_ok=True)
 composition=compositions/route;composition.mkdir()
 save(composition/'kimi-composition.json',dict(schema=1,kind='kimi-history-composition',combo=route,model='kimi-k3',baseline=baselineRef,supplement=supplementRef))
 q=readonly('candidate-inspections/'+route,[str(R/f['releasegate']['path']),'-inspect-bundle',str(composition),'-mode','live','-stage','candidate'])
 assert q['mode']=='live' and q['combo']==route and q['qualification']=='CANDIDATE_COMPLETE'
 assert len(q['measurements'])==2 and len(q['coverage'])==len(q['expected'])==10
 # Preserve every transitive original without copying or rewriting capsules.
 inventory={}
 def pin(path):inventory[str(path.relative_to(R))]=sha(path)
 pin(composition/'kimi-composition.json')
 for er in [baselineRef,supplementRef]:
  native=R/er['bundle']
  for path in native.rglob('*'):
   if path.is_file():pin(path)
  pp=R/er['producer']['path'];pin(pp);pr=load(pp)
  for name in pr['files']:pin(pp.parent/name)
  pin(Path(pr['argv'][0]))
  ng=load(native/'regional-grant.json');ar=load(native/'regional-authorization-receipt.json')
  for name in ['id-'+hashlib.sha256(ng['id'].encode()).hexdigest()+'.json','scope-'+ar['scopeSha256']+'.json']:
   pin(R/'.git/barness-kimi-grants'/name)
 inspection=B/'candidate-inspections'/route
 save(inspection/'transitive-inputs.json',inventory)
 assert all(sha(R/n)==h for n,h in inventory.items())
 frozen()
 results=B/'results';results.mkdir(exist_ok=True)
 save(results/f'{route}.json',{'pass':True,'route':route,'mode':'live','stage':'candidate','grant':ref(grantPath),'producer':ref(receiptPath),'supplement':supplementRef,'baseline':baselineRef,'composition':ref(composition/'kimi-composition.json'),'candidateReceipt':ref(inspection/'receipt.json'),'candidateResult':ref(inspection/'stdout.jsonl'),'transitiveInputs':ref(inspection/'transitive-inputs.json'),'qualification':q['qualification'],'coverage':q['coverage'],'measurements':q['measurements'],'singleComplete':False,'catalogActive':False})
 done.append(route)
 state('已通过'+str(len(done))+'/6路线：'+'、'.join(done)+'。各路线唯一付费单项PASS且complete=false，固定9＋1候选组合CANDIDATE_COMPLETE，分别保留两份测量与十项覆盖。未开始路线未签发grant；下一步'+('停止并收尾' if len(done)==6 else '仅'+ORDER[len(done)])+'。目录active=false；不运行正式消费/晋升/race/压力/暂停恢复。')
 print(json.dumps(dict(route=route,producer='PASS',candidate=q['qualification'],coverage=len(q['coverage']),measurements=2,completed=len(done),callsThisRoute=1,httpAttemptsThisRoute=1)))
except BaseException as error:
 if isinstance(error,SystemExit) and error.code==0:raise
 message=repr(error)
 if key:message=message.replace(key,'[REDACTED]')
 diagnostics=B/'failures';diagnostics.mkdir(exist_ok=True)
 if not (diagnostics/(route+'.json')).exists():save(diagnostics/(route+'.json'),dict(route=route,exception=message,grantIssued=(B/'grants'/f'{route}.json').exists(),retryForbidden=True))
 stop('route execution/acceptance stopped: '+message,route)
 print(json.dumps(dict(stopped=True,route=route,reason=message,remainingGrantsForbidden=True)))
 raise SystemExit(1)
