package tools

import (
	"c2c/internal/config"
	"path"

	// "github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
)

type logger struct{ *logrus.Logger }

var Logger *logger

func NewLogger(config *config.Config) *logger {
	level, err := logrus.ParseLevel(config.Log.ConsoleLevel)
	if err != nil {
		logrus.Fatalf("Failed to parse level (panic,fatal,error,warn,info,debug,trace): %v", err)
	}

	logr := logrus.New()
	logr.SetLevel(level)

	// tools.Logger.SetOutput(colorable.NewColorableStdout())
	logr.SetFormatter(&logrus.TextFormatter{
		PadLevelText:    true,
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if config.Log.UseFile {
		level, err := logrus.ParseLevel(config.Log.FileLevel)
		if err != nil {
			logrus.Fatalf("Failed to parse level (panic,fatal,error,warn,info,debug,trace): %v", err)
		}

		rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
			Filename:   path.Join(config.Log.FilePath, "c2c-search-api.log"),
			MaxSize:    config.Log.FileMaxSize,    // megabytes
			MaxBackups: config.Log.FileMaxBackups, // amouts
			MaxAge:     config.Log.FileMaxAge,     //days
			Level:      level,
			Formatter: &logrus.TextFormatter{
				PadLevelText:    true,
				ForceColors:     true,
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
			},
		})

		if err != nil {
			logrus.Fatalf("Failed to initialize file rotate hook: %v", err)
		}
		logr.AddHook(rotateFileHook)
	}

	Logger = &logger{logr}

	return Logger
}
