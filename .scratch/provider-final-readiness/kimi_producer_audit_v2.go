// Analysis-only access to the existing native old-nine and process checks.
package supportmatrix
import "path/filepath"
func ReadinessOriginalProducer(bundle string,ref KimiFileRef) error {
 root,e:=kimiRepositoryRoot();if e!=nil{return e};r,e:=VerifyKimiBaselineCapture(bundle,"live");if e!=nil{return e};_,e=verifyKimiProducer(root,filepath.Join(root,bundle),ref,r);return e
}
