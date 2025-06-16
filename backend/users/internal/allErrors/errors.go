package allerrors

import (
	"errors"
	"strconv"
)

const (
	minPasswordLen = 8
	maxPasswordLen
)

// REPO
var (
	ErrUsernameExists = errors.New("username exists")
	ErrEmailExists    = errors.New("email exists")
	ErrUserNotExists  = errors.New("user not exists")
)

// CACHE
var (
	ErrNotFoundInCache = errors.New("not found in cache")
	ErrTokenInBlackList = errors.New("refresh token in blacklist")
)

// SERVICE
var (
	ErrWrongPass     = errors.New("wrong password")
	ErrPasswordSmall = errors.New("password smaller than " + strconv.Itoa(minPasswordLen))
	ErrPasswordBig   = errors.New("password bigger than " + strconv.Itoa(maxPasswordLen))
	ErrWrongUUID     = errors.New("wrong uuid")
	ErrDifferentVersionCredentials = errors.New("version in token is different of current version")
)
