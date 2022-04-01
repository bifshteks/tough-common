package tunneling

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestTransmitterSendsMessagesToEverySource(t *testing.T) {
	cases := [][3]*SourceMock{
		{
			NewSourceMock([]string{}, false, false),
			NewSourceMock([]string{"asd", "qwe"}, false, false),
			NewSourceMock([]string{}, false, false)},
		{
			NewSourceMock([]string{"a"}, false, false),
			NewSourceMock([]string{"b"}, false, false),
			NewSourceMock([]string{"c"}, false, false),
		},
	}
	for _, sources := range cases {
		trans := NewTransmitter()
		trans.AddSources(sources[0], sources[1], sources[2])
		ctx, cancel := context.WithCancel(context.Background())
		testCtx, testCancel := context.WithCancel(context.Background())

		go func() {
			trans.Run(ctx, cancel)
			testCancel()
		}()
		time.Sleep(time.Millisecond) // let transmitter time to transmit messages
		cancel()
		<-testCtx.Done()

		for i, src := range sources {
			for j, s := range sources {
				isSelf := s == src
				if isSelf {
					continue
				}
				for _, msg := range src.readMsgs {
					if !s.gotMessage(msg) {
						t.Errorf("source %d did not recieve msg from source %d", j, i)
					}
				}
			}
		}
	}
}

func TestTransmitterStopsWhenCannotStartSource(t *testing.T) {
	s := NewSourceMock([]string{}, true, false)
	trans := NewTransmitter()
	trans.AddSources(s)
	ctx, cancel := context.WithCancel(context.Background())

	go trans.Run(ctx, cancel)

	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Millisecond):
		t.Errorf("transmitter did not stop after source exited .Start() method")
	}
}

func TestTransmitterCanAddSourcesWhileRunning(t *testing.T) {
	s1 := NewSourceMock([]string{}, false, false)
	s2 := NewSourceMock([]string{"asd"}, false, false)
	trans := NewTransmitter()
	ctx, cancel := context.WithCancel(context.Background())

	go trans.Run(ctx, cancel)
	trans.AddSources(s1)
	time.Sleep(time.Millisecond) // let trans start Writing goroutine
	trans.AddSources(s2)
	time.Sleep(time.Millisecond) // let trans write message

	if !reflect.DeepEqual(s1.gotMsgs, s2.readMsgs) {
		t.Errorf("transmitter did not send message from added in runtime source")
	}
}

func TestTransmitterStopsAfterContextCancel(t *testing.T) {
	trans := NewTransmitter()
	ctx, cancel := context.WithCancel(context.Background())
	stopped := false

	go func() {
		trans.Run(ctx, cancel)
		stopped = true
	}()

	time.Sleep(3 * time.Millisecond) // wait for trans to unexpectedly stop running
	if stopped {
		t.Errorf("trans stopped right after start or did not start at all")
	}
	cancel()
	time.Sleep(time.Millisecond) // let trans release and close things
	if !stopped {
		t.Errorf("trans did not stop after context canelation")
	}

}
