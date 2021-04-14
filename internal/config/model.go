package config

type Config struct {
	logger logger
	stream stream
	net    net
}

type logger struct {
	level int
}

type net struct {
	port int
}

type stream struct {
	numThreads int
	displays   []int
	delay      int
	chunkSize  int
}
