package server

import (
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/oklog/ulid"
)

type RestServer struct {
	Logger  *slog.Logger
	Entropy *rand.Rand // Entropy source for generating ULIDs.
}

type RestServerOptions struct {
	Address                  string
	OptLogReqeust            bool
	CallbackEndpointRegister func(muxer *http.ServeMux)
}

// NewRestServer creates a new RestServer instance. It requires a logger instance. All other options are set to their defaults.
func NewRestServer(logger *slog.Logger) *RestServer {

	return &RestServer{
		Logger:  logger,
		Entropy: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RestServerOptionsDefaults returns the default options for a RestServer.
func RestServerOptionsDefaults(callbackEndpointRegister func(muxer *http.ServeMux)) *RestServerOptions {

	return &RestServerOptions{
		Address:                  "localhost:8080",
		OptLogReqeust:            true,
		CallbackEndpointRegister: callbackEndpointRegister,
	}
}

// NewULID generates a new ULID using the server's entropy source.
func (srv *RestServer) NewULID() (ulid.ULID, error) {

	return ulid.New(ulid.Timestamp(time.Now()), srv.Entropy)
}

// Start starts the RestServer on the given address/options. Set options to nil to use the defaults.
func (srv *RestServer) Start(opts *RestServerOptions) error {

	if opts == nil {
		opts = RestServerOptionsDefaults(nil)
	}

	muxer := http.NewServeMux()

	if opts.CallbackEndpointRegister != nil {
		opts.CallbackEndpointRegister(muxer)
	}

	// === create http server ===
	httpserver := &http.Server{
		Addr: opts.Address,
	}

	if opts.OptLogReqeust {
		httpserver.Handler = srv.logRequestMiddleware(muxer)
	} else {
		httpserver.Handler = muxer
	}

	address := opts.Address

	// if address starts with a colon, prepend localhost
	if address[0] == ':' {
		address = "localhost" + address
	}

	srv.Logger.Info("starting server on http://" + address)
	return httpserver.ListenAndServe()
}
