package main
import (
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
 "reflect"
 "time"
 "github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
 "github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
 "github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)
func main(){
 root:=os.Args[1];os.Chdir(root);paths,_:=filepath.Glob(".scratch/kimi-composition-promotion/pending-scopes-v5/*.json");ok:=len(paths)==7;var rows []map[string]any
 for _,p:=range paths {
  raw,e:=os.ReadFile(p);if e!=nil {panic(e)}
  var d struct{DocumentType,Route string;Approved,Executable,GrantIssued bool;MachineScope livegrant.Scope;FrozenProducer struct{Path,SHA256 string};Baseline struct{Manifest struct{SHA256 string}};NativeSourceSHA256 string}
  if json.Unmarshal(raw,&d)!=nil {panic("proposal")}
  var check error
  if d.Route=="minimax-messages" {
   _,check=livegrant.Decode(raw)
   id,_:=minimaxlive.Select("messages",minimaxlive.M3,"standard");s,e:=minimaxlive.ThinkingOnScope(id,"live",d.FrozenProducer.SHA256,d.Baseline.Manifest.SHA256)
   h,e2:=livegrant.FileHash(d.FrozenProducer.Path)
   ok=ok && e==nil && e2==nil && h==d.FrozenProducer.SHA256 && reflect.DeepEqual(s,d.MachineScope)
  } else {var g supportmatrix.RegionalGrant;check=json.Unmarshal(raw,&g);if check==nil {check=g.Check(d.Route,time.Now().UTC())}}
  pass:=check!=nil && !d.Approved && !d.Executable && !d.GrantIssued
  ok=ok&&pass;rows=append(rows,map[string]any{"route":d.Route,"proposalCannotAuthorize":pass,"vendorCalls":0,"claims":0})
 }
 json.NewEncoder(os.Stdout).Encode(rows);if !ok {fmt.Fprintln(os.Stderr,"pending scope validation failed");os.Exit(1)}
}
