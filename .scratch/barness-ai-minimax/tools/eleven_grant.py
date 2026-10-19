"""Canonical native M3 grant, issued once after the fresh remaining-phase review."""
import os
import paid_core as c
import routed_evidence as e
import renewal_common as u
import eleven_contract as r

def base():return c.common_root()/'.evidence/minimax-renewals'/r.ROUND

def capture(out,phase):
 if phase=='glm':c.dump(out/'native-renewal-facts.json',dict(state='NOT_ISSUED'));return
 if c.execute(out,'native-renewal-facts',[out/'minimaxverify','-renewal-source'])!=0:raise RuntimeError('native keyless renewal facts failed')
 value=e.document(u.file_bytes(out/'native-renewal-facts.log'))
 c.dump(out/'native-renewal-facts.json',value)
 validate_facts(out)

def validate_facts(out):
 value=u.json_doc(out/'native-renewal-facts.json');frozen=u.json_doc(out/'source-at-start.json')
 required={'renewalId','commonRootSha256','sourceCommit','executionSourceSha256','originalLedgerSha256','originalLedgerTailSha256','originalCalls','originalAttempts'}
 if (set(value)!=required or value['renewalId']!=r.ROUND or value['commonRootSha256']!=u.sha(str(c.common_root()).encode()) or value['sourceCommit']!=frozen['sourceCommit'] or
     any(u.HEX.fullmatch(value.get(n,'')) is None for n in ('executionSourceSha256','originalLedgerSha256','originalLedgerTailSha256')) or
     type(value['originalCalls']) is not int or value['originalCalls']!=39 or type(value['originalAttempts']) is not int or value['originalAttempts']!=36 or
     value['originalLedgerSha256']!=u.json_doc(out/'history.json')['originalFormalLedgerSHA256']):raise RuntimeError('exact native current-source/original39/36 facts required')
 return value

def expected(out):
 facts=validate_facts(out)
 return dict(schema=1,**facts,decision='APPROVED',maxCalls=60,maxAttempts=72,perIdentityCalls=20,perIdentityAttempts=24,maxOutputTokens=4096,maxRequestedOutput=40000,maxReservedInput=120000,humanApprovalSha256=c.digest(base()/'human-approval.json'),sourceReviewSha256=c.digest(base()/'source-review.json'),sourceManifestSha256=c.digest(out/'source-at-start.json'),binariesManifestSha256=c.digest(out/'binaries.json'),planSha256=c.digest(out/'plan.json'))

def create(out):
 from eleven_validation import approved,human
 if c.doc(out/'plan.json')['phase']!='remaining':raise RuntimeError('native grant requires fresh remaining phase')
 for job in r.jobs(out,'remaining').values():approved(out,job)
 human(out)
 raw=u.file_bytes(out/'prerelease-source-review.json')
 with (base()/'source-review.json').open('xb') as stream:
  stream.write(raw);stream.flush();os.fsync(stream.fileno())
 grant=expected(out);c.dump(base()/'grant.json',grant,exclusive=True)
 return dict(credentialsRead=False,renewalId=r.ROUND,grantSHA256=c.digest(base()/'grant.json'))

def require(out):
 with e.Root(c.common_root()) as reader:
  actual=e.document(reader.read((base()/'grant.json').relative_to(c.common_root()).as_posix()))
  review=reader.read((base()/'source-review.json').relative_to(c.common_root()).as_posix())
 if any(type(actual.get(n)) is not int for n in ('schema','originalCalls','originalAttempts','maxCalls','maxAttempts','perIdentityCalls','perIdentityAttempts','maxOutputTokens','maxRequestedOutput','maxReservedInput')):raise RuntimeError('exact integer native grant caps required before selected credential')
 if actual!=expected(out) or review!=u.file_bytes(out/'prerelease-source-review.json'):raise RuntimeError('canonical native grant/source review changed')
 return c.digest(base()/'grant.json')
