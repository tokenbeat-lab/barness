"""Fixed service routing and immutable prior accounting for ticket13."""
from pathlib import Path
import re
import paid_core as c
import paid_plan as p
import paid

ROUND='routed-validation-1'
PROXY='http://127.0.0.1:7890'
REQUEST='继续完成缺口部分的验证；其中，中国区走直连，国际站通过本地代理（127.0.0.1:7890）进行访问'
ACCOUNTING_SCOPE='current MiniMax-release renewal: original GLM66 + corrective GLM66 + legacy168 + routed-validation-1; earlier standalone GLM-task history remains separate and immutable'
BUDGET=dict(priorCalls=300,priorAttempts=300,additionalMaxCalls=374,
    additionalMaxAttempts=398,totalMaxCalls=674,totalMaxAttempts=698)
PRIOR={'openai-responses':15,'openai-chat':17,'anthropic-messages':11,
    'google-gemini':11,'deepseek-responses':10,'deepseek-chat':13,
    'openai-images':2,'google-interactions-image':4,'openai-decisions':9,
    'typesafe-classifier':3,'protocol114':1,'text-jev':24,'text-luna':24,
    'visual-jev':12,'visual-decisions':12,
    **{prefix+r:22 for prefix in ('glm-','glm-corrective-1-') for r in ('chat','responses','messages')}}


def prior():
    result={}
    for name,count in PRIOR.items():
        path=paid.marker(name);record=c.doc(path);budget=record.get('budget',{})
        if (record.get('state')!='COMPLETED' or record.get('route')!=name or
            record.get('credentialsRead') is not True or record.get('launcherAudit',{}).get('passed') is not True or
            budget.get('state')!='REPORTED' or budget.get('status') not in ('PASS','FAIL') or
            type(budget.get('calls')) is not int or budget['calls']!=count or
            type(budget.get('attempts')) is not int or budget['attempts']!=count or
            re.fullmatch('[a-f0-9]{40}',record.get('sourceCommit','')) is None or
            re.fullmatch('[a-f0-9]{64}',record.get('reportedFileSha256','')) is None):
            raise RuntimeError('immutable prior300/300 complete receipts required; unknown cannot reset')
        result[name]=dict(receiptSHA256=c.digest(path),sourceCommit=record['sourceCommit'],
            reportSHA256=record['reportedFileSha256'],calls=count,attempts=count,status=budget['status'])
    return result


def jobs(out,phase=None):
    original=p.jobs(c.ROOT,out,c.common_root());result={}
    names=[*('glm-'+r for r in ('chat','responses','messages')),
        *(r[0] for r in p.GENERIC),'protocol114','text-jev','text-luna','visual-jev','visual-decisions']
    for base in names:
        if phase is not None and (base.startswith('glm-'))!=(phase=='glm'):continue
        job=original[base];name=ROUND+'-'+base
        job.update(name=name,baseName=base,validationRound=ROUND)
        if base.startswith('glm-'):job['glmRoute']=base.removeprefix('glm-')
        env=job['environment']
        direct=job['provider'] in ('zhipu','deepseek')
        job['networkRoute']='china-direct' if direct else 'international-loopback-proxy'
        if not direct:
            if base.startswith(('text-','visual-')):job['argv']+=['-proxy',PROXY]
            else:env['BARNESS_AI_LIVE_PROXY']=PROXY
        result[name]=job
    return result


def report_path(out,job):
    return p.report_path(out,job|dict(name=job['baseName']))
