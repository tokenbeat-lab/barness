package ai

// Operation is the model operation a binding authorizes and an entry performs.
// It is independent of the Provider and API protocol.
type Operation string

const (
	OperationChat       Operation = "chat"
	OperationImage      Operation = "image"
	OperationClassifier Operation = "classifier"
)
