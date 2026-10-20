"""Fixed recovery6: corrected policy2 estimator, nineteen own full processes."""
from pathlib import Path
import final_contract as old
from recovery_flow_source import SourceFreeze
c=old.c
ROUND='provider-sites-final-acceptance-6'
SPENDING_GUARD_POLICY_SELECTOR=old.ROUND
BASE=Path('.evidence/provider-sites/final-acceptance-20261020')
MAX_CALLS,MAX_ATTEMPTS=378,414
CAMPAIGN_CALLS,CAMPAIGN_ATTEMPTS=1538,1646
RETAINED_CALLS,RETAINED_ATTEMPTS,RETAINED_USD=1160,1232,32050000
NEW_USD=6900000
MAX_USD=39
BASE_ROLE_COUNT=14
ROUTES=old.ROUTES
REGIONAL=old.REGIONAL
EFFECTS=old.EFFECTS
PROXY='http://127.0.0.1:7890'
EXECUTION_ORDER=('formal-zai-responses',)+tuple('mm-global-formal-m3-'+v for v in ROUTES)+tuple('formal-'+v for v in REGIONAL if v!='zai-responses')+tuple('glm-'+v for v in ROUTES)+('openai-decisions','typesafe-classifier','protocol114')+EFFECTS
GUARD_FILES=('ai/internal/testkit/finalbudget/guard.go','ai/internal/testkit/finalbudget/estimate.go')
RESERVES={'formal-'+p+'-'+route:650000 if p=='zai' else 350000 for p in ('zai','minimax-cn') for route in ROUTES}
RESERVES.update({'mm-global-formal-m3-'+route:200000 for route in ROUTES})
RESERVES.update({'glm-chat':1000000,'glm-responses':500000,'glm-messages':500000,'openai-decisions':200000,'typesafe-classifier':100000,'protocol114':200000})
RESERVES.update({name:100000 if name.startswith('text-') else 300000 for name in EFFECTS})
RETAINED_GENERIC=tuple(v[0] for v in old.old.GENERIC[:8])

def policy():return dict(spendingGuardPolicyId=SPENDING_GUARD_POLICY_SELECTOR,spendingGuardSourceSHA256={n:SourceFreeze(c).digest(c.ROOT/n) for n in GUARD_FILES})
def jobs(out):
    result={n:j for n,j in old.jobs(out).items() if n in EXECUTION_ORDER}
    for name,j in result.items():
        j.update(round=ROUND,validationRound=ROUND,nativeRenewalId=ROUND,phase='recovery6',usdLedger=str(out/(name+'-usd.json')),reservedMicrousd=RESERVES[name],networkRoute='direct' if j['provider'] in ('minimax-cn','zhipu','deepseek') else PROXY,**policy())
        if name.startswith('mm-'):
            j['consumerTimeoutSeconds']=360
            j['controllerTimeoutSeconds']=j['timeoutSeconds']+720
            j['environment'].update(BARNESS_MINIMAX_FORMAL_RENEWAL=ROUND,BARNESS_MINIMAX_LEDGER=str(c.common_root()/'.evidence/minimax-renewals'/ROUND/'formal.jsonl'))
    return {n:result[n] for n in EXECUTION_ORDER}

def inventory(out):
    return dict(schema=1,round=ROUND,validationRound=ROUND,nativeRenewalId=ROUND,maxCalls=MAX_CALLS,maxAttempts=MAX_ATTEMPTS,campaignMaxCalls=CAMPAIGN_CALLS,campaignMaxAttempts=CAMPAIGN_ATTEMPTS,newReservedMicrousd=NEW_USD,retainedReservedMicrousd=RETAINED_USD,maxMicrousd=39000000,automaticProcessReruns=False,credentialsRead=False,originalCampaignApprovalSHA256=SourceFreeze(c).digest(c.common_root()/BASE/'authorization.json'),executionOrder=list(EXECUTION_ORDER),**policy(),jobs=jobs(out))

# The23 public support identities remain the same; collection authority does not.
live_identities=old.live_identities

# Exact closed-fifth accounting and independently bounded sixth scope.
RETAINED_ACTUAL_CALLS,RETAINED_ACTUAL_ATTEMPTS=702,692
ORIGINAL_PROCESS_COUNT,UNKNOWN_PROCESSES=48,5
PREDECESSOR_SEAL_FIELD='fifthCohortSealSHA256'
ACTUAL_HUMAN_LEAF='recovery6-authorization.json'
REGIONAL_GRANT_PREFIX='ps-final-6-'
NATIVE_PREFLIGHT_FLAG='-recovery6-preflight='
FINAL_SELECTION_FILE='effective33-retained67.json'

PHASE='recovery6'

RETAINED_INPUT_FILE='retained-generic.json'
RETAINED_INPUT_FIELD='retainedGenericSHA256'
LEDGER_BASIS='Corrected operation-specific guard under funding selector2'
