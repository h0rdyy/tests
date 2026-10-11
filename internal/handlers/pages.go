package handlers

import (
	"net/http"
	"regexp"
	"strconv"

	"cybersec/internal/content"
	"github.com/flosch/pongo2/v6"
)

var servicePath = regexp.MustCompile(`^/services/([1-5])$`)

// Pages renders the site's public HTML pages.
func (a *App) Pages(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		a.Render.Render(w, "index.html", pongo2.Context{})
	case "/about":
		a.Render.Render(w, "about.html", pongo2.Context{})
	case "/privacy":
		a.Render.Render(w, "privacy.html", pongo2.Context{})
	case "/expertise":
		a.Render.Render(w, "expertise.html", pongo2.Context{})
	case "/articles/kibernsk":
		a.Render.Render(w, "articles/kibernsk.html", pongo2.Context{})
	default:
		a.renderServiceOrNotFound(w, r)
	}
}

func (a *App) renderServiceOrNotFound(w http.ResponseWriter, r *http.Request) {
	matches := servicePath.FindStringSubmatch(r.URL.Path)
	if matches != nil {
		id, _ := strconv.Atoi(matches[1])
		if service, ok := content.FindService(id); ok {
			a.Render.Render(w, "service.html", pongo2.Context{"service": service})
			return
		}
	}
	a.Render.NotFound(w)
}
