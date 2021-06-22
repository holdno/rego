package rego

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	DefaultResetDuration = 3 * time.Second
	DefaultJitter        = 0.0
	DefaultBackoffFactor = 1.5
	DefaultRetryTimes    = 5
	DefaultPeriod        = 100 * time.Millisecond
)

type errlist []error

func (e errlist) Error() string {
	sb := strings.Builder{}
	for i, v := range e {
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(":")
		sb.WriteString(v.Error())
		sb.WriteString(";")
	}
	return sb.String()
}

func (e errlist) Latest() error {
	return e[len(e)-1]
}

func Retry(f func() error, opts ...option) errlist {
	ctx := context.Background()
	return RetryWithContext(ctx, func(ctx context.Context) error { return f() }, opts...)
}

type rego struct {
	maxTimes      int
	period        time.Duration
	jitter        float64
	backoffFector float64
	sliding       bool
	resetDuration time.Duration
}

type option func(r *rego)

func WithPeriod(period time.Duration) option {
	return func(r *rego) {
		r.period = period
	}
}

func WithJitter(jitter float64) option {
	return func(r *rego) {
		r.jitter = jitter
	}
}

func WithBackoffFector(backoffFactor float64) option {
	return func(r *rego) {
		r.backoffFector = backoffFactor
	}
}

func WithSliding(sliding bool) option {
	return func(r *rego) {
		r.sliding = sliding
	}
}

func WithResetDuration(reset time.Duration) option {
	return func(r *rego) {
		r.resetDuration = reset
	}
}

func WithTimes(times int) option {
	return func(r *rego) {
		r.maxTimes = times
	}
}

func RetryWithContext(ctx context.Context, f func(ctx context.Context) error, opts ...option) errlist {
	rg := &rego{
		maxTimes:      DefaultRetryTimes,
		period:        DefaultPeriod,
		jitter:        DefaultJitter,
		backoffFector: DefaultBackoffFactor,
		resetDuration: DefaultResetDuration,
	}

	for _, opt := range opts {
		opt(rg)
	}

	ctx, cancel := context.WithCancel(ctx)
	var (
		errs  errlist
		index int
	)
	withCtx := func() {
		defer func() {
			index++
			if r := recover(); r != nil {
				errs = append(errs, fmt.Errorf("%v", r))
			}
		}()
		err := f(ctx)
		if err != nil {
			errs = append(errs, err)
			if rg.maxTimes > index {
				return
			}
		}
		cancel()
	}

	wait.BackoffUntil(withCtx, wait.NewExponentialBackoffManager(rg.period, 0, rg.resetDuration, rg.backoffFector, rg.jitter, &clock.RealClock{}), rg.sliding, ctx.Done())
	return errs
}