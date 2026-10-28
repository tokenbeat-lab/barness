// Offline inspection of the two exact retained Responses bundles. No dialer,
// credentials, grant writer or vendor execution is available in this program.
package main

import (
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/json"
 "fmt"
 "io"
 "net/http"
 "os"
 "path/filepath"
 "reflect"
 "strings"
 "time"

 "github.com/tokenbeat-lab/barness/ai"
)

type exchange struct { Method, URL, RequestBody, ResponseBody string; Status int; ResponseHeaders map[string]string }
type storedMessage struct { ai.AssistantMessage; Content []json.RawMessage `json:"content"` }
type storedResult struct { Message storedMessage; Metadata ai.CallMetadata }
type host struct { binding ai.Binding; x exchange; equal bool; calls int }
func (h *host) ResolveBinding(context.Context, ai.CallScope, string) (ai.Binding,error) { return h.binding,nil }
func (h *host) ResolveCredential(context.Context, ai.CallScope, ai.Binding) (ai.Credential,error) {
 return ai.Credential{OwnerTenantID:h.binding.TenantID,CredentialID:h.binding.CredentialRef,Version:"offline",BindingVersion:h.binding.Version,AccountScopeID:h.binding.AccountScopeID,Active:true,APIKey:ai.NewSecret("synthetic-offline")},nil
}
func canonical(v any) []byte { b,_:=json.Marshal(v); var x any; d:=json.NewDecoder(bytes.NewReader(b)); d.UseNumber(); _=d.Decode(&x); b,_=json.Marshal(x); return b }
func sameRaw(a,b []byte) bool { var x,y any; d:=json.NewDecoder(bytes.NewReader(a));d.UseNumber();if d.Decode(&x)!=nil{return false};d=json.NewDecoder(bytes.NewReader(b));d.UseNumber();return d.Decode(&y)==nil && reflect.DeepEqual(x,y) }
func read(p string, v any) { b,e:=os.ReadFile(p); if e!=nil {panic(e)}; if e=json.Unmarshal(b,v);e!=nil{panic(e)} }
func (h *host) RoundTrip(r *http.Request) (*http.Response,error) {
 h.calls++; b,e:=io.ReadAll(io.LimitReader(r.Body,1<<20));r.Body.Close()
 h.equal=e==nil&&r.Method==h.x.Method&&r.URL.String()==h.x.URL&&sameRaw(b,[]byte(h.x.RequestBody))
 if !h.equal{return nil,fmt.Errorf("current request differs")}
 headers:=http.Header{};for k,v:=range h.x.ResponseHeaders{headers.Set(k,v)}
 return &http.Response{StatusCode:h.x.Status,Header:headers,Body:io.NopCloser(strings.NewReader(h.x.ResponseBody)),Request:r},nil
}
func assistant(m storedMessage) ai.AssistantMessage {
 a:=m.AssistantMessage
 for _,b:=range m.Content{
  var f map[string]json.RawMessage;_ = json.Unmarshal(b,&f)
  switch { case f["text"]!=nil: var x ai.Text;_ = json.Unmarshal(b,&x);a.Content=append(a.Content,x)
  case f["thinking"]!=nil:var x ai.Thinking;_=json.Unmarshal(b,&x);a.Content=append(a.Content,x)
  case f["id"]!=nil:var x ai.ToolCall;_=json.Unmarshal(b,&x);a.Content=append(a.Content,x)
  default:panic("unknown assistant content") }
 }
 return a
}
func request(raw []byte, meta ai.CallMetadata, known map[string]ai.CallMetadata) ai.Request {
 var in struct {SystemPrompt string; Messages []json.RawMessage; Tools []ai.Tool};if json.Unmarshal(raw,&in)!=nil{panic("invalid host input")}
 out:=ai.Request{SystemPrompt:in.SystemPrompt,Tools:in.Tools}
 for _,r:=range in.Messages{
  var f map[string]json.RawMessage;_=json.Unmarshal(r,&f)
  switch {
  case f["api"]!=nil:
   var m storedMessage;_=json.Unmarshal(r,&m);a:=assistant(m)
   old,ok:=known[string(canonical(a))]
   if !ok||old.TenantID!=meta.TenantID||old.AccountScopeID!=meta.AccountScopeID {panic("assistant lacks same-bundle result provenance")}
   env:=ai.NativeStateEnvelope{TenantID:old.TenantID,AccountScopeID:old.AccountScopeID,ProviderID:a.Provider,API:a.API,ModelID:a.Model}
   var e error;a.NativeState,e=ai.TrustNativeState(ai.CallScope{TenantID:meta.TenantID},env);if e!=nil{panic(e)}
   out.Messages=append(out.Messages,a)
  case f["toolCallId"]!=nil:
   var m struct {ToolCallID,ToolName string; IsError bool; Content []ai.Text};_=json.Unmarshal(r,&m)
   t:=ai.ToolResultMessage{ToolCallID:m.ToolCallID,ToolName:m.ToolName,IsError:m.IsError};for _,x:=range m.Content{t.Content=append(t.Content,x)};out.Messages=append(out.Messages,t)
  default:
   var m struct{Content []ai.Text};_=json.Unmarshal(r,&m);u:=ai.UserMessage{};for _,x:=range m.Content{u.Content=append(u.Content,x)};out.Messages=append(out.Messages,u)
  }
 }
 return out
}
func main(){
 var selected []struct{Route,Bundle string};read(".evidence/provider-source-scope/selected-evidence-v1.json",&selected)
 var output []map[string]any;failed:=false
 for _,s:=range selected[1:]{
  var report struct{Endpoint string;Scenarios []struct{ID,CaseID,Model string}};read(filepath.Join(s.Bundle,"live-report.json"),&report)
  known:=map[string]ai.CallMetadata{};results:=map[string]storedResult{}
  for _,sc:=range report.Scenarios{
   p:=filepath.Join(s.Bundle,sc.CaseID,"result.json");var result storedResult
   if s.Route=="zhipu-responses"{var w struct{Result storedResult};read(filepath.Join(s.Bundle,sc.CaseID,"attempt-1-result.json"),&w);result=w.Result}else{read(p,&result)}
   results[sc.CaseID]=result;known[string(canonical(assistant(result.Message)))]=result.Metadata
  }
  for _,sc:=range report.Scenarios{
   dir:=filepath.Join(s.Bundle,sc.CaseID);saved:=results[sc.CaseID];meta:=saved.Metadata
   inputName,xname:="request.json","exchanges.json";if s.Route=="zhipu-responses"{inputName,xname="public-request.json","attempt-1-exchanges.json"}
   raw,_:=os.ReadFile(filepath.Join(dir,inputName));input:=request(raw,meta,known)
   var opts ai.ResponsesOptions;read(filepath.Join(dir,"options.json"),&opts)
   var xs []exchange;read(filepath.Join(dir,xname),&xs);if len(xs)!=1{panic("one exchange required")}
   h:=&host{binding:ai.Binding{TenantID:meta.TenantID,BindingID:meta.BindingID,Version:"offline",Enabled:true,Operation:ai.OperationChat,ProviderID:meta.ProviderID,API:meta.API,Endpoint:report.Endpoint,AuthKind:ai.AuthAPIKey,AccountScopeID:meta.AccountScopeID,CredentialRef:"offline",AllowedModels:[]string{sc.Model}},x:xs[0]}
   policy:=&ai.ResourcePolicy{MaxRequestBytes:256<<10,MaxImageBytes:1<<10,MaxFrameBytes:512<<10,MaxToolJSONBytes:256<<10,MaxErrorBodyBytes:16<<10,MaxOutputBytes:1<<20,MaxQueuedEvents:8192,MaxQueuedEventBytes:2<<20,MaxConcurrentPerTenant:1,MaxConcurrentProcess:1,MaxAdmissionWaiters:1,AdmissionWait:time.Second,CallTimeout:5*time.Second,ConnectTimeout:time.Second,ResponseHeaderTimeout:time.Second,ReadIdleTimeout:time.Second}
   client,e:=ai.NewClient(ai.Config{Policy:policy,Bindings:h,Credentials:h,Transport:h});if e!=nil{panic(e)}
   ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
   hooks:=ai.Hooks{}
   if strings.HasSuffix(sc.ID,"/control-low")||strings.HasSuffix(sc.ID,"/control-max") {
    effort:="low";if strings.HasSuffix(sc.ID,"/control-max"){effort="max"}
    hooks.OnPayload=func(_ context.Context,_ ai.CallScope,p *ai.Payload)(ai.PayloadDecision,error){p.Body["reasoning"].(map[string]any)["effort"]=effort;return ai.KeepPayload(),nil}
   }
   result,callErr:=client.WithHooks(hooks).Complete(ctx,ai.CallScope{TenantID:meta.TenantID,RequestID:"offline-source-review"},ai.Target{BindingID:meta.BindingID,ModelID:sc.Model},input,opts);cancel()
   isCancel:=strings.HasSuffix(sc.ID,"/cancel");expected:=assistant(saved.Message)
   responseEqual:=reflect.DeepEqual(result.Message.Content,expected.Content)&&result.Message.Usage==expected.Usage&&result.Message.StopReason==expected.StopReason&&result.Message.ResponseID==expected.ResponseID
   pass:=h.equal&&h.calls==1&&(isCancel||callErr==nil&&responseEqual)
   row:=map[string]any{"route":s.Route,"caseID":sc.CaseID,"requestEqual":h.equal,"responseEqual":responseEqual,"attempts":h.calls,"cancelCapture":isCancel,"sessionEmpty":opts.SessionID=="","pass":pass,"requestSHA256":fmt.Sprintf("%x",sha256.Sum256([]byte(xs[0].RequestBody)))}
   if callErr!=nil{row["error"]=callErr.Error()};output=append(output,row);if !pass{failed=true}
  }
 }
 b,_:=json.MarshalIndent(output,"","  ");fmt.Println(string(b));if failed{os.Exit(1)}
}
