package prompt_profile

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/httpserver/middleware"
	"calllens/monolit/internal/models"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type Store interface {
	Industries(context.Context, string) ([]models.PromptIndustry, error)
	Topics(context.Context, string, string) ([]models.PromptTopic, error)
	Recommend(context.Context, string, string) ([]models.PromptTopic, error)
	ListProfiles(context.Context, uuid.UUID) ([]models.PromptProfile, error)
	SaveProfile(context.Context, models.PromptProfile) (models.PromptProfile, error)
	DeleteProfile(context.Context, uuid.UUID, uuid.UUID) error
	PutCallContext(context.Context, models.CallPromptContext) error
	CallContext(context.Context, uuid.UUID, uuid.UUID) (models.CallPromptContext, error)
	UserSettings(context.Context, uuid.UUID) (models.PromptUserSettings, error)
	SaveUserSettings(context.Context, models.PromptUserSettings) (models.PromptUserSettings, error)
}
type CallReader interface {
	GetByUUID(context.Context, uuid.UUID, uuid.UUID) (models.Call, error)
}
type Handler struct {
	store Store
	calls CallReader
}

func NewHandler(s Store, calls CallReader) *Handler { return &Handler{store: s, calls: calls} }
func user(r *http.Request) (uuid.UUID, bool)        { return middleware.UserIDFromContext(r.Context()) }
func (h *Handler) Industries(w http.ResponseWriter, r *http.Request) {
	if _, ok := user(r); !ok {
		response.WriteError(w, 401, "unauthorized", "unauthorized")
		return
	}
	out, e := h.store.Industries(r.Context(), r.URL.Query().Get("perspective"))
	if e != nil {
		response.WriteError(w, 500, "prompt_catalog_failed", "failed to list industries")
		return
	}
	response.WriteJSON(w, 200, out)
}
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	id, ok := user(r)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	settings, err := h.store.UserSettings(r.Context(), id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "prompt_settings_failed", "failed to load prompt settings")
		return
	}
	_ = response.WriteJSON(w, http.StatusOK, settings)
}
func (h *Handler) SaveSettings(w http.ResponseWriter, r *http.Request) {
	id, ok := user(r)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var input struct {
		Description  string   `json:"description"`
		IndustryKeys []string `json:"industry_keys"`
		TopicKeys    []string `json:"topic_keys"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_prompt_settings", "invalid settings")
		return
	}
	settings := models.PromptUserSettings{UserID: id, Description: strings.TrimSpace(input.Description), Industries: make([]models.PromptIndustry, 0, len(input.IndustryKeys)), Topics: make([]models.PromptTopic, 0, len(input.TopicKeys))}
	for _, key := range input.IndustryKeys {
		settings.Industries = append(settings.Industries, models.PromptIndustry{Key: key})
	}
	for _, key := range input.TopicKeys {
		settings.Topics = append(settings.Topics, models.PromptTopic{Key: key, Source: "manual"})
	}
	result, err := h.store.SaveUserSettings(r.Context(), settings)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_prompt_settings", "failed to save prompt settings")
		return
	}
	_ = response.WriteJSON(w, http.StatusOK, result)
}
func (h *Handler) Topics(w http.ResponseWriter, r *http.Request) {
	if _, ok := user(r); !ok {
		response.WriteError(w, 401, "unauthorized", "unauthorized")
		return
	}
	industry := strings.TrimSpace(r.URL.Query().Get("industry_key"))
	if industry == "" {
		response.WriteError(w, 400, "invalid_prompt_catalog_input", "industry_key is required")
		return
	}
	out, e := h.store.Topics(r.Context(), industry, r.URL.Query().Get("q"))
	if e != nil {
		response.WriteError(w, 500, "prompt_catalog_failed", "failed to list topics")
		return
	}
	response.WriteJSON(w, 200, out)
}
func (h *Handler) Recommend(w http.ResponseWriter, r *http.Request) {
	if _, ok := user(r); !ok {
		response.WriteError(w, 401, "unauthorized", "unauthorized")
		return
	}
	var in struct {
		Perspective string `json:"perspective"`
		Description string `json:"description"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.Description) == "" {
		response.WriteError(w, 400, "invalid_prompt_catalog_input", "perspective and description are required")
		return
	}
	out, e := h.store.Recommend(r.Context(), in.Perspective, in.Description)
	if e != nil {
		response.WriteError(w, 500, "prompt_catalog_failed", "failed to recommend topics")
		return
	}
	response.WriteJSON(w, 200, out)
}
func (h *Handler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	id, ok := user(r)
	if !ok {
		response.WriteError(w, 401, "unauthorized", "unauthorized")
		return
	}
	out, e := h.store.ListProfiles(r.Context(), id)
	if e != nil {
		response.WriteError(w, 500, "prompt_profiles_failed", "failed to list profiles")
		return
	}
	response.WriteJSON(w, 200, out)
}
func (h *Handler) SaveProfile(w http.ResponseWriter, r *http.Request) {
	id, ok := user(r)
	if !ok {
		response.WriteError(w, 401, "unauthorized", "unauthorized")
		return
	}
	var in struct {
		ID          string               `json:"id"`
		Title       string               `json:"title"`
		Perspective string               `json:"perspective"`
		IndustryKey string               `json:"industry_key"`
		Description string               `json:"description"`
		IsDefault   bool                 `json:"is_default"`
		Topics      []models.PromptTopic `json:"topics"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.IndustryKey) == "" || (in.Perspective != "business" && in.Perspective != "personal") {
		response.WriteError(w, 400, "invalid_prompt_profile", "title, perspective and industry_key are required")
		return
	}
	p := models.PromptProfile{OwnerUserID: id, Title: strings.TrimSpace(in.Title), Perspective: in.Perspective, IndustryKey: in.IndustryKey, Description: in.Description, IsDefault: in.IsDefault, Topics: in.Topics}
	if raw := chi.URLParam(r, "uuid"); raw != "" {
		var e error
		p.ID, e = uuid.Parse(raw)
		if e != nil {
			response.WriteError(w, 400, "invalid_prompt_profile", "invalid profile uuid")
			return
		}
	} else if in.ID != "" {
		p.ID, _ = uuid.Parse(in.ID)
	}
	out, e := h.store.SaveProfile(r.Context(), p)
	if e != nil {
		response.WriteError(w, 500, "prompt_profiles_failed", "failed to save profile")
		return
	}
	response.WriteJSON(w, 200, out)
}
func (h *Handler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	owner, ok := user(r)
	if !ok {
		response.WriteError(w, 401, "unauthorized", "unauthorized")
		return
	}
	id, e := uuid.Parse(chi.URLParam(r, "uuid"))
	if e != nil {
		response.WriteError(w, 400, "invalid_prompt_profile", "invalid profile uuid")
		return
	}
	if e = h.store.DeleteProfile(r.Context(), id, owner); e != nil {
		response.WriteError(w, 404, "prompt_profile_not_found", "profile not found")
		return
	}
	response.WriteNoContent(w)
}
func (h *Handler) GetCallContext(w http.ResponseWriter, r *http.Request) {
	owner, ok := user(r)
	if !ok {
		response.WriteError(w, 401, "unauthorized", "unauthorized")
		return
	}
	id, e := uuid.Parse(chi.URLParam(r, "uuid"))
	if e != nil {
		response.WriteError(w, 400, "invalid_call_uuid", "invalid call uuid")
		return
	}
	if h.calls != nil {
		if _, e = h.calls.GetByUUID(r.Context(), id, owner); e != nil {
			response.WriteError(w, http.StatusNotFound, "call_not_found", "call not found")
			return
		}
	}
	out, e := h.store.CallContext(r.Context(), id, owner)
	if e != nil {
		out = models.CallPromptContext{CallID: id, OwnerUserID: owner, TopicKeys: []string{}}
	}
	response.WriteJSON(w, 200, out)
}
func (h *Handler) PutCallContext(w http.ResponseWriter, r *http.Request) {
	owner, ok := user(r)
	if !ok {
		response.WriteError(w, 401, "unauthorized", "unauthorized")
		return
	}
	callID, e := uuid.Parse(chi.URLParam(r, "uuid"))
	if e != nil {
		response.WriteError(w, 400, "invalid_call_uuid", "invalid call uuid")
		return
	}
	if h.calls != nil {
		if _, e = h.calls.GetByUUID(r.Context(), callID, owner); e != nil {
			response.WriteError(w, http.StatusNotFound, "call_not_found", "call not found")
			return
		}
	}
	var in struct {
		ProfileID string   `json:"profile_id"`
		TopicKeys []string `json:"topic_keys"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		response.WriteError(w, 400, "invalid_prompt_context", "invalid context")
		return
	}
	p := uuid.Nil
	if in.ProfileID != "" {
		p, e = uuid.Parse(in.ProfileID)
		if e != nil {
			response.WriteError(w, 400, "invalid_prompt_context", "invalid profile id")
			return
		}
	}
	e = h.store.PutCallContext(r.Context(), models.CallPromptContext{CallID: callID, ProfileID: p, OwnerUserID: owner, TopicKeys: in.TopicKeys})
	if e != nil {
		response.WriteError(w, 500, "prompt_context_failed", "failed to save context")
		return
	}
	response.WriteJSON(w, 200, map[string]any{"call_uuid": callID, "profile_uuid": p, "topic_keys": in.TopicKeys})
}
