//go:build live

package live

import (
	"net/http"
	"os"
)

func configureGLMProxy(rec *recorder) error {
	if raw := os.Getenv("BARNESS_GLM_PROXY"); raw != "" {
		u, e := smokeProxy(raw)
		if e != nil {
			return e
		}
		rec.next.(*http.Transport).Proxy = http.ProxyURL(u)
	}
	return nil
}
