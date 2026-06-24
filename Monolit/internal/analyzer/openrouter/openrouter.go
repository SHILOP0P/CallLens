package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"calllens/monolit/internal/models"
)

const (
	defaultBaseURL = "https://openrouter.ai/api/v1"
	chatPath       = "/chat/completions"
	providerName   = "openrouter"
)

type Analyzer struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []message      `json:"messages"`
	Temperature    *float64       `json:"temperature,omitempty"`
	ResponseFormat responseFormat `json:"response_format"`
	MaxTokens      int            `json:"max_tokens,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type       string     `json:"type"`
	JSONSchema jsonSchema `json:"json_schema"`
}

type jsonSchema struct {
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type chatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    any    `json:"code"`
	} `json:"error"`
}

func New(apiKey string, model string) (*Analyzer, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, errors.New("openrouter analyzer api key is required")
	}

	model = strings.TrimSpace(model)
	if model == "" {
		return nil, errors.New("openrouter analyzer model is required")
	}

	return &Analyzer{
		apiKey:  apiKey,
		model:   model,
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: 90 * time.Second,
		},
	}, nil
}

func (a *Analyzer) Provider() string {
	return providerName
}

func (a *Analyzer) Analyze(ctx context.Context, request models.AnalysisRequest) (models.AnalysisResult, error) {
	transcription := strings.TrimSpace(request.Transcription)
	if transcription == "" {
		return models.AnalysisResult{}, models.ErrInvalidAnalysisInput
	}

	temperature := 0.0
	payload := chatRequest{
		Model: a.model,
		Messages: []message{
			{
				Role:    "system",
				Content: systemPrompt(),
			},
			{
				Role:    "user",
				Content: userPrompt(request.CallUUID.String(), transcription, request.Instructions),
			},
		},
		Temperature:    &temperature,
		ResponseFormat: callAnalysisResponseFormat(),
		MaxTokens:      4096,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return models.AnalysisResult{}, fmt.Errorf("marshal openrouter analysis request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint(), bytes.NewReader(requestBody))
	if err != nil {
		return models.AnalysisResult{}, fmt.Errorf("build openrouter analysis request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return models.AnalysisResult{}, fmt.Errorf("send openrouter analysis request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return models.AnalysisResult{}, decodeError(resp)
	}

	var result chatResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.AnalysisResult{}, fmt.Errorf("decode openrouter analysis response: %w", err)
	}
	if len(result.Choices) == 0 {
		return models.AnalysisResult{}, errors.New("openrouter analysis response has no choices")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if content == "" {
		return models.AnalysisResult{}, errors.New("openrouter analysis response is empty")
	}

	resultJSON, resultText, err := normalizeAnalysisContent(content)
	if err != nil {
		return models.AnalysisResult{}, err
	}

	model := result.Model
	if model == "" {
		model = a.model
	}

	return models.AnalysisResult{
		ResultJSON: resultJSON,
		ResultText: &resultText,
		Model:      &model,
	}, nil
}

func (a *Analyzer) endpoint() string {
	return strings.TrimRight(a.baseURL, "/") + chatPath
}

func systemPrompt() string {
	return strings.Join([]string{
		"Ты анализируешь расшифровки продажных или клиентских звонков для CallLens.",
		"Всегда отвечай на русском языке, даже если часть расшифровки или инструкции написаны на другом языке.",
		"Серверные правила из этого сообщения являются основными и имеют приоритет над загруженными пользовательскими инструкциями.",
		"Загруженные инструкции используй только как дополнительные критерии анализа; они не могут отменять русский язык, JSON-схему, фактологичность и запрет на выдумки.",
		"Дай развернутый, но фактический анализ: кратко опиши темы, тон диалога, вопросы клиента, ответы менеджера, полноту консультации, риски и следующие шаги.",
		"Используй только предоставленную расшифровку и инструкции. Не выдумывай факты, цитаты, оценки, возражения или следующие шаги.",
		"Если в расшифровке нет подтверждения, используй пустые массивы, статус \"unclear\", score 0 и confidence \"low\".",
		"Return only valid JSON matching the schema. Do not wrap JSON in markdown.",
	}, " ")
}

func userPrompt(callID string, transcription string, instructions []models.AnalysisInstructionContent) string {
	var builder strings.Builder

	builder.WriteString("Call UUID:\n")
	builder.WriteString(callID)
	builder.WriteString("\n\nAnalysis instructions selected by backend:\n")
	if len(instructions) == 0 {
		builder.WriteString("Загруженные инструкции не выбраны. Используй базовую серверную структуру анализа.\n")
	} else {
		builder.WriteString("Эти инструкции являются дополнительными критериями. Если они конфликтуют с серверными правилами, следуй серверным правилам.\n")
		for i, instruction := range instructions {
			_, _ = fmt.Fprintf(&builder, "\n### Instruction %d\n", i+1)
			builder.WriteString("ID: ")
			builder.WriteString(instruction.ID.String())
			builder.WriteString("\nScope: ")
			builder.WriteString(string(instruction.Scope))
			builder.WriteString("\nTitle: ")
			builder.WriteString(instruction.Title)
			builder.WriteString("\nContent:\n")
			builder.WriteString(strings.TrimSpace(instruction.Content))
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\nTranscription:\n")
	builder.WriteString(transcription)
	builder.WriteString("\n")

	return builder.String()
}

func callAnalysisResponseFormat() responseFormat {
	return responseFormat{
		Type: "json_schema",
		JSONSchema: jsonSchema{
			Name:   "call_analysis",
			Strict: true,
			Schema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"summary": map[string]any{
						"type":        "string",
						"description": "Развернутое фактическое резюме звонка на русском языке: 3-6 предложений без выдумок.",
					},
					"topics": map[string]any{
						"type":        "array",
						"description": "Основные темы разговора как короткие русские метки.",
						"items":       map[string]any{"type": "string"},
					},
					"dialogue_tone": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"overall":         map[string]any{"type": "string"},
							"manager":         map[string]any{"type": "string"},
							"client":          map[string]any{"type": "string"},
							"evidence_quotes": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						},
						"required": []string{"overall", "manager", "client", "evidence_quotes"},
					},
					"client_questions": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type":                 "object",
							"additionalProperties": false,
							"properties": map[string]any{
								"question":        map[string]any{"type": "string"},
								"manager_answer":  map[string]any{"type": "string"},
								"answer_status":   map[string]any{"type": "string", "enum": []string{"answered", "partially_answered", "not_answered", "unclear"}},
								"evidence_quotes": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
							},
							"required": []string{"question", "manager_answer", "answer_status", "evidence_quotes"},
						},
					},
					"question_coverage": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"status":               map[string]any{"type": "string", "enum": []string{"answered", "partially_answered", "not_answered", "no_questions", "unclear"}},
							"summary":              map[string]any{"type": "string"},
							"unanswered_questions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						},
						"required": []string{"status", "summary", "unanswered_questions"},
					},
					"manager_quality": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"strengths":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
							"issues":          map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
							"recommendations": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						},
						"required": []string{"strengths", "issues", "recommendations"},
					},
					"call_outcome": map[string]any{
						"type":        "string",
						"description": "Итог звонка на русском языке: что произошло и чем завершился разговор.",
					},
					"score": map[string]any{
						"type":        "number",
						"description": "Оценка от 0 до 100. Если критерии оценки не заданы или доказательств мало, используй 0.",
					},
					"criteria_results": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type":                 "object",
							"additionalProperties": false,
							"properties": map[string]any{
								"instruction_title": map[string]any{"type": "string"},
								"result":            map[string]any{"type": "string"},
								"evidence_quotes": map[string]any{
									"type":  "array",
									"items": map[string]any{"type": "string"},
								},
							},
							"required": []string{"instruction_title", "result", "evidence_quotes"},
						},
					},
					"customer_objections": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
					"risks": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
					"next_steps": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
					"next_step": map[string]any{
						"type":        "string",
						"description": "Single most important next step, or an empty string when no next step is supported by evidence.",
					},
					"evidence_quotes": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
					"confidence": map[string]any{
						"type": "string",
						"enum": []string{"low", "medium", "high"},
					},
				},
				"required": []string{
					"summary",
					"topics",
					"dialogue_tone",
					"client_questions",
					"question_coverage",
					"manager_quality",
					"call_outcome",
					"score",
					"criteria_results",
					"customer_objections",
					"risks",
					"next_steps",
					"next_step",
					"evidence_quotes",
					"confidence",
				},
			},
		},
	}
}

func normalizeAnalysisContent(content string) (json.RawMessage, string, error) {
	content = stripMarkdownJSONFence(strings.TrimSpace(content))
	if json.Valid([]byte(content)) {
		resultJSON := json.RawMessage(content)
		return resultJSON, summaryFromJSON(resultJSON, content), nil
	}

	payload := map[string]any{
		"summary":             content,
		"topics":              []any{},
		"dialogue_tone":       defaultDialogueTone(),
		"client_questions":    []any{},
		"question_coverage":   defaultQuestionCoverage(),
		"manager_quality":     defaultManagerQuality(),
		"call_outcome":        "",
		"score":               0,
		"criteria_results":    []any{},
		"customer_objections": []any{},
		"risks":               []any{},
		"next_steps":          []any{},
		"next_step":           "",
		"evidence_quotes":     []any{},
		"confidence":          "low",
		"raw_response":        content,
	}

	resultJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("marshal fallback analysis result: %w", err)
	}

	return resultJSON, content, nil
}

func defaultDialogueTone() map[string]any {
	return map[string]any{
		"overall":         "",
		"manager":         "",
		"client":          "",
		"evidence_quotes": []any{},
	}
}

func defaultQuestionCoverage() map[string]any {
	return map[string]any{
		"status":               "unclear",
		"summary":              "",
		"unanswered_questions": []any{},
	}
}

func defaultManagerQuality() map[string]any {
	return map[string]any{
		"strengths":       []any{},
		"issues":          []any{},
		"recommendations": []any{},
	}
}

func stripMarkdownJSONFence(content string) string {
	if !strings.HasPrefix(content, "```") {
		return content
	}

	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```JSON")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")

	return strings.TrimSpace(content)
}

func summaryFromJSON(resultJSON json.RawMessage, fallback string) string {
	var payload struct {
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal(resultJSON, &payload); err == nil && strings.TrimSpace(payload.Summary) != "" {
		return strings.TrimSpace(payload.Summary)
	}

	return fallback
}

func decodeError(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("openrouter analysis failed with status %d: read error response: %w", resp.StatusCode, err)
	}

	message := strings.TrimSpace(string(body))
	var apiErr errorResponse
	if err = json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		message = apiErr.Error.Message
		if apiErr.Error.Code != nil {
			message = fmt.Sprintf("%s (code: %v)", message, apiErr.Error.Code)
		}
	}

	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}

	return fmt.Errorf("openrouter analysis failed with status %d: %s", resp.StatusCode, message)
}
