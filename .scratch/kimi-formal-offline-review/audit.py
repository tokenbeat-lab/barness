"""Offline data/source review. No producer, test, activation or grant is run."""
from pathlib import Path
import collections,datetime,gzip,hashlib,json,re,sys
from decimal import Decimal,ROUND_HALF_EVEN
ROOT=Path(__file__).resolve().parents[2];OUT=ROOT/'.evidence/kimi-formal-offline-review';BASE=ROOT/'.evidence/minimax-thinking-supplement';SRC=ROOT/'ai/internal/testkit/kimi'
sys.path.insert(0,str(ROOT/'tools/verification'));from receipt_files import source_snapshot
sha=lambda p:hashlib.sha256(Path(p).read_bytes()).hexdigest()
load=lambda p:json.loads(Path(p).read_text())
def save(name,value):
 with (OUT/name).open('x') as f:json.dump(value,f,indent=2,ensure_ascii=False);f.write('\n')
def ref(p):
 p=Path(p);p=p if p.is_absolute() else ROOT/p
 return {'path':str(p.relative_to(ROOT)),'sha256':sha(p)}
def checkref(r):
 p=ROOT/r['path'];assert p.is_file() and not p.is_symlink() and sha(p)==r['sha256'],r['path'];return p
freeze=load(BASE/'build-v1/freeze.json');ordinaryRef=load(BASE/'verification-and-retention-v1.json')['ordinary'];ordinary=load(checkref(ordinaryRef));before=source_snapshot(ROOT)
assert before==freeze['source']==ordinary['source'];assert ordinary['status']=='PASS' and ordinary['exitCode']==0
assert all(sha(ROOT/n)==h for n,h in ordinary['selectedInputs'].items())
storeBefore=load(BASE/'initial/public-stores-before.json');roots=load(ROOT/'.evidence/minimax-thinking-supplement-plan/store-roots.json')
def store():return {str(p):sha(p) for r in roots for p in Path(r).rglob('*') if p.is_file()}
assert store()==storeBefore
sources=load(SRC/'source-files.json');texts={};verified=[]
for row in sources['sources']:
 p=SRC/'sources'/row['path'];packed=p.read_bytes();raw=gzip.decompress(packed)
 assert hashlib.sha256(packed).hexdigest()==row['compressedSHA256'] and hashlib.sha256(raw).hexdigest()==row['sha256'] and len(raw)==row['bytes']
 texts[row['id']]=raw.decode();verified.append({'sourceID':row['id'],'file':ref(p),'originalSHA256':row['sha256'],'originalBytes':len(raw),'retrievedAt':row['retrievedAt']})
assert len(verified)==17
snapshot=hashlib.sha256()
for n in ['candidate.json','prices.json','source-files.json']:snapshot.update((n+'\n').encode());snapshot.update((SRC/n).read_bytes())
snapshot=snapshot.hexdigest();runtime=load(ROOT/'ai/catalog_kimi.json');loader=(ROOT/'ai/catalog_kimi_data.go').read_text();promoter=(ROOT/'ai/internal/testkit/supportmatrix/kimi_promotion.go').read_text()
assert runtime=={'schema':1,'active':False,'version':'','models':[]}
assert snapshot==re.search(r'payload.ProvenanceSHA256 != "([a-f0-9]{64})"',loader).group(1)
assert sha(ROOT/'ai/catalog_kimi.json')==re.search(r'kimiInactiveSHA256 = "([a-f0-9]{64})"',promoter).group(1)
version=re.search(r'kimiActivationVersion = "([^"]+)"',promoter).group(1);assert version in loader
catalog=load(SRC/'candidate.json');prices=load(SRC/'prices.json');models=catalog['models'];assert len(models)==6
nativeTables={}
for sourceID,provider in [('cn-pricing-chat','moonshot-cn'),('intl-pricing-chat','moonshot')]:
 text=texts[sourceID];r=re.search(r'rows:\[\[`kimi-k3`.*?\]\]',text).group(0)
 if provider=='moonshot-cn':nums=[Decimal(v) for v in re.findall('¥([0-9.]+)',r)];currency='CNY'
 else:nums=[Decimal(v) for v in re.findall(r'children:\[`\$`,`([0-9.]+)`\]',r)];currency='USD'
 assert len(nums)==5 and '1,048,576 tokens' in r
 nativeTables[provider]={'currency':currency,'cacheWrite5m':nums[0],'cacheWrite1h':nums[1],'cacheRead':nums[2],'input':nums[3],'output':nums[4],'contextWindow':1048576,'sourceID':sourceID}
