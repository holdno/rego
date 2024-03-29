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

type Result struct {
	errors  []error
	succeed bool
}

func (e Result) Error() string {
	sb := strings.Builder{}
	for i, v := range e.errors {
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(":")
		sb.WriteString(v.Error())
		sb.WriteString(";")
	}
	return sb.String()
}

func (e Result) Latest() error {
	size := len(e.errors)
	if size == 0 {
		return nil
	}
	return e.errors[len(e.errors)-1]
}

func (e Result) Errors() []error {
	return e.errors
}

func (e Result) Succeed() bool {
	return e.succeed
}

type rego struct {
	maxTimes      int
	period        time.Duration
	jitter        float64
	backoffFactor float64
	sliding       bool
	resetDuration time.Duration
}

type Option func(r *rego)

func WithPeriod(period time.Duration) Option {
	return func(r *rego) {
		r.period = period
	}
}

func WithJitter(jitter float64) Option {
	return func(r *rego) {
		r.jitter = jitter
	}
}

func WithBackoffFector(backoffFactor float64) Option {
	return func(r *rego) {
		r.backoffFactor = backoffFactor
	}
}

func WithSliding(sliding bool) Option {
	return func(r *rego) {
		r.sliding = sliding
	}
}

func WithResetDuration(reset time.Duration) Option {
	return func(r *rego) {
		r.resetDuration = reset
	}
}

func WithTimes(times int) Option {
	return func(r *rego) {
		r.maxTimes = times
	}
}

func Retry(f func() error, opts ...Option) Result {
	ctx := context.Background()
	return RetryWithContext(ctx, func(ctx context.Context) error { return f() }, opts...)
}

func RetryWithContext(ctx context.Context, f func(ctx context.Context) error, opts ...Option) (rv Result) {
	rg := &rego{
		maxTimes:      DefaultRetryTimes,
		period:        DefaultPeriod,
		jitter:        DefaultJitter,
		backoffFactor: DefaultBackoffFactor,
		resetDuration: DefaultResetDuration,
	}

	for _, opt := range opts {
		opt(rg)
	}

	ctx, cancel := context.WithCancel(ctx)
	var (
		index   int
	)
	withCtx := func() {
		if index >= rg.maxTimes {
			cancel()
			return
		}
		defer func() {
			index++
			if r := recover(); r != nil {
				rv.errors = append(rv.errors, fmt.Errorf("%v", r))
			}
		}()
		err := f(ctx)
		if err != nil {
			rv.errors = append(rv.errors, err)
			return
		}
		rv.succeed = true
		cancel()
	}

	wait.BackoffUntil(withCtx, wait.NewExponentialBackoffManager(rg.period, 0, rg.resetDuration, rg.backoffFactor, rg.jitter, &clock.RealClock{}), rg.sliding, ctx.Done())

	return
}
