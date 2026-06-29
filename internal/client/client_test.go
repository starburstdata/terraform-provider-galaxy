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
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestComputeRetryBackoff_RetryAfterSeconds(t *testing.T) {
	resp := &http.Response{Header: http.Header{"Retry-After": []string{"7"}}}
	got := computeRetryBackoff(resp, 0)
	if got != 7*time.Second {
		t.Fatalf("Retry-After=7 should yield 7s, got %v", got)
	}
}

func TestComputeRetryBackoff_RetryAfterSecondsCappedAtBackoffCap(t *testing.T) {
	resp := &http.Response{Header: http.Header{"Retry-After": []string{"3600"}}}
	got := computeRetryBackoff(resp, 0)
	if got != backoffCap {
		t.Fatalf("Retry-After=3600 should be capped at %v, got %v", backoffCap, got)
	}
}

func TestComputeRetryBackoff_RetryAfterHTTPDateCappedAtBackoffCap(t *testing.T) {
	far := time.Now().Add(1 * time.Hour).UTC().Format(http.TimeFormat)
	resp := &http.Response{Header: http.Header{"Retry-After": []string{far}}}
	got := computeRetryBackoff(resp, 0)
	if got != backoffCap {
		t.Fatalf("Retry-After HTTP-date +1h should be capped at %v, got %v", backoffCap, got)
	}
}

func TestComputeRetryBackoff_RetryAfterHTTPDate(t *testing.T) {
	future := time.Now().Add(5 * time.Second).UTC().Format(http.TimeFormat)
	resp := &http.Response{Header: http.Header{"Retry-After": []string{future}}}
	got := computeRetryBackoff(resp, 0)
	// Allow generous slack: http.TimeFormat has second precision, so the
	// returned duration is between ~4s and ~5s in practice.
	if got <= 3*time.Second || got > 6*time.Second {
		t.Fatalf("Retry-After HTTP-date +5s should be ~5s, got %v", got)
	}
}

func TestComputeRetryBackoff_RetryAfterPastDateFallsThrough(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour).UTC().Format(http.TimeFormat)
	resp := &http.Response{Header: http.Header{"Retry-After": []string{past}}}
	got := computeRetryBackoff(resp, 0)
	// Past date is ignored, falls through to exponential backoff at attempt 0:
	// base = backoffBase = 10s, jitter range [5s, 15s).
	if got < backoffBase/2 || got >= 3*backoffBase/2 {
		t.Fatalf("past Retry-After should fall through to exponential, got %v", got)
	}
}

func TestComputeRetryBackoff_RetryAfterNegativeFallsThrough(t *testing.T) {
	resp := &http.Response{Header: http.Header{"Retry-After": []string{"-5"}}}
	got := computeRetryBackoff(resp, 0)
	if got < backoffBase/2 || got >= 3*backoffBase/2 {
		t.Fatalf("negative Retry-After should fall through to exponential, got %v", got)
	}
}

func TestComputeRetryBackoff_RetryAfterGarbageFallsThrough(t *testing.T) {
	resp := &http.Response{Header: http.Header{"Retry-After": []string{"not-a-thing"}}}
	got := computeRetryBackoff(resp, 0)
	if got < backoffBase/2 || got >= 3*backoffBase/2 {
		t.Fatalf("unparseable Retry-After should fall through to exponential, got %v", got)
	}
}

func TestComputeRetryBackoff_ExponentialWithJitterByAttempt(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
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
		// Run several iterations because of randomness in jitter; assert range.
		for i := 0; i < 50; i++ {
			got := computeRetryBackoff(resp, c.attempt)
			minWait := c.wantBase / 2
			maxWait := 3 * c.wantBase / 2
			if got < minWait || got >= maxWait {
				t.Fatalf("attempt %d: got %v, want range [%v, %v)", c.attempt, got, minWait, maxWait)
			}
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

// Compile-time guard: ensure strconv stays imported via the test file even if
// the package code later stops using it (the production code uses it for
// Retry-After parsing; this is a defensive pin).
var _ = strconv.Atoi
