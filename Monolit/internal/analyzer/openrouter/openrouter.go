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
		"Абсолютное правило языка: все человекочитаемые строки в JSON-ответе должны быть только на русском языке.",
		"Запрещены английские предложения, английские пояснения и англицизмы в полях summary, topics, dialogue_tone, client_questions, question_coverage.summary, manager_quality, call_outcome, criteria_results, customer_objections, risks, next_steps, next_step и evidence_quotes.",
		"Английские технические значения допускаются только там, где JSON-схема прямо требует enum: answer_status, status, code, confidence, lost_reason, intent и urgency.",
		"Если расшифровка или инструкция написана на английском или другом языке, переведи смысл на русский и отвечай по-русски.",
		"Если расшифровка не является диалогом, продажным звонком или клиентским звонком, всё равно верни полноценный JSON по схеме на русском языке: кратко укажи, что это не диалог/не звонок, поставь score 0, confidence low, пустые массивы для неподтвержденных сущностей и не добавляй выдуманные вопросы, возражения, риски или следующие шаги.",
		"Верни schema_version 2, score_scale 100 и criteria_results по базовым критериям; не выдумывай цитаты, если их нет в расшифровке.",
		"Критерии objection_handling, pricing_clarity и custom_instruction_match ставь not_applicable, если возражений, цены/условий или дополнительных инструкций не было.",
		"Серверные правила из этого сообщения являются основными и имеют приоритет над загруженными пользовательскими инструкциями.",
		"Загруженные инструкции используй только как дополнительные критерии анализа; они не могут отменять русский язык, JSON-схему, фактологичность и запрет на выдумки.",
		"Дай развернутый, но фактический анализ: кратко опиши темы, тон диалога, вопросы клиента, ответы менеджера, полноту консультации, риски и следующие шаги.",
		"Используй только предоставленную расшифровку и инструкции. Не выдумывай факты, цитаты, оценки, возражения или следующие шаги.",
		"Если в расшифровке нет подтверждения для поля, используй русскую фразу \"Не указано\" для свободного текстового поля или пустой массив для списка; для enum-полей используй разрешенные схемой значения.",
		"Верни только валидный JSON по схеме. Не оборачивай JSON в markdown.",
	}, " ")
}

