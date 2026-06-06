package models

import (
	"errors"
)

// CALL
var ErrCallNotFound = errors.New("call not found")
var ErrCallConvert = errors.New("call convert error")
var ErrUnsupportedAudioType = errors.New("unsupported audio type")
var ErrInvalidCallTitle = errors.New("invalid call title")
var ErrInvalidCallOwner = errors.New("invalid call owner")
var ErrInvalidCallPlacement = errors.New("invalid call placement")
var ErrInvalidCallStatus = errors.New("invalid call status")
var ErrInvalidCallStatusTransition = errors.New("invalid call status transition")

// USER
var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidUserInput = errors.New("invalid user input")

// COMPANY
var ErrCompanyNotFound = errors.New("company not found")
var ErrInvalidCompanyInput = errors.New("invalid company input")
var ErrUserAlreadyManagesCompany = errors.New("user already manages company")

// DEPARTMENT
var ErrDepartmentNotFound = errors.New("department not found")
var ErrInvalidDepartmentInput = errors.New("invalid department input")
var ErrForbidden = errors.New("forbidden")

// REFRESH SESSION
var ErrRefreshSessionNotFound = errors.New("refresh session not found")
var ErrInvalidRefreshToken = errors.New("invalid refresh token")
