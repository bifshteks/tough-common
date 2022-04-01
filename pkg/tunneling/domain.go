package tunneling

import "context"

type Source interface {
	//GetName() string
	//GetUrl() string
	//Connect(ctx context.Context) (err error) // dial connection
	// Start - start reading from instance
	Consume(ctx context.Context) (err error)
	// GetReader - get channel to retrieve messages from this source
	GetReader() chan []byte
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
