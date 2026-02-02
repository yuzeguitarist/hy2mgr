package web

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yuzeguitarist/hy2mgr/internal/audit"
	"github.com/yuzeguitarist/hy2mgr/internal/crypto"
	"github.com/yuzeguitarist/hy2mgr/internal/service"
	"github.com/yuzeguitarist/hy2mgr/internal/state"
	"github.com/yuzeguitarist/hy2mgr/internal/systemd"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	Store *sessions.CookieStore
	State *state.State
}

func NewServer(st *state.State, sessionKey []byte) *Server {
	cs := sessions.NewCookieStore(sessionKey)
	cs.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	return &Server{Store: cs, State: st}
}

func (s *Server) Router() http.Handler {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.FileServerFS(FS))
	r.HandleFunc("/login", s.loginPage).Methods("GET")
	r.HandleFunc("/login", s.loginPost).Methods("POST")
	r.HandleFunc("/logout", s.logout).Methods("GET")

	// public subscription (token protected)
	r.HandleFunc("/sub/{token}", s.subscription).Methods("GET")

	authed := r.NewRoute().Subrouter()
	authed.Use(s.requireLogin)
	authed.HandleFunc("/", s.appShell).Methods("GET")
	authed.HandleFunc("/api/dashboard", s.apiDashboard).Methods("GET")
	authed.HandleFunc("/api/logs", s.apiLogs).Methods("GET")
	authed.HandleFunc("/api/nodes", s.apiNodes).Methods("GET")
	authed.HandleFunc("/api/nodes", s.apiNodesCreate).Methods("POST")
	authed.HandleFunc("/api/nodes/{id}", s.apiNodesDelete).Methods("DELETE")
	authed.HandleFunc("/api/nodes/{id}/disable", s.apiNodeDisable).Methods("POST")
	authed.HandleFunc("/api/nodes/{id}/enable", s.apiNodeEnable).Methods("POST")
	authed.HandleFunc("/api/nodes/{id}/reset", s.apiNodeReset).Methods("POST")
	authed.HandleFunc("/api/nodes/{id}/uri", s.apiNodeURI).Methods("GET")
	authed.HandleFunc("/api/nodes/{id}/qrcode.png", s.apiNodeQRPNG).Methods("GET")
	authed.HandleFunc("/api/nodes/{id}/qrcode.svg", s.apiNodeQRSVG).Methods("GET")
	authed.HandleFunc("/api/subscription", s.apiSubscriptionInfo).Methods("GET")
	authed.HandleFunc("/api/subscription/rotate", s.apiSubscriptionRotate).Methods("POST")
	authed.HandleFunc("/api/settings", s.apiSettings).Methods("GET")
	authed.HandleFunc("/api/settings", s.apiSettingsSave).Methods("POST")
	authed.HandleFunc("/api/cert/rotate", s.apiCertRotate).Methods("POST")
	authed.HandleFunc("/api/admin/password", s.apiAdminPassword).Methods("POST")

	// CSRF for all POSTs
	return csrf.Protect([]byte("change-this-in-prod"), csrf.Secure(false))(r)
}

func (s *Server) appShell(w http.ResponseWriter, r *http.Request) {
	b, _ := FS.ReadFile("templates/layout.html")
	w.Header().Set("content-type", "text/html; charset=utf-8")
	_, _ = w.Write(b)
}

