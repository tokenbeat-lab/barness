// Analysis only: expose existing checks for the fixed failed measurement.
// This file is never installed in the product or used to grant qualification.
package minimaxlive

import (
 "encoding/json"
 "fmt"
 "path/filepath"
 "reflect"
 "github.com/tokenbeat-lab/barness/ai"
 "github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)
func ReadinessAudit(dir string) (map[string]any,error) {
 r,e:=Load(dir);if e!=nil{return nil,e}
 checks:=map[string]string{}
 add:=func(n string,e error){if e!=nil{checks[n]=e.Error()}else{checks[n]="PASS"}}
 raw,e:=ReadFile(filepath.Join(dir,SealFile),2<<20);var seal Seal
 if e==nil{e=StrictJSON(raw,&seal)};inventory,ie:=Inventory(dir)
 if e==nil && (ie!=nil||seal.Schema!=1||!reflect.DeepEqual(inventory,seal.Files)){e=fmt.Errorf("sealed inventory differs")};add("seal",e)
 add("artifacts",verifyArtifacts(dir,r,inventory));add("originalAuthority",verifyProcessAuthorization(dir,r));add("originalLedger",VerifyLedger(dir,r))
 current,e:=ExecutionSource();if e==nil{e=evidence.VerifyRouteSource(r.ExecutionSource,current,"minimax",string(r.Identity.API))};add("currentSource",e)
 sdk,e:=SDKVersion(r.Identity);if e==nil&&sdk!=r.SDK{e=fmt.Errorf("SDK differs")};add("SDK",e)
 cat,e:=evidence.MeasuredCatalog(dir,r.CatalogVersion,r.CatalogHash,ai.BuiltinCatalog());if e==nil{e=evidence.VerifyRouteCatalog(cat,ai.BuiltinCatalog(),ai.OperationChat,ai.ProviderMiniMax,r.Identity.API,[]string{r.Identity.Model})};add("usedModelCatalog",e)
 plan:=Plan(r.Identity);raw,_=json.Marshal(plan);if !reflect.DeepEqual(plan,r.Expected)||Hash(raw)!=r.PlanHash||r.SourceHash!=SourcePin{add("plan",fmt.Errorf("original fixed plan differs"))}else{add("plan",nil)}
 var cases []map[string]any
 for i,c:=range r.Calls {
  if c.Outcome!="PASS" {
   if Recovered(r.Calls,i){var p Case;for _,x:=range plan{if x.ID==c.ID{p=x}};add("recoveredInitial:"+c.CaseID,verifyCall(dir,r,c,p))}
   continue
  }
  var p Case;for _,x:=range plan{if x.ID==c.ID{p=x}}
  e:=verifyCall(dir,r,c,p);msg:="";if e!=nil{msg=e.Error()}
  cases=append(cases,map[string]any{"id":c.ID,"caseId":c.CaseID,"pass":e==nil,"error":msg,"actualAttempts":len(c.Attempts)})
 }
 pass:=len(cases)==13;for _,c:=range cases{pass=pass&&c["pass"].(bool)};for _,v:=range checks{pass=pass&&v=="PASS"}
 _,full:=VerifyCapture(dir,true);fullError:="";if full!=nil{fullError=full.Error()};pass=pass&&full!=nil
 return map[string]any{"pass":pass,"kind":"partial-evidence-analysis-only","checks":checks,"cases":cases,"nativeFullQualification":false,"nativeFullRejection":fullError,"originalOutcome":r.Outcome,"originalComplete":r.Complete,"missingActualCases":[]string{"thinking-on"}},nil
}
