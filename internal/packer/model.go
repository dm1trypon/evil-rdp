package packer

type Packer struct {
	lc string
}

type MsgInit struct {
	Method     string     `json:"method"`
	PckgSize   int        `json:"package_size"`
	Resolution Resolution `json:"resolution"`
}

type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}
