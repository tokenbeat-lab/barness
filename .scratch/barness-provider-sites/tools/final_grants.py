"""Only current native component templates and pre-key checks for this round."""
import datetime as dt
import json
from pathlib import Path
import final_contract as r
import final_state as s
c=r.c


def templates(out, phase):
    result={}
    catalog=c.doc(c.ROOT/'ai/release/catalog-snapshot.json')
    for leaf,value in r.old.authorizations(out,catalog['hash']).items():
        value.update(accountAlias='<actual-own-account>',regionAlias='<actual-own-region>')
        target=out/(leaf+'.template.json');c.dump(target,value);result[str(target.relative_to(out))]=c.digest(target)
    for name,job in r.jobs(out).items():
        if job['phase']!=phase or 'combo' not in job:
            continue
        label=name+'-grant-template'
        if s.run(out,label,[out/'regionalverify','-grant-template',job['combo'],'-stage',job['stage']],timeout=120)!=0:
            raise RuntimeError('current native own regional grant template failed')
        value=json.loads((out/(label+'.log')).read_text())
        target=out/'grants'/(name+'.template.json');c.dump(target,value);result[str(target.relative_to(out))]=c.digest(target)
    return result


def require_component(out,job,auth,scope):
    name=job['name']
    if 'combo' in job:
        path=Path(job['environment']['BARNESS_REGIONAL_GRANT'])
        template=c.doc(out/'grants'/(name+'.template.json'));value=c.doc(path)
        mutable={'id','accountAlias','regionAlias','expiresAt'}
        if (set(value)!=set(template) or {k:v for k,v in value.items() if k not in mutable}!={k:v for k,v in template.items() if k not in mutable} or
            value.get('mode')!='live' or value.get('accountAlias')!=scope['accountAlias'] or value.get('regionAlias')!=scope['regionAlias'] or
            not value.get('id','').startswith('ps-final-1-')):
            raise RuntimeError('new exact own native grant/scope required before selected key')
        expires=dt.datetime.fromisoformat(value['expiresAt'].replace('Z','+00:00'));now=dt.datetime.now(dt.timezone.utc)
        if expires<=now or expires-now>dt.timedelta(hours=24):
            raise RuntimeError('new regional grant must expire within24 hours')
    elif name=='protocol114' or name.startswith('visual-'):
        path=out/('protocol-authorization.json' if name=='protocol114' else name+'-authorization.json')
        value=c.doc(path)
        if (value.get('accountAlias')!=scope['accountAlias'] or value.get('regionAlias')!=scope['regionAlias'] or
            value.get('dedicatedCredential') is not True or value.get('maxCalls')!=12 or value.get('maxAttempts')!=12):
            raise RuntimeError('fresh dedicated own protocol/effect finite authorization required')
    else:
        return
    if c.digest(path)!=auth.get('componentAuthorizationSHA256',{}).get(name):
        raise RuntimeError('reviewed human component grant bytes must be pinned before key')


def native_preflight(out,job,record):
    if not job['name'].startswith('mm-'):
        return record
    name=job['name']+'-native-preflight'
    if s.run(out,name,[out/'minimaxverify','-renewal-preflight='+job['route']],timeout=120)!=0:
        raise RuntimeError('new canonical native renewal authority/source/history refused before key')
    facts=json.loads((out/(name+'.log')).read_text())
    if facts.get('route')!=job['route'] or facts.get('facts',{}).get('renewalId')!=r.ROUND:
        raise RuntimeError('native current own renewal preflight identity invalid')
    return record|dict(nativePreflightReceiptSha256=c.digest(out/(name+'-command.json')),
        renewalId=r.ROUND,renewalGrantSha256=facts['grantSha256'])
