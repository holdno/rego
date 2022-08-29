package rego

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	err := Retry(func() error {
		t.Log(time.Now().Unix())
		return errors.New("fake error")
	}, WithBackoffFector(2), WithPeriod(time.Second), WithResetDuration(time.Minute))
	t.Log(err)
}

func TestRetry(t *testing.T) {
	beforeTime := time.Now()
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	errs := RetryWithContext(ctx, func(ctx context.Context) error {
		t.Log(time.Now().Sub(beforeTime))
		beforeTime = time.Now()
		panic("for retry")
	})
	t.Log(errs)
}

func TestSuccess(t *testing.T) {
	err := Retry(func() error {
		t.Log("only saw once")
		return nil
	})
	if err != nil {
		t.Fatal("emm")
	}
	t.Log("double success")
}

func TestPanic(t *testing.T) {
	var index int64
	err := Retry(func() error {
		index++
		panic("panic")
	}, WithTimes(3))

	if err != nil && index == 3 {
		t.Log("success with three panics:", err)
		return
	}
	t.Log("some wrong")
}
