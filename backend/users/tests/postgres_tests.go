package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Cwby333/user-microservice/internal/adapters/repository/postgres"
	allerrors "github.com/Cwby333/user-microservice/internal/allErrors"
	"github.com/Cwby333/user-microservice/internal/models"
	"github.com/stretchr/testify/require"
)

type Config struct {
	Host     string
	Port     uint16
	User     string
	Password string
	DB       string
	MaxConns int
	MinConns int
}

type PostgresTest struct {
	pg   postgres.Postgres
	isUp bool
	mustSkipTest bool
}

//Здесь для тестов нужно менять под свою локальную базу, чтобы тесты не падали из-за неправильного конфига добавил mustSkipTest В PostgresTest, который проскипает тесты, связанные с Postgres, если соединение не было установлено
var cfgForTestDBLocal Config = Config{
	Host:     "localhost",
	Port:     5432,
	Password: "qwerty1234",
	User:     "postgres",
	DB:       "users",
	MaxConns: 20,
	MinConns: 5,
}

var pgTest PostgresTest

func InitPostgresForTest(t *testing.T) {
	cfg := postgres.Config{
		Host:     cfgForTestDBLocal.Host,
		Port:     cfgForTestDBLocal.Port,
		User:     cfgForTestDBLocal.User,
		Password: cfgForTestDBLocal.Password,
		DB:       cfgForTestDBLocal.DB,
		MaxConns: cfgForTestDBLocal.MaxConns,
		MinConns: cfgForTestDBLocal.MinConns,
	}
	if !pgTest.isUp {
		pg, err := postgres.New(context.Background(), cfg)
		if err != nil {
			pgTest.mustSkipTest = true
		}

		pgTest.pg = pg
		pgTest.isUp = true
	}

	if pgTest.mustSkipTest == true {
		t.Skip()
	}
}

func (pg PostgresTest) Clean() {
	pg.pg.Pool.Query(context.Background(), "TRUNCATE users")
	pg.pg.Pool.Query(context.Background(), "TRUNCATE deffered_tasks")
}

func TestCreateUser(t *testing.T) {
	InitPostgresForTest(t)

	testCases := []struct {
		name          string
		user          models.User
		expectedError error
	}{
		{
			name: "success create",
			user: models.User{
				Username: "someusername",
				Email:    "someemail",
				Password: "somepassword",
			},
			expectedError: nil,
		},
		{
			name: "username duplicate",
			user: models.User{
				Username: "someusername",
				Email:    "someemail2",
				Password: "somepassword",
			},
			expectedError: allerrors.ErrUsernameExists,
		},
		{
			name: "email duplicate",
			user: models.User{
				Username: "someusername3",
				Email:    "someemail",
				Password: "somepassword",
			},
			expectedError: allerrors.ErrUsernameExists,
		},
	}

	for _, tc := range testCases {
		userFromPg, err := pgTest.pg.CreateUser(context.Background(), tc.user)

		if tc.expectedError != nil {
			require.Error(t, err, tc.expectedError)
		} else {
			require.Equal(t, tc.user.Username, userFromPg.Username)
			require.Equal(t, tc.user.Password, userFromPg.Password)
			require.Equal(t, tc.user.Email, userFromPg.Email)
		}
	}

	pgTest.Clean()
}

func TestGetUserByID(t *testing.T) {
	InitPostgresForTest(t)

	userForTest, _ := pgTest.pg.CreateUser(context.Background(), models.User{
		Username: "fumr90w8vwtuwe8m90tguew8v90umrw90c",
		Password: "r02vmjr2cjr02gdfgd",
		Email:    "481r8ei0fjsdm0vpcfji9s",
	})

	testCases := []struct {
		name          string
		user          models.User
		expectedError error
	}{
		{
			name: "success",
			user: models.User{
				ID: userForTest.ID,
			},
			expectedError: nil,
		},
		{
			name: "user not found",
			user: models.User{
				ID: "not found id",
			},
			expectedError: allerrors.ErrUserNotExists,
		},
	}

	for _, tc := range testCases {
		userFromPg, err := pgTest.pg.GetUserByID(context.Background(), tc.user.ID)

		if tc.expectedError != nil {
			require.Error(t, err, tc.expectedError)
		} else {
			require.Equal(t, userForTest.Username, userFromPg.Username)
			require.Equal(t, userForTest.Email, userFromPg.Email)
			require.Equal(t, userForTest.Password, userFromPg.Password)
		}
	}

	pgTest.Clean()
}

func TestGetUserByUsername(t *testing.T) {
	InitPostgresForTest(t)

	userForTest, _ := pgTest.pg.CreateUser(context.Background(), models.User{
		Username: "fumr90w8vwtuwe8m90tguew8v90umrw90c333",
		Password: "r02vmjr2cjr02gdfgd",
		Email:    "481r8ei0fjsdm0vpcfji9s",
	})
	userForTest, _ = pgTest.pg.CreateUser(context.Background(), models.User{
		Username: "fumr90w8vwtuwe8m90tguew8v90umrw90c333",
		Password: "r02vmjr2cjr02gdfgd",
		Email:    "481r8ei0fjsdm0vpcfji9s",
	})

	testCases := []struct {
		name          string
		user          models.User
		expectedError error
	}{
		{
			name: "success",
			user: models.User{
				Username: userForTest.Username,
			},
			expectedError: nil,
		},
		{
			name: "user not found",
			user: models.User{
				Username: "not exist username",
			},
			expectedError: allerrors.ErrUserNotExists,
		},
	}

	for _, tc := range testCases {
		userFromPg, err := pgTest.pg.GetUserByUsername(context.Background(), tc.user.ID)
		fmt.Println(err)
		if tc.expectedError != nil {
			require.Error(t, err, tc.expectedError)
		} else {
			require.Equal(t, userForTest.Username, userFromPg.Username)
			require.Equal(t, userForTest.Email, userFromPg.Email)
			require.Equal(t, userForTest.Password, userFromPg.Password)
		}
	}

	pgTest.Clean()
}

