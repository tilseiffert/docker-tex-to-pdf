package textopdfa

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/tilseiffert/docker-tex-to-pdf/internal/contextkeys"
)

// Log returns the logger from the context (for more convenient logging)
func Log(ctx context.Context) *zerolog.Logger {

	if logger, ok := ctx.Value(contextkeys.LoggerKey).(zerolog.Logger); ok {
		return &logger
	}

	panic("unable to retrieve logger from context")
}

// assureCommand checks if a command is available and returns an error if not
// ctx is the context
// name is the name of the command
func assureCommand(ctx context.Context, name string) error {
	_, err := exec.LookPath(name)

	if err != nil {
		return fmt.Errorf("could not assure command '%s': %w", name, err)
	}

	return nil
}

// CompileTexToPDFA compiles a TeX file to a PDF/A file
// It returns the path to the PDF/A file
// ctx is the context
// texfile_name is the name to the TeX file (it is expected in the current working directory)
// builddir_template is the template for the build directory (e.g. "tex-to-pdfa_build_*")
func CompileTexToPDFA(ctx context.Context, texfile_name string, builddir_template string) (string, error) {
	var cmd_stdout, cmd_stderr bytes.Buffer

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
	builddir, err := os.MkdirTemp("", builddir_template)
	defer os.RemoveAll(builddir)

	if err != nil {
		Log(ctx).Fatal().Err(err).Msg("Could not create temp dir, aborting...")
	}

	Log(ctx).Debug().Str("builddir", builddir).Msg("Created temp dir")

	// === Build PDF from TeX ===

	// get absolute path of main.tex
	texfile, err := filepath.Abs(texfile_name)

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

	return resultpath, nil
}
