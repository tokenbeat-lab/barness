"""Reviewable fixed3 freeze with separate historical2 USD policy metadata."""
import json
from pathlib import Path
import shutil
import subprocess
import sys
import recovery_contract as r
import recovery_history as h
import final_state as shared
c=r.c
run=shared.run

def path(value):
    if not value:raise RuntimeError('owned recovery3 output required')
    p=Path(value);p=p if p.is_absolute() else c.ROOT/p
    p=h.expand(str(p));parts=p.relative_to(c.common_root()).parts
    if len(parts)<3 or parts[:2]!=('.evidence','provider-sites'):raise RuntimeError('new owned .evidence/provider-sites output required')
    return p

def doc(p):return h.doc(p)
def pin(p):return h.pin(p)

def prepare(out):
    predecessor=h.facts();preserved=h.preserved();frozen=shared.freeze()
    out.mkdir(parents=True,exist_ok=False,mode=0o700)
    for leaf,value in [('predecessor.json',predecessor),('preserved-glm.json',preserved),('source-at-start.json',frozen)]:c.dump(out/leaf,value)
    if run(out,'go-version',['go','version'])!=0:raise RuntimeError('build tool identity unavailable')
    toolchain=dict(goVersion=(out/'go-version.log').read_text().strip(),pythonVersion=sys.version,executables={str(Path(shutil.which('go')).resolve()):c.digest(Path(shutil.which('go')).resolve()),str(Path(sys.executable).resolve()):c.digest(Path(sys.executable).resolve())})
    c.dump(out/'toolchain.json',c.redact(toolchain))
    if run(out,'regional-source',['go','run','./ai/live/cmd/regionalverify','-source'])!=0:raise RuntimeError('native current source unavailable')
    source=json.loads((out/'regional-source.log').read_text());c.dump(out/'regional-source.json',source)
    builds=[('live.test',['go','test','-c','-tags','live','-ldflags','-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource='+source['sourceSha256'],'-o',out/'live.test','./ai/live'])]
    packages=dict(regionalpromote='./ai/live/cmd/regionalpromote',regionalverify='./ai/live/cmd/regionalverify',minimaxverify='./ai/live/cmd/minimaxverify',glmformal='./ai/live/cmd/glmformal',jevformal='./ai/live/cmd/jevformal',supportmatrix='./ai/live/cmd/supportmatrix',chineseeval='./ai/examples/chineseeval/cmd/chineseeval',visualeval='./ai/examples/visualeval',releasegate='./ai/release/cmd/releasegate')
    builds += [(n,['go','build','-o',out/n,p]) for n,p in packages.items()]
    for name,argv in builds:
        if run(out,'build-'+name,argv,timeout=600)!=0:raise RuntimeError('keyless current reviewed build failed')
    c.dump(out/'binaries.json',{name:c.digest(out/name) for name,_ in builds})
    import recovery_native
    templates=recovery_native.templates(out)
    c.dump(out/'component-templates.json',templates)
    c.dump(out/'plan.json',c.redact(r.inventory(out)))
    prepared=dict(schema=1,round=r.ROUND,validationRound=r.ROUND,nativeRenewalId=r.ROUND,credentialsRead=False,sourceCommit=frozen['sourceCommit'],secondCohortSealSHA256=predecessor['secondCohortSealSHA256'],originalCampaignApprovalSHA256=predecessor['authorizationSHA256'],**r.policy())
    for leaf,field in PINNED.items():prepared[field]=pin(out/leaf)
    c.dump(out/'prepared.json',prepared)
    c.dump(out/'prerelease-source-review.template.json',prepared|dict(decision='PENDING',reviewedBy='',standardsReview=dict(path='',sha256=''),specReview=dict(path='',sha256=''),authorizedRoutes=[]))
    auth=prepared|dict(decision='PENDING',requestText='',requestReference='',maxCalls=506,maxAttempts=518,campaignMaxCalls=962,campaignMaxAttempts=1010,newReservedMicrousd=19200000,retainedReservedMicrousd=10700000,maxUSD=30,automaticProcessReruns=False,predecessor=predecessor,eligibilityAuthorizationSHA256=predecessor['authorizationSHA256'],raceAuthorizationSHA256=predecessor['authorizationSHA256'],processReservationsMicrousd=r.RESERVES,usdLedgers={n:c.redact(j['usdLedger']) for n,j in r.jobs(out).items()},networkRoutes={n:j['networkRoute'] for n,j in r.jobs(out).items()},standardCredentialScopes=predecessor['credentialScopes'],componentAuthorizationSHA256={},actualHumanApprovalSHA256='')
    c.dump(out/'execution-authorization.template.json',auth)
    c.dump(out/'recovery3-authorization.template.json',{k:v for k,v in auth.items() if k not in ('standardCredentialScopes','componentAuthorizationSHA256','networkRoutes','predecessor')})
    guard(out);return prepared

