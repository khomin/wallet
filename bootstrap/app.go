package bootstrap

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tracker/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Application struct {
	Cfg     *config.Config
	LogFile *os.File
}

func App() Application {
	app := &Application{}
	app.Cfg = config.NewConfig()
	app.InitLog()
	return *app
}

func (a *Application) InitLog() {
	if strings.Contains(a.Cfg.Server.Environment, "dev") {
		logFormatter := new(logrus.TextFormatter)
		logFormatter.ForceColors = true
		logrus.SetFormatter(logFormatter)
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetOutput(os.Stderr)
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		lumberjackLogger := &lumberjack.Logger{
			Filename:   filepath.ToSlash(a.Cfg.Server.LogPath),
			MaxSize:    1, // MB
			MaxBackups: 2,
			MaxAge:     3,    // days
			Compress:   true, // disabled by default
		}
		multiWriter := io.MultiWriter(os.Stderr, lumberjackLogger)
		logFormatter := new(logrus.TextFormatter)
		logFormatter.TimestampFormat = time.RFC1123Z
		logFormatter.FullTimestamp = true
		logFormatter.ForceColors = true
		logrus.SetFormatter(logFormatter)
		logrus.SetOutput(multiWriter)
		logrus.SetLevel(logrus.WarnLevel)
	}
}
