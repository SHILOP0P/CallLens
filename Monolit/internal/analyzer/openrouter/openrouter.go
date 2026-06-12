package openrouter

import (
	"bytes"
	"calllens/monolit/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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
	defer resp.Body.Close()

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
		"You analyze sales or support call transcriptions for CallLens.",
		"Use only the provided transcription and analysis instructions.",
		"Do not invent facts, quotes, scores, objections, or next steps.",
		"If evidence is missing, use empty arrays, score 0, and confidence \"low\".",
		"Follow all provided instructions. Department instructions may refine company instructions.",
		"Return only valid JSON matching the schema. Do not wrap JSON in markdown.",
	}, " ")
}

func userPrompt(callID string, transcription string, instructions []models.AnalysisInstructionContent) string {
	var builder strings.Builder

	builder.WriteString("Call UUID:\n")
	builder.WriteString(callID)
	builder.WriteString("\n\nAnalysis instructions selected by backend:\n")
	if len(instructions) == 0 {
		builder.WriteString("No uploaded instructions were selected. Use the generic schema fields only.\n")
	} else {
		for i, instruction := range instructions {
			builder.WriteString(fmt.Sprintf("\n### Instruction %d\n", i+1))
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
						"description": "Short factual call summary.",
					},
					"topics": map[string]any{
						"type":        "array",
						"description": "Main call topics as short labels.",
						"items":       map[string]any{"type": "string"},
					},
					"score": map[string]any{
						"type":        "number",
						"description": "Numeric score from 0 to 100. Use 0 when instructions do not define scoring or evidence is missing.",
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
