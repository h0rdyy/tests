package templates

import (
	"fmt"
	"log"
	"net/http"

	"cybersec/internal/content"
	"github.com/flosch/pongo2/v6"
)

type Renderer struct {
	set *pongo2.TemplateSet
	log *log.Logger
}

func New(set *pongo2.TemplateSet, logger *log.Logger) *Renderer {
	return &Renderer{set: set, log: logger}
}

func (r *Renderer) URLFor(name string, args ...interface{}) string {
	switch name {
	case "index":
		return "/"
	case "about":
		return "/about"
	case "privacy_page":
		return "/privacy"
	case "expertise_page":
		return "/expertise"
	case "service_page":
		if len(args) > 0 {
			return fmt.Sprintf("/services/%v", args[0])
		}
	case "article_page":
		if len(args) > 0 {
			return "/articles/" + fmt.Sprint(args[0])
		}
	case "static":
		if len(args) > 1 {
			return "/static/" + fmt.Sprint(args[1])
		}
	}
	return "#"
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data pongo2.Context) {
	data["services"] = content.Services
	data["url_for"] = r.URLFor
	tpl, err := r.set.FromFile(name)
	if err != nil {
		r.log.Printf("template %s parse: %v", name, err)
		if name == "404.html" {
			http.Error(w, "template error", http.StatusInternalServerError)
			return
		}
		r.NotFound(w)
		return
	}
	if err = tpl.ExecuteWriter(data, w); err != nil {
		r.log.Printf("template %s execute: %v", name, err)
	}
}

func (r *Renderer) NotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	r.Render(w, "404.html", pongo2.Context{})
}
