"""Fixed task-10 collection inventory; executable argv never comes from a receipt."""
from pathlib import Path

GENERIC = (
 ('openai-responses','openai','OPENAI_KEY',32),('openai-chat','openai','OPENAI_KEY',32),
 ('anthropic-messages','anthropic','ANTHROPIC_KEY',32),('google-gemini','google','GEMINI_KEY',32),
 ('deepseek-responses','deepseek','DEEPSEEK_KEY',32),('deepseek-chat','deepseek','DEEPSEEK_KEY',32),
 ('openai-images','openai','OPENAI_KEY',4),('google-interactions-image','google','GEMINI_KEY',4),
 ('openai-decisions','openai','OPENAI_KEY',12),('typesafe-classifier','typesafe','TYPESAFE_KEY',6))
LIVE_IDENTITIES = [r[0] for r in GENERIC]+['typesafe-classifier-jev-1.14.0']+[
 'zhipu-'+r for r in ('chat','responses','messages')]+['minimax-cn-'+r for r in ('chat','responses','messages')]
PROVIDER_CAPS = {'openai':116,'anthropic':32,'google':36,'deepseek':64,'typesafe':54,'zhipu':72}


def jobs(root, out, common):
    result={}
    def add(name,provider,key,injected,argv,env,calls,attempts=None,timeout=2100,cwd=None,**extra):
        evidence=env.get('BARNESS_AI_EVIDENCE_DIR')
        if evidence is None and '-out' in argv:evidence=str(argv[argv.index('-out')+1])
        result[name]=dict(name=name,provider=provider,sourceVariable=key,injectedCredential=injected,evidenceRoot=evidence,
            argv=list(map(str,argv)),environment=env,maxCalls=calls,maxAttempts=attempts or calls,
            timeoutSeconds=timeout,cwd=str(cwd or root),**extra)
    live=out/'live.test'
    for route,provider,key,cap in GENERIC:
        env={'BARNESS_AI_LIVE':'1','BARNESS_AI_LIVE_COMBO':route,
             'BARNESS_AI_LIVE_ACCOUNT_ALIAS':provider+'-local-unverified@region-unverified',
             'BARNESS_AI_EVIDENCE_DIR':str(out/'live'/route)}
        selector='^TestLive$/^'+route+'$';limit='30m';timeout=2100
        if cap in (4,6): limit,timeout='10m',660
        if route=='openai-decisions':
            selector,limit,timeout='^TestFormalDecisions$','180s',210
            env['BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT']='1'
        add(route,provider,key,'BARNESS_AI_LIVE_KEY_'+route.upper().replace('-','_'),
            [live,'-test.run='+selector,'-test.timeout='+limit],env,cap,timeout=timeout,cwd=root/'ai/live')
    add('protocol114','typesafe','TYPESAFE_KEY','BARNESS_MM_KEY',
        [live,'-test.run=^TestMultimodalFormal$','-test.timeout=180s'],
        {'BARNESS_MM_ROUTE':'jev','BARNESS_MM_FORMAL':'1','BARNESS_MM_LIVE':'1',
         'BARNESS_MM_AUTHORIZATION':str(out/'protocol-authorization.json'),
         'BARNESS_AI_EVIDENCE_DIR':str(out/'protocol114')},12,timeout=210,cwd=root/'ai/live')
    for route,key in (('jev','TYPESAFE_KEY'),('luna','OPENAI_KEY')):
        provider='typesafe' if route=='jev' else 'openai'
        add('text-'+route,provider,key,'BARNESS_AI_CHINESE_EVAL_'+('TYPESAFE' if route=='jev' else 'OPENAI')+'_KEY',
            [out/'chineseeval','-route',route,'-live','-out',out/('text-'+route)],
            {'BARNESS_AI_CHINESE_EVAL':'1','BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS':
             provider+'-local-unverified@region-unverified'},24,timeout=630)
    for route,key in (('jev','TYPESAFE_KEY'),('decisions','OPENAI_KEY')):
        add('visual-'+route,'typesafe' if route=='jev' else 'openai',key,'BARNESS_VISUAL_KEY',
            [out/'visualeval','-formal','-mode','live','-route',route,'-entry','plain',
             '-authorization',out/('visual-'+route+'-authorization.json'),'-out',out/('visual-'+route)],{},12,timeout=270)
    for route in ('chat','responses','messages'):
        add('glm-'+route,'zhipu','GLM_KEY','BARNESS_GLM_KEY',
            [live,'-test.run=^TestGLMFormal$','-test.timeout=1800s','-test.v'],
            {'BARNESS_AI_LIVE':'1','BARNESS_GLM_ROUTE':route,'BARNESS_GLM_FORMAL':'1',
             'BARNESS_AI_EVIDENCE_DIR':str(out/('glm-'+route))},24,32,timeout=1800,cwd=root/'ai/live',plannedCalls=22)
    for short,model,purpose,key in (('m3','MiniMax-M3','standard','MINIMAX_KEY'),
                                   ('m31','MiniMax-M3.1-Flash-Preview','mplan','MINIMAX_MPLAN_KEY')):
        for phase in ('research','candidate','formal'):
            if short=='m31' and phase=='formal':continue
            stage='formal' if phase=='formal' else 'candidate'
            for route in ('chat','responses','messages'):
                name='mm-'+phase+'-'+short+'-'+route
                env={'BARNESS_AI_LIVE':'1','BARNESS_MINIMAX_ROUTE':route,'BARNESS_MINIMAX_MODEL':model,
                     'BARNESS_MINIMAX_PURPOSE':purpose,'BARNESS_MINIMAX_ENTITLEMENT':'confirmed',
                     'BARNESS_MINIMAX_ACCOUNT_ALIAS':'minimax-cn-'+purpose,
                     'BARNESS_MINIMAX_REGION_ALIAS':'cn','BARNESS_AI_EVIDENCE_DIR':str(out/name),
                     'BARNESS_MINIMAX_LEDGER':str(common/'.evidence/minimax-stages'/(stage+'.jsonl'))}
                if phase=='research':env['BARNESS_MINIMAX_RESEARCH']='1'
                if phase=='candidate':env['BARNESS_MINIMAX_PREDECESSOR_ROOT']=str(common/'.evidence/minimax-research-captures'/('mm-research-'+short+'-'+route))
                if phase=='formal':env['BARNESS_MINIMAX_FORMAL']='1'
                selector='^TestMiniMax'+phase.title()+'$'
                add(name,'minimax',key,'BARNESS_MINIMAX_KEY',
                    [live,'-test.run='+selector,'-test.timeout=1800s','-test.v'],env,20,24,
                    timeout=1800,cwd=root/'ai/live',phase=phase,stage=stage,model=model,purpose=purpose,
                    route=route,plannedCalls=3 if phase=='research' else 13,maxOutputTokens=4096,
                    maxRequestedOutput=40000,maxReservedInput=120000,maxToolReruns=0 if phase=='research' else 1,maxRequestReruns=0 if phase=='research' else 3)
    return result


