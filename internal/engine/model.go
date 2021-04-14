package engine

import (
	"github.com/dm1trypon/evil-rdp/internal/chunker"
	"github.com/dm1trypon/evil-rdp/internal/devices"
	"github.com/dm1trypon/evil-rdp/internal/packer"
	"github.com/dm1trypon/evil-rdp/internal/screener"
	"github.com/dm1trypon/evil-rdp/internal/server"
)

type Engine struct {
	lc                 string
	serverInst         *server.Server
	packerInst         *packer.Packer
	screenerInst       *screener.Screener
	devicesInst        *devices.Devices
	chunkerInst        *chunker.Chunker
	chunkSize          int
	activeDisplay      int
	frame              int
	buf                chan [][]byte
	screenshotThreads  int
	screenshotInterval int
	screenshotDuration int
	frameDuration      int
	msgs               *chan []byte
	newClient          *chan string
}

type KeyBoardMsg struct {
	Method   string   `json:"method"`
	IsHold   bool     `json:"is_hold"`
	Position Position `json:"position"`
}

type MouseMsg struct {
	Method   string   `json:"method"`
	State    int      `json:"state"`
	Position Position `json:"position"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}
