package screener

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	logger "github.com/dm1trypon/easy-logger"
	"github.com/kbinani/screenshot"
)

var countImg = 1

func (s *Screener) Create() *Screener {
	s = &Screener{
		lc: "SCREENER",
	}

	return s
}

func (s *Screener) GetActiveDisplays() int {
	return screenshot.NumActiveDisplays()
}

func (s *Screener) GetResolutionDisplay(display int) image.Point {
	return screenshot.GetDisplayBounds(display).Max
}

func (s *Screener) GetScreenImage(display int) ([]byte, error) {
	bounds := screenshot.GetDisplayBounds(display)
	rgba, err := screenshot.CaptureRect(bounds)
	if err != nil {
		logger.Error(s.lc, fmt.Sprint("Error capture rect: ", err.Error()))
		return nil, err
	}

	buffer := new(bytes.Buffer)
	defer buffer.Reset()

	enc := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	if err := enc.Encode(buffer, rgba); err != nil {
		logger.Error(s.lc, fmt.Sprint("Error png encode: ", err.Error()))
		return nil, err
	}

	countImg++

	return buffer.Bytes(), nil
}