rows=[]
for m in models:
 provider,api=m['provider'],m['api'];market='cn' if provider=='moonshot-cn' else 'intl';kind={'openai-completions':'chat','openai-responses':'responses','anthropic-messages':'messages'}[api];route='kimi-'+market+'-'+kind
 assert m['id']=='kimi-k3' and m['input']==['text'] and m['contextWindow']==1048576 and m['maxTokens']==131072 and m['reasoning'] and m['defaultThinkingLevel']=='max' and m['compat']=={'requiresThinking':True}
 assert m['thinkingLevelMap']=={'off':None,'minimal':None,'low':'low','medium':None,'high':'high','xhigh':None,'max':'max'} and 'tiers' not in m['cost']
 p=next(x for x in prices['prices'] if x['provider']==provider and x['api']==api);n=nativeTables[provider]
 assert p['model']=='kimi-k3' and p['currency']==n['currency']
 for f in ['input','output','cacheRead','cacheWrite5m','cacheWrite1h']:assert Decimal(str(p[f]))==n[f]
 sourceRow=next(x for x in sources['sources'] if x['id']==p['sourceID']);assert p['sourceSHA256']==sourceRow['sha256']
 fx=Decimal('6.7220') if provider=='moonshot-cn' else Decimal(1)
 if provider=='moonshot-cn':assert p['fxID']=='fx-2026-10-16' and Decimal(str(p['fxUSDCNY']))==fx and '6.7220' in texts[p['fxID']]
 expected={k:(n[v]/fx).quantize(Decimal('0.0000000001'),rounding=ROUND_HALF_EVEN) for k,v in [('input','input'),('output','output'),('cacheRead','cacheRead'),('cacheWrite','cacheWrite5m')]}
 assert all(Decimal(str(m['cost'][k]))==v for k,v in expected.items())
 usdReserve=Decimal(4096)*(max(expected['input'],expected['cacheRead'],expected['cacheWrite'],expected['input']*2)+expected['output'])/Decimal(1000000)
 assert usdReserve<Decimal('.10')
 proposalPath=ROOT/'.scratch/minimax-thinking-supplement/pending-scopes-v3'/(route+'.json');d=load(proposalPath)
 for f in ['manifest','report','seal']:checkref(d['baseline'][f])
 assert d['facts']['provider']==provider and d['facts']['api']==api and d['model']=='kimi-k3';checkref(d['frozenProducer'])
 rows.append({'route':route,'provider':provider,'api':api,'model':m['id'],'input':m['input'],'contextWindow':m['contextWindow'],'supportedMaxOutput':m['maxTokens'],'reasoningDefault':'max','levels':['low','high','max'],'compat':m['compat'],'candidateUSDPerMillion':m['cost'],'nativePricesPerMillion':{k:str(v) if isinstance(v,Decimal) else v for k,v in n.items()},'fx':str(fx),'conservativeSingleUSD':str(usdReserve),'proposedCapUSD':'0.10','metadataReview':'PASS_FOR_FROZEN_SUBSET','actualNewHistoryScene':'NOT_RUN','proposal':ref(proposalPath)})
