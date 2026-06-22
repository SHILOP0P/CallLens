package analysis_instruction

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"calllens/monolit/internal/httpserver/middleware"
	"calllens/monolit/internal/models"
	serviceMocks "calllens/monolit/internal/service/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestHandlersSuccess(t *testing.T) {
	userID := uuid.New()
	instructionID := uuid.New()
	item := models.AnalysisInstruction{
		ID: instructionID, Scope: models.AnalysisInstructionScopePersonal,
		UserUUID: uuid.NullUUID{UUID: userID, Valid: true}, OriginalFilename: "guide.md",
		MimeType: "text/markdown", SizeBytes: 5, CreatedByUserUUID: userID,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	service := serviceMocks.NewAnalysisInstructionService(t)
	service.EXPECT().Create(mock.Anything, mock.Anything).Return(item, nil).Once()
	service.EXPECT().List(mock.Anything, mock.Anything).Return([]models.AnalysisInstruction{item}, nil).Once()
	service.EXPECT().GetFile(mock.Anything, mock.Anything, mock.Anything).Return(models.File{
		Content: io.NopCloser(strings.NewReader("guide")), OriginalFilename: "guide.md",
		MimeType: "text/markdown", SizeBytes: 5,
	}, nil).Once()
	service.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	handler := NewHandler(service)

	body, contentType := instructionMultipart(t, map[string]string{"scope": "personal"}, "guide.md", "guide")
	rec, req := instructionRequest(http.MethodPost, "/", body.String(), userID, nil)
	req.Header.Set("Content-Type", contentType)
	handler.Create(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Create status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec, req = instructionRequest(http.MethodGet, "/?scope=personal", "", userID, nil)
	handler.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("List status = %d", rec.Code)
	}

	rec, req = instructionRequest(http.MethodGet, "/", "", userID, map[string]string{"uuid": instructionID.String()})
	handler.GetFile(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "guide" {
		t.Fatalf("GetFile status=%d body=%q", rec.Code, rec.Body.String())
	}

	rec, req = instructionRequest(http.MethodDelete, "/", "", userID, map[string]string{"uuid": instructionID.String()})
	handler.Delete(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Delete status = %d", rec.Code)
	}
}

func TestValidationHelpersAndErrorMappings(t *testing.T) {
	userID := uuid.New()
	companyID := uuid.New()
	departmentID := uuid.New()
	tests := []struct {
		scope, company, department string
		ok                         bool
	}{
		{"personal", "", "", true},
		{"company", companyID.String(), "", true},
		{"department", companyID.String(), departmentID.String(), true},
		{"personal", companyID.String(), "", false},
		{"company", "", "", false},
		{"department", companyID.String(), "", false},
		{"unknown", "", "", false},
	}
	for _, tt := range tests {
		_, _, _, _, err := parseInstructionPlacement(tt.scope, userID, tt.company, tt.department)
		if (err == nil) != tt.ok {
			t.Fatalf("placement %+v err=%v", tt, err)
		}
	}
	if _, err := parseInstructionUUID("bad"); err == nil {
		t.Fatal("expected invalid UUID error")
	}
	if got, err := parseInstructionUUID(companyID.String()); err != nil || got != companyID {
		t.Fatalf("parseInstructionUUID = %v, %v", got, err)
	}

	errorCases := []struct {
		err  error
		code int
	}{
		{models.ErrInvalidAnalysisInstructionInput, http.StatusBadRequest},
		{models.ErrUnsupportedInstructionType, http.StatusBadRequest},
		{models.ErrInstructionLimitExceeded, http.StatusBadRequest},
		{models.ErrSubscriptionRequired, http.StatusPaymentRequired},
		{models.ErrAnalysisInstructionNotFound, http.StatusNotFound},
		{models.ErrCompanyNotFound, http.StatusNotFound},
		{models.ErrDepartmentNotFound, http.StatusNotFound},
		{models.ErrForbidden, http.StatusForbidden},
		{errors.New("db"), http.StatusInternalServerError},
	}
	for _, tt := range errorCases {
		rec := httptest.NewRecorder()
		writeInstructionError(rec, tt.err, "fallback", "failed")
		if rec.Code != tt.code {
			t.Fatalf("instruction error %v: %d", tt.err, rec.Code)
		}
	}

	fileErrors := []struct {
		err  error
		code int
	}{
		{models.ErrInvalidAnalysisInstructionInput, http.StatusBadRequest},
		{models.ErrAnalysisInstructionNotFound, http.StatusNotFound},
		{models.ErrInstructionFileNotFound, http.StatusNotFound},
		{models.ErrCompanyNotFound, http.StatusNotFound},
		{models.ErrDepartmentNotFound, http.StatusNotFound},
		{models.ErrForbidden, http.StatusForbidden},
		{errors.New("db"), http.StatusInternalServerError},
	}
	for _, tt := range fileErrors {
		rec := httptest.NewRecorder()
		writeInstructionFileError(rec, tt.err)
		if rec.Code != tt.code {
			t.Fatalf("file error %v: %d", tt.err, rec.Code)
		}
	}
}

func TestHandlersRejectInvalidRequests(t *testing.T) {
	handler := NewHandler(serviceMocks.NewAnalysisInstructionService(t))
	for _, method := range []func(http.ResponseWriter, *http.Request){handler.Create, handler.List, handler.GetFile, handler.Delete} {
		rec, req := instructionRequest(http.MethodGet, "/", "", uuid.Nil, nil)
		method(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("unauthorized status = %d", rec.Code)
		}
	}

	rec, req := instructionRequest(http.MethodPost, "/", "invalid", uuid.New(), nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=x")
	handler.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid multipart status = %d", rec.Code)
	}

	rec, req = instructionRequest(http.MethodGet, "/?scope=unknown", "", uuid.New(), nil)
	handler.List(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid list status = %d", rec.Code)
	}

	for _, method := range []func(http.ResponseWriter, *http.Request){handler.GetFile, handler.Delete} {
		rec, req = instructionRequest(http.MethodGet, "/", "", uuid.New(), map[string]string{"uuid": "bad"})
		method(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("invalid UUID status = %d", rec.Code)
		}
	}
}

func instructionRequest(method, path, body string, userID uuid.UUID, params map[string]string) (*httptest.ResponseRecorder, *http.Request) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if userID != uuid.Nil {
		req = req.WithContext(middleware.ContextWithUserID(req.Context(), userID))
	}
	route := chi.NewRouteContext()
	for key, value := range params {
		route.URLParams.Add(key, value)
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, route))
	return httptest.NewRecorder(), req
}

func instructionMultipart(t *testing.T, fields map[string]string, name, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte(content))
	_ = writer.Close()
	return body, writer.FormDataContentType()
}
