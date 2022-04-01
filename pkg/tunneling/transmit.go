package tunneling

import (
	"context"
	"sync"
)

type Message struct {
	content []byte
	author  Source
}

type Pool struct {
	mx      sync.Mutex
	sources []Source
}

func (pool *Pool) All() (sources []Source) {
	return pool.sources
}

func (pool *Pool) Add(sources ...Source) {
	pool.mx.Lock()
	pool.sources = append(pool.sources, sources...)
	pool.mx.Unlock()
}

func (pool *Pool) findSourceIndex(source Source) (i int, exists bool) {
	for i, s := range pool.sources {
		if s == source {
			return i, true
		}
	}
	return 0, false
}

func (pool *Pool) Remove(source Source) {
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
}

func NewTransmitter() *Transmitter {
	return &Transmitter{
		pool: &Pool{
			sources: make([]Source, 0),
		},
		messagesCh: make(chan Message),
	}
}

func (t *Transmitter) processSources(sources ...Source) {
	for _, s := range sources {
		t.wg.Add(1)
		go t.read(s)
	}
}

// Run starts reading from all currently added sources and writing their output
// to other sources.
// If source returns error on source.Start() - calls context's cancel function
func (t *Transmitter) Run(ctx context.Context, cancel context.CancelFunc) {
	t.ctx = ctx
	t.cancel = cancel
	t.processSources(t.pool.All()...)
	go t.WriteToSources()

	<-ctx.Done()
	t.wg.Wait()
	t.ctx = nil
	t.cancel = nil
	close(t.messagesCh) // after waited for every source to finish - close channel to stop Write goroutine
}

func (t *Transmitter) AddSources(sources ...Source) {
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

func (t *Transmitter) read(source Source) {
	readCtx, readCancel := context.WithCancel(context.Background())
	go func() {
		err := source.Consume(t.ctx)
		if err != nil {
			t.cancel()
		}
		t.pool.Remove(source)
		readCancel() // stop reading for-loop processing current source
		t.wg.Done()  // call only after .Start() ends
	}()
	reader := source.GetReader()
	for {
		select {
		case msg := <-reader:
			t.messagesCh <- Message{content: msg, author: source}
		case <-readCtx.Done():
			return
		}
	}
}

func (t *Transmitter) WriteToSources() {
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
