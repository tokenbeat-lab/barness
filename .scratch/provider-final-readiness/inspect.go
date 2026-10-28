// Fixed read-only analysis command, not a qualification consumer.
package main
import("encoding/json";"fmt";"os";"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive")
func main(){os.Chdir("/Users/cyber/RestoX/barness");r,e:=minimaxlive.ReadinessAudit(".evidence/barness-ai-grok/live/minimax-messages/20261027T000249.194767Z-krdog6id/bundles/20261027T000249.211139000Z");if e!=nil{fmt.Fprintln(os.Stderr,e);os.Exit(1)};raw,_:=json.MarshalIndent(r,"","  ");fmt.Println(string(raw));if r["pass"]!=true{os.Exit(1)}}
