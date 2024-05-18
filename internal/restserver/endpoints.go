package restserver

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/tilseiffert/docker-tex-to-pdf/pkg/server"
)

type RequestCreateJob struct {
	Name       string `json:"name"`
	TexContent string `json:"tex_content"`
}

type ResponseCreateJob struct {
	JobID string `json:"job_id"`
}

func (srv *Server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	// var err error

	logger := r.Context().Value("logger").(*slog.Logger)
	logger = logger.With("func", "restserver.handleCreateJob")

	logger.Debug("Got request to create job")

	// ===== Parse request body =====

	var req RequestCreateJob
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		if err.Error() == "EOF" {
			_ = server.WriteError(w, http.StatusBadRequest, "Request body is empty [6W3VLJ97]", logger)

			return
		}

		_ = server.WriteError(w, http.StatusBadRequest, err.Error()+" [ZPA89CTE]", logger)

		return
	}

	// ===== Validate request =====

	if req.Name == "" {
		_ = server.WriteError(w, http.StatusBadRequest, "name is empty [HY85SV7R]", logger)
		return
	}

	if req.TexContent == "" {
		_ = server.WriteError(w, http.StatusBadRequest, "tex_content is empty [RYA39AGA]", logger)
		return
	}

	// ===== Create job =====

	job_id, err := srv.NewULID()

	if err != nil {
		_ = server.WriteError(w, http.StatusInternalServerError, "failed to generate job ID [J7WL1VI4]", logger)
		return
	}

	// ===== Write response =====

	resp := ResponseCreateJob{
		JobID: job_id.String(),
	}

	_ = server.WriteResponse(w, resp, logger)

	// _ = server.WriteError(w, http.StatusNotImplemented, "", logger)
}
