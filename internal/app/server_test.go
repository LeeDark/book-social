package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestRunServerGracefullyShutsDownWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listenStarted := make(chan struct{})
	stopListening := make(chan struct{})
	shutdownDeadline := make(chan time.Time, 1)
	server := fakeHTTPServer{
		listenAndServe: func() error {
			close(listenStarted)
			<-stopListening
			return http.ErrServerClosed
		},
		shutdown: func(shutdownCtx context.Context) error {
			deadline, _ := shutdownCtx.Deadline()
			shutdownDeadline <- deadline
			close(stopListening)
			return nil
		},
	}

	result := make(chan error, 1)
	go func() {
		result <- runServer(ctx, testServerLogger(), server)
	}()

	<-listenStarted
	cancel()

	if err := <-result; err != nil {
		t.Fatalf("runServer() error = %v, want nil", err)
	}

	deadline := <-shutdownDeadline
	if deadline.IsZero() {
		t.Fatal("shutdown context has no deadline")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > 5*time.Second {
		t.Fatalf("shutdown deadline remaining = %s, want (0s, 5s]", remaining)
	}
}

func TestRunServerReturnsNilWhenServerIsAlreadyClosed(t *testing.T) {
	server := fakeHTTPServer{
		listenAndServe: func() error { return http.ErrServerClosed },
		shutdown: func(context.Context) error {
			t.Fatal("Shutdown() must not be called when ListenAndServe() has already stopped")
			return nil
		},
	}

	if err := runServer(context.Background(), testServerLogger(), server); err != nil {
		t.Fatalf("runServer() error = %v, want nil", err)
	}
}

func TestRunServerReturnsListenError(t *testing.T) {
	wantErr := errors.New("listen failed")
	server := fakeHTTPServer{
		listenAndServe: func() error { return wantErr },
		shutdown:       func(context.Context) error { return nil },
	}

	if err := runServer(context.Background(), testServerLogger(), server); !errors.Is(err, wantErr) {
		t.Fatalf("runServer() error = %v, want %v", err, wantErr)
	}
}

func TestRunServerReturnsShutdownError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wantErr := errors.New("shutdown timed out")
	listenStarted := make(chan struct{})
	stopListening := make(chan struct{})
	server := fakeHTTPServer{
		listenAndServe: func() error {
			close(listenStarted)
			<-stopListening
			return http.ErrServerClosed
		},
		shutdown: func(context.Context) error {
			close(stopListening)
			return wantErr
		},
	}

	result := make(chan error, 1)
	go func() {
		result <- runServer(ctx, testServerLogger(), server)
	}()

	<-listenStarted
	cancel()

	if err := <-result; !errors.Is(err, wantErr) {
		t.Fatalf("runServer() error = %v, want %v", err, wantErr)
	}
}

func TestRunServerReturnsListenerErrorAfterShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wantErr := errors.New("listener failed while shutting down")
	listenStarted := make(chan struct{})
	stopListening := make(chan struct{})
	server := fakeHTTPServer{
		listenAndServe: func() error {
			close(listenStarted)
			<-stopListening
			return wantErr
		},
		shutdown: func(context.Context) error {
			close(stopListening)
			return nil
		},
	}

	result := make(chan error, 1)
	go func() {
		result <- runServer(ctx, testServerLogger(), server)
	}()

	<-listenStarted
	cancel()

	if err := <-result; !errors.Is(err, wantErr) {
		t.Fatalf("runServer() error = %v, want %v", err, wantErr)
	}
}

type fakeHTTPServer struct {
	listenAndServe func() error
	shutdown       func(context.Context) error
}

func (s fakeHTTPServer) ListenAndServe() error {
	return s.listenAndServe()
}

func (s fakeHTTPServer) Shutdown(ctx context.Context) error {
	return s.shutdown(ctx)
}

func testServerLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
