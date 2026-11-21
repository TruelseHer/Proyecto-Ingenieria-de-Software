package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type ToastMessage struct {
	Type    string
	Title   string
	Message string
}

type Renderer struct {
	templatesDir string
	pages        map[string]*template.Template
	partials     *template.Template
}

func NewRenderer(templatesDir string) (*Renderer, error) {
	funcMap := template.FuncMap{
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires even number of arguments")
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				k, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				m[k] = values[i+1]
			}
			return m, nil
		},
		"formatDate": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			months := []string{"", "Ene", "Feb", "Mar", "Abr", "May", "Jun", "Jul", "Ago", "Sep", "Oct", "Nov", "Dic"}
			return fmt.Sprintf("%02d %s %d", t.Day(), months[int(t.Month())], t.Year())
		},
		"formatDateTime": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			months := []string{"", "Ene", "Feb", "Mar", "Abr", "May", "Jun", "Jul", "Ago", "Sep", "Oct", "Nov", "Dic"}
			return fmt.Sprintf("%02d %s %02d:%02d", t.Day(), months[int(t.Month())], t.Hour(), t.Minute())
		},
		"formatDateInput": func(t time.Time) string {
			if t.IsZero() {
				t = time.Now().Add(48 * time.Hour)
			}
			return t.Format("2006-01-02T15:04")
		},
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "--:--"
			}
			return t.Format("15:04")
		},
		"formatMoney": func(amount float64, currency string) string {
			if currency == "" {
				currency = "USD"
			}
			return fmt.Sprintf("%s %.2f", strings.ToUpper(currency), amount)
		},
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"joinStrings": func(items []string) string {
			return strings.Join(items, ", ")
		},
		"toJSON": func(v any) string {
			b, err := json.Marshal(v)
			if err != nil {
				return "[]"
			}
			return string(b)
		},
	}

	partialsGlob := filepath.Join(templatesDir, "partials", "*.html")
	partialsTpl, err := template.New("partials").Funcs(funcMap).ParseGlob(partialsGlob)
	if err != nil {
		return nil, fmt.Errorf("parse partials: %w", err)
	}

	pageFiles, err := filepath.Glob(filepath.Join(templatesDir, "pages", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("glob pages: %w", err)
	}

	baseFile := filepath.Join(templatesDir, "layouts", "base.html")
	pages := make(map[string]*template.Template, len(pageFiles))
	for _, pf := range pageFiles {
		name := filepath.Base(pf)
		files := []string{baseFile, pf}
		partialMatches, _ := filepath.Glob(partialsGlob)
		files = append(files, partialMatches...)

		t, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			return nil, fmt.Errorf("parse page %s: %w", name, err)
		}
		pages[name] = t
	}

	return &Renderer{
		templatesDir: templatesDir,
		pages:        pages,
		partials:     partialsTpl,
	}, nil
}

func (r *Renderer) RenderPage(w http.ResponseWriter, pageName string, data any) {
	t, ok := r.pages[pageName]
	if !ok {
		http.Error(w, "Vista no encontrada: "+pageName, http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base.html", data); err != nil {
		http.Error(w, "Error al renderizar página: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

func (r *Renderer) RenderPartial(w http.ResponseWriter, partialName string, data any, toast *ToastMessage) {
	var buf bytes.Buffer
	if partialName != "" {
		if err := r.partials.ExecuteTemplate(&buf, partialName, data); err != nil {
			http.Error(w, "Error al renderizar fragmento: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if toast != nil {
		_ = r.partials.ExecuteTemplate(&buf, "toast.html", toast)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

func (r *Renderer) RenderToastOnly(w http.ResponseWriter, status int, toast ToastMessage) {
	var buf bytes.Buffer
	_ = r.partials.ExecuteTemplate(&buf, "toast.html", toast)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}
