package tunneling

import (
	"context"
	"github.com/bifshteks/tough_common/pkg/tunneling/source"
)

// ITransmitter is an object that is connected by any number of Sources,
// that it reads and transmits messages from a sources to any other sources,
// connected to ITransmitter
type ITransmitter interface {
	Run(ctx context.Context, cancel context.CancelFunc)
	AddSources(sources ...source.Source)
}

type IPool interface {
	Add(sources ...source.Source)
	Remove(src source.Source)
	All() (sources []source.Source)
}
