package gateway

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/openclio/openclio/internal/config"
	"github.com/openclio/openclio/internal/control"
)

type browserActionRequest struct {
	Enabled *bool `json:"enabled"`
}

type approvalsActionRequest struct {
	AllowAll *bool `json:"allow_all"`
}

type execProfileActionRequest struct {
	Profile string `json:"profile"`
}

type modelActionRequest struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url,omitempty"`
}

type mcpServerActionRequest struct {
	Name    string `json:"name"`
	Enabled *bool  `json:"enabled"`
}

type channelActionRequest struct {
	Name           string `json:"name"`
	Action         string `json:"action"`
	ForceReconnect bool   `json:"force_reconnect,omitempty"`
}

type sessionActionRequest struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

type cronActionRequest struct {
	Name    string `json:"name"`
	Action  string `json:"action"`
	Enabled *bool  `json:"enabled,omitempty"`
}

func (h *Handlers) controlActionEnv() control.ActionEnv {
	var deleteSession func(id string) error
	if h.sessions != nil {
		deleteSession = h.sessions.Delete
	}
	var runCron func(name string) error
	var setCronEnabled func(name string, enabled bool) error
	var deleteCron func(name string) error
	if h.scheduler != nil {
		runCron = h.scheduler.RunNow
		setCronEnabled = func(name string, enabled bool) error {
			_, err := h.scheduler.SetPersistentEnabled(name, enabled)
			return err
		}
		deleteCron = h.scheduler.DeletePersistent
	}
	return control.ActionEnv{
		Config:           h.cfg,
		DataDir:          h.dataDir,
		Allowlist:        h.allowlist,
		Runtime:          h,
		ChannelConnector: h.channelControl,
		ChannelLifecycle: h.channelLife,
		DeleteSession:    deleteSession,
		RunCron:          runCron,
		SetCronEnabled:   setCronEnabled,
		DeleteCron:       deleteCron,
		WriteTools:       true,
	}
}

func (h *Handlers) SyncBrowserTool() {
	h.syncRuntimeBrowserTool()
}

func (h *Handlers) SyncExecTool() {
	h.syncRuntimeExecTool()
}

// ControlsCatalog returns the shared CLI/UI control catalog.
func (h *Handlers) ControlsCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"groups": control.Catalog(),
	})
}

// StatusSummary returns the shared runtime status summary used by CLI and UI.
func (h *Handlers) StatusSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	setupRequired, setupReason := h.setupState()
	sessionCount := 0
	if h.sessions != nil {
		if n, err := h.sessions.Count(); err == nil {
			sessionCount = n
		}
	}
	channelCount := 0
	if h.manager != nil {
		channelCount = len(h.manager.Statuses())
	}
	cronJobCount := 0
	if h.scheduler != nil {
		cronJobCount = len(h.scheduler.ListJobs())
	}
	uptime := int64(0)
	if !h.startedAt.IsZero() {
		uptime = int64(time.Since(h.startedAt).Seconds())
	}
	summary := control.BuildStatusSummary("ok", h.cfg, setupRequired, setupReason, uptime, sessionCount, channelCount, cronJobCount)
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"text":    control.FormatStatusSummaryText(summary),
	})
}

// AuthSummary returns the shared auth summary used by CLI and UI.
func (h *Handlers) AuthSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	oc := h.cfg.Auth.OpenAIOAuth
	configured := oc.Enabled &&
		strings.TrimSpace(oc.ClientID) != "" &&
		strings.TrimSpace(oc.AuthorizationURL) != "" &&
		strings.TrimSpace(oc.TokenURL) != ""
	signedIn := false
	var expiresAt time.Time
	message := ""
	tok, err := ReadOpenAIOAuthToken(h.dataDir)
	if err != nil || tok == nil {
		message = "no token stored"
	} else {
		expiresAt = tok.ExpiresAt
		signedIn = time.Until(tok.ExpiresAt) > 60*time.Second
		if !signedIn {
			message = "stored token expired or near expiry"
		}
	}
	summary := control.BuildAuthSummary(configured, signedIn, expiresAt, message)
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"text":    control.FormatAuthSummaryText(summary),
	})
}

// PluginsSummary returns the shared runtime plugin summary used by CLI and UI.
func (h *Handlers) PluginsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	items := []control.PluginSummaryItem{}
	if h.manager != nil {
		for _, st := range h.manager.Statuses() {
			message := strings.TrimSpace(st.LastHealthError)
			if message == "" {
				message = strings.TrimSpace(st.LastError)
			}
			items = append(items, control.PluginSummaryItem{
				Name:      st.Name,
				Running:   st.Running,
				Healthy:   st.Healthy,
				Restarted: st.RestartCount,
				Message:   message,
			})
		}
	}
	summary := control.BuildPluginSummary(items)
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"text":    control.FormatPluginSummaryText(summary),
	})
}

