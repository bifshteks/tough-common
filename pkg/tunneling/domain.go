package tunneling

import "context"

type Source interface {
	// Consume starts process of reading from the
	// source and writing the messages to the reader
	Consume(ctx context.Context) (err error)
	// GetReader - get channel to retrieve messages from this source
	GetReader() (reader chan []byte)
	Write([]byte) (err error)
}

// ITransmitter is an object that is connected by any number of Sources,
// that it reads and transmit messages from a sources to any other sources,
// connected to Transmitter
type ITransmitter interface {
	Run(ctx context.Context, cancel context.CancelFunc)
	AddSources(sources ...Source)
}

type IPool interface {
	Add(sources ...Source)
	Remove(source Source)
	All() (sources []Source)
}
