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

// USER
var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidUserInput = errors.New("invalid user input")

// REFRESH SESSION
var ErrRefreshSessionNotFound = errors.New("refresh session not found")
var ErrInvalidRefreshToken = errors.New("invalid refresh token")
