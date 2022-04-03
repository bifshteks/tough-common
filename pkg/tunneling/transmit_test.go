package tunneling

import (
	"context"
	"github.com/bifshteks/tough_common/pkg/logutil"
	"github.com/bifshteks/tough_common/pkg/tunneling/source"
	requirement "github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSourceMockImplementsInterfaces(t *testing.T) {
	mock := NewNormalSourceMock([]string{})
	var _ source.Source = mock
	var _ source.NetworkSource = mock
}
func TestTransmitterSendsMessagesToEverySource(t *testing.T) {
	require := requirement.New(t)
	cases := [][3]*SourceMock{
		{
			NewNormalSourceMock([]string{}),
			NewNormalSourceMock([]string{"asd", "qwe"}),
			NewNormalSourceMock([]string{})},
		{
			NewNormalSourceMock([]string{"a"}),
			NewNormalSourceMock([]string{"b"}),
			NewNormalSourceMock([]string{"c"}),
		},
	}
	for _, sources := range cases {
		trans := NewTransmitter(logutil.DummyLogger)
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
					require.True(s.gotMessage(msg), "source %d did not receive msg from source %d", j, i)
				}
			}
		}
	}
}

func TestTransmitterStopsWhenCannotStartSource(t *testing.T) {
	s := NewSourceMock(
		[]string{},
		true, false,
		0)
	trans := NewTransmitter(logutil.DummyLogger)
	trans.AddSources(s)
	ctx, cancel := context.WithCancel(context.Background())

	go trans.Run(ctx, cancel)

	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Millisecond):
		t.Errorf("transmitter did not stop after source exited .Consume() method")
	}
}

func TestTransmitterCanAddSourcesWhileRunning(t *testing.T) {
	require := requirement.New(t)
	s1 := NewNormalSourceMock([]string{})
	s2 := NewNormalSourceMock([]string{"asd"})
	trans := NewTransmitter(logutil.DummyLogger)
	ctx, cancel := context.WithCancel(context.Background())

	go trans.Run(ctx, cancel)
	trans.AddSources(s1)
	time.Sleep(time.Millisecond) // let trans start Writing goroutine
	trans.AddSources(s2)
	time.Sleep(time.Millisecond) // let trans write message

	require.Equal(s1.gotMsgs, s2.readMsgs, "transmitter did not send message from added in runtime source")
}

func TestTransmitterStopsAfterContextCancel(t *testing.T) {
	require := requirement.New(t)
	trans := NewTransmitter(logutil.DummyLogger)
	ctx, cancel := context.WithCancel(context.Background())
	stopped := false

	go func() {
		trans.Run(ctx, cancel)
		stopped = true
	}()

	time.Sleep(3 * time.Millisecond) // wait for trans to unexpectedly stop running
	require.False(stopped, "trans stopped right after start")

	cancel()
	time.Sleep(time.Millisecond) // let trans release and close things
	require.True(stopped, "trans did not stop after context cancelation")
}

func TestTransmitterRunEndsAfterSourcesStopConsuming(t *testing.T) {
	require := requirement.New(t)
	stopDelay := 100
	s := NewSourceMock(
		[]string{},
		false, false,
		stopDelay)
	trans := NewTransmitter(logutil.DummyLogger)
	ctx, cancel := context.WithCancel(context.Background())
	trans.AddSources(s)

	go func() {
		time.Sleep(2 * time.Millisecond) // wait for trans to start
		cancel()
	}()
	trans.Run(ctx, cancel)

	require.False(s.consuming, "source still consuming after trans.Run() ends")
}
