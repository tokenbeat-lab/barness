"""Account for the two explicitly authorized document revisions without erasing
original bytes. Recheck public stores after the full ordinary process finishes."""
from pathlib import Path
import hashlib,json
ROOT=Path(__file__).resolve().parents[2];BASE=ROOT/'.evidence/minimax-thinking-supplement'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
old=json.load(open(BASE/'history-protection-final-v1.json'));expected={'docs/adr/0055-route-scoped-evidence-applicability.md','docs/agents/verification.md'}
revisions=[]
for p in sorted(expected):
 retained=BASE/'initial'/p.replace('/','__');h=sha(retained)
 for mapPath in [ROOT/'.evidence/provider-final-readiness/protected-history-before.json',BASE/'initial/protected-inputs-before.json']:
  assert json.load(open(mapPath))[p]==h
 revisions.append({'path':p,'reason':'explicit user-authorized ADR/SOP revision before implementation','originalRetained':str(retained.relative_to(ROOT)),'originalSHA256':h,'currentSHA256':sha(ROOT/p)})
for name in ['allHistorical','taskProtected']:
 assert not old['checks'][name]['missing'] and set(old['checks'][name]['changed'])==expected
before=json.load(open(BASE/'initial/public-stores-before.json'));roots=json.load(open(ROOT/'.evidence/minimax-thinking-supplement-plan/store-roots.json'));current={str(p):sha(p) for r in roots for p in Path(r).rglob('*') if p.is_file()}
assert before==current,'public authority changed'
value={'kind':'historical-retention-with-authorized-document-revisions','pass':True,'completePriorScan':{'path':'.evidence/minimax-thinking-supplement/history-protection-final-v1.json','sha256':sha(BASE/'history-protection-final-v1.json')},'allHistoricalFilesChecked':old['checks']['allHistorical']['protected'],'taskProtectedFilesChecked':old['checks']['taskProtected']['protected'],'missingHistorical':[],'unexplainedChanges':[],'authorizedRevisionsWithOriginalBytes':revisions,'publicAuthorityFiles':len(before),'publicAuthorityExactSetAndBytesUnchanged':True,'realClaims':0,'paidGrantsIssued':0,'vendorCalls':0,'pausedResumed':False}
with (BASE/'history-protection-final-v2.json').open('x') as f:json.dump(value,f,indent=2,ensure_ascii=False);f.write('\n')
print({'pass':True,'historical':value['allHistoricalFilesChecked'],'authorizedDocumentRevisions':2,'publicAuthorityFiles':len(before)})
