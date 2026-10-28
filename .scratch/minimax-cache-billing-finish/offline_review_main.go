package main

import (
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
 "github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func main(){
 root:=os.Args[1];if err:=os.Chdir(root);err!=nil {panic(err)}
 raw,err:=os.ReadFile(filepath.Join(root,".evidence/provider-minimal-paid-supplement/delivery-v1.json"));if err!=nil {panic(err)}
 var delivery struct{Attempts []struct{NativeReport,Producer minimaxlive.ThinkingFileRef}}
 if err:=json.Unmarshal(raw,&delivery);err!=nil {panic(err)}
 pass:=len(delivery.Attempts)==2;rows:=[]map[string]any{}
 for _,a:=range delivery.Attempts {
  dir:=filepath.Dir(a.NativeReport.Path)
  r:=minimaxlive.ReviewFailedThinkingCapture(dir,a.Producer);rows=append(rows,r)
  pass=pass && r["semanticReplayReusable"]==true && r["nativeQualificationPass"]==false && r["successfulSupplementProducerPass"]==false
 }
 if err:=json.NewEncoder(os.Stdout).Encode(map[string]any{"decisions":rows,"reviewPass":pass,"vendorCalls":0,"qualificationReusable":false});err!=nil {panic(err)}
 if !pass {fmt.Fprintln(os.Stderr,"offline review expectations differ; retain gap, never patch originals");os.Exit(1)}
}
