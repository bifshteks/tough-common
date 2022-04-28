package tunneling

import (
	"context"
	"github.com/bifshteks/tough_common/pkg/tunneling/source"
	"github.com/sirupsen/logrus"
	"sync"
)

type Message struct {
	content []byte
	author  source.Source
}

type Pool struct {
	mx      sync.Mutex
	sources []source.Source
}

func (pool *Pool) All() (sources []source.Source) {
	return pool.sources
}

func (pool *Pool) Add(sources ...source.Source) {
	pool.mx.Lock()
	pool.sources = append(pool.sources, sources...)
	pool.mx.Unlock()
}

func (pool *Pool) findSourceIndex(source source.Source) (i int, exists bool) {
	for i, s := range pool.sources {
		if s == source {
			return i, true
		}
	}
	return 0, false
}

func (pool *Pool) Remove(source source.Source) {
	pool.mx.Lock()
	i, exists := pool.findSourceIndex(source)
	if !exists {
		return
	}
	lastIndex := len(pool.sources) - 1
	pool.sources[i] = pool.sources[lastIndex]
	pool.sources = pool.sources[:lastIndex]
	pool.mx.Unlock()
}

type Transmitter struct {
	pool       IPool
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	messagesCh chan Message
	logger     *logrus.Logger
}

func NewTransmitter(logger *logrus.Logger) *Transmitter {
	return &Transmitter{
		pool: &Pool{
			sources: make([]source.Source, 0),
		},
		messagesCh: make(chan Message),
		logger:     logger,
	}
}

func (t *Transmitter) processSources(sources ...source.Source) {
	for _, s := range sources {
		t.wg.Add(1)
		go t.read(s)
	}
}

// Run starts reading from all currently added sources and writing their output
// to other sources.
func (t *Transmitter) Run(ctx context.Context, cancel context.CancelFunc) {
	defer t.logger.Infoln("transmitter.Run() ends")
	t.ctx = ctx
	t.cancel = cancel
	t.processSources(t.pool.All()...)
	go t.WriteToSources()

	<-ctx.Done()
	t.wg.Wait()
	close(t.messagesCh) // after waited for every source to finish - close channel to stop Write goroutine
}

func (t *Transmitter) AddSources(sources ...source.Source) {
	t.pool.Add(sources...)
	isRunning := t.ctx != nil
	if !isRunning {
		return
	}
	select {
	case <-t.ctx.Done():
		panic("cannot add sources to closed Transmitter")
	default:
		t.processSources(sources...)
	}
}

func (t *Transmitter) read(source source.Source) {
	t.logger.Debugln("transfmitter.read start", source, "ctx = ", t.ctx)
	defer t.logger.Infof("transmitter.read() ends")
	var sourceReadWg sync.WaitGroup
	sourceReadWg.Add(1)
	go func() {
		err := source.Consume(t.ctx)
		if err != nil {
			t.logger.Errorf("source had error consuming: %s", err.Error())
		}
		sourceReadWg.Done()
		t.cancel()
	}()
	reader := source.GetReader()
	for {
		select {
		case msg := <-reader:
			t.messagesCh <- Message{content: msg, author: source}
		case <-t.ctx.Done():
			sourceReadWg.Wait()
			t.pool.Remove(source)
			t.wg.Done()
			return
		}
	}
}

func (t *Transmitter) WriteToSources() {
	defer t.logger.Info("transmitter.WriteToSources() ends")
	for {
		msg, more := <-t.messagesCh
		if !more {
			return
		}
		for _, src := range t.pool.All() {
			isAuthorOfMsg := src == msg.author
			if isAuthorOfMsg {
				continue
			}
			// don't care if it ended with error - is not ours responsibility
			_ = src.Write(msg.content)
		}
	}
}
