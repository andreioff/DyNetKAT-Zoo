package util

import (
	"bytes"
	"embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	PDFLATEX_COMMAND = "pdflatex"
	JOBNAME_ARG      = "-jobname"
	FILE_PERM        = 0755
	DEFAULT_DOC_TMPL = "default_doc.tmpl"
)

// Ensure the util/templates folder is embedded into the final binary
//
//go:embed templates
var tmplsFS embed.FS

var tmpls *template.Template

// load templates
func init() {
	templates, err := template.New(DEFAULT_DOC_TMPL).ParseFS(tmplsFS, "templates/*.tmpl")
	if err != nil {
		log.Panicf("Failed to load templates!\n%s\n", err.Error())
	}
	tmpls = templates
}

func GetFileBasename(path string) string {
	_, fName := filepath.Split(path)
	return strings.Split(fName, ".")[0]
}

func createDir(dir string) error {
	_, err := os.Stat(dir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		err = os.Mkdir(dir, FILE_PERM)
		if err != nil {
			return err
		}
	}

	return nil
}

func WriteToNewFile(dir, fileName, data string) error {
	switch {
	case dir == "":
		return NewError(ErrEmptyStringVar, "dir")
	case fileName == "":
		return NewError(ErrEmptyStringVar, "fileName")
	}

	err := createDir(dir)
	if err != nil {
		return err
	}
	filePath := filepath.Join(dir, fileName)

	return os.WriteFile(filePath, []byte(data), FILE_PERM)
}

func WriteToNewPdf(dir, name, content string) error {
	switch {
	case dir == "":
		return NewError(ErrEmptyStringVar, "dir")
	case name == "":
		return NewError(ErrEmptyStringVar, "name")
	}

	err := createDir(dir)
	if err != nil {
		return err
	}

	withTmpl, err := applyDefaultDocTemplate(content)
	if err != nil {
		return err
	}

	return runPdfLatex(dir, name, withTmpl)
}

func applyDefaultDocTemplate(content string) (string, error) {
	newContent := bytes.NewBuffer([]byte{})
	err := tmpls.Execute(newContent, content)
	if err != nil {
		return "", nil
	}

	return newContent.String(), nil
}

func runPdfLatex(dir, name, content string) error {
	// required by pdflatex
	content = "\"" + content + "\""

	errOut := bytes.NewBuffer([]byte{})
	cmd := exec.Command(PDFLATEX_COMMAND, JOBNAME_ARG, name)
	cmd.Dir = dir
	cmd.Stdin = bytes.NewBuffer([]byte(content))
	cmd.Stderr = errOut

	runErr := cmd.Run()
	if runErr != nil {
		return NewError("%s\n%s\n", runErr.Error(), errOut.String())
	}

	return nil
}
