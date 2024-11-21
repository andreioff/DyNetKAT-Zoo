package util

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"text/template"
)

const (
	PDFLATEX_COMMAND = "pdflatex"
	JOBNAME_ARG      = "-jobname"
	FILE_PERM        = 0755
	DEFAULT_DOC_TMPL = "default_doc.tmpl"
)

var tmpls *template.Template

func init() {
	templates, err := template.New(DEFAULT_DOC_TMPL).ParseGlob("./templates/*.tmpl")
	if err != nil {
		log.Panicf("Failed to load templates!\n%s\n", err.Error())
	}
	tmpls = templates
}

func getFilePath(dir, fileName string) string {
	filePath := dir + fileName
	if dir[len(dir)-1] != '/' {
		filePath = dir + "/" + fileName
	}

	return filePath
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
	filePath := getFilePath(dir, fileName)

	withTmpl, err := applyDefaultDocTemplate(data)
	if err != nil {
		return err
	}

	os.WriteFile(filePath, []byte(withTmpl), FILE_PERM)
	return nil
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
	// required by pdflatex
	withTmpl = "\"" + withTmpl + "\""

	errOut := bytes.NewBuffer([]byte{})
	cmd := exec.Command(PDFLATEX_COMMAND, JOBNAME_ARG, name)
	cmd.Dir = dir
	cmd.Stdin = bytes.NewBuffer([]byte(withTmpl))
	cmd.Stderr = errOut

	runErr := cmd.Run()
	if runErr != nil {
		return NewError("%s\n%s\n", runErr.Error(), errOut.String())
	}

	return nil
}

func applyDefaultDocTemplate(content string) (string, error) {
	newContent := bytes.NewBuffer([]byte{})
	err := tmpls.Execute(newContent, content)
	if err != nil {
		return "", nil
	}

	return newContent.String(), nil
}
