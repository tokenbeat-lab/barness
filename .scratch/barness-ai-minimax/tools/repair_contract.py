"""The single six-process release-repair authority; no arbitrary round or caps."""
import hashlib
import paid_core as c
import paid_plan as p
import routed_contract as old
ROUND='release-repair-validation-1'
PROXY=old.PROXY
ACCOUNTING_SCOPE='current MiniMax-release renewal: prior39 closed545/545 plus release-repair-validation-1; earlier standalone GLM and MiniMax histories stay separate and immutable'
BUDGET=dict(priorCalls=545,priorAttempts=545,additionalMaxCalls=108,additionalMaxAttempts=132,totalMaxCalls=653,totalMaxAttempts=677)
NAMES=('glm-chat','glm-responses','glm-messages','protocol114','visual-jev','visual-decisions')
PROPOSAL='Approve one fixed release-repair-validation-1: GLM3 + protocol114 + visual2; additional maximum108 calls/132 HTTP Attempts, known prior545/545, combined maximum653/677; no automatic retry or next round.'
PROPOSAL_SHA256=hashlib.sha256(PROPOSAL.encode()).hexdigest()
EXTRA_PINS={'history.json':'historySHA256','reuse.json':'reuseSHA256','scope.json':'scopeSHA256'}

def phase_for(base):return 'glm' if base.startswith('glm-') else 'effects'

def jobs(out,phase=None):
 original=p.jobs(c.ROOT,out,c.common_root());result={}
 for base in NAMES:
  if phase is not None and phase_for(base)!=phase:continue
  job=original[base];name=ROUND+'-'+base;job.update(name=name,baseName=base,validationRound=ROUND,plannedCalls=22 if base.startswith('glm-') else 12)
  if base.startswith('glm-'):job.update(glmRoute=base.removeprefix('glm-'),networkRoute='china-direct')
  else:
   job['networkRoute']='international-loopback-proxy'
   if base.startswith('visual-'):job['argv']+=['-proxy',PROXY]
   else:job['environment']['BARNESS_AI_LIVE_PROXY']=PROXY
  result[name]=job
 return result

def report_path(out,job):return p.report_path(out,job|dict(name=job['baseName']))

def expected_authorization(value,phase,pins,checks,review):
 text=value.get('requestText');reference=value.get('requestReference')
 if not isinstance(text,str) or not text.strip() or text==old.REQUEST or not isinstance(reference,str) or not reference.strip():raise RuntimeError('new explicit additional108/132 approval request/reply required')
 return dict(decision='APPROVED',authority='explicit-additional-budget-reply',requestReference=reference,requestText=text,
  approvalProposalSHA256=PROPOSAL_SHA256,approvedAdditionalMaxCalls=108,approvedAdditionalMaxAttempts=132,
  round=ROUND,phase=phase,budget=BUDGET,accountingScope=ACCOUNTING_SCOPE,**pins,checksSHA256=checks,sourceReviewSHA256=review)

def validate_checks(reader,relative,checks):
 import paid_checks
 paid_checks.verify_at(reader,relative,checks)
