package logger

import (
	"encoding/json"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Blogs struct {
	Logger *zap.Logger
}

func NewLog(lodFileName string) *Blogs {
	rawJSON, err := os.ReadFile(lodFileName)
	if err != nil {
		rawJSON = []byte(`{
		"level": "debug",
		"encoding": "json",
		"outputPaths": ["./clientLogs.log"],
		"errorOutputPaths": ["stderr"],
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "levelEncoder": "lowercase"
		}
	  }`)
	}

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("Jan 02 15:04:05.000000")
	cfg.EncoderConfig.StacktraceKey = "" // to hide stacktrace info

	logger := zap.Must(cfg.Build())
	defer logger.Sync()

	return &Blogs{logger}
}