func (s *Server) loginPage(w http.ResponseWriter, r *http.Request) {
	b, _ := FS.ReadFile("templates/login.html")
	html := string(b)
	html = strings.ReplaceAll(html, "{{.CSRFToken}}", csrf.Token(r))
	if s.State.Admin.TOTPEnabled {
		html = strings.ReplaceAll(html, "{{if .TOTPEnabled}}", "")
	} else {
		html = strings.ReplaceAll(html, "{{if .TOTPEnabled}}", "")
		html = strings.ReplaceAll(html, "{{end}}", "")
		html = strings.ReplaceAll(html, "    <label>TOTP</label>\n    <input name=\"totp\" inputmode=\"numeric\" autocomplete=\"one-time-code\" />\n    ", "")
	}
	// no error
	html = strings.ReplaceAll(html, "{{if .Error}}", "")
	html = strings.ReplaceAll(html, "{{.Error}}", "")
	html = strings.ReplaceAll(html, "{{end}}", "")
	w.Header().Set("content-type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

func (s *Server) loginPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	u := r.FormValue("username")
	p := r.FormValue("password")
	totp := r.FormValue("totp")

	if u != s.State.Admin.Username || bcrypt.CompareHashAndPassword([]byte(s.State.Admin.PasswordBcrypt), []byte(p)) != nil {
		s.loginFail(w, r, "invalid credentials")
		return
	}
	if s.State.Admin.TOTPEnabled && !verifyTOTP(s.State.Admin.TOTPSecret, totp) {
		s.loginFail(w, r, "invalid totp")
		return
	}

	sess, _ := s.Store.Get(r, "hy2mgr")
	sess.Values["auth"] = true
	sess.Values["ts"] = time.Now().Unix()
	_ = sess.Save(r, w)
	audit.Write(audit.Entry{Time: time.Now().UTC().Format(time.RFC3339), IP: clientIP(r), User: u, Action: "login"})
	http.Redirect(w, r, "/#dashboard", http.StatusFound)
}

func (s *Server) loginFail(w http.ResponseWriter, r *http.Request, msg string) {
	b, _ := FS.ReadFile("templates/login.html")
	html := string(b)
	html = strings.ReplaceAll(html, "{{.CSRFToken}}", csrf.Token(r))
	if s.State.Admin.TOTPEnabled {
		html = strings.ReplaceAll(html, "{{if .TOTPEnabled}}", "")
	} else {
		html = strings.ReplaceAll(html, "{{if .TOTPEnabled}}", "")
		html = strings.ReplaceAll(html, "{{end}}", "")
		html = strings.ReplaceAll(html, "    <label>TOTP</label>\n    <input name=\"totp\" inputmode=\"numeric\" autocomplete=\"one-time-code\" />\n    ", "")
	}
	html = strings.ReplaceAll(html, "{{if .Error}}", "")
	html = strings.ReplaceAll(html, "{{.Error}}", msg)
	html = strings.ReplaceAll(html, "{{end}}", "")
	w.Header().Set("content-type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := s.Store.Get(r, "hy2mgr")
	sess.Options.MaxAge = -1
	_ = sess.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *Server) requireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, _ := s.Store.Get(r, "hy2mgr")
		if v, ok := sess.Values["auth"].(bool); !ok || !v {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) apiDashboard(w http.ResponseWriter, r *http.Request) {
	active, _ := systemd.IsActive("hysteria-server.service")
	pin, _ := crypto.ParseCertPin("/etc/hysteria/cert.crt")
	logs, _ := systemd.JournalTail("hysteria-server.service", 200)
	recent := filterErrors(logs)
	resp := map[string]any{
		"hysteriaStatus": map[bool]string{true: "active", false: "inactive"}[active],
		"listen":         fmt.Sprintf(":%d", s.State.Settings.ListenPort),
		"port":           s.State.Settings.ListenPort,
		"pin":            pin,
		"recentErrors":   recent,
	}
	writeJSON(w, resp)
}

func (s *Server) apiLogs(w http.ResponseWriter, r *http.Request) {
	lines, _ := strconv.Atoi(r.URL.Query().Get("lines"))
	if lines <= 0 || lines > 2000 {
		lines = 200
	}
	logs, _ := systemd.JournalTail("hysteria-server.service", lines)
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(logs))
}

func (s *Server) apiSettings(w http.ResponseWriter, r *http.Request) { writeJSON(w, s.State.Settings) }

func (s *Server) apiSettingsSave(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ListenPort        int    `json:"listenPort"`
		SNI               string `json:"sni"`
		MasqueradeURL     string `json:"masqueradeUrl"`
		MasqueradeRewrite bool   `json:"masqueradeRewrite"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if in.ListenPort <= 0 || in.ListenPort > 65535 {
		http.Error(w, "invalid port", 400)
		return
	}
	if in.SNI == "" {
		in.SNI = "www.bing.com"
	}
	if in.MasqueradeURL == "" {
		in.MasqueradeURL = "https://www.bing.com"
	}
	s.State.Settings.ListenPort = in.ListenPort
	s.State.Settings.SNI = in.SNI
	s.State.Settings.MasqueradeURL = in.MasqueradeURL
	s.State.Settings.MasqueradeRewrite = in.MasqueradeRewrite

	if err := service.Apply(s.State, false); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	audit.Write(audit.Entry{Time: time.Now().UTC().Format(time.RFC3339), IP: clientIP(r), User: s.State.Admin.Username, Action: "settings.save"})
	writeJSON(w, map[string]any{"ok": true, "port": s.State.Settings.ListenPort})
}

func (s *Server) apiNodes(w http.ResponseWriter, r *http.Request) {
	type nodeOut struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Enabled  bool   `json:"enabled"`
	}
	var out []nodeOut
	for _, n := range s.State.NodesSorted() {
		out = append(out, nodeOut{ID: n.ID, Name: n.Name, Username: n.Username, Enabled: n.Enabled})
	}
	writeJSON(w, map[string]any{"nodes": out})
}

func (s *Server) apiNodesCreate(w http.ResponseWriter, r *http.Request) {
	var in struct{ Name string `json:"name"` }
	_ = json.NewDecoder(r.Body).Decode(&in)
	if strings.TrimSpace(in.Name) == "" {
		http.Error(w, "name required", 400)
		return
	}
	n, err := service.NodeAdd(s.State, in.Name, "", "")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	audit.Write(audit.Entry{Time: time.Now().UTC().Format(time.RFC3339), IP: clientIP(r), User: s.State.Admin.Username, Action: "node.add", Object: n.ID})
	writeJSON(w, map[string]any{"ok": true, "id": n.ID})
}

func (s *Server) apiNodesDelete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := service.NodeDelete(s.State, id); err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	audit.Write(audit.Entry{Time: time.Now().UTC().Format(time.RFC3339), IP: clientIP(r), User: s.State.Admin.Username, Action: "node.delete", Object: id})
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) apiNodeDisable(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := service.NodeSetEnabled(s.State, id, false); err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	audit.Write(audit.Entry{Time: time.Now().UTC().Format(time.RFC3339), IP: clientIP(r), User: s.State.Admin.Username, Action: "node.disable", Object: id})
	writeJSON(w, map[string]any{"ok": true})
}
func (s *Server) apiNodeEnable(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := service.NodeSetEnabled(s.State, id, true); err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	audit.Write(audit.Entry{Time: time.Now().UTC().Format(time.RFC3339), IP: clientIP(r), User: s.State.Admin.Username, Action: "node.enable", Object: id})
	writeJSON(w, map[string]any{"ok": true})
}
func (s *Server) apiNodeReset(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := service.NodeResetPassword(s.State, id); err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	audit.Write(audit.Entry{Time: time.Now().UTC().Format(time.RFC3339), IP: clientIP(r), User: s.State.Admin.Username, Action: "node.reset", Object: id})
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) apiNodeURI(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	uri, err := service.NodeURI(s.State, id)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	writeJSON(w, map[string]any{"uri": uri})
}

func (s *Server) apiNodeQRPNG(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	uri, err := service.NodeURI(s.State, id)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	png, err := QrPNG(uri, 256)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "image/png")
	_, _ = w.Write(png)
}
func (s *Server) apiNodeQRSVG(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	uri, err := service.NodeURI(s.State, id)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	svg, err := QrSVG(uri, 6)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "image/svg+xml")
	_, _ = w.Write(svg)
}

func (s *Server) apiSubscriptionInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"url":  "/sub/<token>",
		"note": "Token is sensitive; view via CLI: `hy2mgr export subscription`",
	})
}

func (s *Server) apiSubscriptionRotate(w http.ResponseWriter, r *http.Request) {
	token, urlPath, err := service.SubscriptionRotate(s.State)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	audit.Write(audit.Entry{Time: time.Now().UTC().Format(time.RFC3339), IP: clientIP(r), User: s.State.Admin.Username, Action: "subscription.rotate"})
	writeJSON(w, map[string]any{"ok": true, "token": token, "url": urlPath})
}

func (s *Server) subscription(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	if !service.SubscriptionVerify(s.State, token) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	lines := []string{}
	for _, n := range s.State.NodesSorted() {
		if !n.Enabled {
			continue
		}
		uri, err := service.NodeURI(s.State, n.ID)
		if err == nil {
			lines = append(lines, uri)
		}
	}
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(strings.Join(lines, "\n") + "\n"))
}

func (s *Server) apiCertRotate(w http.ResponseWriter, r *http.Request) {
	if err := service.RotateCert(s.State, false); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = service.Apply(s.State, false)
	audit.Write(audit.Entry{Time: time.Now().UTC().Format(time.RFC3339), IP: clientIP(r), User: s.State.Admin.Username, Action: "cert.rotate"})
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) apiAdminPassword(w http.ResponseWriter, r *http.Request) {
	var in struct{ Password string `json:"password"` }
	_ = json.NewDecoder(r.Body).Decode(&in)
	if len(in.Password) < 8 {
		http.Error(w, "too short", 400)
		return
	}
	h, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	s.State.Admin.PasswordBcrypt = string(h)
	_ = s.State.SaveAtomic()
	audit.Write(audit.Entry{Time: time.Now().UTC().Format(time.RFC3339), IP: clientIP(r), User: s.State.Admin.Username, Action: "admin.password.rotate"})
	writeJSON(w, map[string]any{"ok": true})
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func clientIP(r *http.Request) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	if host == "" {
		return r.RemoteAddr
	}
	return host
}

func filterErrors(logs string) string {
	lines := strings.Split(logs, "\n")
	var out []string
	for _, l := range lines {
		low := strings.ToLower(l)
		if strings.Contains(low, "error") || strings.Contains(low, "failed") || strings.Contains(low, "permission denied") {
			out = append(out, l)
		}
	}
	if len(out) == 0 {
		return ""
	}
	if len(out) > 30 {
		out = out[len(out)-30:]
	}
	return strings.Join(out, "\n")
}
