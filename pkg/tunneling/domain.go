package tunneling

import (
	"context"
	"github.com/bifshteks/tough_common/pkg/tunneling/source"
)

// ITransmitter is an object that is connected by any number of Sources,
// that it reads and transmit messages from a sources to any other sources,
// connected to Transmitter
type ITransmitter interface {
	Run(ctx context.Context, cancel context.CancelFunc)
	AddSources(sources ...source.Source)
}

type IPool interface {
	Add(sources ...source.Source)
	Remove(src source.Source)
	All() (sources []source.Source)
}

type Test interface {
	Qwe()
}
type Test2 interface {
	Asd()
}
type Test3 interface {
	Zxc()
}
