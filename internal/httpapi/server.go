package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yecharlot/NexoWork/internal/domain"
	"github.com/yecharlot/NexoWork/internal/store"
)

const WhatsAppLink = "https://wa.me/5350937508"

type Server struct {
	store      *store.Store
	adminToken string
	webDir     string
}

func New(st *store.Store, adminToken, webDir string) *Server {
	return &Server{store: st, adminToken: adminToken, webDir: webDir}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/api/meta", s.meta)
	mux.HandleFunc("/api/requests", s.requests)
	mux.HandleFunc("/api/requests/", s.requestByID)
	mux.HandleFunc("/api/providers", s.providers)
	mux.HandleFunc("/api/matches", s.matches)
	mux.HandleFunc("/api/admin/", s.adminRoute)

	// Static frontend
	fs := http.FileServer(http.Dir(s.webDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		// SPA-ish: serve files if exist, else index.html
		path := filepath.Join(s.webDir, filepath.Clean(r.URL.Path))
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(s.webDir, "index.html"))
			return
		}
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		// admin path
		if strings.HasPrefix(r.URL.Path, "/admin") {
			http.ServeFile(w, r, filepath.Join(s.webDir, "admin", "index.html"))
			return
		}
		http.ServeFile(w, r, filepath.Join(s.webDir, "index.html"))
	})
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-NEXO-Admin-Token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{
		"status": "ok", "service": "NEXO", "engine": "PrismaTec-Core",
		"time": time.Now().UTC(),
	})
}

func (s *Server) meta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{
		"name": "NEXO",
		"tagline": "Conectamos lo que se necesita con quien lo puede hacer",
		"whatsapp": WhatsAppLink,
		"phone": "+53 50937508",
		"email": "lhmolam@gmail.com",
		"payments": []string{"CUP", "USD", "Zelle", "MLC"},
		"service_types": domain.ServiceTypes,
		"municipalities": domain.Municipalities,
	})
}

func (s *Server) requests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var in struct {
			ServiceType       string `json:"service_type"`
			Description       string `json:"description"`
			Municipality      string `json:"municipality"`
			PreferredDate     string `json:"preferred_date"`
			ContactName       string `json:"contact_name"`
			ContactPhone      string `json:"contact_phone"`
			FromAbroad        bool   `json:"from_abroad"`
			AbroadName        string `json:"abroad_name"`
			AbroadContact     string `json:"abroad_contact"`
			LocalContactName  string `json:"local_contact_name"`
			LocalContactPhone string `json:"local_contact_phone"`
		}
		if err := decodeJSON(r, &in); err != nil {
			writeJSON(w, 400, map[string]string{"error": "JSON inválido"})
			return
		}
		// Normalize abroad form fields
		name := strings.TrimSpace(in.ContactName)
		phone := strings.TrimSpace(in.ContactPhone)
		if in.FromAbroad {
			if name == "" {
				name = strings.TrimSpace(in.LocalContactName)
			}
			if phone == "" {
				phone = strings.TrimSpace(in.LocalContactPhone)
			}
		}
		if in.ServiceType == "" || name == "" || phone == "" {
			writeJSON(w, 400, map[string]string{"error": "service_type, nombre y teléfono son obligatorios"})
			return
		}
		req, err := s.store.CreateRequest(domain.ServiceRequest{
			ServiceType:       strings.TrimSpace(in.ServiceType),
			Description:       strings.TrimSpace(in.Description),
			Municipality:      strings.TrimSpace(in.Municipality),
			PreferredDate:     strings.TrimSpace(in.PreferredDate),
			ContactName:       name,
			ContactPhone:      phone,
			FromAbroad:        in.FromAbroad,
			AbroadName:        strings.TrimSpace(in.AbroadName),
			AbroadContact:     strings.TrimSpace(in.AbroadContact),
			LocalContactName:  strings.TrimSpace(in.LocalContactName),
			LocalContactPhone: strings.TrimSpace(in.LocalContactPhone),
		})
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		waText := buildRequestWAText(req)
		writeJSON(w, 201, map[string]any{
			"request": req,
			"whatsapp_url": WhatsAppLink + "?text=" + urlQueryEscape(waText),
			"message": "Solicitud registrada. Continúa por WhatsApp para coordinar.",
		})
	case http.MethodGet:
		if !s.isAdmin(r) {
			writeJSON(w, 401, map[string]string{"error": "admin token required"})
			return
		}
		list, err := s.store.ListRequests()
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"requests": list, "total": len(list)})
	default:
		writeJSON(w, 405, map[string]string{"error": "method not allowed"})
	}
}

