"""Synthetic closed first-cohort facts for public process tests, never live proof."""
from pathlib import Path
import hashlib
import json

BASE=Path('.evidence/provider-sites/final-acceptance-20261020')
OLD='provider-sites-final-acceptance-1'


def seed(root,source_commit):
    files={}
    def put(relative,value,raw=False):
        path=root/relative;path.parent.mkdir(parents=True,exist_ok=True)
        content=value if raw else (json.dumps(value,sort_keys=True)+'\n').encode()
        path.write_bytes(content);files[str(relative)]=hashlib.sha256(content).hexdigest()
        return files[str(relative)]
    scopes={v:dict(provider=p,purpose='standard',confirmed=True,accountAlias='controlled-own-standard',regionAlias='controlled-own-region') for v,p in [('OPENAI_KEY','openai'),('ANTHROPIC_KEY','anthropic'),('GEMINI_KEY','google'),('DEEPSEEK_KEY','deepseek'),('TYPESAFE_KEY','typesafe'),('ZHIPU_KEY','zhipu'),('ZAI_KEY','zai'),('MINIMAX_KEY','minimax'),('MINIMAX_CN_KEY','minimax-cn')]}
    approval=put(BASE/'authorization.json',dict(decision='APPROVED',roundId=OLD,approvedScope=dict(budgetUSD=30,nineProviderStandardAPITestAccountEligibilityConfirmed=True,race="one final releasegate race after READY"),researchPlanSha256='a'*64))
    parent=Path('.evidence/minimax-renewals')/OLD
    put(parent/'operator-authority.json',dict(round=OLD,synthetic=True))
    (root/parent/'launch.lock').write_bytes(b'')
    out=BASE/'candidate-final'
    for name in ('prepared.json','plan.json','binaries.json','source-at-start.json','regional-source.json','prerelease-source-review.json'):
        put(out/name,dict(sourceCommit=source_commit,synthetic=True))
    put(out/'execution-authorization.json',dict(standardCredentialScopes=scopes))
    processes=[]
    for name,exit_code in [('candidate-zai-messages',0),('candidate-zai-chat',1)]:
        native=out/name/'controlled-session';marker=Path('.evidence/minimax-launches')/(OLD+'-'+name+'.json')
        report=dict(combo=name.removeprefix('candidate-'),gitCommit=source_commit,budget=dict(callsUsed=18,httpAttempts=18),expected=[str(i) for i in range(18)],scenarios=[dict(id=str(i),outcome='FAIL' if exit_code and i==0 else 'PASS') for i in range(18)],regional=dict(mode='live',stage='candidate',auditPassed=True))
        put(native/'regional-report.json',report);put(native/'audit.json',dict(findings=[]))
        put(out/(name+'-command.json'),dict(state='COMPLETED',exitCode=exit_code,launcherAudit=dict(passed=True,findings=[])))
        put(out/(name+'-usd.json'),dict(round=OLD,attempts=18,maxMicrousd=650000,reservedMicrousd=350000))
        pin=put(marker,dict(state='COMPLETED',exitCode=exit_code,validationRound=OLD,sourceCommit=source_commit,maxCalls=24,maxAttempts=24,reservedMicrousd=650000))
        processes.append(dict(name=name,sourceCommit=source_commit,exitCode=exit_code,actualCalls=18,actualHTTPAttempts=18,claimedCalls=24,claimedAttempts=24,wholeReservedMicrousd=650000,scenarios=18,failedScenarios=['0'] if exit_code else [],nativeBundle=str(native),canonicalClaim=str(marker),canonicalClaimSHA256=pin,command=str(out/(name+'-command.json')),qualification='FAIL' if exit_code else 'PASS'))
    research=dict(state='COMPLETED',calls=4,attempts=4,reservedMicrousd=200000,authorizationSHA256=approval)
    put(BASE/'research-index.json',dict(calls=4,httpAttempts=4,synthetic=True))
    put(BASE/'closed-cohort-1/source/operator.py',b'# immutable synthetic closed source\n',raw=True)
    absent=['openai-responses','openai-chat','anthropic-messages','google-gemini','deepseek-responses','deepseek-chat','openai-images','google-interactions-image','openai-decisions','typesafe-classifier','protocol114','glm-chat','glm-responses','glm-messages','mm-global-formal-m3-chat','mm-global-formal-m3-responses','mm-global-formal-m3-messages','text-jev','text-luna','visual-jev','visual-decisions','candidate-zai-responses','candidate-minimax-cn-chat','candidate-minimax-cn-responses','candidate-minimax-cn-messages',*('formal-'+p+'-'+r for p in ('zai','minimax-cn') for r in ('chat','responses','messages'))]
    native_absent=[str(parent/n) for n in ('grant.json','human-approval.json','source-review.json','formal.jsonl','chat.json','responses.json','messages.json')]
    index=dict(schema=1,kind='CLOSED_FAILED_REGIONAL_CANDIDATE_COHORT',roundId=OLD,sourceCommit=source_commit,state='CLOSED_NO_FURTHER_PAID_LAUNCH',research=research,processes=processes,actualCalls=40,actualHTTPAttempts=40,claimedCalls=52,claimedHTTPAttempts=52,wholeReservedMicrousd=1500000,refundMicrousd=0,noProcessRestarts=True,unstartedCanonicalClaims=[str(Path('.evidence/minimax-launches')/(OLD+'-'+n+'.json')) for n in absent],nativeMiniMaxRound1Absent=native_absent,originalNativeM3HistoryUnchanged=True,sourceArchive='source',files=dict(files),paidContinuationAuthorized=False)
    index_pin=put(BASE/'closed-cohort-1/index.json',index)
    history={}
    for n in range(15):
        path=Path('.evidence')/('synthetic-native-history-'+str(n)+'.json');pin=put(path,dict(synthetic=True,index=n));history[str(path.relative_to('.evidence'))]=pin
    facts=dict(originalCalls=39,originalAttempts=36,previousCalls=39,previousAttempts=36,closedHistory=history)
    native_base=BASE/'closed-cohort-1'
    facts_pin=put(native_base/'native-minimax-history.log',facts);command_pin=put(native_base/'native-minimax-history-command.json',dict(exitCode=0,credentialsRead=False))
    seal=dict(schema=1,kind='CLOSED_FIRST_COHORT_HISTORY_SEAL',cohortIndexSHA256=index_pin,cohortIndex='index.json',nativeHistory=dict(facts='native-minimax-history.log',factsSHA256=facts_pin,command='native-minimax-history-command.json',commandSHA256=command_pin,closedHistory=history,closedHistorySHA256=hashlib.sha256(json.dumps(history,sort_keys=True,separators=(',',':')).encode()).hexdigest(),originalCalls=39,originalAttempts=36,previousCalls=39,previousAttempts=36,countsAreHistoricalNotAdditionalCampaignConsumption=True),nativeRound1NeverIssued=True,campaignActualCalls=40,campaignActualHTTPAttempts=40,campaignClaimedCalls=52,campaignClaimedHTTPAttempts=52,campaignWholeReservedMicrousd=1500000,remainingPaidScopeRequiresNewApproval=True)
    put(native_base/'seal.json',seal)