// Doctor returns a shared health and readiness report used by CLI and UI.
func (h *Handlers) Doctor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}

	configPath := ""
	if strings.TrimSpace(h.dataDir) != "" {
		configPath = filepath.Join(h.dataDir, "config.yaml")
	}

	var mcpStatuses []control.MCPRuntimeStatus
	if h.mcpStatus != nil {
		for _, st := range h.mcpStatus.SnapshotMCPStatus() {
			mcpStatuses = append(mcpStatuses, control.MCPRuntimeStatus{
				Name:    st.Name,
				Healthy: st.Healthy,
				Status:  st.Status,
				Error:   st.LastHealthError,
			})
		}
	}

	var channelStats []control.ChannelDoctorStatus
	if h.manager != nil {
		for _, st := range h.manager.Statuses() {
			message := st.LastHealthError
			if strings.TrimSpace(message) == "" {
				message = st.LastError
			}
			channelStats = append(channelStats, control.ChannelDoctorStatus{
				Name:       st.Name,
				Configured: true,
				Running:    st.Running,
				Healthy:    st.Healthy,
				Message:    message,
			})
		}
	}

	report := control.BuildDoctorReport(control.DoctorInput{
		Config:       h.cfg,
		DataDir:      h.dataDir,
		ConfigPath:   configPath,
		MCPStatuses:  mcpStatuses,
		ChannelStats: channelStats,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"report": report,
		"text":   control.FormatDoctorReportText(report),
	})
}

// ModelsSummary returns the shared model summary used by CLI and UI.
func (h *Handlers) ModelsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	summary := control.BuildModelSummary(h.cfg)
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"text":    control.FormatModelSummaryText(summary),
	})
}

// ChannelsSummary returns the shared channel summary used by CLI and UI.
func (h *Handlers) ChannelsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	summary := control.BuildChannelSummary(h.cfg)
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"text":    control.FormatChannelSummaryText(summary),
	})
}

// ToolsSummary returns the shared tooling summary used by CLI and UI.
func (h *Handlers) ToolsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	summary := control.BuildToolingSummary(h.cfg)
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"text":    control.FormatToolingSummaryText(summary),
	})
}

// ApprovalsSummary returns the shared approvals summary used by CLI and UI.
func (h *Handlers) ApprovalsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	allowAll := true
	var approved []string
	if h.allowlist != nil {
		allowAll = h.allowlist.AllowAll()
		approved = h.allowlist.List()
	}
	summary := control.BuildApprovalsSummary(h.cfg, allowAll, approved)
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"text":    control.FormatApprovalsSummaryText(summary),
	})
}

// LogsSummary returns the shared logs summary used by CLI and UI.
func (h *Handlers) LogsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	summary := control.BuildLogsSummary(h.cfg)
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"text":    control.FormatLogsSummaryText(summary),
	})
}

// BrowserSummary returns the shared browser summary used by CLI and UI.
func (h *Handlers) BrowserSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	summary := control.BuildBrowserSummary(h.cfg)
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"text":    control.FormatBrowserSummaryText(summary),
	})
}

// SessionsSummary returns the shared sessions summary used by CLI and UI.
func (h *Handlers) SessionsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	total, err := h.sessions.Count()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count sessions: "+err.Error())
		return
	}
	recent, err := h.sessions.List(10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sessions: "+err.Error())
		return
	}
	items := make([]control.SessionSummaryItem, 0, len(recent))
	for _, s := range recent {
		lastActive := s.LastActive
		if lastActive.IsZero() {
			lastActive = s.CreatedAt
		}
		items = append(items, control.SessionSummaryItem{
			ID:         s.ID,
			Channel:    s.Channel,
			LastActive: formatSummaryTime(lastActive),
		})
	}
	summary := control.BuildSessionSummary(total, items)
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"text":    control.FormatSessionSummaryText(summary),
	})
}