assert len({(m['provider'],m['api'],m['id']) for m in models})==6
# Verify the static handoff gap, without invoking the mutating activation code.
regional=(ROOT/'ai/internal/testkit/supportmatrix/regional.go').read_text();selection=(ROOT/'ai/internal/testkit/supportmatrix/regional_selection.go').read_text()
assert 'regionalReport(dir, false, mode)' in promoter and 'VerifyKimiComposition' not in promoter and 'IsKimiComposition' not in promoter
assert 'source != e.SourceSHA256' in promoter and 'binary != e.BinarySHA256' in promoter and 'commit != r.GitCommit' in promoter
assert 'regionalSelectedReport(dir, formal, mode, regionalFull)' in regional and 'filepath.Join(dir, RegionalReportFile)' in regional
assert 'return []string{"kimi-k3/" + grokhistory.Scenario}' in selection
mm=load(ROOT/'.scratch/minimax-thinking-supplement/pending-scopes-v3/minimax-messages.json');checkref(mm['frozenProducer']);mmSources=load(ROOT/'ai/internal/testkit/minimax/price-sources.json');mmScope=mmSources['scope']
assert mmScope['lowUsdPerMillion']=={'input':'0.30','output':'1.20','cacheRead':'0.06'} and mmScope['highUsdPerMillion']=={'input':'0.60','output':'2.40','cacheRead':'0.12'}
mmReserve=Decimal(2048)*(Decimal(mmScope['highUsdPerMillion']['input'])+Decimal(mmScope['highUsdPerMillion']['output']))/Decimal(1000000);assert mmReserve==Decimal('0.006144')
index=load(BASE/'pending-scopes-v3.json');assert len(index['scopeDocuments'])==7
for r in index['scopeDocuments']:
 d=load(checkref(r));assert not d['approved'] and not d['grantIssued'] and not d['executable'] and d['authorizationStillRequired']['id']=='NOT_ISSUED' and d['authorizationStillRequired']['expiresAt']=='NOT_ISSUED'
groups=[{'group':'minimax-international','routes':['minimax-messages'],'aliases':{'accountAlias':'MINIMAX','regionAlias':'global'},'tenantID':'live-smoke','accountScopeID':'live-minimax-messages','endpoint':'https://api.minimax.io/anthropic','model':'MiniMax-M3','accountPurpose':'own international standard API; exclude M Plan and priority','identityConfirmation':'PENDING_TRUSTED_HOST','rateConfirmation':'PENDING_TRUSTED_HOST','standardRateBasis':mmScope,'calls':1,'proposedUSD':'0.01','knownRateReservationUSD':str(mmReserve)},
 {'group':'kimi-cn','routes':['kimi-cn-chat','kimi-cn-responses','kimi-cn-messages'],'aliases':{'accountAlias':'KIMI','regionAlias':'CN'},'tenantID':'live-smoke','accountScopeID':'regional:KIMI','endpointChatAndResponses':'https://api.moonshot.cn/v1','endpointMessages':'https://api.moonshot.cn/anthropic','model':'kimi-k3','accountPurpose':'own standard metered API; no web/Kimi Code subscription or proxy','identityConfirmation':'PENDING_TRUSTED_HOST','sameUnderlyingAccountAcrossOwnThreeProtocols':'PENDING_TRUSTED_HOST','rateConfirmation':'PENDING_TRUSTED_HOST','currency':'CNY','nativeRatesPerMillion':{k:str(v) for k,v in nativeTables['moonshot-cn'].items() if isinstance(v,Decimal)},'historicalUSDConversion':'1 USD = 6.7220 CNY, ADR-0052; not a new confirmation request or current FX quote','calls':3,'proposedUSD':'0.30'},
 {'group':'kimi-intl','routes':['kimi-intl-chat','kimi-intl-responses','kimi-intl-messages'],'aliases':{'accountAlias':'KIMI','regionAlias':'GLOBAL'},'tenantID':'live-smoke','accountScopeID':'regional:KIMI','endpointChatAndResponses':'https://api.moonshot.ai/v1','endpointMessages':'https://api.moonshot.ai/anthropic','model':'kimi-k3','accountPurpose':'own standard metered API; no web/Kimi Code subscription or proxy','identityConfirmation':'PENDING_TRUSTED_HOST','sameUnderlyingAccountAcrossOwnThreeProtocols':'PENDING_TRUSTED_HOST','rateConfirmation':'PENDING_TRUSTED_HOST','currency':'USD','nativeRatesPerMillion':{k:str(v) for k,v in nativeTables['moonshot'].items() if isinstance(v,Decimal)},'calls':3,'proposedUSD':'0.30'}]
