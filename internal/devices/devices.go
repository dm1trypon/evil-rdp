package devices

import (
	"fmt"

	logger "github.com/dm1trypon/easy-logger"
	"github.com/dm1trypon/mks-win/mouse"
)

func (d *Devices) Create() *Devices {
	d = &Devices{
		lc: "DEVICES",
	}
	return d
}

func (d *Devices) Mouse(x, y, state int) {
	logger.Info(d.lc, fmt.Sprint("MOUSE EVENT: position [", x, ":", y, "] state: ", state))
	if state == 0 {
		mouse.Move(x, y)
	} else if state == 1 {
		mouse.Press(x, y, mouse.LeftButton)
	} else if state == 2 {
		mouse.Release(x, y, mouse.LeftButton)
	}
}

func (d *Devices) Keyboard() {
	logger.Info(d.lc, fmt.Sprint("MOUSE EVENT"))
}
