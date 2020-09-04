package logger

//go:generate go run github.com/golang/mock/mockgen -destination=../../mocks/logger/logger_mock.go -source=../logger/logger.go -package=logger_mock -mock_names=Logger=Logger

// Logger is the interface for logging content
type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warning(msg string)
	Error(msg string)
}
