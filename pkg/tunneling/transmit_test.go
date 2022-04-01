package tunneling

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type SourceMock struct {
	out                        chan []byte
	gotMsgs                    []string
	readMsgs                   []string
	failsConsumeBeforeMessages bool
	failsConsumeAfterMessages  bool
}

func NewSourceMock(readMsgs []string, failsConsumeBeforeMessages bool, failsConsumeAfterMessages bool) *SourceMock {
	return &SourceMock{
		out:                        make(chan []byte),
		readMsgs:                   readMsgs,
		gotMsgs:                    make([]string, 0),
		failsConsumeBeforeMessages: failsConsumeBeforeMessages,
		failsConsumeAfterMessages:  failsConsumeAfterMessages,
	}
}

func (s SourceMock) Consume(ctx context.Context) (err error) {
	if s.failsConsumeBeforeMessages {
		return errors.New("failed before")
	}
	for _, msg := range s.readMsgs {
		s.out <- []byte(msg)
	}
	if s.failsConsumeAfterMessages {
		return errors.New("failed after")
	}
	<-ctx.Done()
	return nil
}

func (s SourceMock) GetReader() chan []byte {
	return s.out
}

func (s *SourceMock) Write(bytes []byte) (err error) {
	s.gotMsgs = append(s.gotMsgs, string(bytes))
	return nil
}

func (s *SourceMock) gotMessage(msg string) bool {
	for _, m := range s.gotMsgs {
		if m == msg {
			return true
		}
	}
	return false
}

func TestTransmitterSendsMessages(t *testing.T) {
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
		time.Sleep(time.Millisecond) // let transmitter time to transmit
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
