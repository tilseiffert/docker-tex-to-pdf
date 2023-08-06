package contextkeys

type contextKey struct {
	name string
}

// LoggerKey is the key for the logger in the context
var LoggerKey = &contextKey{"logger"}
