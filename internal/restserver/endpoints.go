package restserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	JOBSTATUS_CREATED       = "0 - created"
	JOBSTATUS_COMPILING     = "1 - compiling"
	JOBSTATUS_FINISHED      = "2 - finished"
	JOBSTATUS_ERROR         = "X - error"
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

type ResponseJobStatus struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Running bool   `json:"running"`
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func (srv *Server) runJob(ctx context.Context, job_id string, texfile_path string) {

	logger := ctx.Value("logger").(*slog.Logger)
	logger = logger.With("func", "restserver.runJob", "job", job_id)

	// === Compile TeX to PDF/A ===

	zerologLogger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	zerologLogger = zerologLogger.With().Str("job", job_id).Str("logger", "job-zerolog").Logger()
	ctx = context.WithValue(ctx, contextkeys.LoggerKey, zerologLogger)

	// update db
	logger.Debug("Compiling TeX to PDF/A")
	tx := srv.db.Model(&Jobs{}).Where("job_id = ?", job_id).Update("status", JOBSTATUS_COMPILING)

	if tx.Error != nil {
		logger.Error("Error updating job status, aborting [A9VXQIFU]", "err", tx.Error)
		return
	}

	builddir_template := srv.Options.BUILDDIR_PREFIX + BUILDDIR_DELIM + job_id + BUILDDIR_DELIM
	result, err := textopdfa.CompileTexToPDFA(ctx, texfile_path, builddir_template+BUILDDIR_PREFIX_COMPILE)

	if err != nil {
		logger.Error("Error compiling TeX to PDF/A [ITGMFXSI]", "err", err)

		// update db
		tx := srv.db.Model(&Jobs{}).Where("job_id = ?", job_id).
			Update("status", JOBSTATUS_ERROR).
			Update("status_running", false).
			Update("status_success", false).
			Update("error", err.Error())

		if tx.Error != nil {
			logger.Error("Error updating job status [JO79QRDU]", "err", tx.Error)
		}

		return
	}

	logger.Info("Successfully compiled TeX to PDF/A", "path", result)
	tx = srv.db.Model(&Jobs{}).Where("job_id = ?", job_id).
		Update("result", result).
		Update("status", JOBSTATUS_FINISHED).
		Update("status_running", false).
		Update("status_success", true)

	if tx.Error != nil {
		logger.Error("Error updating job status [EW8QVQF0]", "err", tx.Error)
	}

	logger.Debug("Bye")
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

	logger = logger.With("job", job_id.String())

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

	// add to db

	tx := srv.db.Create(&Jobs{
		JobID:         job_id.String(),
		Name:          req.Name,
		Status:        JOBSTATUS_CREATED,
		StatusRunning: true,
		StatusSuccess: false,
		Path:          builddir,
	})

	if tx.Error != nil {
		_ = server.WriteError(w, http.StatusInternalServerError, "failed to write job to db [PBYSS5CV]", logger)
		return
	}

	logger.Debug("Added job to db")

	// === Compile TeX to PDF/A ===

	logger.Debug("Compiling TeX to PDF/A")
	srv.runJob(r.Context(), job_id.String(), texfile_path)

	// ===== Write response =====

	resp := ResponseCreateJob{
		JobID:   job_id.String(),
		Message: "Job created with builddir: " + builddir,
	}

	_ = server.WriteResponse(w, resp, logger)

	// _ = server.WriteError(w, http.StatusNotImplemented, "", logger)
}

func (srv *Server) handleJobStatus(w http.ResponseWriter, r *http.Request) {

	logger := r.Context().Value("logger").(*slog.Logger)
	logger = logger.With("func", "restserver.handleJobStatus")

	// ===== Parse request body =====

	// job_id := r.URL.Query().Get("job")
	job_id := r.PathValue("id")

	if job_id == "" {
		_ = server.WriteError(w, http.StatusBadRequest, "job_id is empty [TK872KG6]", logger)
		return
	}

	logger.Debug("Got request to get job status for job " + job_id)

	// ===== Get job status =====

	var job Jobs
	tx := srv.db.First(&job, "job_id = ?", job_id)

	if tx.Error != nil {
		_ = server.WriteError(w, http.StatusNotFound, "job not found [79E9DQFC]", logger)
		return
	}

	// ===== Write response =====

	resp := ResponseJobStatus{
		JobID:   job.JobID,
		Status:  job.Status,
		Running: job.StatusRunning,
		Success: job.StatusSuccess,
		Error:   job.Error,
	}

	_ = server.WriteResponse(w, resp, logger)
}

func (srv *Server) handleJobGetResult(w http.ResponseWriter, r *http.Request) {
	// var err error

	logger := r.Context().Value("logger").(*slog.Logger)
	logger = logger.With("func", "restserver.handleJobGetResult")

	logger.Debug("Got request to get job result")

	// ===== Parse request body =====

	// job_id := r.URL.Query().Get("job")
	job_id := r.PathValue("id")

	if job_id == "" {
		_ = server.WriteError(w, http.StatusBadRequest, "job_id is empty [KD57LWTL]", logger)
		return
	}

	logger.Debug("Got request to get job result for job " + job_id)

	// ===== Get job status =====

	var job Jobs
	tx := srv.db.First(&job, "job_id = ?", job_id)

	if tx.Error != nil {
		_ = server.WriteError(w, http.StatusNotFound, "job not found [I2JH1HX5]", logger)
		return
	}

	if job.StatusRunning {
		_ = server.WriteError(w, http.StatusAccepted, "job is still running [B3QSX2RD]", logger)
		return
	}

	if !job.StatusSuccess {
		_ = server.WriteError(w, http.StatusInternalServerError, "job failed [B0K8PMG8]: "+job.Error, logger)
		return
	}

	// check if file exists
	resultfile := job.Result

	logger.Debug("Checking result file", "resultfile", resultfile)

	if _, err := os.Stat(resultfile); os.IsNotExist(err) {
		logger.Error("Result file not found", "resultfile", resultfile)
		_ = server.WriteError(w, http.StatusInternalServerError, "result file not found [730P4NCL]: "+err.Error(), logger)
		return
	}

	// Open the PDF file
	file, err := os.Open(resultfile)
	if err != nil {
		_ = server.WriteError(w, http.StatusInternalServerError, "could not open PDF file [23TN7WM9]: "+err.Error(), logger)
		return
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		_ = server.WriteError(w, http.StatusInternalServerError, "could not get file info [I6OT68Q1]: "+err.Error(), logger)
		return
	}

	// Set headers
	// w.Header().Set("Content-Disposition", "attachment; filename="+job.Name+".pdf")
	w.Header().Set("Content-Disposition", "inline; filename="+job.Name+".pdf")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Write the file content to the response
	logger.Debug("Sending file", "path", resultfile)
	if _, err := io.Copy(w, file); err != nil {
		_ = server.WriteError(w, http.StatusInternalServerError, "could not send file [XY123GHI]: "+err.Error(), logger)
		return
	}

	logger.Debug("Sent file")

	// _ = server.WriteError(w, http.StatusNotImplemented, "not implemented yet", logger)
}
