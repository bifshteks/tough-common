package tunneling

import (
	"context"
	"sync"
)

type BridgeMessage struct {
	msg    []byte
	author Source
}

type Bridge struct {
	Active     bool
	Pool       []Source
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	mx         sync.Mutex
	messagesCh chan BridgeMessage
}

func NewBridge(ctx context.Context, cancel context.CancelFunc, sources ...Source) *Bridge {
	return &Bridge{
		Pool:       sources,
		Active:     false,
		ctx:        ctx,
		cancel:     cancel,
		messagesCh: make(chan BridgeMessage),
	}
}

func (b *Bridge) processSources(sources ...Source) {
	for _, s := range sources {
		b.wg.Add(1)
		go b.SourceHandle(s)
	}
}

func (b *Bridge) Start() {
	b.Active = true
	b.processSources(b.Pool...)
	ctx, cancel := context.WithCancel(context.Background())
	go b.WriteToSources(ctx)
	b.wg.Wait()
	cancel()
}

func (b *Bridge) AddSources(sources ...Source) {
	b.Pool = append(b.Pool, sources...)
	if b.Active {
		b.processSources(sources...)
	}
}

func (b *Bridge) findSourceIndex(source Source) (i int, exists bool) {
	for i, s := range b.Pool {
		if s == source {
			return i, true
		}
	}
	return 0, false
}

func (b *Bridge) removeSource(source Source) {
	b.mx.Lock()
	i, exists := b.findSourceIndex(source)
	if !exists {
		return
	}
	lastIndex := len(b.Pool) - 1
	b.Pool[i] = b.Pool[lastIndex]
	b.Pool = b.Pool[:lastIndex]
	b.mx.Unlock()
}

func (b *Bridge) SourceHandle(source Source) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := source.Start(b.ctx)
		if err != nil {
			b.cancel()
		}
		b.removeSource(source)
		cancel()
		b.wg.Done() // call only after .Start() ends
	}()
	reader := source.GetReader()
	for {
		select {
		case msg := <-reader:
			b.messagesCh <- BridgeMessage{msg: msg, author: source}
		case <-ctx.Done():
			return
		}
	}
}

func (b *Bridge) WriteToSources(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case bmsg := <-b.messagesCh:
			for _, src := range b.Pool {
				if src == bmsg.author {
					continue
				}
				// don't care if it ended with error - not ours responsibility
				_ = src.Write(bmsg.msg)
			}
		}
	}
}
