package source

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"time"
)

type RetrierConfig struct {
	Tries      *int
	Delay      float64
	MaxTimeout float64 // seconds, use math.Inf(1) if you don't want to have a limit
	JitterFunc func() float64
}

type Retrier struct {
	NetworkSource
	cfg RetrierConfig
}

func NewRetrier(source NetworkSource, cfg RetrierConfig) *Retrier {
	return &Retrier{NetworkSource: source, cfg: cfg}
}

func (retrier *Retrier) getTimeoutFunc() func() (next float64) {
	var i float64 = 0
	return func() (next float64) {
		jitter := retrier.cfg.JitterFunc()
		i += retrier.cfg.Delay * jitter
		if i > retrier.cfg.MaxTimeout {
			return retrier.cfg.MaxTimeout + 1*jitter
		}
		return i
	}
}

func (retrier *Retrier) Connect(ctx context.Context) (err error) {
	i := 0
	url := retrier.NetworkSource.GetUrl()
	getTimeout := retrier.getTimeoutFunc()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			logrus.Debugf(
				"connecting to ws on %s attempt %d/%d", url, i+1, retrier.cfg.Tries)
			err = retrier.NetworkSource.Connect(ctx)
			if err != nil {
				if fatalErr, ok := err.(*FatalConnectError); ok {
					logrus.Debugln("retirer error is fatal", err)
					return fatalErr
				}
				logrus.Errorln("connect to ws failed:", err)
				time.Sleep(time.Duration(getTimeout()) * time.Second)
				if retrier.cfg.Tries == nil {
					continue
				}
				i++
				exceeded := i >= *retrier.cfg.Tries
				if exceeded {
					return errors.New("max reconnect attempts exceeded")
				}
			}
			logrus.Debugf("connected to ws on %s", url)
			return nil
		}
	}
}

func (retrier *Retrier) Start(sessionCtx context.Context) error {
	url := retrier.NetworkSource.GetUrl()
	getTimeout := retrier.getTimeoutFunc()
	for {
		select {
		case <-sessionCtx.Done():
			return nil
		default:
			err := retrier.NetworkSource.Consume(sessionCtx)
			if err == nil {
				return nil
			}
			logrus.Debugf("could not read ws on %s", url)
			err = retrier.Connect(sessionCtx)
			if err != nil {
				return err
			}
			time.Sleep(time.Duration(getTimeout()) * time.Second)
		}
	}
}
