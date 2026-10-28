package main
import (
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
 "github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
 "github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)
func main(){
 dir:=os.Args[1];receipt:=os.Args[2]
 r,e:=minimaxlive.VerifyThinkingBaseline(dir,true);if e!=nil {fmt.Fprintln(os.Stderr,e);os.Exit(1)}
 raw,e:=os.ReadFile(receipt);if e!=nil {panic(e)}
 inputs,e:=minimaxlive.VerifyThinkingProducer(dir,minimaxlive.ThinkingFileRef{Path:receipt,SHA256:livegrant.Hash(raw)},r,true);if e!=nil {fmt.Fprintln(os.Stderr,e);os.Exit(1)}
 g,b,c,e:=livegrant.ReadEvidence(dir);if e!=nil {panic(e)}
 if e=livegrant.Verify(g,b,c);e!=nil {panic(e)}
 var before map[string]any;if e=json.Unmarshal(b,&before);e!=nil {panic(e)}
 afterRaw,_:=json.Marshal(g);var after map[string]any;json.Unmarshal(afterRaw,&after)
 if after["scope"].(map[string]any)["maxUSD"]!=nil || after["scope"].(map[string]any)["baselineManifestSha256"]!=nil {panic("old canonical scope changed")}
 pass:=[]string{};for _,v:=range r.Calls {if v.Outcome=="PASS" && v.ID!="thinking-on" {pass=append(pass,v.ID)}}
 _,fullErr:=minimaxlive.VerifyCapture(dir,true)
 json.NewEncoder(os.Stdout).Encode(map[string]any{"pass":len(pass)==13 && fullErr!=nil,"kind":"current-fixed-thirteen-applicability","bundle":dir,"producer":receipt,"originalOutcome":r.Outcome,"originalExitCode":1,"originalCalls":len(r.Calls),"originalBudget":r.Budget,"applicable":pass,"originalClaimKeyVerified":true,"oldOptionalFieldsOmitted":true,"originalScopeProcessSeconds":g.Scope.ProcessSeconds,"sourceAndProducerInputs":inputs,"fullQualificationRefused":fmt.Sprint(fullErr),"reportFile":filepath.Join(dir,minimaxlive.ReportFile),"vendorCalls":0,"grantsIssued":0,"claims":0})
}
