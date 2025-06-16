package userservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"

	allerrors "github.com/Cwby333/user-microservice/internal/allErrors"
	"github.com/Cwby333/user-microservice/internal/models"
	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user models.User) (models.User, error)

	GetUserByID(ctx context.Context, ID string) (models.User, error)
	GetUserByUsername(ctx context.Context, username string) (models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)

	DeleteUserByID(ctx context.Context, ID string) error

	UpdateUserByID(ctx context.Context, ID string, newUserInfo models.User) (models.User, error)
}

type DefferedTaksRepo interface {
	Create(ctx context.Context, task models.DefferedTask) error
}

type RefreshInvalidator interface {
	InvalidRefresh(ctx context.Context, tokenID string, expired time.Time) error
	CheckTokenInBlackList(ctx context.Context, tokenID string) error
}

type UserCache interface {
	Set(ctx context.Context, usersID string, user models.User) error
	Get(ctx context.Context, userID string) (models.User, error)
	Delete(ctx context.Context, userID string) error
}

type JWTConfig struct {
	SecretKey      string
	Issuer         string
	AccessExpired  time.Duration
	RefreshExpired time.Duration
}

type Service struct {
	userRepo         UserRepo
	defferedTaskRepo DefferedTaksRepo
	invalidator      RefreshInvalidator
	userCache        UserCache
	config           JWTConfig
}

func New(userRepo UserRepo, taskRepo DefferedTaksRepo, userCache UserCache, invalidator RefreshInvalidator, cfg JWTConfig) Service {
	return Service{
		userRepo:         userRepo,
		defferedTaskRepo: taskRepo,
		invalidator:      invalidator,
		userCache:        userCache,
		config:           cfg,
	}
}

func (s Service) Register(ctx context.Context, user models.User) (models.User, error) {
	const op = "./internal/service/userService/service.go.Register.go"

	if utf8.RuneCountInString(user.Password) < 8 {
		return models.User{}, fmt.Errorf("%s: %w", op, allerrors.ErrPasswordSmall)
	}
	if utf8.RuneCountInString(user.Password) > 72 {
		return models.User{}, fmt.Errorf("%s: %w", op, allerrors.ErrPasswordBig)
	}

	psw, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user.Password = string(psw)
	if user.Role == "" {
		user.Role = "user"
	}

	user, err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s Service) Login(ctx context.Context, user models.User) (access models.JWTAccess, refresh models.JWTRefresh, err error) {
	const op = "./internal/service/userService/service.go.Login"

	userFromRepo, err := s.userRepo.GetUserByUsername(ctx, user.Username)
	if err != nil {
		return models.JWTAccess{}, models.JWTRefresh{}, fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(userFromRepo.Password), []byte(user.Password))
	if err != nil {
		return models.JWTAccess{}, models.JWTRefresh{}, fmt.Errorf("%s: %w", op, allerrors.ErrWrongPass)
	}

	user = userFromRepo
	access, refresh, err = s.createTokens(ctx, user)
	if err != nil {
		return models.JWTAccess{}, models.JWTRefresh{}, fmt.Errorf("%s: %w", op, err)
	}

	return access, refresh, nil
}

func (s Service) Logout(ctx context.Context, tokenID string, tokenExpiredUnix time.Time) error {
	const op = "./internal/service/userService/service.go.Logout"

	err := s.invalidator.InvalidRefresh(ctx, tokenID, tokenExpiredUnix)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Service) FindUserByID(ctx context.Context, ID string) (models.User, error) {
	const op = "./internal/service/userService/service.go.FindUserByID"

	if err := uuid.Validate(ID); err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, allerrors.ErrWrongUUID)
	}

	user, err := s.userCache.Get(ctx, ID)
	switch err {
	case nil:
		slog.Info("user cache hit", slog.String("userID", ID))

		return user, nil
	default:
		if errors.Is(err, allerrors.ErrNotFoundInCache) {
			slog.Info("user cache miss", slog.String("userID", ID))
			break
		}

		slog.Info("user cache", slog.String("error", err.Error()))
	}

	user, err = s.userRepo.GetUserByID(ctx, ID)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	err = s.userCache.Set(ctx, ID, user)
	if err != nil {
		slog.Info("user cache", slog.String("error", err.Error()))
	}

	return user, err
}

func (s Service) GetAllUsers(ctx context.Context) ([]models.User, error) {
	const op = "./internal/service/userService/service.go.GetAllUsers"

	slice, err := s.userRepo.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return slice, nil
}

func (s Service) DeleteUser(ctx context.Context, ID string) error {
	const op = "./internal/service/userService/service.go.GetAllUsers"

	if err := uuid.Validate(ID); err != nil {
		return fmt.Errorf("%s: %w", op, allerrors.ErrWrongUUID)
	}

	err := s.userRepo.DeleteUserByID(ctx, ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.userCache.Delete(ctx, ID)
	if err != nil {
		slog.Info("cache", slog.String("error", err.Error()))
	}

	return nil
}

func (s Service) UpdateUser(ctx context.Context, ID string, newUserInfo models.User) (models.User, error) {
	const op = "./internal/service/userService/service.go.UpdateUser"

	user, err := s.userRepo.GetUserByID(ctx, ID)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	if newUserInfo.Email == "" {
		newUserInfo.Email = user.Email
	}
	if newUserInfo.Username == "" {
		newUserInfo.Username = user.Username
	}
	newUserInfo.Role = user.Role
	newUserInfo.VersionCredentials = user.VersionCredentials + 1

	switch newUserInfo.Password {
	case "":
		newUserInfo.Password = user.Password
	default:
		psw, err := bcrypt.GenerateFromPassword([]byte(newUserInfo.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.User{}, fmt.Errorf("%s: %w", op, err)
		}

		newUserInfo.Password = string(psw)
	}

	user, err = s.userRepo.UpdateUserByID(ctx, ID, newUserInfo)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	err = s.userCache.Set(ctx, ID, newUserInfo)
	if err != nil {
		slog.Info("cache", slog.String("error", err.Error()))
	}

	return user, err
}

func (s Service) ActionWithSong(ctx context.Context, task models.DefferedTask) error {
	const op = "./internal/service/userService/service.go.LikeSong"

	err := s.defferedTaskRepo.Create(ctx, task)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}