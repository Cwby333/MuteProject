package tests

import (
	"context"
	"testing"
	"time"

	allerrors "github.com/Cwby333/user-microservice/internal/allErrors"
	"github.com/Cwby333/user-microservice/internal/models"
	userservice "github.com/Cwby333/user-microservice/internal/service/userService"
	mock_userservice "github.com/Cwby333/user-microservice/internal/service/userService/mock_userService"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRegisterNegative(t *testing.T) {
	testTable := []struct {
		name          string
		in            models.User
		expectedError error
	}{
		struct {
			name          string
			in            models.User
			expectedError error
		}{
			name: "to small password",
			in: models.User{
				Password: "12345",
			},
			expectedError: allerrors.ErrPasswordSmall,
		},
		struct {
			name          string
			in            models.User
			expectedError error
		}{
			name: "to big password",
			in: models.User{
				Password: "!11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
			},
			expectedError: allerrors.ErrPasswordBig,
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepoMock := mock_userservice.NewMockUserRepo(ctrl)
	service := userservice.New(userRepoMock, nil, nil, nil, userservice.JWTConfig{})

	for i := range testTable {
		t.Run(testTable[i].name, func(t *testing.T) {
			_, err := service.Register(context.Background(), testTable[i].in)
			require.ErrorIs(t, err, testTable[i].expectedError)
		})
	}
}

func TestRegisterPositive(t *testing.T) {
	testTable := []struct {
		name string
		in   models.User
	}{
		{
			name: "success register1",
			in: models.User{
				Username: "username",
				Password: "12345678",
			},
		},
		{
			name: "success register2",
			in: models.User{
				Username: "username2",
				Password: "somepassworddddd",
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repoMock := mock_userservice.NewMockUserRepo(ctrl)

	service := userservice.New(repoMock, nil, nil, nil, userservice.JWTConfig{})

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			repoMock.EXPECT().CreateUser(
				context.Background(),
				gomock.AssignableToTypeOf(models.User{}),
			).Times(1).DoAndReturn(func(ctx context.Context, user models.User) (models.User, error) {
				err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tt.in.Password))
				require.NoError(t, err)

				require.Equal(t, tt.in.Username, user.Username)
				require.Equal(t, "user", user.Role)

				return user, nil
			})

			_, err := service.Register(context.Background(), tt.in)
			require.NoError(t, err)
		})
	}
}

func TestLogin(t *testing.T) {
	testTable := []struct {
		name             string
		inputUser        models.User
		mockUserFromRepo models.User
		expectedError    error
	}{
		{
			name: "successful login",
			inputUser: models.User{
				Username: "validuser",
				Password: "correctpassword",
			},
			mockUserFromRepo: models.User{
				Username: "validuser",
				Password: func() string {
					hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
					return string(hash)
				}(),
				Role: "user",
			},
			expectedError: nil,
		},
		{
			name: "wrong password",
			inputUser: models.User{
				Username: "validuser",
				Password: "wrongpassword",
			},
			mockUserFromRepo: models.User{
				Username: "validuser",
				Password: func() string {
					hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
					return string(hash)
				}(),
				Role: "user",
			},
			expectedError: allerrors.ErrWrongPass,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mock_userservice.NewMockUserRepo(ctrl)
	service := userservice.New(repoMock, nil, nil, nil, userservice.JWTConfig{})

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			repoMock.EXPECT().
				GetUserByUsername(context.Background(), tt.inputUser.Username).
				Return(tt.mockUserFromRepo, nil)

			_, _, err := service.Login(context.Background(), tt.inputUser)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	testTable := []struct {
		TokenID string
		Expired time.Time
	}{
		struct {
			TokenID string
			Expired time.Time
		}{
			TokenID: "sometokenid",
			Expired: time.Unix(32513207339, 0),
		},
		struct {
			TokenID string
			Expired time.Time
		}{
			TokenID: "1111-1111-1111-111-1-1-1-",
			Expired: time.Unix(32513207339, 0),
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	invalidatorMock := mock_userservice.NewMockRefreshInvalidator(ctrl)

	service := userservice.New(nil, nil, nil, invalidatorMock, userservice.JWTConfig{})

	for i := range testTable {
		invalidatorMock.EXPECT().InvalidRefresh(context.Background(), testTable[i].TokenID, testTable[i].Expired).Return(nil)
		err := service.Logout(context.Background(), testTable[i].TokenID, testTable[i].Expired)
		require.NoError(t, err)
	}
}

func TestFindUserByIDPositive(t *testing.T) {
	testTable := []struct {
		ID string
	}{
		struct{ ID string }{
			ID: "fe1a3ff1-a586-431e-b0a5-835813e6e030",
		},
		struct{ ID string }{
			ID: "9bd7afd6-627c-47d9-9bbe-922e7a3b6929",
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mock_userservice.NewMockUserRepo(ctrl)
	cacheMock := mock_userservice.NewMockUserCache(ctrl)

	service := userservice.New(repoMock, nil, cacheMock, nil, userservice.JWTConfig{})

	for i := range testTable {
		cacheMock.EXPECT().Get(context.Background(), testTable[i].ID).Return(models.User{}, nil)
		_, err := service.FindUserByID(context.Background(), testTable[i].ID)
		require.NoError(t, err)
	}
}

func TestFindUserByIDNegative(t *testing.T) {
	testTable := []struct {
		ID string
	}{
		struct{ ID string }{
			ID: "321342412",
		},
		struct{ ID string }{
			ID: "wrongUUID",
		},
	}

	service := userservice.New(nil, nil, nil, nil, userservice.JWTConfig{})

	for i := range testTable {
		_, err := service.FindUserByID(context.Background(), testTable[i].ID)
		require.Error(t, err)
	}
}

func TestGetAllUsersService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mock_userservice.NewMockUserRepo(ctrl)

	service := userservice.New(repoMock, nil, nil, nil, userservice.JWTConfig{})
	repoMock.EXPECT().GetAllUsers(context.Background()).Return([]models.User{}, nil)
	_, err := service.GetAllUsers(context.Background())
	require.NoError(t, err)
}

func TestDeleteUserPositive(t *testing.T) {
	testTable := []struct {
		ID string
	}{
		struct{ ID string }{
			ID: "fe1a3ff1-a586-431e-b0a5-835813e6e030",
		},
		struct{ ID string }{
			ID: "9bd7afd6-627c-47d9-9bbe-922e7a3b6929",
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mock_userservice.NewMockUserRepo(ctrl)
	cacheMock := mock_userservice.NewMockUserCache(ctrl)

	service := userservice.New(repoMock, nil, cacheMock, nil, userservice.JWTConfig{})

	for i := range testTable {
		repoMock.EXPECT().DeleteUserByID(context.Background(), testTable[i].ID).Return(nil)
		cacheMock.EXPECT().Delete(context.Background(), testTable[i].ID).Return(nil)

		err := service.DeleteUser(context.Background(), testTable[i].ID)
		require.NoError(t, err)
	}
}

func TestDeleteUserNegative(t *testing.T) {
	testTable := []struct {
		ID string
	}{
		struct{ ID string }{
			ID: "invlidauuid",
		},
		struct{ ID string }{
			ID: "111233421",
		},
	}

	service := userservice.New(nil, nil, nil, nil, userservice.JWTConfig{})

	for i := range testTable {
		err := service.DeleteUser(context.Background(), testTable[i].ID)
		require.Error(t, err)
	}
}

func TestUpdateUser(t *testing.T) {
	testTable := []struct {
		name         string
		userID       string
		oldUser      models.User
		newUserInfo  models.User
		expectedUser models.User
	}{
		{
			name:   "update all fields",
			userID: "id1",
			oldUser: models.User{
				Username: "oldusername",
				Password: "oldpassword",
				Email:    "oldemail@example.com",
				Role:     "user",
			},
			newUserInfo: models.User{
				Username: "newusername",
				Password: "newPassword",
				Email:    "newemail@example.com",
			},
			expectedUser: models.User{
				Username: "newusername",
				Email:    "newemail@example.com",
				Role:     "user",
			},
		},
		{
			name:   "partial update",
			userID: "id2",
			oldUser: models.User{
				Username: "oldusername",
				Password: "oldpassword",
				Email:    "oldemail@example.com",
				Role:     "user",
			},
			newUserInfo: models.User{
				Email: "newemail@example.com",
			},
			expectedUser: models.User{
				Username: "oldusername",
				Password: "oldpassword",
				Email:    "newemail@example.com",
				Role:     "user",
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := mock_userservice.NewMockUserRepo(ctrl)
	cacheMock := mock_userservice.NewMockUserCache(ctrl)
	service := userservice.New(repoMock, nil, cacheMock, nil, userservice.JWTConfig{})

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			repoMock.EXPECT().
				GetUserByID(context.Background(), tt.userID).
				Return(tt.oldUser, nil)

			repoMock.EXPECT().
				UpdateUserByID(context.Background(), tt.userID, gomock.Any()).
				DoAndReturn(func(ctx context.Context, id string, user models.User) (models.User, error) {
					if tt.newUserInfo.Username != "" {
						require.Equal(t, tt.newUserInfo.Username, user.Username)
					} else {
						require.Equal(t, tt.oldUser.Username, user.Username)
					}

					if tt.newUserInfo.Email != "" {
						require.Equal(t, tt.newUserInfo.Email, user.Email)
					} else {
						require.Equal(t, tt.oldUser.Email, user.Email)
					}

					require.Equal(t, tt.oldUser.Role, user.Role)

					if tt.newUserInfo.Password != "" {
						err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tt.newUserInfo.Password))
						require.NoError(t, err)
						tt.expectedUser.Password = user.Password
					} else {
						require.Equal(t, tt.oldUser.Password, user.Password)
					}

					return user, nil
				})

			cacheMock.EXPECT().
				Set(context.Background(), tt.userID, gomock.Any()).
				Do(func(ctx context.Context, id string, user models.User) {
					require.Equal(t, tt.expectedUser.Username, user.Username)
					require.Equal(t, tt.expectedUser.Email, user.Email)
					require.Equal(t, tt.expectedUser.Role, user.Role)

					if tt.newUserInfo.Password != "" {
						err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tt.newUserInfo.Password))
						require.NoError(t, err)
					} else {
						require.Equal(t, tt.expectedUser.Password, user.Password)
					}
				}).
				Return(nil)

			_, err := service.UpdateUser(context.Background(), tt.userID, tt.newUserInfo)
			require.NoError(t, err)
		})
	}
}

func TestActionWithSong(t *testing.T) {
	testTable := []struct {
		Task models.DefferedTask
	}{
		struct{ Task models.DefferedTask }{
			Task: models.DefferedTask{
				ID:        uuid.NewString(),
				Topic:     "sometopic",
				Data:      []byte("data for sender"),
				CreatedAt: time.Now(),
			},
		},
	}

	ctrl := gomock.NewController(t)
	repoMock := mock_userservice.NewMockDefferedTaksRepo(ctrl)

	service := userservice.New(nil, repoMock, nil, nil, userservice.JWTConfig{})

	for i := range testTable {
		repoMock.EXPECT().Create(context.Background(), testTable[i].Task).Return(nil)

		err := service.ActionWithSong(context.Background(), testTable[i].Task)
		require.NoError(t, err)
	}
}
