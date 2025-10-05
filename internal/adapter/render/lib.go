package render

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"
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
