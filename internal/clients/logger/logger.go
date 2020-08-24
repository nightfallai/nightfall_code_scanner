package logger

// Logger is the interface for logging content
type Logger interface {
	Debug(msg string)
	Warning(msg string)
	Error(msg string)
	ErrorAnnotation(msg, location string)
}
