package restserver

import (
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/oklog/ulid"
	"gorm.io/gorm"
)

type Server struct {
	db      *gorm.DB
	Entropy *rand.Rand // Entropy source for generating ULIDs.
	Options *ServerOptions
}

type ServerOptions struct {
	BUILDDIR_TEMPLATE string
}

func NewServer(db *gorm.DB, options *ServerOptions) (*Server, error) {

	err := AutoMigrate(db)

	if err != nil {
		return nil, err
	}

	server := &Server{
		db:      db,
		Entropy: rand.New(rand.NewSource(time.Now().UnixNano())),
		Options: options,
	}

	return server, nil
}

func (srv *Server) RegisterEndpoints(muxer *http.ServeMux) {

	path := "/api/v1/"

	// Differences between HTTP methods: GET, POST, PUT, DELETE

	/*
	   GET:
	   - Purpose: Retrieve data from a server.
	   - Idempotency: Yes, making multiple identical requests has the same effect as a single request.
	   - Cacheable: Yes, responses can be cached.
	   - Use Case: Fetching a webpage, querying a database.

	   POST:
	   - Purpose: Send data to a server to create or update a resource.
	   - Idempotency: No, making multiple identical requests may result in multiple resources being created.
	   - Cacheable: No, responses are not cacheable.
	   - Use Case: Submitting a form, uploading a file.

	   PUT:
	   - Purpose: Update a resource or create it if it doesn't exist.
	   - Idempotency: Yes, making multiple identical requests has the same effect as a single request.
	   - Cacheable: No, responses are not cacheable.
	   - Use Case: Updating a user profile, replacing a file.

	   DELETE:
	   - Purpose: Remove a resource from the server.
	   - Idempotency: Yes, making multiple identical requests has the same effect as a single request.
	   - Cacheable: No, responses are not cacheable.
	   - Use Case: Deleting a record, removing a file.
	*/

	muxer.HandleFunc(path+"ping", func(w http.ResponseWriter, r *http.Request) {

		logger := r.Context().Value("logger").(*slog.Logger)
		logger = logger.With("func", "restserver.RegisterEndpoints")
		logger.Debug("Hello from /ping handler")

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("pong"))

		if err != nil {
			logger.Error("Failed to write response", "err", err)
		}
	})

	muxer.HandleFunc("POST "+path+"createJob", srv.handleCreateJob)

	// muxer.HandleFunc("GET "+path+"test", func(w http.ResponseWriter, r *http.Request) {

	// 	logger := r.Context().Value("logger").(*slog.Logger)
	// 	logger.Debug("Test - GET")

	// 	w.WriteHeader(http.StatusOK)
	// 	_, err := w.Write([]byte("GET"))

	// 	if err != nil {
	// 		logger.Error("Failed to write response", "err", err)
	// 	}
	// })

	// muxer.HandleFunc("POST "+path+"test", func(w http.ResponseWriter, r *http.Request) {

	// 	logger := r.Context().Value("logger").(*slog.Logger)
	// 	logger.Debug("Test - POST")

	// 	w.WriteHeader(http.StatusOK)
	// 	_, err := w.Write([]byte("POST"))

	// 	if err != nil {
	// 		logger.Error("Failed to write response", "err", err)
	// 	}
	// })

}

func (srv *Server) NewULID() (ulid.ULID, error) {

	return ulid.New(ulid.Timestamp(time.Now()), srv.Entropy)
}
