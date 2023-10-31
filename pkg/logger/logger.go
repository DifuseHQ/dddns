package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var Log *logrus.Logger

const (
	AsciiArt = `
██████╗ ██████╗ ██████╗ ███╗   ██╗███████╗
██╔══██╗██╔══██╗██╔══██╗████╗  ██║██╔════╝
██║  ██║██║  ██║██║  ██║██╔██╗ ██║███████╗
██║  ██║██║  ██║██║  ██║██║╚██╗██║╚════██║
██████╔╝██████╔╝██████╔╝██║ ╚████║███████║
╚═════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝`
)

func InitLogger(logPath string, logLevel int, version string) {
	fmt.Println(AsciiArt)
	fmt.Println("\t\t\t\t", version, "\n")

	Log = logrus.New()

	switch logLevel {
	case 0:
		Log.SetLevel(logrus.InfoLevel)
	case 1:
		Log.SetLevel(logrus.DebugLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Log.Fatal("Unable to open log file for writing: ", err.Error())
	}

	formatter := &logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	}

	Log.SetFormatter(formatter)
	Log.SetOutput(io.MultiWriter(os.Stdout, logFile))
}