PINNED={'source-at-start.json':'sourceManifestSHA256','regional-source.json':'nativeSourceManifestSHA256','predecessor.json':'predecessorSHA256','preserved-glm.json':'preservedGLMSHA256','toolchain.json':'toolchainSHA256','binaries.json':'binariesManifestSHA256','component-templates.json':'componentTemplatesSHA256','plan.json':'planSHA256'}
def guard(out):
    prepared=doc(out/'prepared.json')
    if any(prepared.get(k)!=v for k,v in dict(round=r.ROUND,validationRound=r.ROUND,nativeRenewalId=r.ROUND,**r.policy()).items()):raise RuntimeError('fixed3 authority and unchanged2 spending policy required')
    for leaf,field in PINNED.items():
        if pin(out/leaf)!=prepared[field]:raise RuntimeError('frozen recovery input changed: '+leaf)
    if doc(out/'predecessor.json')!=h.facts() or doc(out/'preserved-glm.json')!=h.preserved():raise RuntimeError('immutable prior facts changed')
    frozen=doc(out/'source-at-start.json')
    if frozen['files']!=shared.sources() or frozen['sourceCommit']!=c.git('rev-parse','HEAD') or prepared['sourceCommit']!=frozen['sourceCommit']:raise RuntimeError('complete source commit/set changed')
    for name,sha in doc(out/'binaries.json').items():
        if pin(out/name)!=sha:raise RuntimeError('current frozen binary changed')
    for executable,sha in doc(out/'toolchain.json')['executables'].items():
        value=executable
        for label,p in (('$REPO',c.ROOT),('$COMMON',c.common_root()),('$HOME',Path.home())):value=value.replace(label,str(p))
        if c.digest(Path(value))!=sha:raise RuntimeError('frozen build executable changed')
    for leaf,sha in doc(out/'component-templates.json').items():
        if pin(out/leaf)!=sha:raise RuntimeError('component template changed')
    current=json.loads(subprocess.check_output([out/'regionalverify','-source'],cwd=c.ROOT,env=c.ENV,timeout=30))
    if current!=doc(out/'regional-source.json') or doc(out/'plan.json')!=c.redact(r.inventory(out)):raise RuntimeError('native source or reconstructed fixed24 plan changed')
    return prepared

def approved(out,job):
    prepared=guard(out);review=doc(out/'prerelease-source-review.json');auth=doc(out/'execution-authorization.json');human=h.actual_approval(prepared)
    for value in (review,auth):
        if value.get('decision')!='APPROVED' or any(value.get(k)!=v for k,v in prepared.items()):raise RuntimeError('new exact independent review and human source approval required')
    if not review.get('reviewedBy') or set(review.get('authorizedRoutes',[]))!=set(r.jobs(out)):raise RuntimeError('all24 exact independently reviewed routes required')
    for field in ('standardsReview','specReview'):
        proof=review[field]
        if pin(path(proof['path']))!=proof['sha256']:raise RuntimeError('retained independent review changed')
    prior=doc(out/'predecessor.json')
    for key,value in dict(maxCalls=506,maxAttempts=518,campaignMaxCalls=962,campaignMaxAttempts=1010,newReservedMicrousd=19200000,retainedReservedMicrousd=10700000,maxUSD=30,automaticProcessReruns=False,predecessor=prior,eligibilityAuthorizationSHA256=prior['authorizationSHA256'],raceAuthorizationSHA256=prior['authorizationSHA256'],processReservationsMicrousd=r.RESERVES,usdLedgers={n:c.redact(j['usdLedger']) for n,j in r.jobs(out).items()},networkRoutes={n:j['networkRoute'] for n,j in r.jobs(out).items()},standardCredentialScopes=prior['credentialScopes'],actualHumanApprovalSHA256=pin(h.HUMAN),requestText=human['requestText'],requestReference=human['requestReference']).items():
        if auth.get(key)!=value:raise RuntimeError('fixed exact new human scope/budget/proxy/prefix required: '+key)
    scope=auth['standardCredentialScopes'][job['sourceVariable']]
    if scope.get('provider')!=job['provider'] or scope.get('purpose')!='standard' or scope.get('confirmed') is not True:raise RuntimeError('original own standard service qualification required')
    return prepared,auth,scope,job['reservedMicrousd']

def checks(out):
    guard(out);codes={};pins={}
    for name,argv,timeout in [('vet',['go','vet','./...'],600),('vet-live',['go','vet','-tags','live','./...'],600),('no-key-live',['go','test','-tags','live','-count=1','-timeout=20m','./ai/live'],1260)]:
        codes[name]=run(out,name,argv,timeout=timeout,env={'BARNESS_AI_EVIDENCE_DIR':str(out/'no-key-live')});pins[name]=pin(out/(name+'-command.json'))
        if codes[name]!=0:break
    result=dict(passed=len(codes)==3 and all(v==0 for v in codes.values()),commands=codes,commandReceipts=pins,credentialsRead=False);c.dump(out/'checks.json',result);guard(out)
    if not result['passed']:raise RuntimeError('keyless checks failed')
    return result
