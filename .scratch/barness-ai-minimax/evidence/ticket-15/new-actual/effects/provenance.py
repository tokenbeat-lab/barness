"""Independent custody checks for one already completed, explicitly approved effects phase."""
SOURCE='7931b163f983a8b362374163b7949d31e68bbec4';ROUND='release-repair-validation-1'
BUDGET=dict(priorCalls=545,priorAttempts=545,additionalMaxCalls=108,additionalMaxAttempts=132,totalMaxCalls=653,totalMaxAttempts=677)
HUMAN='request_user_input_async/call_4440aa130a7c45eda945dcb7589a3291/0'
COMMANDS={'vet':['go','vet','./...'],'vet-live':['go','vet','-tags','live','./...'],'no-key-live':['go','test','-tags','live','-count=1','-timeout=20m','./ai/live'],'no-key-visual-proxy':['go','test','./ai/e2e','-run','^TestVisualEvaluationProxyProcess$','-count=1','-timeout=5m'],'no-key-release-contracts':['go','test','./ai/e2e','-count=1','-run','^(TestMiniMaxAdditiveReleaseContract|TestGLMReleaseTraceContract|TestReleaseRecordedBundleDiagnosticProcess)$'],'no-key-release-package':['go','test','./ai/internal/testkit/release','-count=1']}

def authority(b,m):
 doc=lambda leaf:b.doc(m['frozen/'+leaf]);source=doc('source-at-start.json');plan=doc('plan.json');review=doc('prerelease-source-review.json');auth=doc('execution-authorization.json');human=doc('human-additional-approval.json');checks=doc('checks.json')
 pins={'sourceCommit':SOURCE,**{field:b.sha(m['frozen/'+leaf]) for field,leaf in [('sourceManifestSHA256','source-at-start.json'),('binariesManifestSHA256','binaries.json'),('planSHA256','plan.json'),('historySHA256','history.json'),('reuseSHA256','reuse.json'),('scopeSHA256','scope.json')]}}
 names={ROUND+'-'+r for r in ('protocol114','visual-jev','visual-decisions')}
 if plan['collection']!={'rootSHA256':'bf574a3ed52ff02819aa91b7a137d9abac7c4ba7333701e584cde352ac83aa72','relativeRoot':'.evidence/minimax-release-repair-effects-1'}:raise ValueError('fixed own original collection differs')
 if source['sourceCommit']!=SOURCE or len(source['files'])!=1056 or plan['round']!=ROUND or plan['phase']!='effects' or plan['budget']!=BUDGET or set(plan['routes'])!=names or plan['automaticProcessReruns'] is not False:raise ValueError('fixed own exact source/phase/plan differs')
 if review['decision']!='APPROVED' or review['billingAuthorized'] is not False or set(review['authorizedRoutes'])!=names or review['credentialScopes']!={'TYPESAFE_KEY':ROUND+'/international-loopback-proxy','OPENAI_KEY':ROUND+'/international-loopback-proxy'} or any(review.get(k)!=v or auth.get(k)!=v for k,v in pins.items()):raise ValueError('independent exact-source authority pins differ')
 if auth['decision']!='APPROVED' or auth['authority']!='explicit-additional-budget-reply' or auth['round']!=ROUND or auth['phase']!='effects' or auth['budget']!=BUDGET or auth['approvedAdditionalMaxCalls']!=108 or auth['approvedAdditionalMaxAttempts']!=132 or auth['approvalProposalSHA256']!=plan['approvalProposalSHA256'] or auth['checksSHA256']!=b.sha(m['frozen/checks.json']) or auth['sourceReviewSHA256']!=b.sha(m['frozen/prerelease-source-review.json']):raise ValueError('new finite human approval differs')
 if human['questionItemId']!=HUMAN or auth['requestReference']!=HUMAN or human['scope']!=ROUND or human['automaticNextRound'] is not False or human['answer']!='批准新增 108/132，继续六项验收（推荐）' or auth['requestText']!=human['question']+'\n用户答复：'+human['answer']:raise ValueError('original new human question/reply provenance differs')
 if checks['passed'] is not True or checks['credentialsRead'] is not False or checks['commands']!={n:0 for n in COMMANDS} or set(checks['commandReceipts'])!=set(COMMANDS) or review['checksSHA256']!=b.sha(m['frozen/checks.json']):raise ValueError('six actual keyless checks differ')
 for name,argv in COMMANDS.items():
  raw=m['frozen/'+name+'-command.json'];v=b.doc(raw)
  if checks['commandReceipts'][name]!=b.sha(raw) or v['state']!='COMPLETED' or type(v['exitCode']) is not int or v['exitCode']!=0 or v['argv']!=argv or v['cwd']!='$REPO' or v['credentialsRead'] is not False or v['credentialsInherited'] is not False or v['sourceVariable'] is not None or v['injectedCredential'] is not None or v['launcherAudit']!={'passed':True,'findings':[]} or any('KEY' in n or 'PROXY' in n for n in v['environmentKeys']):raise ValueError('original numeric keyless receipt differs')
 for axis in ('standards','spec'):
  row=review['independentReports'][axis];raw=m['frozen/'+row['file']];v=b.doc(raw)
  if b.sha(raw)!=row['sha256'] or v['findings']!=[]:raise ValueError('original independent source review differs')
  table=v['pins'] if axis=='standards' else v['pinsSHA256']
  if any(b.sha(m['frozen/'+name])!=sha for name,sha in table.items()):raise ValueError('independent reviewer file pins differ')
  if (v['sourceCommit'] if axis=='standards' else v['reviewedSourceCommit'])!=SOURCE:raise ValueError('independent reviewed source differs')
 history=doc('history.json');reuse=doc('reuse.json');scope=doc('scope.json')
 if len(history['receipts'])!=39 or history['known']!={'calls':545,'attempts':545} or sum(v['calls'] for v in history['receipts'].values())!=545 or sum(v['attempts'] for v in history['receipts'].values())!=545 or len(reuse['inputs'])!=15 or len(reuse['files'])!=990 or scope['matrixSHA256']!='7fbe72c37324a5700b0cff9416bf2a05543167ca1eed88186741669e39c6190f' or any(plan[k]!=v for k,v in [('history',history),('reuse',reuse),('scope',scope)]):raise ValueError('original closed39/reusable990 references differ')
 return source,plan,dict(sourcePins=pins,oldClosedReceipts=39,oldKnownCalls=545,oldKnownAttempts=545,reusedNativeFilesReferenced=990,reusedNativeTreesCopied=0,humanQuestionItemId=HUMAN,sourceReviewBillingAuthority=False,checks={n:0 for n in COMMANDS})
