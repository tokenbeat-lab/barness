package minimaxlive

// Verify checks the current single-process contract. Historical authorities are
// consumed with their fixed archived verifier, never a current fallback.
func Verify(dir string, formal bool) (Report, error) { return VerifyCapture(dir, formal) }
