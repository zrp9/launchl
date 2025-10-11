package noti

import (
	"embed"
	"html/template"
	"path/filepath"
	"strings"
	texttmpl "text/template"
)

//go:embed templates/html templates/text
var fs embed.FS

type Renderer struct {
	html *template.Template
	txt  *texttmpl.Template
}

func NewRenderer() (*Renderer, error) {
	r := &Renderer{}
	if err := r.reload(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Renderer) reload() error {
	htmlFns, txtFns := HTMLFuncMap(), TextFuncMap()

	// Parse HTML
	h := template.New("html").Funcs(htmlFns)
	for _, p := range []string{
		"templates/html/*.html.tmpl",
		"templates/html/*.html.tmpl",
		"templates/**/*.html.tmpl",
	} {
		var err error
		h, err = h.ParseFS(fs, p)
		if err != nil {
			return err
		}
	}

	// Parse Text
	t := texttmpl.New("txt").Funcs(txtFns)
	for _, p := range []string{
		"templates/text/*.txt.tmpl",
		"templates/text/*.txt.tmpl",
		"templates/**/*.txt.tmpl",
	} {
		var err error
		t, err = t.ParseFS(fs, p)
		if err != nil {
			return err
		}
	}

	r.html, r.txt = h, t
	return nil
}

// Render name like "account/verify"
func (r *Renderer) Render(name string, data any) (htmlBody, textBody string, err error) {
	htmlName := strings.TrimSuffix(name, filepath.Ext(name)) + ".html.tmpl"
	txtName := strings.TrimSuffix(name, filepath.Ext(name)) + ".txt.tmpl"

	var b strings.Builder
	if err = r.html.ExecuteTemplate(&b, htmlName, data); err != nil {
		return "", "", err
	}
	htmlBody = b.String()
	b.Reset()
	if err = r.txt.ExecuteTemplate(&b, txtName, data); err != nil {
		return "", "", err
	}
	textBody = b.String()
	return
}
