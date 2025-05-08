package web

import (
	"context"
	"github.com/resend/resend-go/v2"
	"github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-limiter/memorystore"
	"lockbox-webserver/db"
	"net/http"
	"os"
	"time"
)

type HTTPServer struct {
	hostname string

	dbPool       *db.Pool
	jwtSecretKey []byte
	resendClient *resend.Client

	createAccountLimiter limiter.Store
}

func NewHTTPServer(hostname string, dbPool *db.Pool, jwtSecretKey []byte) (server *HTTPServer, err error) {
	createAccountLimiter, err := memorystore.New(&memorystore.Config{
		Tokens:   1,
		Interval: 15 * time.Minute,
	})
	if err != nil {
		return
	}

	resendClient := resend.NewClient(os.Getenv("RESEND_API_KEY"))

	server = &HTTPServer{
		hostname:             hostname,
		dbPool:               dbPool,
		jwtSecretKey:         jwtSecretKey,
		resendClient:         resendClient,
		createAccountLimiter: createAccountLimiter,
	}

	return
}

// Run runs a Server.
func (s *HTTPServer) Run(ctx context.Context, addr string) (err error) {
	ginEngine, err := s.setupRoutes()
	if err != nil {
		return
	}

	// Spin up the server
	srv := &http.Server{Addr: addr, Handler: ginEngine}
	errChan := make(chan error)
	go func() { errChan <- srv.ListenAndServe() }()

	// Wait for a close event
	for {
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			srv.Shutdown(context.Background())
			<-errChan
			return nil
		}
	}
}
