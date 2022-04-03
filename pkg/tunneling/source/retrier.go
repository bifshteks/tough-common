package source

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type RetryPolicy struct {
	Tries      *int
	Delay      float64
	MaxTimeout float64 // seconds, use math.Inf(1) if you don't want to have a limit
	jitterFunc func() float64
}

type Retrier struct {
	NetworkSource
	policy RetryPolicy
}

func NewRetrier(source NetworkSource, policy RetryPolicy) *Retrier {
	return &Retrier{NetworkSource: source, policy: policy}
}

func (retrier *Retrier) getTimeoutFunc() func() (next float64) {
	var i float64 = 0
	return func() (next float64) {
		jitter := retrier.policy.jitterFunc()
		i += retrier.policy.Delay * jitter
		if i > retrier.policy.MaxTimeout {
			return retrier.policy.MaxTimeout + 1*jitter
		}
		return i
	}
}

func (retrier *Retrier) Connect(ctx context.Context) (err error) {
	i := 0
	url := retrier.NetworkSource.GetUrl()
	getTimeout := retrier.getTimeoutFunc()
	var maxTriesStr string
	if retrier.policy.Tries == nil {
		maxTriesStr = "inf"
	} else {
		maxTriesStr = strconv.Itoa(*retrier.policy.Tries)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			logrus.Debugf(
				"retrier connecting to source on %s, attempt %d/%s",
				url, i+1, maxTriesStr)
			err = retrier.NetworkSource.Connect(ctx)
			// handle case when err is "context expiration"
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			if err != nil {
				if fatalErr, ok := err.(*FatalConnectError); ok {
					logrus.Debugf("retrier error is fatal: %s", err)
					return fatalErr
				}
				logrus.Errorf("retrier connect to source on %s failed: %s", url, err)
				timeout := time.Duration(getTimeout())
				time.Sleep(timeout * time.Second)
				endlessRetry := retrier.policy.Tries == nil
				if endlessRetry {
					continue
				}
				i++
				triesExceeded := i >= *retrier.policy.Tries
				if triesExceeded {
					return errors.New("max reconnect attempts exceeded")
				}
			}
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
			select {
			case <-sessionCtx.Done():
				return nil
			default:
			}
			if err != nil {
				logrus.Debugf("retier could not consume source on %s: %s", url, err.Error())
				err = retrier.Connect(sessionCtx)
				if err != nil {
					return err
				}
				timeout := time.Duration(getTimeout())
				time.Sleep(timeout * time.Second)
				continue
			}
			return nil
		}
	}
}
