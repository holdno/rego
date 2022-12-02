package rego

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	rv := Retry(func() error {
		t.Log(time.Now().Unix())
		return errors.New("fake error")
	}, WithBackoffFector(2), WithPeriod(time.Second), WithResetDuration(time.Minute))
	t.Log(rv)
	if rv.Succeed() {
		t.Error("should not success")
	}
}

func TestRetry(t *testing.T) {
	beforeTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	rv := RetryWithContext(ctx, func(ctx context.Context) error {
		t.Log(time.Since(beforeTime))
		beforeTime = time.Now()
		panic("for retry")
	})
	t.Log(rv)
	if rv.Succeed() || rv.Latest() == nil || len(rv.Errors()) == 0 {
		t.Error("should not success")
	}
}

func TestSuccess(t *testing.T) {
	rv := Retry(func() error {
		t.Log("only saw once")
		return nil
	})
	t.Log(rv)
	if !rv.Succeed() || rv.Latest() != nil || len(rv.Errors()) > 0 {
		t.Error("should success")
	}

}

func TestPanic(t *testing.T) {
	var index int64
	rv := Retry(func() error {
		index++
		panic("panic")
	}, WithTimes(3))

	t.Log(rv)
	if rv.Succeed() || rv.Latest() == nil || len(rv.Errors()) == 0 {
		t.Error("should not success")
	}
	if len(rv.Errors()) != 3 {
		t.Error("wrong errors count")
	}
	if index != 3 {
		t.Error("wrong count")
	}
}

func TestReturnIfRetrySuccess(t *testing.T) {
	var index int64
	rv := Retry(func() error {
		if index > 1 {
			return nil
		}
		index++
		panic("panic")
	}, WithTimes(3))
	t.Log(rv)
	if !rv.Succeed() {
		t.Error("should success")
	}
	if len(rv.Errors()) != 2 {
		t.Error("wrong errors count")
	}
	if index != 2 {
		t.Error("wrong count")
	}
}
