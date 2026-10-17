//go:build live

package live

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// Pace only cancellation frames. Complete calls consume the exact bytes without
// pacing. No response byte is added/rewritten, and the original partial tail is
// retained. A read blocks after each visible frame until the public consumer
// advances, making the original observed cancellation content reproducible.
func (s *glmCaptureReplay) cancellationBody(req *http.Request, body string) io.ReadCloser {
	pr, pw := io.Pipe()
	for {
		select {
		case <-s.ack:
			continue
		default:
			goto drained
		}
	}
drained:
	go func() {
		reader := bufio.NewReader(strings.NewReader(body))
		frame := ""
		for {
			line, e := reader.ReadString('\n')
			frame += line
			if strings.TrimSpace(line) == "" && frame != "" {
				if _, err := io.WriteString(pw, frame); err != nil {
					return
				}
				if glmVisibleFrame(frame) {
					select {
					case <-s.ack:
					case <-req.Context().Done():
						pw.CloseWithError(req.Context().Err())
						return
					}
				}
				frame = ""
			}
			if e != nil {
				if frame != "" {
					if _, err := io.WriteString(pw, frame); err != nil {
						return
					}
				}
				break
			}
		}
		<-req.Context().Done()
		pw.CloseWithError(req.Context().Err())
	}()
	return pr
}
func glmVisibleFrame(frame string) bool {
	for _, line := range strings.Split(frame, "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var value map[string]any
		if json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &value) != nil {
			continue
		}
		if kind, _ := value["type"].(string); kind == "response.reasoning_text.delta" || kind == "response.output_text.delta" {
			text, _ := value["delta"].(string)
			return text != ""
		}
		if delta, ok := value["delta"].(map[string]any); ok {
			for _, key := range []string{"thinking", "text"} {
				if text, _ := delta[key].(string); text != "" {
					return true
				}
			}
		}
		if choices, ok := value["choices"].([]any); ok {
			for _, v := range choices {
				choice, _ := v.(map[string]any)
				delta, _ := choice["delta"].(map[string]any)
				for _, key := range []string{"content", "reasoning_content", "reasoning"} {
					if text, _ := delta[key].(string); text != "" {
						return true
					}
				}
			}
		}
	}
	return false
}
