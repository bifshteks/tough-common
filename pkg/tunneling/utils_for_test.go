package tunneling

import (
	"context"
	"errors"
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
