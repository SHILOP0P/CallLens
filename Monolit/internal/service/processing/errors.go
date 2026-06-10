package processing

import (
	"calllens/monolit/internal/models"
	"errors"
)

func isPermanentProcessingError(err error) bool {
	return errors.Is(err, models.ErrInvalidProcessingJobType) ||
		errors.Is(err, models.ErrCallNotFound) ||
		errors.Is(err, models.ErrInvalidCallStatus) ||
		errors.Is(err, models.ErrInvalidCallStatusTransition) ||
		errors.Is(err, models.ErrAudioFileNotFound) ||
		errors.Is(err, models.ErrInvalidAudioPath) ||
		errors.Is(err, models.ErrTranscriberNotConfigured)
}
