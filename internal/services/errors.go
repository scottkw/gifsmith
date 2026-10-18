package services

// ActionError is the one error type ActionService returns to the webview (besides
// ErrClipsCopy). Its text is fixed per kind: never a URL, path or transport error.
type ActionError struct {
	Kind string `json:"kind"`
}

func (e *ActionError) Error() string { return "" }

// MarshalError makes an *ActionError err.cause in JS; nil leaves every other error to Wails.
func MarshalError(err error) []byte { return nil }
