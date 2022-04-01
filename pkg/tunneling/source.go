package tunneling

import "context"

type Source interface {
	GetName() string
	GetUrl() string
	Connect(ctx context.Context) (err error) // dial connection
	// Start - start reading from instance
	Start(ctx context.Context) (err error)
	// GetReader - get channel to retrieve messages from this source
	GetReader() chan []byte
	Write([]byte) (err error)
}
