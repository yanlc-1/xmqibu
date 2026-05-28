package web

import (
	"embed"
	"html/template"
	"io/fs"
	"strings"
	"time"
)

//go:embed templates/*.gohtml static/*
var assets embed.FS

func ParseTemplates() (*template.Template, error) {
	funcs := template.FuncMap{
		"dict": func(values ...any) map[string]any {
			result := make(map[string]any, len(values)/2)
			for i := 0; i+1 < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					continue
				}
				result[key] = values[i+1]
			}
			return result
		},
		"formatDate": func(value time.Time) string {
			if value.IsZero() {
				return ""
			}
			return value.Format("2006-01-02")
		},
		"formatMonthDay": func(value time.Time) string {
			if value.IsZero() {
				return ""
			}
			return value.Format("01/02")
		},
		"joinLinks": func(items []string) string {
			return strings.Join(items, " · ")
		},
	}
	return template.New("site").Funcs(funcs).ParseFS(assets, "templates/*.gohtml")
}

func StaticFS() fs.FS {
	static, err := fs.Sub(assets, "static")
	if err != nil {
		panic(err)
	}
	return static
}
