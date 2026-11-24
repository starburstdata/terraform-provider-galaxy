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
	"errors"
	"io"
	"net/http"
	"strings"
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

type mockRoundTripper struct {
	requestCount int
	statusCode   int
	method       string
	attempts     []string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.requestCount++
	m.method = req.Method
	m.attempts = append(m.attempts, req.Method)
	return &http.Response{
		StatusCode: m.statusCode,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("{}")),
		Request:    req,
	}, nil
}

func TestPost500NotRetried(t *testing.T) {
	mock := &mockRoundTripper{statusCode: http.StatusInternalServerError}
	client := &GalaxyClient{
		BaseURL:         "http://localhost",
		ClientID:        "test",
		ClientSecret:    "test",
		ProviderVersion: "1.0.0",
		HTTPClient: &http.Client{
			Transport: mock,
		},
		accessToken: "test-token",
		tokenExpiry: time.Now().Add(1 * time.Hour),
	}

	ctx := context.Background()
	var result map[string]interface{}
	err := client.doRequestWithRetry(ctx, http.MethodPost, "/test", nil, &result, 3)

	if err == nil {
		t.Fatalf("expected error for 500 POST, got nil")
	}
	if mock.requestCount != 1 {
		t.Fatalf("POST 500 should not be retried; expected 1 request, got %d", mock.requestCount)
	}
}

func TestRequestTimeout(t *testing.T) {
	if got := requestTimeout(http.MethodPost, dataQualityCreatePath); got != dataQualityCreateTimeout {
		t.Fatalf("DQ create timeout = %s, want %s", got, dataQualityCreateTimeout)
	}
	if got := requestTimeout(http.MethodGet, dataQualityCreatePath); got != defaultRequestTimeout {
		t.Fatalf("GET timeout = %s, want %s", got, defaultRequestTimeout)
	}
	if got := requestTimeout(http.MethodPost, "/public/api/v1/other"); got != defaultRequestTimeout {
		t.Fatalf("other POST timeout = %s, want %s", got, defaultRequestTimeout)
	}
}

func TestResponseRetryDelay(t *testing.T) {
	now := time.Date(2026, time.November, 23, 18, 0, 0, 0, time.UTC)
	cases := []struct {
		name       string
		method     string
		status     int
		retryAfter string
		wantRetry  bool
	}{
		{"get 500", http.MethodGet, 500, "", true},
		{"get 502", http.MethodGet, 502, "", true},
		{"put 503", http.MethodPut, 503, "", true},
		{"patch 504", http.MethodPatch, 504, "", true},
		{"delete 503", http.MethodDelete, 503, "", true},
		{"hinted conflict", http.MethodPatch, 409, "20", true},
		{"permanent conflict", http.MethodPatch, 409, "", false},
		{"zero conflict hint", http.MethodPatch, 409, "0", false},
		{"invalid conflict hint", http.MethodPatch, 409, "invalid", false},
		{"post 500", http.MethodPost, 500, "", false},
		{"post 503", http.MethodPost, 503, "20", false},
		{"post conflict", http.MethodPost, 409, "20", false},
		{"post rate limit", http.MethodPost, 429, "0", true},
		{"not found", http.MethodGet, 404, "", false},
		{"forbidden", http.MethodGet, 403, "", false},
		{"unsupported method", http.MethodConnect, 503, "", false},
		{"oversized hint", http.MethodGet, 503, "61", false},
		{"overflow hint", http.MethodGet, 503, "9223372036854775807", false},
		{"date hint", http.MethodGet, 503, now.Add(20 * time.Second).Format(http.TimeFormat), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			delay, retry := responseRetryDelay(tc.method, tc.status, tc.retryAfter, 0, backoffBase, now)
			if retry != tc.wantRetry {
				t.Fatalf("retry = %v, want %v", retry, tc.wantRetry)
			}
			if retry && (delay < backoffBase/2 || delay > backoffCap) {
				t.Fatalf("delay = %s, outside bounded backoff", delay)
			}
			if retry && (tc.retryAfter == "20" || tc.name == "date hint") && delay < 20*time.Second {
				t.Fatalf("delay = %s, shorter than Retry-After", delay)
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, time.November, 23, 18, 0, 0, 0, time.UTC)
	for _, value := range []string{"", "0", "-1", "invalid", now.Add(-time.Second).Format(http.TimeFormat)} {
		if _, ok := parseRetryAfter(value, now); ok {
			t.Fatalf("accepted invalid or nonpositive Retry-After %q", value)
		}
	}
	if delay, ok := parseRetryAfter(" 5 ", now); !ok || delay != 5*time.Second {
		t.Fatalf("delay = %s, valid = %v", delay, ok)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type trackedResponseBody struct {
	io.Reader
	closes  int
	onClose func()
}

func (b *trackedResponseBody) Close() error {
	b.closes++
	if b.onClose != nil {
		b.onClose()
	}
	return nil
}

func retryTestClient(transport http.RoundTripper) *GalaxyClient {
	return &GalaxyClient{
		BaseURL:         "http://localhost",
		ProviderVersion: "test",
		HTTPClient:      &http.Client{Transport: transport},
		accessToken:     "test-token",
		tokenExpiry:     time.Now().Add(time.Hour),
	}
}

func TestRetryClosesResponseBeforeBackoffAndHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	body := &trackedResponseBody{Reader: strings.NewReader("{}"), onClose: cancel}
	requests := 0
	client := retryTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		return &http.Response{StatusCode: 503, Header: http.Header{}, Body: body, Request: req}, nil
	}))
	err := client.doRequestWithRetry(ctx, http.MethodGet, "/test", nil, nil, maxRequestRetries)
	if !errors.Is(err, context.Canceled) || requests != 1 || body.closes != 1 {
		t.Fatalf("error = %v, requests = %d, closes = %d", err, requests, body.closes)
	}
}

