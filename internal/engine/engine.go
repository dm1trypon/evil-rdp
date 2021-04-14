package engine

import (
	"encoding/json"
	"fmt"
	"time"

	logger "github.com/dm1trypon/easy-logger"
	"github.com/dm1trypon/evil-rdp/internal/chunker"
	"github.com/dm1trypon/evil-rdp/internal/devices"
	"github.com/dm1trypon/evil-rdp/internal/packer"
	"github.com/dm1trypon/evil-rdp/internal/screener"
	"github.com/dm1trypon/evil-rdp/internal/server"
)

func (e *Engine) Create(numThreads, delay, chunkSize, port int, serverError *chan error) *Engine {
	e = &Engine{
		lc:                 "ENGINE",
		serverInst:         new(server.Server).Create(serverError, port),
		packerInst:         new(packer.Packer).Create(),
		screenerInst:       new(screener.Screener).Create(),
		devicesInst:        new(devices.Devices).Create(),
		chunkerInst:        new(chunker.Chunker).Create(chunkSize),
		chunkSize:          chunkSize,
		activeDisplay:      0,
		frame:              0,
		buf:                make(chan [][]byte),
		screenshotThreads:  0,
		screenshotInterval: 0,
		screenshotDuration: 0,
		frameDuration:      delay,
		msgs:               nil,
		newClient:          nil,
	}

	e.msgs = e.serverInst.GetMsgsChan()
	e.newClient = e.serverInst.GetNewClient()

	if err := e.setup(); err != nil {
		return nil
	}

	go e.worker()
	go e.monitoringFPS()
	go e.sender()
	go e.serverEventer()

	return e
}

func (e *Engine) serverEventer() {
	for {
		select {
		case msg := <-*e.msgs:
			// logger.Info(e.lc, fmt.Sprint("RECV: ", string(msg)))
			e.onMessage(msg)
		case address := <-*e.newClient:
			e.onNewClient(address)
		}
	}
}

func (e *Engine) onNewClient(address string) {
	resolution := e.screenerInst.GetResolutionDisplay(0)
	logger.Info(e.lc, fmt.Sprint("RES: ", resolution.X, "x", resolution.Y))
	initMsg, err := e.packerInst.MakeInit(resolution.X, resolution.Y, e.chunkSize)
	if err != nil {
		return
	}

	e.serverInst.InteractiveSend(initMsg, address)
}

func (e *Engine) onMessage(msg []byte) {
	var kbMsg KeyBoardMsg
	var mouseMsg MouseMsg

	if err := json.Unmarshal(msg, &kbMsg); err == nil && kbMsg.Method == "keyboard" {
		e.devicesInst.Keyboard()
	} else if err := json.Unmarshal(msg, &mouseMsg); err == nil && kbMsg.Method == "mouse" {
		e.devicesInst.Mouse(mouseMsg.Position.X, mouseMsg.Position.Y, mouseMsg.State)
	} else {
		logger.Error(e.lc, fmt.Sprint("Message parsing error: no supported method"))
	}
}

func (e *Engine) SetActiveDisplay(activeDisplay int) {
	displays := e.screenerInst.GetActiveDisplays()

	if activeDisplay >= displays || activeDisplay < 0 {
		logger.Critical(e.lc, fmt.Sprint("Out of range display ", activeDisplay))
		return
	}

	e.activeDisplay = 1
}

func (e *Engine) screenshotBenchmark() error {
	screenBenchmark := time.Now()
	_, err := e.screenerInst.GetScreenImage(e.activeDisplay)
	if err != nil {
		return err
	}

	e.screenshotDuration = int(time.Since(screenBenchmark).Milliseconds())

	return nil
}

func (e *Engine) setup() error {
	if err := e.screenshotBenchmark(); err != nil {
		return err
	}

	e.screenshotThreads = e.screenshotDuration / e.frameDuration

	if e.screenshotThreads*e.frameDuration < e.screenshotDuration {
		e.screenshotThreads++
	}

	e.screenshotInterval = e.screenshotThreads*e.frameDuration - e.screenshotDuration

	logger.Info(e.lc, fmt.Sprint("Screenshot's duration is ", e.screenshotDuration))
	logger.Info(e.lc, fmt.Sprint("Count threads is ", e.screenshotThreads))
	logger.Info(e.lc, fmt.Sprint("Screenshot's interval by thread is ", e.screenshotInterval))

	return nil
}

func (e *Engine) worker() {
	for thread := 0; thread < e.screenshotThreads; thread++ {
		go func() {
			for {
				if e.serverInst.GetNumConnectedClients() < 1 {
					continue
				}

				data, err := e.screenerInst.GetScreenImage(e.activeDisplay)
				if err != nil {
					continue
				}

				chunks := e.chunkerInst.MakeParts(data)

				e.buf <- chunks

				time.Sleep(time.Duration(e.screenshotInterval) * time.Millisecond)
			}
		}()

		time.Sleep(time.Duration(e.frameDuration) * time.Millisecond)
	}
}

func (e *Engine) monitoringFPS() {
	for {
		time.Sleep(time.Second)
		logger.Info(e.lc, fmt.Sprint("FPS: ", e.frame))
		e.frame = 0
	}
}

func (e *Engine) sender() {
	for {
		chunks := <-e.buf

		for _, chunk := range chunks {
			e.serverInst.StreamSend(chunk)
		}

		e.frame++
	}
}
