package source

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
	"time"
)

type RetryPolicy struct {
	// Tries is max number of attempts that retrier will do before return error.
	// If null - retrier will try to connect infinitely
	Tries *int
	// Delay is a number in seconds that used to increase the timeout between tries.
	// (retrier uses arithmetical progression)
	Delay float64
	// MaxTimeout is a number in seconds that defines maximum delay between tries.
	// Use math.Inf(1) if you don't want to have such limit
	MaxTimeout float64
	// JitterFunc is a function that returns a small divergent number by which
	// each delay timeout is multiplied.
	// It is used not to let DDOS manager server by a lot of simultaneous connections
	// after manager restarts.
	// Use
	JitterFunc func() float64
}

// DefaultPolicy is a policy with all fields set to the default values
var DefaultPolicy = &RetryPolicy{
	Tries:      nil, // infinite tries limit
	Delay:      1,   // increase delay by 1 each time
	MaxTimeout: 60,  // don't let timeout be more than 60 sec
	JitterFunc: rand.ExpFloat64,
}

type Retrier struct {
	NetworkSource
	policy RetryPolicy
	logger *logrus.Logger
}

func NewRetrier(source NetworkSource, policy RetryPolicy, logger *logrus.Logger) *Retrier {
	return &Retrier{NetworkSource: source, policy: policy, logger: logger}
}

func (retrier *Retrier) getTimeoutFunc() func() (next float64) {
	var timeout float64 = 0
	return func() (next float64) {
		jitter := retrier.policy.JitterFunc()
		timeout += retrier.policy.Delay * jitter
		if timeout > retrier.policy.MaxTimeout {
			return retrier.policy.MaxTimeout + 1*jitter
		}
		return timeout
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
			retrier.logger.Debugf(
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
					retrier.logger.Debugf("retrier error is fatal: %s", err)
					return fatalErr
				}
				i++
				retrier.logger.Errorf("retrier connect to source on %s failed: %s", url, err)
				timeout := time.Duration(getTimeout())
				time.Sleep(timeout * time.Second)
				endlessRetry := retrier.policy.Tries == nil
				if endlessRetry {
					continue
				}
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
				retrier.logger.Debugf("retier could not consume source on %s: %s", url, err.Error())
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
