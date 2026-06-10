package processing

import (
	"calllens/monolit/internal/models"
	"errors"
	"fmt"
	"testing"
)

func TestIsPermanentProcessingError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "invalid job type",
			err:  models.ErrInvalidProcessingJobType,
			want: true,
		},
		{
			name: "wrapped call not found",
			err:  fmt.Errorf("get call for processing: %w", models.ErrCallNotFound),
			want: true,
		},
		{
			name: "invalid status transition",
			err:  fmt.Errorf("validate call status: %w", models.ErrInvalidCallStatusTransition),
			want: true,
		},
		{
			name: "invalid status",
			err:  fmt.Errorf("validate call status: %w", models.ErrInvalidCallStatus),
			want: true,
		},
		{
			name: "missing audio",
			err:  fmt.Errorf("open audio: %w", models.ErrAudioFileNotFound),
			want: true,
		},
		{
			name: "invalid audio path",
			err:  fmt.Errorf("open audio: %w", models.ErrInvalidAudioPath),
			want: true,
		},
		{
			name: "transcriber not configured",
			err:  models.ErrTranscriberNotConfigured,
			want: true,
		},
		{
			name: "temporary provider error",
			err:  errors.New("provider timeout"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPermanentProcessingError(tt.err); got != tt.want {
				t.Fatalf("isPermanentProcessingError() = %v, want %v", got, tt.want)
			}
		})
	}
}
