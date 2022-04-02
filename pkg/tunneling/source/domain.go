package source

import "context"

type Source interface {
	// Consume starts process of reading from the
	// source and writing the messages to the reader
	Consume(ctx context.Context, cancel context.CancelFunc) (err error)
	// GetReader - get channel to retrieve messages from this source
	GetReader() (reader chan []byte)
	Write([]byte) (err error)
}

type NetworkSource interface {
	Source
	Connect(ctx context.Context, cancel context.CancelFunc) (err error)
	GetUrl() (url string)
}
