package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	BUILDDIR_TEMPLATE = "tex-to-pdfa_build_*"
	TEXFILE           = "main.tex"
)

type contextKey struct {
	name string
}

var loggerKey = &contextKey{"logger"}

// initLogger initializes a logger and adds it to the context
func initLogger(ctx context.Context) context.Context {

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Caller().Logger()

	return context.WithValue(ctx, loggerKey, logger)
}

// Log returns the logger from the context
func Log(ctx context.Context) *zerolog.Logger {

	// logger, ok := ctx.Value(loggerKey).(*log.Logger)

	if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return &logger
	}

	panic("unable to retrieve logger from context")
}

func assureCommand(ctx context.Context, name string) error {
	_, err := exec.LookPath(name)

	if err != nil {
		return fmt.Errorf("could not assure command '%s': %w", name, err)
	}

	return nil
}

func main() {
	var cmd_stdout, cmd_stderr bytes.Buffer

	// === Initialize context ===

	starttime := time.Now()

	// Initialize context
	ctx := initLogger(context.Background())

	Log(ctx).Debug().Msg("Hello world 👋")

	// === Check for essential commands ===

	// slice of strings
	commands := []string{"pdflatex", "rubber", "gs"}

	ok := true
	for _, command := range commands {
		err := assureCommand(ctx, command)

		if err != nil {
			ok = false
			Log(ctx).Err(err).Str("command", command).Msg("Command not found")
		}
	}

	if !ok {
		Log(ctx).Fatal().Msg("Could not assure essential commands, see errors above, aborting...")
	}

	Log(ctx).Trace().Msg("All essential commands found")

	// === Prepare build ===

	// create temp dir
	builddir, err := os.MkdirTemp("", BUILDDIR_TEMPLATE)
	defer os.RemoveAll(builddir)

	if err != nil {
		Log(ctx).Fatal().Err(err).Msg("Could not create temp dir, aborting...")
	}

	Log(ctx).Debug().Str("builddir", builddir).Msg("Created temp dir")

	// === Build PDF from TeX ===

	// get absolute path of main.tex
	texfile, err := filepath.Abs(TEXFILE)

	if err != nil {
		Log(ctx).Fatal().Err(err).Msg("Could not get absolute path of tex-file, aborting...")
	}

	// check if file main.tex exists
	if _, err := os.Stat(texfile); os.IsNotExist(err) {
		Log(ctx).Fatal().Err(err).Msgf("tex-file does not exist, expected '%s', aborting...", texfile)
	}

	basename := strings.TrimSuffix(filepath.Base(texfile), ".tex")
	maindir := filepath.Dir(texfile)
	Log(ctx).Debug().Str("texfile", texfile).Msgf("Found tex-file %s", basename)

	// cmd := exec.Command("pdflatex", "-output-directory="+builddir, "-interaction=nonstopmode", TEXFILE)
	cmd := exec.Command("rubber", "--pdf", "--into="+builddir, texfile)
	cmd.Stdout = &cmd_stdout
	cmd.Stderr = &cmd_stderr

	Log(ctx).Info().Msg("Compiling TeX file")
	Log(ctx).Debug().Str("cmd", cmd.String()).Msg("Running command")
	err = cmd.Run()

	if err != nil {
		Log(ctx).Fatal().Err(err).Str("stdout", cmd_stdout.String()).Str("stderr", cmd_stderr.String()).Msg("Could not run command, aborting...")
	}

	Log(ctx).Trace().Str("stdout", cmd_stdout.String()).Str("stderr", cmd_stderr.String()).Msgf("Command %s finished", cmd.Path)

	// check pdf file
	pdffile := builddir + "/" + basename + ".pdf"

	if _, err := os.Stat(pdffile); os.IsNotExist(err) {
		Log(ctx).Fatal().Err(err).Msgf("pdffile does not exist, expected '%s', aborting...", pdffile)
	}

	Log(ctx).Trace().Str("path", pdffile).Msg("Found pdffile")

	// === Convert PDF to PDF/A-1 ===
	Log(ctx).Info().Msg("Converting PDF to PDF/A")

	pdffile_pdfa1 := builddir + "/" + basename + "_pdfa1.pdf"

	cmd_stdout.Reset()
	cmd_stderr.Reset()

	cmd = exec.Command("gs", "-sDEVICE=pdfwrite", "-dPDFA=1", "-sColorConversionStrategy=UseDeviceIndependentColor", "-dPDFACompatibilityPolicy=2", "-o", pdffile_pdfa1, pdffile)
	cmd.Stdout = &cmd_stdout
	cmd.Stderr = &cmd_stderr

	Log(ctx).Debug().Str("cmd", cmd.String()).Msg("Running command")
	err = cmd.Run()

	if err != nil {
		Log(ctx).Fatal().Err(err).Str("stdout", cmd_stdout.String()).Str("stderr", cmd_stderr.String()).Msg("Could not run command, aborting...")
	}

	Log(ctx).Trace().Str("stdout", cmd_stdout.String()).Str("stderr", cmd_stderr.String()).Msgf("Command %s finished", cmd.Path)

	if _, err := os.Stat(pdffile_pdfa1); os.IsNotExist(err) {
		Log(ctx).Fatal().Err(err).Msgf("pdffile does not exist, expected '%s', aborting...", pdffile_pdfa1)
	}

	Log(ctx).Trace().Str("path", pdffile_pdfa1).Msg("Found pdffile_pdfa1")

	// === Convert PDF to PDF/A-3 ===

	pdffile_pdfa3 := builddir + "/" + basename + "_pdfa3.pdf"

	cmd_stdout.Reset()
	cmd_stderr.Reset()

	cmd = exec.Command("gs", "-sDEVICE=pdfwrite", "-dPDFA=3", "-sColorConversionStrategy=UseDeviceIndependentColor", "-dPDFACompatibilityPolicy=2", "-o", pdffile_pdfa3, pdffile_pdfa1)
	cmd.Stdout = &cmd_stdout
	cmd.Stderr = &cmd_stderr

	Log(ctx).Debug().Str("cmd", cmd.String()).Msg("Running command")
	err = cmd.Run()

	if err != nil {
		Log(ctx).Fatal().Err(err).Str("stdout", cmd_stdout.String()).Str("stderr", cmd_stderr.String()).Msg("Could not run command, aborting...")
	}

	Log(ctx).Trace().Str("stdout", cmd_stdout.String()).Str("stderr", cmd_stderr.String()).Msgf("Command %s finished", cmd.Path)

	if _, err := os.Stat(pdffile_pdfa3); os.IsNotExist(err) {
		Log(ctx).Fatal().Err(err).Msgf("pdffile does not exist, expected '%s', aborting...", pdffile_pdfa3)
	}

	Log(ctx).Trace().Str("path", pdffile_pdfa3).Msg("Found pdffile_pdfa2")

	// === Move PDF to output dir ===

	resultpath := maindir + "/" + basename + ".pdf"

	cmd_stdout.Reset()
	cmd_stderr.Reset()

	cmd = exec.Command("cp", "-v", pdffile_pdfa3, resultpath)
	cmd.Stdout = &cmd_stdout
	cmd.Stderr = &cmd_stderr

	err = cmd.Run()

	if err != nil {
		Log(ctx).Fatal().Err(err).Str("stdout", cmd_stdout.String()).Str("stderr", cmd_stderr.String()).Msg("Could not run command, aborting...")
	}

	Log(ctx).Trace().Str("stdout", cmd_stdout.String()).Str("stderr", cmd_stderr.String()).Msgf("Command %s finished", cmd.Path)

	if _, err := os.Stat(resultpath); os.IsNotExist(err) {
		Log(ctx).Fatal().Err(err).Msgf("pdffile does not exist, expected '%s', aborting...", resultpath)
	}

	Log(ctx).Trace().Str("path", resultpath).Msg("Found resultpath")

	Log(ctx).Info().Str("path", resultpath).Msg("PDF/A-3 file created")

	// === Tidy up ===

	// log runtime
	Log(ctx).Debug().Str("runtime", time.Since(starttime).String()).Msg("Done 👋")
}
