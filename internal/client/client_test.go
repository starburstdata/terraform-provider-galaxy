// Copyright Starburst Data, Inc. All rights reserved.
//
// The source code is the proprietary and confidential information of Starburst Data, Inc. and
// may be used only for reference purposes in connection with the Terraform Registry. All rights,
// title, interest and ownership of the code and any derivatives, updates, upgrades, enhancements
// and modifications thereof remain with Starburst Data, Inc. You are not permitted to distribute,
// disclose, sell, lease, transfer, assign, modify, create derivative works of, or sublicense the
// code, or use the code to create or develop any products or services.

package client

import (
	"context"
	"testing"
	"time"
)

func TestComputeRetryBackoff_ExponentialWithJitterByAttempt(t *testing.T) {
	cases := []struct {
		attempt  int
		wantBase time.Duration
	}{
		{0, backoffBase},      // 10s
		{1, backoffBase << 1}, // 20s
		{2, backoffBase << 2}, // 40s
		{3, backoffCap},       // would be 80s, capped at 60s
		{4, backoffCap},       // capped
	}
	for _, c := range cases {
		for i := 0; i < 50; i++ {
			got := computeRetryBackoff(c.attempt, backoffBase)
			minWait := c.wantBase / 2
			maxWait := 3 * c.wantBase / 2
			if got < minWait || got >= maxWait {
				t.Fatalf("attempt %d: got %v, want range [%v, %v)", c.attempt, got, minWait, maxWait)
			}
		}
	}
}

func TestComputeRetryBackoff_ClusterBase(t *testing.T) {
	// clusterBackoffBase=60s means attempt 0 yields [30s, 60s), capped at backoffCap.
	for i := 0; i < 50; i++ {
		got := computeRetryBackoff(0, clusterBackoffBase)
		if got < clusterBackoffBase/2 || got > backoffCap {
			t.Fatalf("cluster attempt 0: got %v, want [%v, %v]", got, clusterBackoffBase/2, backoffCap)
		}
	}
}

func TestSleepCtx_NonPositiveReturnsImmediately(t *testing.T) {
	start := time.Now()
	if err := sleepCtx(context.Background(), 0); err != nil {
		t.Fatalf("sleepCtx(0) returned error: %v", err)
	}
	if err := sleepCtx(context.Background(), -1*time.Second); err != nil {
		t.Fatalf("sleepCtx(-1s) returned error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
		t.Fatalf("sleepCtx with non-positive duration blocked for %v", elapsed)
	}
}

func TestSleepCtx_CompletesAfterDuration(t *testing.T) {
	start := time.Now()
	if err := sleepCtx(context.Background(), 20*time.Millisecond); err != nil {
		t.Fatalf("sleepCtx returned error: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 20*time.Millisecond {
		t.Fatalf("sleepCtx returned before duration elapsed: %v", elapsed)
	}
}

func TestSleepCtx_RespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	err := sleepCtx(ctx, 10*time.Second)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("sleepCtx did not return promptly on cancel: %v", elapsed)
	}
}
