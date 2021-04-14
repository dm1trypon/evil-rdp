package main

import (
	"os"

	logger "github.com/dm1trypon/easy-logger"
	"github.com/dm1trypon/evil-rdp/internal/config"
	"github.com/dm1trypon/evil-rdp/internal/engine"
)

// LC - logging category
const LC = "MAIN"

func main() {
	configInst := new(config.Config).Create()

	cfg := logger.Cfg{
		AppName: "EVIL_RDP",
		LogPath: "",
		Level:   configInst.GetLoggerLevel(),
	}

	logger.SetConfig(cfg)
	logger.Info(LC, "STARTING SERVICE")

	delay := configInst.GetStreamDelay()
	chunkSize := configInst.GetStreamChunkSize()
	port := configInst.GetNetPort()
	numThreads := configInst.GetNumThreads()

	serverError := make(chan error)

	if engineInst := new(engine.Engine).Create(numThreads, delay, chunkSize, port, &serverError); engineInst == nil {
		logger.Info(LC, "STOPING SERVICE")
		os.Exit(0)
	}

	<-serverError
	logger.Info(LC, "STOPING SERVICE")
}
