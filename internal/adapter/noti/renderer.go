package noti

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	texttmpl "text/template"

	zhtml "github.com/zrp9/launchl/internal/adapter/noti/templates/html"
)

//go:embed templates/text/*.tmpl
var fs embed.FS

type Renderer struct {
	txt *texttmpl.Template
}

func NewRenderer() (*Renderer, error) {
	txt, err := texttmpl.ParseFS(fs, "templates/text/*.tmpl")
	if err != nil {
		return nil, err
	}

	return &Renderer{
		txt: txt,
	}, nil
}

// Render name like "account/verify"
func (r *Renderer) Render(templateName string, data EmailJob) (htmlBody, textBody string, err error) {
	ctx := context.Background()
	var htmlBuf bytes.Buffer
	// ---- TEXT via text/template ----
	// assumes:
	// templates/text/surveyThanks.txt.tmpl
	// templates/text/welcome.txt.tmpl
	var txtBuf bytes.Buffer
	switch templateName {
	case "surveyThanks":
		var d SurveyThanksData
		if err = json.Unmarshal(data.Data, &d); err != nil {
			return "", "", fmt.Errorf("failed to parse json %w", err)
		}

		// html template
		comp := zhtml.SurveyThanksEmail(d.Name, d.ReferralURL)
		if err = comp.Render(ctx, &htmlBuf); err != nil {
			return "", "", fmt.Errorf("render html template (surveyThanks) error: %w", err)
		}

		if err = r.renderTxtTemplate(&txtBuf, templateName, d); err != nil {
			return "", "", fmt.Errorf("render txt template (surveyThanks) error :%w", err)
		}

	case "welcome":
		var d WelcomeData
		if err = json.Unmarshal(data.Data, &d); err != nil {
			return "", "", fmt.Errorf("failed to parse json %w", err)
		}

		comp := zhtml.WelcomeEmail(d.Name, d.ReferralURL)
		if err = comp.Render(ctx, &htmlBuf); err != nil {
			return "", "", fmt.Errorf("render html template (welcome) error: %w", err)
		}

		if err = r.renderTxtTemplate(&txtBuf, templateName, d); err != nil {
			return "", "", fmt.Errorf("render txt template (welcome) error :%w", err)
		}

	default:
		return "", "", fmt.Errorf("unknown email template: %q", templateName)
	}

	htmlStr := htmlBuf.String()

	return htmlStr, txtBuf.String(), nil
}

func (r *Renderer) renderTxtTemplate(buffer *bytes.Buffer, templateName string, data any) error {
	// txt template
	txtName := strings.TrimSuffix(templateName, filepath.Ext(templateName)) + ".txt.tmpl"
	if err := r.txt.ExecuteTemplate(buffer, txtName, data); err != nil {
		return err
	}
	return nil
}
