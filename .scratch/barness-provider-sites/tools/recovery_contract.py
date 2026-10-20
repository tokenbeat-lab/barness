"""Fixed recovery3 execution scope; the unchanged USD guard is policy2 only."""
from pathlib import Path
import final_contract as old
c=old.c
ROUND='provider-sites-final-acceptance-3'
SPENDING_GUARD_POLICY_SELECTOR=old.ROUND
BASE=Path('.evidence/provider-sites/final-acceptance-20261020')
MAX_CALLS,MAX_ATTEMPTS=506,518
CAMPAIGN_CALLS,CAMPAIGN_ATTEMPTS=962,1010
RETAINED_CALLS,RETAINED_ATTEMPTS,RETAINED_USD=456,492,10700000
NEW_USD=19200000
ROUTES=old.ROUTES
REGIONAL=old.REGIONAL
EFFECTS=old.EFFECTS
PROXY='http://127.0.0.1:7890'
EXECUTION_ORDER=tuple('mm-global-formal-m3-'+v for v in ROUTES)+tuple('formal-'+v for v in REGIONAL)+tuple(v[0] for v in old.old.GENERIC)+('protocol114',)+EFFECTS
GUARD_FILES=('ai/internal/testkit/finalbudget/guard.go','ai/internal/testkit/finalbudget/estimate.go')
RESERVES=dict(zip([v[0] for v in old.old.GENERIC],(1500000,1500000,1800000,500000,500000,500000,3500000,4500000,200000,100000)))
RESERVES.update(protocol114=200000,**{'mm-global-formal-m3-'+r:200000 for r in ROUTES},**{n:100000 if n.startswith('text-') else 300000 for n in EFFECTS},**{'formal-'+p+'-'+r:650000 if p=='zai' else 350000 for p in ('zai','minimax-cn') for r in ROUTES})

def policy():return dict(spendingGuardPolicyId=SPENDING_GUARD_POLICY_SELECTOR,spendingGuardSourceSHA256={n:c.digest(c.ROOT/n) for n in GUARD_FILES})
def jobs(out):
    result={n:j for n,j in old.jobs(out).items() if j['phase']=='remaining'}
    for name,j in result.items():
        j.update(round=ROUND,validationRound=ROUND,nativeRenewalId=ROUND,phase='recovery3',usdLedger=str(out/(name+'-usd.json')),reservedMicrousd=RESERVES[name],networkRoute='direct' if j['provider'] in ('minimax-cn','zhipu','deepseek') else PROXY,**policy())
        if name.startswith('mm-'):
            j['environment'].update(BARNESS_MINIMAX_FORMAL_RENEWAL=ROUND,BARNESS_MINIMAX_LEDGER=str(c.common_root()/'.evidence/minimax-renewals'/ROUND/'formal.jsonl'))
    return {n:result[n] for n in EXECUTION_ORDER}

def inventory(out):
    return dict(schema=1,round=ROUND,validationRound=ROUND,nativeRenewalId=ROUND,maxCalls=MAX_CALLS,maxAttempts=MAX_ATTEMPTS,campaignMaxCalls=CAMPAIGN_CALLS,campaignMaxAttempts=CAMPAIGN_ATTEMPTS,newReservedMicrousd=NEW_USD,retainedReservedMicrousd=RETAINED_USD,maxMicrousd=30000000,automaticProcessReruns=False,credentialsRead=False,originalCampaignApprovalSHA256=c.digest(c.common_root()/BASE/'authorization.json'),executionOrder=list(EXECUTION_ORDER),**policy(),jobs=jobs(out))

# The23 public support identities remain the same; collection authority does not.
live_identities=old.live_identities

# Fixed-context metadata for the two reviewed recovery workflows.
RETAINED_ACTUAL_CALLS,RETAINED_ACTUAL_ATTEMPTS=303,302
ORIGINAL_PROCESS_COUNT,UNKNOWN_PROCESSES=17,2
PREDECESSOR_SEAL_FIELD='secondCohortSealSHA256'
ACTUAL_HUMAN_LEAF='recovery3-authorization.json'
REGIONAL_GRANT_PREFIX='ps-final-3-'
NATIVE_PREFLIGHT_FLAG='-recovery-preflight='
FINAL_SELECTION_FILE='effective33-retained41.json'

PHASE='recovery3'