// CronSummary returns the shared cron summary used by CLI and UI.
func (h *Handlers) CronSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	if h.scheduler == nil {
		summary := control.BuildCronSummary(0, 0, 0, nil)
		writeJSON(w, http.StatusOK, map[string]any{
			"summary": summary,
			"text":    control.FormatCronSummaryText(summary),
		})
		return
	}
	jobs := h.scheduler.ListJobs()
	items := make([]control.CronSummaryItem, 0, len(jobs))
	enabled := 0
	for _, job := range jobs {
		if job.Enabled {
			enabled++
		}
		items = append(items, control.CronSummaryItem{
			Name:    job.Name,
			Enabled: job.Enabled,
			NextRun: formatSummaryTime(job.NextRun),
			Source:  job.Source,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	history, err := h.scheduler.History(20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load cron history: "+err.Error())
		return
	}
	recentFailures := 0
	for _, entry := range history {
		if !entry.Success {
			recentFailures++
		}
	}
	summary := control.BuildCronSummary(len(jobs), enabled, recentFailures, items)
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": summary,
		"text":    control.FormatCronSummaryText(summary),
	})
}

// BrowserAction mutates the browser automation enabled state through the shared control layer.
func (h *Handlers) BrowserAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "use PUT")
		return
	}
	var payload browserActionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	if payload.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}
	result, err := control.SetBrowserEnabled(h.controlActionEnv(), *payload.Enabled)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func formatSummaryTime(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}

// ApprovalsAction mutates channel allowlist mode through the shared control layer.
func (h *Handlers) ApprovalsAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "use PUT")
		return
	}
	var payload approvalsActionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	if payload.AllowAll == nil {
		writeError(w, http.StatusBadRequest, "allow_all is required")
		return
	}
	result, err := control.SetAllowAllMode(h.controlActionEnv(), *payload.AllowAll)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// ExecProfileAction mutates the local exec profile through the shared control layer.
func (h *Handlers) ExecProfileAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "use PUT")
		return
	}
	var payload execProfileActionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	result, err := control.SetExecProfile(h.controlActionEnv(), payload.Profile)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// ModelAction mutates the active provider/model selection through the shared control layer.
func (h *Handlers) ModelAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "use PUT")
		return
	}
	var payload modelActionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	previous := h.cfg.Model
	result, err := control.SetActiveModelConfig(h.controlActionEnv(), payload.Provider, payload.Model, payload.BaseURL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.agent != nil {
		providerImpl, err := buildProviderStackFromConfig(h.cfg, h.dataDir)
		if err != nil {
			h.cfg.Model = previous
			_ = config.Save(filepath.Join(h.dataDir, "config.yaml"), h.cfg)
			writeError(w, http.StatusBadRequest, "provider initialization failed: "+err.Error())
			return
		}
		h.agent.SetProvider(providerImpl, h.cfg.Model.Model)
		h.setSetupState(false, "")
		if h.cfg.Model.Provider == "ollama" && h.cfg.Embeddings.Provider != "openai" {
			ensureOllamaEmbeddingModelPulled(h.cfg.Embeddings.Model)
		}
	}
	writeJSON(w, http.StatusOK, result)
}

// MCPServerAction toggles one configured MCP server stub through the shared control layer.
func (h *Handlers) MCPServerAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "use PUT")
		return
	}
	var payload mcpServerActionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	if payload.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}
	result, err := control.SetMCPServerEnabled(h.controlActionEnv(), payload.Name, *payload.Enabled)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// ChannelsAction runs runtime channel actions through the shared control layer.
func (h *Handlers) ChannelsAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "use PUT")
		return
	}
	var payload channelActionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	if h.manager == nil {
		writeError(w, http.StatusServiceUnavailable, "channel manager is not configured")
		return
	}
	result, err := control.RunChannelAction(h.controlActionEnv(), payload.Name, payload.Action, payload.ForceReconnect)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	name := strings.TrimSpace(strings.ToLower(payload.Name))
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"name":    name,
		"action":  strings.TrimSpace(strings.ToLower(payload.Action)),
		"status":  latestChannelStatus(h.manager, name),
		"updated": result.Updated,
		"message": result.Note,
	})
}

// SessionsAction runs session mutations through the shared control layer.
func (h *Handlers) SessionsAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "use PUT")
		return
	}
	var payload sessionActionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	if strings.ToLower(strings.TrimSpace(payload.Action)) != "delete" {
		writeError(w, http.StatusBadRequest, "action must be delete")
		return
	}
	result, err := control.DeleteSession(h.controlActionEnv(), payload.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"id":      strings.TrimSpace(payload.ID),
		"action":  "delete",
		"updated": result.Updated,
		"message": result.Note,
	})
}

// CronAction runs cron mutations through the shared control layer.
func (h *Handlers) CronAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "use PUT")
		return
	}
	var payload cronActionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	result, err := control.RunCronMutation(h.controlActionEnv(), payload.Name, payload.Action, payload.Enabled)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"name":    strings.TrimSpace(payload.Name),
		"action":  strings.TrimSpace(strings.ToLower(payload.Action)),
		"updated": result.Updated,
		"message": result.Note,
	})
}
