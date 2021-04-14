package packer

import (
	"encoding/json"
	"fmt"

	logger "github.com/dm1trypon/easy-logger"
)

func (p *Packer) Create() *Packer {
	p = &Packer{
		lc: "PACKER",
	}

	return p
}

func (p *Packer) MakeInit(width, height, chunkSize int) ([]byte, error) {
	bodyObj := MsgInit{
		Method:   "init",
		PckgSize: chunkSize,
		Resolution: Resolution{
			Width:  width,
			Height: height,
		},
	}

	body, err := json.Marshal(bodyObj)
	if err != nil {
		logger.Error(p.lc, fmt.Sprint("Packing init error: ", err.Error()))
		return nil, err
	}

	return body, nil
}
