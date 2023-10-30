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

func InitLogger(logPath string) {
	fmt.Println(AsciiArt, "\n")

	Log = logrus.New()
	Log.SetLevel(logrus.InfoLevel)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Log.Fatal("Unable to open log file for writing: ", err)
	}

	fileFormatter := &logrus.JSONFormatter{}
	fileWriter := logrus.New()
	fileWriter.SetFormatter(fileFormatter)
	fileWriter.SetOutput(logFile)

	stdoutFormatter := &logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	}
	stdoutWriter := logrus.New()
	stdoutWriter.SetFormatter(stdoutFormatter)
	stdoutWriter.SetOutput(os.Stdout)

	Log.SetOutput(io.MultiWriter(stdoutWriter.Writer(), fileWriter.Writer()))
}
