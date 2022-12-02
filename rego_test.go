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
	if rv := err.(Result); rv.Succeed() {
		t.Error("should not success")
	}
}

func TestRetry(t *testing.T) {
	beforeTime := time.Now()
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	err := RetryWithContext(ctx, func(ctx context.Context) error {
		t.Log(time.Now().Sub(beforeTime))
		beforeTime = time.Now()
		panic("for retry")
	})
	t.Log(err)
	if rv := err.(Result); rv.Succeed() {
		t.Error("should not success")
	}
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
	t.Fatal("should not success")
}

func TestReturnLatestErrorIfFail(t *testing.T) {
	var index int64
	err := Retry(func() error {
		index++
		panic("panic")
	}, WithTimes(3), WithLatestError())

	if err != nil && index == 3 {
		t.Log("success with panic:", err)
		return
	}
	t.Fatal("should not success")
}

func TestReturnNilIfRetrySuccess(t *testing.T) {
	var index int64
	err := Retry(func() error {
		index++
		if index > 1 {
			return nil
		}
		panic("panic")
	}, WithTimes(3), WithLatestError())

	if err != nil || index != 2 {
		t.Fatal("something went wrong", err, index)
	}
}

func TestReturnIfRetrySuccess(t *testing.T) {
	var index int64
	err := Retry(func() error {
		index++
		if index > 1 {
			return nil
		}
		panic("panic")
	}, WithTimes(3))

	if err == nil || index != 2 {
		t.Fatal("something went wrong", err, index)
	}

	if rv := err.(Result); !rv.Succeed() {
		t.Error("should success")
	}
}
