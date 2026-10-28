// Analysis-only exported access to current native read-only checks. It adds no
// accepted historical path to production verifiers and changes no old packet.
package minimaxlive

import (
 "fmt"
 "path/filepath"

 "github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func ReviewFailedThinkingCapture(dir string,producer ThinkingFileRef) map[string]any {
 out:=map[string]any{"bundle":dir,"vendorCalls":0,"originalsModified":false,"qualificationReusable":false}
 r,err:=Load(dir)
 if err!=nil {out["loadError"]=err.Error();return out}
 out["originalOutcome"]=r.Outcome;out["originalStage"]=r.Stage;out["originalMode"]=r.Mode;out["originalComplete"]=r.Complete
 check:=func(name string,e error) {out[name+"Pass"]=e==nil;if e!=nil {out[name+"Error"]=e.Error()}}
 current,e:=ExecutionSource();if e==nil {e=evidence.VerifyRouteSource(r.ExecutionSource,current,"minimax",string(r.Identity.API))};check("routeSource",e)
 check("originalAuthority",verifyProcessAuthorization(dir,r))
 sdk,e:=SDKVersion(r.Identity);if e==nil && sdk!=r.SDK {e=fmt.Errorf("SDK changed")};check("sdk",e)
 b,e:=AssessThinkingBilling(dir,r);check("billing",e);out["currentBillingAssessment"]=b
 if len(r.Calls)==1 {check("currentWireReplay",verifyThinkingReplay(dir,r,r.Calls[0]));check("budgetUsageResources",verifyThinkingBudget(dir,r))}
 _,e=VerifyThinkingSupplement(dir,false);check("nativeQualification",e)
 _,e=VerifyThinkingProducer(dir,producer,r,false);check("successfulSupplementProducer",e)
 out["semanticReplayReusable"]=out["routeSourcePass"]==true && out["originalAuthorityPass"]==true && out["sdkPass"]==true && out["billingPass"]==true && out["currentWireReplayPass"]==true && out["budgetUsageResourcesPass"]==true
 out["regressionRawFixture"]=filepath.Join(dir,r.Calls[0].CaseID,"response.sse")
 out["decision"]="RAW_REPLAY_APPLICABLE_QUALIFICATION_BLOCKED_BY_ORIGINAL_FAIL_AND_SEAL_GAP"
 return out
}
