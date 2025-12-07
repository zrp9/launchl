package noti

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/zrp9/launchl/internal/domain"
)

func HTMLFuncMap() template.FuncMap {
	return template.FuncMap{
		"absURL": func(base, path string) (template.URL, error) {
			u, err := url.JoinPath(base, path)
			return template.URL(u), err
		},
		"safeHTML": func(s string) template.HTML { return template.HTML(s) }, // only for trusted content
		"nl2br": func(s string) template.HTML {
			return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
		},
		"money": func(cents int64) string { return fmt.Sprintf("$%.2f", float64(cents)/100) },
	}
}

func TextFuncMap() map[string]any {
	return map[string]any{
		"wrapAt": func(s string, n int) string {
			// minimal word-wrap for plain text
			var out, line string
			for _, w := range strings.Fields(s) {
				if len(line)+1+len(w) > n && line != "" {
					out += line + "\n"
					line = w
				} else {
					if line != "" {
						line += " "
					}
					line += w
				}
			}
			if line != "" {
				out += line
			}
			return out
		},
		"money": func(cents int64) string { return fmt.Sprintf("$%.2f", float64(cents)/100) },
	}
}

func parseEventType(m domain.Message) (string, error) {
	get := func(k string) (string, bool) {
		for _, f := range m.Values {
			if f.Name == k {
				return f.Value, true
			}
		}
		return "", false
	}

	if s, ok := get("eventType"); ok {
		return s, nil
	}

	return "", nil
}

func parseEmailJob(m domain.Message) (EmailJob, error) {
	// m.Values is []domain.Field { Name, Value }
	get := func(k string) (string, bool) {
		for _, f := range m.Values {
			if f.Name == k {
				return f.Value, true
			}
		}
		return "", false
	}

	// Preferred: whole JSON in "payload"
	if raw, ok := get("payload"); ok && raw != "" {
		var job EmailJob
		if err := json.Unmarshal([]byte(raw), &job); err != nil {
			return EmailJob{}, err
		}
		return job, nil
	}

	// Fallback: field-wise values
	var job EmailJob
	if s, ok := get("template"); ok {
		job.Template = s
	}
	if s, ok := get("subject"); ok {
		job.Subject = s
	}
	if s, ok := get("to"); ok && s != "" {
		// allow "to" to be JSON array or comma-separated
		var arr []string
		if json.Unmarshal([]byte(s), &arr) == nil {
			job.To = arr
		} else {
			job.To = splitComma(s)
		}
	}
	if s, ok := get("data"); ok && s != "" {
		_ = json.Unmarshal([]byte(s), &job.Data) // ignore error; zero value acceptable
	}

	if len(job.To) == 0 || job.Template == "" || job.Subject == "" {
		return EmailJob{}, errors.New("invalid email job: missing to/template/subject")
	}
	if job.Data == nil {
		job.Data = json.RawMessage{}
	}
	return job, nil
}

func splitComma(s string) []string {
	out := []string{}
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if start < i {
				out = append(out, trimSpace(s[start:i]))
			}
			start = i + 1
		}
	}
	return out
}

func trimSpace(s string) string {
	i, j := 0, len(s)-1
	for i <= j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	for j >= i && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n' || s[j] == '\r') {
		j--
	}
	if i > j {
		return ""
	}
	return s[i : j+1]
}

func mapToStruct[T any](m map[string]any) (T, error) {
	var out T
	b, err := json.Marshal(m)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(b, &out)
	return out, err
}
