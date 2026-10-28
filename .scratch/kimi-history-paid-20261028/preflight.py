"""Re-read all original bytes/proofs and regenerate only non-executable scopes."""
from common import *
from decimal import Decimal
try:
 assert not (B/'preflight.json').exists() and not (B/'STOP.json').exists()
 f=frozen();v=load(R/'.evidence/minimax-cache-billing-finish/verification-v1.json');op=check_ref(v['ordinary']);o=load(op)
 assert o['status']=='PASS' and o['exitCode']==0 and o['source']==f['source']
 for n,h in o['selectedInputs'].items():assert sha(R/n)==h,n
 for n,h in o['files'].items():assert sha(op.parent/n)==h,n
 for row in v['nativeInventories']:
  check_ref(row['manifest']);assert load(check_ref(row['audit']))['findings']==[]
 index=load(R/'.evidence/kimi-composition-promotion/pending-scopes-v5-confirmed-v1.json')
 facts=load(check_ref(index['confirmationEvidence']))
 docs={}
 for vref in index['scopeDocuments']:
  if 'kimi-' not in vref['path']:continue
  d=load(check_ref(vref));assert d['route'] in ORDER
  for key in ['accountAndRegion','ratesAndServiceTier','modelPermissionsAndQuota']:assert d['confirmations'][key]=='USER_CONFIRMED'
  docs[d['route']]=d
 assert set(docs)==set(ORDER)
 groups=[x for x in facts['accountGroups'] if x['group'].startswith('kimi-')]
 assert len(groups)==2
 for g in groups:
  for k in ['ownStandardMeteredAPI','frozenRatesStillApplicable','noAdditionalFees','sameActualAccountAcrossThreeProtocolsAtThisSite','sameActualAccountAsOriginalEvidence','modelPermissionNormal','availableQuotaNormal','regionUnchanged']:assert g[k]
 native=readonly('preflight/regional-source',[str(R/f['processgrant']['path']),'-source','regional'])
 assert native==f['regionalSource']
 # Existing native semantic/producer proofs are retained measurements. Re-read
 # their full bound inputs; current scope below uses only the new frozen build.
 proofs=[];protected={};proof_diffs={}
 for n in ['.evidence/kimi-history-supplement/native-baselines-v2/receipt.json','.evidence/kimi-history-supplement/public-replay-v2/receipt.json','.evidence/kimi-composition-promotion/kimi-original-producers-v2/receipt.json']:
  p=R/n;d=load(p);assert d['status']=='PASS' and d['exitCode']==0
  for name,h in d['inputs'].items():assert sha(R/name)==h,name;protected[name]=h
  for name,h in d.get('files',{}).items():assert sha(p.parent/name)==h,name
  rows=load(p.parent/'stdout.jsonl');assert all(x.get('pass',x.get('producerProofPass',False)) for x in rows)
  assert len(rows)==(54 if 'public-replay' in n else 6)
  diffs=[name for name in f['source']['files'].keys()|d['source']['files'].keys() if f['source']['files'].get(name)!=d['source']['files'].get(name)]
  proof_diffs[n]=diffs;proofs.append(ref(p));protected[n]=sha(p)
 # Full original capsules, original sources, actual process receipts/logs and
 # authoritative reservations are re-read without changing old seals or time.
 selected=load(R/'.evidence/provider-delivery-offline/selected-evidence-v1.json');old={x['route']:x for x in selected['kimi']};baseline=[];sealed=0
 for route in ORDER:
  d=docs[route];row=old[route];b=R/d['baseline']['bundle'];m=load(check_ref(d['baseline']['manifest']));rep=load(check_ref(d['baseline']['report']));seal=load(check_ref(d['baseline']['seal']))
  assert sha(b/'manifest.json')==PINS[route] and row['bundle']==d['baseline']['bundle']
  inventory={str(p.relative_to(b)):sha(p) for p in b.rglob('*') if p.is_file() and p.name not in ['regional-integrity.json','live-report.json']}
  assert inventory==seal['files'],route+' seal inventory'
  assert len(rep['scenarios'])==9 and all(x['outcome']=='PASS' and x['calls']==x['attempts']==1 for x in rep['scenarios'])
  assert len(m['cases'])==9 and all(x['status']=='PASS' for x in m['cases'])
  assert rep['regional']['mode']=='live' and rep['regional']['stage']=='candidate'
  for case in m['cases']:
   for name,h in case['artifacts'].items():assert sha(b/case['dir']/name)==h
  for p in b.rglob('*'):
   if p.is_file():protected[str(p.relative_to(R))]=sha(p)
  sealed+=len(inventory)
  src=load(b/'regional-execution-source.json');nativehash=hashlib.sha256(json.dumps(src,sort_keys=True,separators=(',',':'),ensure_ascii=False).encode()).hexdigest();assert nativehash==rep['regional']['sourceSha256']==rep['regional']['buildSourceSha256']
  pr=load(check_ref(row['original']['receipt']));pp=R/row['original']['receipt']['path']
  assert pr['status']=='PASS' and pr['exitCode']==0 and pr['stage']=='live'
  assert pr['argv'][1:]==['-test.run=^TestRegionalCandidate$','-test.count=1','-test.timeout=31m']
  assert sha(Path(pr['argv'][0]))==rep['regional']['binarySha256']
  for name,h in src.items():assert pr['source']['files'].get(name,pr['selectedInputs'].get(name))==h
  for name,h in pr['files'].items():assert sha(pp.parent/name)==h;protected[str((pp.parent/name).relative_to(R))]=h
  assert (pp.parent/'exit-code').read_text().strip()=='0'
  assert any(str((pp.parent/x['path']).resolve())==str(b.resolve()) and x['manifestSHA256']==PINS[route] for x in pr['bundles'])
  grant=load(b/'regional-grant.json');ar=load(b/'regional-authorization-receipt.json')
  assert sha(b/'regional-grant.json')==rep['regional']['grantSha256']==ar['grantSha256']
  assert grant['maxCalls']==grant['maxAttempts']==9 and grant['id']==ar['grantId']
  assert datetime.fromisoformat(grant['expiresAt'].replace('Z','+00:00'))>datetime.fromisoformat(ar['startedAt'].replace('Z','+00:00'))
  for name in ['id-'+hashlib.sha256(grant['id'].encode()).hexdigest()+'.json','scope-'+ar['scopeSha256']+'.json']:
   q=R/'.git/barness-kimi-grants'/name;assert load(q)==ar;protected[str(q.relative_to(R))]=sha(q)
  for vref in row['original'].values():
   if isinstance(vref,dict) and 'path' in vref and 'sha256' in vref:check_ref(vref)
  lp=load(check_ref(row['original']['ledgerProof']))
  for rec in lp['records']:
   assert sha(R/'.git'/rec['trustedGitCommonRelativePath'])==sha(R/rec['preservedCopy'])==rec['sha256']
  baseline.append(dict(route=route,cases=9,manifestSHA256=PINS[route],sourceSHA256=nativehash,producer=row['original']['receipt'],budget=rep['budget'],authorization=ref(b/'regional-grant.json'),ledger=ref(b/'regional-authorization-receipt.json')))
  protected[str(pp.relative_to(R))]=sha(pp);protected[str(Path(pr['argv'][0]).relative_to(R))]=sha(Path(pr['argv'][0]))
 # Generate six INCOMPLETE scopes. No IDs, expiry or executable authority.
 (B/'scopes').mkdir();prices=load(R/'ai/internal/testkit/kimi/prices.json');scopes=[]
 for route in ORDER:
  d=docs[route];g=readonly('preflight/templates/'+route,[str(R/f['processgrant']['path']),'-template',route,'-mode','live','-stage','candidate','-purpose','xai-history-supplement','-baseline-manifest-sha256',PINS[route]])
  assert g['schema']==2 and g['sourceSha256']==NATIVE and g['expiresAt']=='0001-01-01T00:00:00Z'
  for k in ['endpoint','sdk','catalogVersion','catalogHash','provider']:assert g[k]==d['facts'][{'catalogVersion':'candidateCatalogVersion','catalogHash':'candidateCatalogHash'}.get(k,k)]
  assert g['contract']=='kimi/'+route+'/xai-origin-history/v1' and g['scenarioIds']==['kimi-k3/xai-origin-history']
  assert g['models']==['kimi-k3'] and g['maxCalls']==g['maxAttempts']==1 and g['maxRequestedOutputTokens']==g['maxReservedInputTokens']==4096 and g['callIntervalMs']==60000
  g['binarySha256']=f['kimi']['sha256'];g['maxUSD']=0.10
  p=next(p for p in prices['prices'] if p['provider']==g['provider'] and p['api']==d['facts']['api'])
  fx=Decimal(1) if p['currency']=='USD' else Decimal(str(p['fxUSDCNY']))
  reserve=(Decimal(4096)*max(Decimal(str(p[k])) for k in ['input','cacheRead','cacheWrite5m','cacheWrite1h'])+Decimal(4096)*Decimal(str(p['output'])))/Decimal(1000000)/fx
  assert reserve<=Decimal('0.10')
  doc=dict(kind='INCOMPLETE_MACHINE_SCOPE_NOT_EXECUTABLE',route=route,machineTemplate=g,baseline=d['baseline'],producer=old[route]['original']['receipt'],identity=d['identityToConfirm'],confirmedFacts=index['confirmationEvidence'],freeze=ref(FREEZE),reservationAt4096InputOutputUSD=str(reserve),retries=0,callSeconds=120,processSeconds=240,grantIssued=False)
  sp=B/'scopes'/f'{route}.json';save(sp,doc);scopes.append(ref(sp))
 common=subprocess.check_output(['git','rev-parse','--path-format=absolute','--git-common-dir'],cwd=R,text=True,env=environment('e2e')[0]).strip();assert Path(common).resolve()==(R/'.git').resolve()
 save(B/'authority-before.json',stores())
 for root in [R/'.evidence/provider-minimal-paid-supplement',R/'.evidence/kimi-composition-promotion',R/'.scratch/kimi-composition-promotion/pending-scopes-v5-confirmed-v1']:
  for p in root.rglob('*'):
   if p.is_file() and not p.is_symlink():protected[str(p.relative_to(R))]=sha(p)
 save(B/'protected-before.json',protected)
 save(B/'preflight.json',dict(pass_=True,kind='offline-reread-original54-and-frozen-inputs',source=f['source'],freeze=ref(FREEZE),ordinary=ref(op),ordinarySelectedInputs=len(o['selectedInputs']),baseline=baseline,sealedBaselineFiles=sealed,semanticProofs=proofs,proofSourceDifferences=proof_diffs,machineScopes=scopes,protectedFiles=len(protected),grantsIssued=0,vendorCalls=0,oldSourceBinaryBindingsNotReused=True))
 state('离线核验通过：当前2051来源/187普通输入、冻结二进制及工具、原54项完整原件/真实producer/原授权和账本均已核对。新六份机器范围无ID/有效期，不可执行。尚未签发grant，厂商请求0；下一步仅①kimi-cn-chat。')
 print(json.dumps(dict(preflight='PASS',baselineCases=54,scopes=len(scopes),sealedFiles=sealed,protectedFiles=len(protected),grantsIssued=0,vendorCalls=0)))
except Exception as e:
 stop('offline preflight failed: '+str(e));raise