def authorizations(out, catalog_hash):
    result={}
    for route,model,provider,service in (
        ('decisions','gpt-6-luna','openai','https://api.openai.com/v1/decisions'),
        ('jev','jev-1.14.0','typesafe','https://api.typesafe.ai/v1/systemone')):
        proof=dict(route=route,model=model,service=service,catalogHash=catalog_hash,
            accountAlias=provider+'-local-unverified',regionAlias='unverified',
            reference='barness-ai-minimax-ticket10',dedicatedCredential=True,maxCalls=12,maxAttempts=12)
        result['visual-'+route+'-authorization.json']=proof
        if route=='jev':
            protocol={k:v for k,v in proof.items() if k not in ('catalogHash','reference')}
            protocol['authorizationReference']=proof['reference']
            result['protocol-authorization.json']=protocol
    return result


def unique(base, leaf):
    matches=list(base.glob('*/'+leaf))
    if len(matches)!=1:raise RuntimeError('unique current bundle missing: '+str(base))
    return matches[0].parent


def report_path(out, job):
    n=job['name']
    if n.startswith('visual-'):return out/n/'report.json'
    if n.startswith('text-'):return unique(out/n,'evaluation-report.json')/'evaluation-report.json'
    if n=='protocol114':return unique(out/n,'multimodal-report.json')/'multimodal-report.json'
    if n.startswith('mm-'):return unique(out/n,'minimax-report.json')/'minimax-report.json'
    if n.startswith('glm-'):return unique(out/n,'formal-glm-report.json')/'formal-glm-report.json'
    return unique(out/'live'/n,'live-report.json')/'live-report.json'


def live_inputs(out):
    result={n:unique(out/'live'/n,'live-report.json') for n,_,_,_ in GENERIC}
    result['typesafe-classifier-jev-1.14.0']=out/'support'
    for r in ('chat','responses','messages'):
        result['zhipu-'+r]=unique(out/('glm-'+r),'live-report.json')
        result['minimax-cn-'+r]=unique(out/('mm-formal-m3-'+r),'live-report.json')
    for path in result.values():
        if not (path/'live-report.json').is_file():raise RuntimeError('own live report missing')
    return [result[n] for n in LIVE_IDENTITIES]


def keyless_plan(root,out):
    commands=[]
    for route in ('jev','luna'):
        commands.append([str(out/'chineseeval'),'-route',route,'-verify',str(unique(out/('text-'+route),'manifest.json'))])
    for route in ('jev','decisions'):
        commands.append([str(out/'visualeval'),'-verify',str(out/('visual-'+route)),'-require-live'])
    commands.append(['go','run','./ai/live/cmd/jevformal','-protocol',str(unique(out/'protocol114','manifest.json')),
                     '-effects',str(out/'visual-jev'),'-out',str(out/'support'),'-require-live'])
    commands.append(['go','run','./ai/live/cmd/jevformal','-verify',str(out/'support'),'-require-live'])
    for r in ('chat','responses','messages'):
        commands.append(['go','run','./ai/live/cmd/minimaxverify','-formal',
                         str(unique(out/('mm-formal-m3-'+r),'minimax-report.json'))])
    inputs=list(map(str,live_inputs(out)))
    merge=['go','run','./ai/live/cmd/supportmatrix',*inputs]
    gate=['go','run','./ai/release/cmd/releasegate','-out',str(out/'gate'),
          '-evaluation-jev',str(unique(out/'text-jev','manifest.json')),
          '-evaluation-luna',str(unique(out/'text-luna','manifest.json')),
          '-visual-evaluation-jev',str(out/'visual-jev'),'-visual-evaluation-decisions',str(out/'visual-decisions')]
    for path in inputs:gate+=['-live',path]
    return dict(cwd=str(root),credentialsRead=False,verify=commands,merge=merge,gate=gate,
                liveIdentities=LIVE_IDENTITIES,liveInputs=inputs,gateOuterTimeoutSeconds=6600)
