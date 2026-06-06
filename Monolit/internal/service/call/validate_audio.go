package call

import (
	"calllens/monolit/internal/models"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var allowedAudioExtensions = map[string]struct{}{
	".mp3": {},
	".wav": {},
	".m4a": {},
}

var allowedAudioMimeTypes = map[string]struct{}{
	"audio/mpeg":     {},
	"audio/wav":      {},
	"audio/x-wav":    {},
	"audio/wave":     {},
	"audio/vnd.wave": {},
	"audio/mp4":      {},
	"audio/x-m4a":    {},
}

func validateAudioInput(input models.CreateCallInput) error {
	if input.UploadedByUserUUID == uuid.Nil {
		return models.ErrInvalidCallOwner
	}

	ext := strings.ToLower(filepath.Ext(input.OriginalFilename))
	if _, ok := allowedAudioExtensions[ext]; !ok {
		return models.ErrUnsupportedAudioType
	}

	if _, ok := allowedAudioMimeTypes[input.MimeType]; !ok {
		return models.ErrUnsupportedAudioType
	}

	if input.SizeBytes <= 0 {
		return models.ErrCallConvert
	}

	if input.Content == nil {
		return models.ErrUnsupportedAudioType
	}

	if err := validateCallPlacement(input); err != nil {
		return err
	}

	return nil
}

func validateCallPlacement(input models.CreateCallInput) error {
	switch input.VisibilityScope {
	case models.CallVisibilityScopePersonal:
		if input.CompanyUUID.Valid || input.DepartmentUUID.Valid {
			return models.ErrInvalidCallPlacement
		}
	case models.CallVisibilityScopeCompany:
		if !input.CompanyUUID.Valid || input.DepartmentUUID.Valid {
			return models.ErrInvalidCallPlacement
		}
	case models.CallVisibilityScopeDepartment:
		if !input.CompanyUUID.Valid || !input.DepartmentUUID.Valid {
			return models.ErrInvalidCallPlacement
		}
	default:
		return models.ErrInvalidCallPlacement
	}

	return nil
}
