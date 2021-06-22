package rego

import (
	"context"
	"testing"
	"time"
)

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