func (s *Server) requestByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/requests/")
	id = strings.Trim(id, "/")
	if id == "" {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	if r.Method == http.MethodPatch {
		if !s.isAdmin(r) {
			writeJSON(w, 401, map[string]string{"error": "admin token required"})
			return
		}
		var in struct {
			Status string `json:"status"`
		}
		if err := decodeJSON(r, &in); err != nil || in.Status == "" {
			writeJSON(w, 400, map[string]string{"error": "status required"})
			return
		}
		req, err := s.store.UpdateRequestStatus(id, in.Status)
		if err != nil {
			writeJSON(w, 404, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, req)
		return
	}
	writeJSON(w, 405, map[string]string{"error": "method not allowed"})
}

func (s *Server) providers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var in struct {
			Services   []string `json:"services"`
			Zone       string   `json:"zone"`
			Name       string   `json:"name"`
			Phone      string   `json:"phone"`
			Experience string   `json:"experience"`
		}
		if err := decodeJSON(r, &in); err != nil {
			writeJSON(w, 400, map[string]string{"error": "JSON inválido"})
			return
		}
		if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Phone) == "" {
			writeJSON(w, 400, map[string]string{"error": "nombre y teléfono obligatorios"})
			return
		}
		p, err := s.store.CreateProvider(domain.ServiceProvider{
			Services:   in.Services,
			Zone:       strings.TrimSpace(in.Zone),
			Name:       strings.TrimSpace(in.Name),
			Phone:      strings.TrimSpace(in.Phone),
			Experience: strings.TrimSpace(in.Experience),
		})
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		waText := "Hola NEXO, me registro como proveedor.\nNombre: " + p.Name + "\nTel: " + p.Phone + "\nZona: " + p.Zone
		writeJSON(w, 201, map[string]any{
			"provider": p,
			"whatsapp_url": WhatsAppLink + "?text=" + urlQueryEscape(waText),
			"message": "Registro recibido. Continúa por WhatsApp.",
		})
	case http.MethodGet:
		if !s.isAdmin(r) {
			writeJSON(w, 401, map[string]string{"error": "admin token required"})
			return
		}
		list, err := s.store.ListProviders()
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"providers": list, "total": len(list)})
	default:
		writeJSON(w, 405, map[string]string{"error": "method not allowed"})
	}
}

func (s *Server) matches(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if !s.isAdmin(r) {
			writeJSON(w, 401, map[string]string{"error": "admin token required"})
			return
		}
		var in struct {
			RequestID  string `json:"request_id"`
			ProviderID string `json:"provider_id"`
			Notes      string `json:"notes"`
		}
		if err := decodeJSON(r, &in); err != nil || in.RequestID == "" || in.ProviderID == "" {
			writeJSON(w, 400, map[string]string{"error": "request_id y provider_id requeridos"})
			return
		}
		m, err := s.store.CreateMatch(domain.Match{
			RequestID:  in.RequestID,
			ProviderID: in.ProviderID,
			Notes:      in.Notes,
		})
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 201, m)
	case http.MethodGet:
		if !s.isAdmin(r) {
			writeJSON(w, 401, map[string]string{"error": "admin token required"})
			return
		}
		list, err := s.store.ListMatches()
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"matches": list, "total": len(list)})
	default:
		writeJSON(w, 405, map[string]string{"error": "method not allowed"})
	}
}

func (s *Server) adminRoute(w http.ResponseWriter, r *http.Request) {
	if !s.isAdmin(r) {
		writeJSON(w, 401, map[string]string{"error": "admin token required"})
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/")
	switch path {
	case "summary":
		reqs, _ := s.store.ListRequests()
		provs, _ := s.store.ListProviders()
		matches, _ := s.store.ListMatches()
		writeJSON(w, 200, map[string]any{
			"requests": len(reqs), "providers": len(provs), "matches": len(matches),
			"engine": "PrismaTec-Core",
		})
	default:
		writeJSON(w, 404, map[string]string{"error": "not found"})
	}
}

func (s *Server) isAdmin(r *http.Request) bool {
	tok := strings.TrimSpace(r.Header.Get("X-NEXO-Admin-Token"))
	if tok == "" {
		h := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(h), "bearer ") {
			tok = strings.TrimSpace(h[7:])
		}
	}
	if s.adminToken == "" || tok == "" {
		return false
	}
	if len(tok) != len(s.adminToken) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(tok), []byte(s.adminToken)) == 1
}

func buildRequestWAText(req domain.ServiceRequest) string {
	var b strings.Builder
	b.WriteString("Hola NEXO, envié una solicitud de servicio.\n")
	b.WriteString("ID: " + req.ID + "\n")
	b.WriteString("Servicio: " + req.ServiceType + "\n")
	if req.Municipality != "" {
		b.WriteString("Municipio: " + req.Municipality + "\n")
	}
	if req.Description != "" {
		b.WriteString("Detalle: " + req.Description + "\n")
	}
	b.WriteString("Contacto: " + req.ContactName + " / " + req.ContactPhone + "\n")
	if req.FromAbroad {
		b.WriteString("(Solicitud desde el exterior)\n")
		if req.AbroadName != "" {
			b.WriteString("Solicitante exterior: " + req.AbroadName + " " + req.AbroadContact + "\n")
		}
	}
	return b.String()
}

func urlQueryEscape(s string) string {
	// minimal escape for wa.me text
	replacer := strings.NewReplacer(
		" ", "%20", "\n", "%0A", "&", "%26", "=", "%3D", "?", "%3F",
		"#", "%23", "+", "%2B",
	)
	return replacer.Replace(s)
}
