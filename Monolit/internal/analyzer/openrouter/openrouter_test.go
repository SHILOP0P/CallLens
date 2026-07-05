package openrouter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

func TestNewRequiresAPIKey(t *testing.T) {
	_, err := New("", "google/gemini-2.5-flash")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewRequiresModel(t *testing.T) {
	_, err := New("sk-or-v1-test", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAnalyzeSendsTranscriptionAndInstructions(t *testing.T) {
	callID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != chatPath {
			t.Fatalf("path = %s, want %s", r.URL.Path, chatPath)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-or-v1-test" {
			t.Fatalf("authorization = %q", got)
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "google/gemini-2.5-flash" {
			t.Fatalf("model = %q", req.Model)
		}
		if req.Temperature == nil || *req.Temperature != 0 {
			t.Fatalf("temperature = %v", req.Temperature)
		}
		if req.ResponseFormat.Type != "json_schema" || req.ResponseFormat.JSONSchema.Name != "call_analysis" {
			t.Fatalf("response format = %#v", req.ResponseFormat)
		}
		if len(req.Messages) != 2 {
			t.Fatalf("messages len = %d", len(req.Messages))
		}
		if !strings.Contains(req.Messages[0].Content, "Абсолютное правило языка") {
			t.Fatalf("system message does not require Russian output:\n%s", req.Messages[0].Content)
		}
		if !strings.Contains(req.Messages[1].Content, "без диалога") {
			t.Fatalf("user message does not contain strict non-dialogue rule:\n%s", req.Messages[1].Content)
		}

		userMessage := req.Messages[1].Content
		for _, want := range []string{
			callID.String(),
			"Проверить приветствие",
			"Менеджер должен поздороваться",
			"Клиент сказал, что цена высокая.",
		} {
			if !strings.Contains(userMessage, want) {
				t.Fatalf("user message does not contain %q:\n%s", want, userMessage)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"model": "google/gemini-2.5-flash",
			"choices": [{
				"message": {
					"role": "assistant",
					"content": "{\"summary\":\"Клиент возражал по цене.\",\"score\":80,\"criteria_results\":[{\"instruction_title\":\"Проверить приветствие\",\"result\":\"Приветствие было\",\"evidence_quotes\":[\"Здравствуйте\"]}],\"customer_objections\":[\"Цена высокая\"],\"risks\":[],\"next_steps\":[\"Отправить расчет\"],\"evidence_quotes\":[\"цена высокая\"],\"confidence\":\"high\"}"
				}
			}]
		}`))
	}))
	defer server.Close()

	analyzer, err := New("sk-or-v1-test", "google/gemini-2.5-flash")
	if err != nil {
		t.Fatalf("new analyzer: %v", err)
	}
	analyzer.baseURL = server.URL
	analyzer.client = server.Client()

	got, err := analyzer.Analyze(context.Background(), models.AnalysisRequest{
		CallUUID:      callID,
		Transcription: "Менеджер: Здравствуйте. Клиент сказал, что цена высокая.",
		Instructions: []models.AnalysisInstructionContent{
			{
				ID:      uuid.New(),
				Scope:   models.AnalysisInstructionScopeCompany,
				Title:   "Проверить приветствие",
				Content: "Менеджер должен поздороваться.",
			},
		},
	})
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if got.Model == nil || *got.Model != "google/gemini-2.5-flash" {
		t.Fatalf("model = %v", got.Model)
	}
	if got.ResultText == nil || *got.ResultText != "Клиент возражал по цене." {
		t.Fatalf("result text = %v", got.ResultText)
	}
	if !json.Valid(got.ResultJSON) {
		t.Fatalf("result json is invalid: %s", string(got.ResultJSON))
	}
}

func TestAnalyzeWrapsNonJSONResponse(t *testing.T) {
	resultJSON, resultText, err := normalizeAnalysisContent("Обычный текстовый ответ")
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if resultText != "Обычный текстовый ответ" {
		t.Fatalf("result text = %q", resultText)
	}

	var payload map[string]any
	if err = json.Unmarshal(resultJSON, &payload); err != nil {
		t.Fatalf("decode result json: %v", err)
	}
	if payload["raw_response"] != "Обычный текстовый ответ" {
		t.Fatalf("raw response = %v", payload["raw_response"])
	}
}

func TestAnalyzeReturnsOpenRouterError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(`{"error":{"message":"insufficient credits","code":402}}`))
	}))
	defer server.Close()

	analyzer, err := New("sk-or-v1-test", "google/gemini-2.5-flash")
	if err != nil {
		t.Fatalf("new analyzer: %v", err)
	}
	analyzer.baseURL = server.URL
	analyzer.client = server.Client()

	_, err = analyzer.Analyze(context.Background(), models.AnalysisRequest{
		CallUUID:      uuid.New(),
		Transcription: "Тестовая транскрипция.",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "status 402") || !strings.Contains(err.Error(), "insufficient credits") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeRejectsEmptyTranscription(t *testing.T) {
	analyzer, err := New("sk-or-v1-test", "google/gemini-2.5-flash")
	if err != nil {
		t.Fatalf("new analyzer: %v", err)
	}

	_, err = analyzer.Analyze(context.Background(), models.AnalysisRequest{
		CallUUID:      uuid.New(),
		Transcription: "   ",
	})
	if !errors.Is(err, models.ErrInvalidAnalysisInput) {
		t.Fatalf("error = %v, want invalid analysis input", err)
	}
}
