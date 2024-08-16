package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/scch94/ins_log"
	"github.com/scch94/micropagosSmtpServer/config"
	"github.com/scch94/micropagosSmtpServer/server"
)

func main() {
	ctx := context.Background()

	if err := config.Upconfig(ctx); err != nil {
		ins_log.Errorf(ctx, "error loading configuration: %v %s ", err)
		return
	}

	go initializeAndWatchLogger(ctx)
	time.Sleep(5 * time.Second)
	//setiamos el nombre del servicio que esta usando el loger y el nivel de logeo
	ins_log.SetService("MicropagosSmtpServer")
	ins_log.SetLevel(config.Config.LogLevel)
	ctx = ins_log.SetPackageNameInContext(ctx, "main")
	defer end(ctx)

	ins_log.Infof(ctx, "starting  micropagos smtp server version: %v", getVersion())

	if err := server.StartServer(ctx); err != nil {
		ins_log.Errorf(ctx, "error starting server: %s", err.Error())
	}

}

// funcion que ira cambiando de log cada hora
func initializeAndWatchLogger(ctx context.Context) {
	var file *os.File
	var logFileName string
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		default:
			logDir := "../log"

			// Create the log directory if it doesn't exist
			if err = os.MkdirAll(logDir, 0755); err != nil {
				ins_log.Errorf(ctx, "error creating log directory: %v", err)
				return
			}

			// Define the log file name
			today := time.Now().Format("2006-01-02 15")
			replacer := strings.NewReplacer(" ", "_")
			today = replacer.Replace(today)
			logFileName = filepath.Join(logDir, config.Config.LogName+today+".log")

			// Open the log file
			file, err = os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				ins_log.Errorf(ctx, "error opening log file: %v", err)
				return
			}

			// Create a writer that writes to both file and console
			multiWriter := io.MultiWriter(os.Stdout, file)
			ins_log.StartLoggerWithWriter(multiWriter)

			// Esperar hasta el inicio de la próxima hora
			nextHour := time.Now().Truncate(time.Hour).Add(time.Hour)
			time.Sleep(time.Until(nextHour))

			// Close the previous log file
			file.Close()
		}
	}
}
func getVersion() string {
	return "1.0.0"
}

func end(ctx context.Context) {
	ins_log.Infof(ctx, "clossing...")
}