func TestGatewayRetrySucceeds(t *testing.T) {
	body := &trackedResponseBody{Reader: strings.NewReader("{}")}
	requests := 0
	waits := 0
	client := retryTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if requests == 1 {
			return &http.Response{StatusCode: 502, Header: http.Header{}, Body: body, Request: req}, nil
		}
		if body.closes != 1 {
			t.Fatal("previous response not closed before retry")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"id":"test"}`)), Request: req}, nil
	}))
	client.retryWait = func(ctx context.Context, delay time.Duration) error {
		waits++
		if delay < backoffBase/2 || delay > backoffCap {
			t.Fatalf("unexpected retry delay %s", delay)
		}
		if body.closes != 1 {
			t.Fatal("previous response not closed before backoff")
		}
		return ctx.Err()
	}
	var result map[string]interface{}
	if err := client.doRequestWithRetry(context.Background(), http.MethodGet, "/test", nil, &result, maxRequestRetries); err != nil {
		t.Fatal(err)
	}
	if requests != 2 || waits != 1 || result["id"] != "test" {
		t.Fatalf("requests = %d, waits = %d, result = %v", requests, waits, result)
	}
}

func TestGatewayRetryBudgetExhausted(t *testing.T) {
	mock := &mockRoundTripper{statusCode: 503}
	client := retryTestClient(mock)
	if err := client.doRequestWithRetry(context.Background(), http.MethodGet, "/test", nil, nil, 0); err == nil || mock.requestCount != 1 {
		t.Fatalf("error = %v, requests = %d", err, mock.requestCount)
	}
}

func TestPostFailuresAreNotReplayed(t *testing.T) {
	for _, status := range []int{409, 500, 502, 503, 504} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			requests := 0
			client := retryTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				requests++
				return &http.Response{StatusCode: status, Header: http.Header{"Retry-After": {"1"}}, Body: io.NopCloser(strings.NewReader("{}")), Request: req}, nil
			}))
			if err := client.doRequestWithRetry(context.Background(), http.MethodPost, dataQualityCreatePath, nil, nil, maxRequestRetries); err == nil || requests != 1 {
				t.Fatalf("error = %v, requests = %d", err, requests)
			}
		})
	}
	requests := 0
	client := retryTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		return nil, context.DeadlineExceeded
	}))
	if err := client.doRequestWithRetry(context.Background(), http.MethodPost, dataQualityCreatePath, nil, nil, maxRequestRetries); !errors.Is(err, context.DeadlineExceeded) || requests != 1 {
		t.Fatalf("error = %v, requests = %d", err, requests)
	}
}

func TestRequestContextDeadlines(t *testing.T) {
	for _, tc := range []struct {
		name    string
		path    string
		timeout time.Duration
	}{
		{"data quality create", dataQualityCreatePath, dataQualityCreateTimeout},
		{"cluster create", "/public/api/v1/cluster?extended=true", defaultRequestTimeout},
		{"token", "/oauth/v2/token", defaultRequestTimeout},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			client := retryTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				deadline, ok := req.Context().Deadline()
				if !ok || deadline.Before(start.Add(tc.timeout)) || deadline.After(time.Now().Add(tc.timeout)) {
					t.Fatalf("unexpected deadline %s, deadline present = %v", deadline, ok)
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"access_token":"test","expires_in":3600}`)), Request: req}, nil
			}))
			var err error
			if tc.name == "token" {
				err = client.getAccessToken(context.Background())
			} else {
				err = client.doRequestWithRetry(context.Background(), http.MethodPost, tc.path, nil, nil, 0)
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	parentDeadline, _ := ctx.Deadline()
	client := retryTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if deadline, _ := req.Context().Deadline(); deadline != parentDeadline {
			t.Fatalf("request deadline %s does not preserve parent %s", deadline, parentDeadline)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("{}")), Request: req}, nil
	}))
	if err := client.doRequestWithRetry(ctx, http.MethodPost, dataQualityCreatePath, nil, nil, 0); err != nil {
		t.Fatal(err)
	}
}
