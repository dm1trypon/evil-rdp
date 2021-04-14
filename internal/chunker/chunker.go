package chunker

func (c *Chunker) Create(chunkSize int) *Chunker {
	c = &Chunker{
		lc:        "CHUNKER",
		chunkSize: chunkSize,
	}

	return c
}

func (c *Chunker) MakeParts(data []byte) [][]byte {
	var chunkList [][]byte

	pack := data

	for {
		if len(pack) < c.chunkSize {
			chunkList = append(chunkList, pack)
			break
		}

		chunkList = append(chunkList, pack[:c.chunkSize])
		pack = pack[c.chunkSize:]
	}

	return chunkList
}