func TestGetAllUsers(t *testing.T) {
	InitPostgresForTest(t)

	sliceUserForTest := []models.User{
		models.User{
			Username: "username1",
			Password: "password1",
			Email:    "email1",
		},
		models.User{
			Username: "username2",
			Password: "password2",
			Email:    "email2",
		},
		models.User{
			Username: "username3",
			Password: "password3",
			Email:    "email3",
		},
	}

	for i := range sliceUserForTest {
		_, err := pgTest.pg.CreateUser(context.Background(), sliceUserForTest[i])
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	_, err := pgTest.pg.GetAllUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	pgTest.Clean()
}

func TestDeleteUser(t *testing.T) {
	InitPostgresForTest(t)

	err := pgTest.pg.DeleteUserByID(context.Background(), "some id")
	if err != nil {
		t.Fatal(err.Error())
	}

	pgTest.Clean()
}

func TestUpdateUserByID(t *testing.T) {
	InitPostgresForTest(t)

	sliceUserForTest := []models.User{
		models.User{
			Username: "username1",
			Password: "password1",
			Email:    "email1",
		},
		models.User{
			Username: "username2",
			Password: "password2",
			Email:    "email2",
		},
		models.User{
			Username: "username3",
			Password: "password3",
			Email:    "email3",
		},
	}

	for i := range sliceUserForTest {
		_, err := pgTest.pg.CreateUser(context.Background(), sliceUserForTest[i])
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	testCases := []struct {
		name        string
		user        models.User
		updatedUser models.User
	}{
		{
			name: "update all fields",
			user: models.User{
				Username: "username1",
				Password: "password1",
				Email:    "email1",
			},
			updatedUser: models.User{
				Username: "updated_username1",
				Password: "updated_password1",
				Email:    "updated_password1",
			},
		},
		{
			name: "update username",
			user: models.User{
				Username: "username2",
				Password: "password1",
				Email:    "email1",
			},
			updatedUser: models.User{
				Username: "updated_username2",
				Password: "password2",
				Email:    "email2",
			},
		},
		{
			name: "update email",
			user: models.User{
				Username: "username3",
				Password: "password3",
				Email:    "email3",
			},
			updatedUser: models.User{
				Username: "username3",
				Password: "password3",
				Email:    "updated_email3",
			},
		},
	}

	for _, tc := range testCases {
		userFromPg, err := pgTest.pg.UpdateUserByID(context.Background(), tc.user.ID, tc.updatedUser)
		if err != nil {
			t.Fatal(err.Error())
		}

		require.Equal(t, tc.updatedUser.Username, userFromPg.Username)
		require.Equal(t, tc.updatedUser.Password, userFromPg.Password)
		require.Equal(t, tc.updatedUser.Email, userFromPg.Email)
	}

	pgTest.Clean()
}

func createTask(action, userID, trackID string) models.DefferedTask {
	m := map[string]any{
		"action":   action,
		"user_id":  userID,
		"track_id": trackID,
	}

	data, _ := json.Marshal(m)

	return models.DefferedTask{
		Topic:     "some-topic",
		Data:      data,
		CreatedAt: time.Now(),
	}
}

func TestCreateDefferedTask(t *testing.T) {
	InitPostgresForTest(t)

	testCases := []struct {
		name string
		task models.DefferedTask
	}{
		{
			name: "task 1",
			task: createTask("like", "123", "333"),
		},
		{
			name: "task 2",
			task: createTask("loke", "111", "777"),
		},
		{
			name: "task 3",
			task: createTask("unlike", "123", "321"),
		},
	}

	for _, tc := range testCases {
		err := pgTest.pg.Create(context.Background(), tc.task)
		if err != nil {
			t.Fatal(err)
		}
	}

	pgTest.Clean()
}

func TestToUserDTO(t *testing.T) {
	testCases := []struct {
		name string
		user models.User
	}{
		{
			name: "case 1",
			user: models.User{
				ID:       "123",
				Username: "username1",
				Email:    "email1",
				Password: "password1",
				Role:     "user",
			},
		}, {
			name: "case 2",
			user: models.User{
				ID:       "124",
				Username: "username2",
				Email:    "email2",
				Password: "password2",
				Role:     "user",
			},
		}, {
			name: "case 3",
			user: models.User{
				ID:       "125",
				Username: "username3",
				Email:    "email3",
				Password: "password3",
				Role:     "user",
			},
		},
	}

	for _, tc := range testCases {
		dto := postgres.ToUserDTO(tc.user)

		require.Equal(t, tc.user.ID, dto.ID)
		require.Equal(t, tc.user.Username, dto.Username)
		require.Equal(t, tc.user.Email, dto.Email)
		require.Equal(t, tc.user.Password, dto.Password)
		require.Equal(t, tc.user.Role, dto.Role)
	}
}
