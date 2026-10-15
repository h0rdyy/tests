package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"time"
)

const (
	maxLeadBodySize = 1 << 20
	maxSourceLength = 120

	invalidJSONResponse     = `{"ok":false,"error":"invalid JSON"}`
	invalidPhoneResponse    = `{"ok":false,"error":"Укажите корректный номер телефона"}`
	internalErrorResponse   = `{"ok":false,"error":"internal error"}`
	tooManyRequestsResponse = `{"ok":false,"error":"too many requests"}`
	leadCreatedResponse     = `{"ok":true}`
	insertLeadSQL           = `INSERT INTO leads (phone, created_at, source) VALUES (?, ?, ?)`
)

type LeadRequest struct {
	Phone  string `json:"phone"`
	Source string `json:"source"`
}

// Lead stores a submitted phone lead and notifies the configured recipient.
func (a *App) Lead(w http.ResponseWriter, r *http.Request) {
	if !a.Limiter.Allow(clientIP(r)) {
		http.Error(w, tooManyRequestsResponse, http.StatusTooManyRequests)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxLeadBodySize)
	var request LeadRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, invalidJSONResponse, http.StatusBadRequest)
		return
	}

	phone, ok := normalizePhone(request.Phone)
	if !ok {
		http.Error(w, invalidPhoneResponse, http.StatusBadRequest)
		return
	}
	request.Source = normalizeSource(request.Source)

	createdAt := time.Now().UTC().Format(time.RFC3339)
	if _, err := a.DB.Exec(insertLeadSQL, phone, createdAt, request.Source); err != nil {
		a.Log.Printf("lead insert: %v", err)
		http.Error(w, internalErrorResponse, http.StatusInternalServerError)
		return
	}

	go a.Mailer.Send(phone, request.Source, createdAt)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(leadCreatedResponse))
}

func clientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || ip == "" {
		return r.RemoteAddr
	}
	return ip
}

func normalizeSource(source string) string {
	if len(source) > maxSourceLength {
		source = source[:maxSourceLength]
	}
	if source == "" {
		return "unknown"
	}
	return source
}
