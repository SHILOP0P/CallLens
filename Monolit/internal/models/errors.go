package models

import (
	"errors"
)

var ErrCallNotFound = errors.New("call not found")
var ErrCallConvert = errors.New("call convert error")
var ErrUnsupportedAudioType = errors.New("unsupported audio type")
