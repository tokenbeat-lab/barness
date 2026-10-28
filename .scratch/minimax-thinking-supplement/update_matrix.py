from pathlib import Path
import hashlib,json
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/minimax-thinking-supplement';SCRATCH=ROOT/'.scratch/minimax-thinking-supplement'
def load(p):return json.loads(Path(p).read_text())
def ref(p):
 p=Path(p);p=p if p.is_absolute() else ROOT/p
 return {'path':str(p.relative_to(ROOT)),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()}
def save(p,v):
 with p.open('x') as f:json.dump(v,f,indent=2,ensure_ascii=False,sort_keys=True);f.write('\n')
old=ROOT/'.evidence/provider-final-readiness/delivery-matrix-v2.json';m=load(old);freeze=load(BASE/'build-v1/freeze.json')
m['previousMatrix']=ref(old);m['source']=freeze['source'];m['pendingScopes']=ref(BASE/'pending-scopes-v3.json');m['newVendorEvidence']=False;m['implementationState']='FIXED_MINIMAX_SINGLE_AND_13_PLUS_1_IMPLEMENTED; ORDINARY RECEIPT IN CURRENT STATUS'
for r in m['paths']:
 if r['route'].startswith('kimi-'):
  r['proposal']=ref(SCRATCH/'pending-scopes-v3'/(r['route']+'.json'));r['executionReadiness']='fixed9+1 unchanged; current normal binary rebound only; 1 real scene remains NOT_RUN'
 if r['route']=='minimax-messages':
  r['actualPaidCallCap']=1;r['USD']='0.01';r['proposal']=ref(SCRATCH/'pending-scopes-v3/minimax-messages.json');r['nativeCaseAnalysis']=ref(BASE/'original-baseline-v5/receipt.json');r['executionReadiness']='fixed thinking-on single and13+1 consumer implemented; normal frozen binary; human authorization and1 real scene NOT_RUN'
m['releaseBlockers']=['MiniMax international Messages: independently authorize and execute thinking-on once, then read-only13+1 real qualification','Kimi six separate authorizations/real xai-origin-history scenes; then9+1 qualification and own catalog review','impact-scoped human race/pressure and final release record; paused rounds stay paused']
m['changeApplicabilityReason']='32 verification/collector/source inputs changed; product runtime,SDK,go.mod/go.sum,formal model facts and prices unchanged. Preserve22 complete paths and original23 model records; no all-provider paid revalidation.'
save(BASE/'delivery-matrix-v3.json',m)
lines=['# Provider 交付矩阵 v3','','12 个 Provider 身份、29 条路径；22 条完整适用，Kimi 六条各9/10，MiniMax 国际Messages13/14。受控组合的14项是机制结论，真实缺口仍共7项。TypeSafe两个版本归为一条路径，完整路线/型号证据共23条。','','| Provider | 路线 | 原真实合同证据 | 当前最小剩余 |','| --- | --- | --- | --- |']
for r in m['paths']:
 if 'actualEvidence' in r:counts='；'.join('/'.join(x['models'])+' '+str(x['requiredContractItems'])+'项' for x in r['actualEvidence']);left='保留原适用性；新增补采0'
 else:
  counts=str(r['applicableActualItems'])+'/'+str(r['requiredContractItems']);left='1项／USD0.10，待独立批准；目录inactive' if r['route'].startswith('kimi-') else 'thinking-on 1项／USD0.01，待独立批准；13＋1已实现'
 lines.append('| '+r['provider']+' | '+r['route']+' | '+counts+' | '+left+' |')
lines+=['','[七份提案](pending-scopes-v3/README.md)合计拟议7调用、7HTTP／USD0.61，权限与预算独立，不借用旧余额。没有签发ID/有效期或执行真实调用。Kimi正式目录与各项目原人工阶段继续单独收口，release.pass=false；其他Provider不因消费者来源改变而付费复验。']
(SCRATCH/'delivery-matrix-v3.md').write_text('\n'.join(lines)+'\n')
print({'paths':len(m['paths']),'counts':m['counts'],'paidCalls':0})
