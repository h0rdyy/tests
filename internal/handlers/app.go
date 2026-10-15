package handlers

import (
	"database/sql"
	"log"

	"cybersec/internal/mailer"
	"cybersec/internal/middleware"
	"cybersec/internal/templates"
)

// App contains the dependencies shared by the HTTP handlers.
type App struct {
	DB      *sql.DB
	Log     *log.Logger
	Limiter *middleware.Limiter
	Render  *templates.Renderer
	Mailer  *mailer.Mailer
}