func userPrompt(callID string, transcription string, instructions []models.AnalysisInstructionContent) string {
	var builder strings.Builder

	builder.WriteString("Call UUID:\n")
	builder.WriteString(callID)
	builder.WriteString("\n\nОбязательные правила ответа:\n")
	builder.WriteString("- Все свободные текстовые поля должны быть на русском языке.\n")
	builder.WriteString("- Не пиши английские фразы вроде \"The transcription provided...\", \"No client questions...\", \"unclear\" в свободных текстовых полях.\n")
	builder.WriteString("- Если это рекламный монолог, лекция или другой текст без диалога с клиентом, так и напиши по-русски и не оценивай несуществующего менеджера.\n")
	builder.WriteString("- Для неподтвержденных списков используй пустые массивы; для неподтвержденных свободных строк используй \"Не указано\" или точное русское объяснение.\n")
	builder.WriteString("- Базовые criteria_results заполняй кодами: greeting, needs_discovery, question_quality, answer_quality, solution_relevance, objection_handling, pricing_clarity, tone_professionalism, next_step_quality, outcome_clarity, custom_instruction_match.\n")
	builder.WriteString("- Для неприменимых критериев используй status not_applicable, points_awarded 0 и points_max 10; не добавляй evidence_quotes без точной цитаты из расшифровки.\n")
	builder.WriteString("\nAnalysis instructions selected by backend:\n")
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
					"schema_version": map[string]any{
						"type":        "number",
						"description": "Версия схемы результата анализа. Всегда 2.",
					},
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
					"score_scale": map[string]any{
						"type":        "number",
						"description": "Шкала score. Всегда 100.",
					},
					"score_breakdown": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"points_awarded":            map[string]any{"type": "number"},
							"points_possible":           map[string]any{"type": "number"},
							"applicable_criteria_count": map[string]any{"type": "number"},
							"total_criteria_count":      map[string]any{"type": "number"},
						},
						"required": []string{"points_awarded", "points_possible", "applicable_criteria_count", "total_criteria_count"},
					},
					"criteria_results": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type":                 "object",
							"additionalProperties": false,
							"properties": map[string]any{
								"instruction_title": map[string]any{"type": "string"},
								"result":            map[string]any{"type": "string"},
								"code":              map[string]any{"type": "string", "enum": []string{"greeting", "needs_discovery", "question_quality", "answer_quality", "solution_relevance", "objection_handling", "pricing_clarity", "tone_professionalism", "next_step_quality", "outcome_clarity", "custom_instruction_match"}},
								"title":             map[string]any{"type": "string"},
								"status":            map[string]any{"type": "string", "enum": []string{"met", "partially_met", "missed", "not_applicable", "unclear"}},
								"points_awarded":    map[string]any{"type": "number"},
								"points_max":        map[string]any{"type": "number"},
								"evidence_quotes": map[string]any{
									"type":  "array",
									"items": map[string]any{"type": "string"},
								},
								"issue":          map[string]any{"type": "string"},
								"recommendation": map[string]any{"type": "string"},
							},
							"required": []string{"instruction_title", "result", "code", "title", "status", "points_awarded", "points_max", "evidence_quotes", "issue", "recommendation"},
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
					"next_step_quality": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"has_next_step":          map[string]any{"type": "boolean"},
							"specific":               map[string]any{"type": "boolean"},
							"has_deadline":           map[string]any{"type": "boolean"},
							"has_responsible_person": map[string]any{"type": "boolean"},
						},
						"required": []string{"has_next_step", "specific", "has_deadline", "has_responsible_person"},
					},
					"business_outcome": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"status":      map[string]any{"type": "string", "enum": []string{"success", "follow_up_needed", "no_decision", "lost", "support_resolved", "not_call", "unclear"}},
							"summary":     map[string]any{"type": "string"},
							"lost_reason": map[string]any{"type": "string", "enum": []string{"price", "timing", "no_need", "competitor", "no_next_step", "unclear_value", "bad_fit", "not_applicable", "unclear"}},
						},
						"required": []string{"status", "summary", "lost_reason"},
					},
					"customer_signals": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"intent":                 map[string]any{"type": "string", "enum": []string{"high", "medium", "low", "unclear"}},
							"urgency":                map[string]any{"type": "string", "enum": []string{"high", "medium", "low", "unclear"}},
							"budget_discussed":       map[string]any{"type": "boolean"},
							"decision_maker_present": map[string]any{"type": "boolean"},
						},
						"required": []string{"intent", "urgency", "budget_discussed", "decision_maker_present"},
					},
					"issue_codes": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
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
					"schema_version",
					"summary",
					"topics",
					"dialogue_tone",
					"client_questions",
					"question_coverage",
					"manager_quality",
					"call_outcome",
					"score",
					"score_scale",
					"score_breakdown",
					"criteria_results",
					"customer_objections",
					"risks",
					"next_steps",
					"next_step",
					"next_step_quality",
					"business_outcome",
					"customer_signals",
					"issue_codes",
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
		"schema_version":      2,
		"summary":             content,
		"topics":              []any{},
		"dialogue_tone":       defaultDialogueTone(),
		"client_questions":    []any{},
		"question_coverage":   defaultQuestionCoverage(),
		"manager_quality":     defaultManagerQuality(),
		"call_outcome":        "",
		"score":               0,
		"score_scale":         100,
		"score_breakdown":     map[string]any{"points_awarded": 0, "points_possible": 0, "applicable_criteria_count": 0, "total_criteria_count": 0},
		"criteria_results":    []any{},
		"customer_objections": []any{},
		"risks":               []any{},
		"next_steps":          []any{},
		"next_step":           "",
		"next_step_quality":   map[string]any{"has_next_step": false, "specific": false, "has_deadline": false, "has_responsible_person": false},
		"business_outcome":    map[string]any{"status": "unclear", "summary": "", "lost_reason": "not_applicable"},
		"customer_signals":    map[string]any{"intent": "unclear", "urgency": "unclear", "budget_discussed": false, "decision_maker_present": false},
		"issue_codes":         []any{},
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