inputs={str(p.relative_to(ROOT)):sha(p) for p in [ROOT/'ai/catalog_kimi.json',ROOT/'ai/catalog_kimi_data.go',ROOT/'ai/kimi_catalog.go',ROOT/'ai/internal/testkit/supportmatrix/kimi_promotion.go',ROOT/'ai/internal/testkit/supportmatrix/kimi_promotion_files.go',ROOT/'ai/internal/testkit/supportmatrix/regional.go',ROOT/'ai/internal/testkit/supportmatrix/regional_kimi.go',ROOT/'ai/internal/testkit/supportmatrix/kimi_composition.go',ROOT/'ai/live/cmd/regionalpromote/main.go',ROOT/'docs/adr/0052-barness-ai-kimi-cost-estimates.md',ROOT/'docs/adr/0055-route-scoped-evidence-applicability.md',ROOT/'docs/agents/verification.md',SRC/'candidate.json',SRC/'prices.json',SRC/'source-files.json',ROOT/'ai/internal/testkit/minimax/price-sources.json',BASE/'pending-scopes-v3.json',BASE/'build-v1/freeze.json',BASE/'verification-and-retention-v1.json']}
for r in index['scopeDocuments']:inputs[r['path']]=r['sha256']
value={'kind':'offline-kimi-formal-catalog-review','reviewedAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'metadataReview':'PASS_FOR_RETAINED_FACTS','productionCatalogStillInactive':True,'productionCatalog':ref(ROOT/'ai/catalog_kimi.json'),'reviewedCandidate':ref(SRC/'candidate.json'),'targetVersionAfterFutureActivation':version,'snapshotSHA256':snapshot,'sourceOriginals':verified,'identities':rows,'promotionHandoff':{'status':'BLOCKED_SOFTWARE_COMPOSITION_DISPATCH','basis':'static source call chain; no CLI reproduction or tests executed this turn','locations':['ai/internal/testkit/supportmatrix/kimi_promotion.go:37','ai/internal/testkit/supportmatrix/kimi_promotion.go:45','ai/internal/testkit/supportmatrix/kimi_promotion.go:51','ai/internal/testkit/supportmatrix/regional.go:55'],'issue':'.scratch/kimi-formal-offline-review/issues/01-kimi-promotion-composition.md','scope':'Kimi-only fixed6 route promotion with each real9+1 qualification; retain12 measurements, no synthetic unified Report or global source equality','implication':'Fix and verify offline before paid work; refreeze affected unapproved machine bindings once; no old vendor evidence expires merely from this collector/consumer change'},'sourceAndOrdinaryUnchanged':{'sourceFiles':len(before['files']),'selectedInputs':len(ordinary['selectedInputs']),'ordinary':ordinaryRef},'accountGroups':groups,'minimalHumanInformation':['same own standard metered actual account/region/model entitlement as each old baseline; Kimi each site three protocols share that corresponding account or list exceptions','actual account standard rates/service tier still fit the retained rate table; report changed or unknown group only'],'notNeededFromHuman':['API keys, tokens, passwords or full account identifiers','repeat route/model/endpoint/SDK/binary/manifest hashes','new real-time FX or reapproval of already accepted historical FX design','grant IDs/expiry or paid execution approval in this turn'],'totalProposedUSD':'0.61','paidApprovalStillSeparate':True,'newActualVendorScenesStillMissing':7,'testsExecuted':False,'catalogActivated':False,'grantsIssued':0,'realClaims':0,'vendorCalls':0,'raceExecuted':False,'pressureExecuted':False,'pausedResumed':False,'readInputs':inputs,'historicalRetentionReference':ref(BASE/'history-protection-final-v2.json')}
assert before==source_snapshot(ROOT) and all(sha(ROOT/n)==h for n,h in ordinary['selectedInputs'].items());assert store()==storeBefore
value['retention']={'publicAuthorityFiles':len(storeBefore),'exactSetAndBytesUnchanged':True,'sevenProposalsUnchanged':True,'frozenBinariesUnchanged':True,'runtimeAndTestSourcesUnchanged':True,'allHistoricalPayloadRescanned':False}
save('review-v1.json',value);save('confirmation-checklist-v1.json',{'documentType':'ACCOUNT_AND_RATE_FACT_CONFIRMATION_NOT_PAID_AUTHORIZATION','groups':groups,'minimalQuestions':value['minimalHumanInformation'],'paidApprovalStillSeparate':True,'grantsIssued':0,'totalProposedUSD':'0.61'})
print(json.dumps({'metadata':value['metadataReview'],'originalSources':len(verified),'identities':len(rows),'promotion':value['promotionHandoff']['status'],'accountGroups':len(groups),'sourceFiles':len(before['files']),'ordinaryInputs':len(ordinary['selectedInputs']),'publicStoreFiles':len(storeBefore),'vendorCalls':0},ensure_ascii=False))
