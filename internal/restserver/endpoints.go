package restserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/tilseiffert/docker-tex-to-pdf/internal/contextkeys"
	"github.com/tilseiffert/docker-tex-to-pdf/internal/textopdfa"
	"github.com/tilseiffert/docker-tex-to-pdf/pkg/server"
)

const (
	BUILDDIR_PREFIX_COMPILE = "compile"
	BUILDDIR_TEXFILE        = "main.tex"
	BUILDDIR_DELIM          = "."
	DEFAULT_JOBDIR          = "./jobs"
)

var (
	jobdir = DEFAULT_JOBDIR
)

type RequestCreateJob struct {
	Name       string `json:"name"`
	TexContent string `json:"tex_content"`
}

type ResponseCreateJob struct {
	JobID   string `json:"job_id"`
	Message string `json:"message"`
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

	// ===== Prepare job =====

	job_id, err := srv.NewULID()

	if err != nil {
		_ = server.WriteError(w, http.StatusInternalServerError, "failed to generate job ID [J7WL1VI4]", logger)
		return
	}

	// create job directory
	builddir := jobdir + "/" + job_id.String()
	err = os.MkdirAll(builddir, 0755)

	// builddir_template := srv.Options.BUILDDIR_PREFIX + BUILDDIR_DELIM + job_id.String() + BUILDDIR_DELIM
	// builddir, err := os.MkdirTemp("", builddir_template)
	// // defer os.RemoveAll(builddir)

	if err != nil {
		_ = server.WriteError(w, http.StatusInternalServerError, "failed to create build directory [J7WL1VI4]", logger)
		return
	}

	logger.Debug("Created build directory", "builddir", builddir)

	// write tex content to file
	texfile_path := builddir + "/" + BUILDDIR_TEXFILE

	err = os.WriteFile(texfile_path, []byte(req.TexContent), 0644)

	if err != nil {
		_ = server.WriteError(w, http.StatusInternalServerError, "failed to write tex content to file [RICMARTU]", logger)
		return
	}

	logger.Debug("Wrote tex content to file", "texfile_path", texfile_path)

	// === Compile TeX to PDF/A ===

	ctx := r.Context()
	//set new zerolog logger to context
	zerologLogger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx = context.WithValue(ctx, contextkeys.LoggerKey, zerologLogger)

	logger.Debug("Compiling TeX to PDF/A")
	builddir_template := srv.Options.BUILDDIR_PREFIX + BUILDDIR_DELIM + job_id.String() + BUILDDIR_DELIM
	result, err := textopdfa.CompileTexToPDFA(ctx, texfile_path, builddir_template+BUILDDIR_PREFIX_COMPILE)

	if err != nil {
		logger.Error("Error compiling TeX to PDF/A", "err", err)
		_ = server.WriteError(w, http.StatusInternalServerError, "failed to compile TeX to PDF/A [GHMTAG6I]: "+err.Error(), logger)
		return
	}

	logger.Info("Successfully compiled TeX to PDF/A", "result", result)

	// ===== Write response =====

	resp := ResponseCreateJob{
		JobID:   job_id.String(),
		Message: "Job created with builddir: " + builddir,
	}

	_ = server.WriteResponse(w, resp, logger)

	// _ = server.WriteError(w, http.StatusNotImplemented, "", logger)
}
