"""One fixed eleven-group authority; its two phases share one finite budget."""
import hashlib
import paid_core as c
import paid_plan as p
import paid_checks
ROUND='current-source-eleven-renewal-1'
PROXY='http://127.0.0.1:7890'
NAMES=tuple('glm-'+r for r in ('chat','responses','messages'))+tuple('mm-global-formal-m3-'+r for r in ('chat','responses','messages'))+('text-jev','text-luna','protocol114','visual-jev','visual-decisions')
BUDGET=dict(priorCalls=648,priorAttempts=648,additionalMaxCalls=216,additionalMaxAttempts=252,totalMaxCalls=864,totalMaxAttempts=900)
ACCOUNTING_SCOPE='prior current-release renewal648/648; original M3 formal39/36 remains separate; this fixed eleven renewal adds at most216/252'
PROPOSAL='Approve current-source-eleven-renewal-1 only: GLM3 then M3formal3+text2+protocol114+visual2; additional maximum216 Logical Calls/252 HTTP Attempts; output4096; registered M3 scenario-local bounded reruns stay within each20/24 and aggregate60/72; no automatic process rerun or next round; prior648/648 and original M3formal39/36 retained separately. Phase2 requires three closed known GLM results: own3PASS permits strict committed GLM selector publication only; any known FAIL permits no publication and unchanged execution source/data. Each phase gets a fresh exact freeze/check/review.'
PROPOSAL_SHA256=hashlib.sha256(PROPOSAL.encode()).hexdigest()
EXTRA_PINS={'history.json':'historySHA256','generic.json':'genericSHA256','transition.json':'transitionSHA256','native-renewal-facts.json':'nativeRenewalFactsSHA256'}
GLM_DATA={'ai/live/testdata/glm-formal/'+n for n in ('live-captures.json','live-captures.sha256','known-usage.json','support-views.json')}

def phase_for(base):return 'glm' if base.startswith('glm-') else 'remaining'

def jobs(out,phase=None):
 original=p.jobs(c.ROOT,out,c.common_root());result={}
 for base in NAMES:
  if phase is not None and phase_for(base)!=phase:continue
  j=original[base];name=ROUND+'-'+base;j.update(name=name,baseName=base,validationRound=ROUND,maxOutputTokens=4096)
  if base.startswith('glm-'):j.update(glmRoute=base[4:],networkRoute='china-direct',plannedCalls=22,plannedAttempts=22)
  else:
   j['networkRoute']='international-loopback-proxy'
   if base.startswith(('text-','visual-')):j['argv']+=['-proxy',PROXY]
   else:j['environment']['BARNESS_MINIMAX_PROXY' if base.startswith('mm-') else 'BARNESS_AI_LIVE_PROXY']=PROXY
   j.update(plannedCalls=13 if base.startswith('mm-') else 24 if base.startswith('text-') else 12,plannedAttempts=12 if base.startswith('mm-') else 24 if base.startswith('text-') else 12)
  if base.startswith('mm-'):
   j['environment'].update(BARNESS_MINIMAX_FORMAL_RENEWAL=ROUND,BARNESS_MINIMAX_LEDGER=str(c.common_root()/'.evidence/minimax-renewals'/ROUND/'formal.jsonl'))
  result[name]=j
 return result

def report_path(out,job):return p.report_path(out,job|dict(name=job['baseName']))

def expected_authorization(value,phase,pins,checks,review):
 if not isinstance(value.get('budget'),dict) or any(type(v) is not int for v in value['budget'].values()):raise RuntimeError('exact integer fixed budget required')
 if any(type(value.get(k)) is not int for k in ('approvedAdditionalMaxCalls','approvedAdditionalMaxAttempts')):raise RuntimeError('exact integer216/252 approval required')
 if not all(isinstance(value.get(k),str) and value[k].strip() for k in ('requestText','requestReference')):raise RuntimeError('new explicit216/252 human reply required')
 return dict(decision='APPROVED',authority='explicit-additional-budget-reply',requestReference=value['requestReference'],requestText=value['requestText'],approvalProposalSHA256=PROPOSAL_SHA256,approvedAdditionalMaxCalls=216,approvedAdditionalMaxAttempts=252,round=ROUND,phase=phase,budget=BUDGET,accountingScope=ACCOUNTING_SCOPE,**pins,checksSHA256=checks,sourceReviewSHA256=review)

def validate_checks(reader,relative,checks):return paid_checks.verify_at(reader,relative,checks)
