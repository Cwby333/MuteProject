package userrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Cwby333/user-microservice/internal/adapters/transport/http/lib"
	"github.com/Cwby333/user-microservice/internal/adapters/transport/http/userRouter/userRouterMocks"
	allerrors "github.com/Cwby333/user-microservice/internal/allErrors"
	"github.com/Cwby333/user-microservice/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := userRouterMocks.NewMockUserService(ctrl)
	mockTaskService := userRouterMocks.NewMockDefferedTaskService(ctrl)

	router := New(mockUserService, mockTaskService, nil)

	testCases := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "successful registration",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "testpassword",
			},
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user models.User) (models.User, error) {
						require.Equal(t, "testuser", user.Username)
						require.Equal(t, "test@example.com", user.Email)
						require.Equal(t, "testpassword", user.Password)
						require.Equal(t, "user", "user")

						return models.User{
							ID:       "123",
							Username: "testuser",
							Email:    "test@example.com",
							Role:     "user",
						}, nil
					})
			},
			expectedStatus: http.StatusOK,
			expectedBody: RegisterResponse{
				Response: lib.Response{
					StatusCode: http.StatusOK,
					Message:    "success register",
				},
				Username: "testuser",
				Email:    "test@example.com",
				ID:       "123",
			},
		},
		{
			name: "validation error - missing username",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "testpassword",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: RegisterResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "field Username is required",
				},
			},
		},
		{
			name: "username already exists",
			requestBody: map[string]string{
				"username": "existinguser",
				"email":    "test@example.com",
				"password": "testpassword",
			},
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), gomock.Any()).
					Return(models.User{}, allerrors.ErrUsernameExists)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: RegisterResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "username already exists",
				},
			},
		},
		{
			name: "email already exists",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "existing@example.com",
				"password": "testpassword",
			},
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), gomock.Any()).
					Return(models.User{}, allerrors.ErrEmailExists)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: RegisterResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "email already exists",
				},
			},
		},
		{
			name: "password too small",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "short",
			},
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), gomock.Any()).
					Return(models.User{}, allerrors.ErrPasswordSmall)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: RegisterResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "password to small",
				},
			},
		},
		{
			name: "password big",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
			},
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(models.User{}, allerrors.ErrPasswordBig)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: RegisterResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "password to big",
				},
			},
		},
		{
			name: "internal server error",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "testpassword",
			},
			mockSetup: func() {
				mockUserService.EXPECT().Register(gomock.Any(), gomock.Any()).
					Return(models.User{}, errors.New("server error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: RegisterResponse{
				Response: lib.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    "server error",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockSetup != nil {
				tc.mockSetup()
			}

			body, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/user/register", bytes.NewBuffer(body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(router.Register)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedBody != nil {
				var response RegisterResponse
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				if tc.expectedStatus == http.StatusOK {
					expected := tc.expectedBody.(RegisterResponse)
					require.Equal(t, expected.Response, response.Response)
					require.Equal(t, expected.Username, response.Username)
					require.Equal(t, expected.Email, response.Email)
					require.Equal(t, expected.ID, response.ID)
				} else {
					expected := tc.expectedBody.(RegisterResponse)
					msg := response.Response.Message
					require.Equal(t, expected.Response.Message, msg)
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := userRouterMocks.NewMockUserService(ctrl)
	mockTaskService := userRouterMocks.NewMockDefferedTaskService(ctrl)

	router := New(mockUserService, mockTaskService, nil)

	testCases := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		// {
		// 	name: "successful login",
		// 	requestBody: map[string]string{
		// 		"username": "testuser",
		// 		"password": "testpassword",
		// 	},
		// 	mockSetup: func() {
		// 		mockUserService.EXPECT().Login(gomock.Any(), gomock.Any()).
		// 			Return(models.JWTAccess{
		// 				RegisteredClaims: jwt.RegisteredClaims{
		// 					ExpiresAt: jwt.NewNumericDate(time.Now()),
		// 					NotBefore: jwt.NewNumericDate(time.Now()),
		// 					IssuedAt:  jwt.NewNumericDate(time.Now()),
		// 					Subject: JWTSubject,
		// 				},
		// 			}, models.JWTRefresh{
		// 				RegisteredClaims: jwt.RegisteredClaims{
		// 					ExpiresAt: jwt.NewNumericDate(time.Now()),
		// 					NotBefore: jwt.NewNumericDate(time.Now()),
		// 					IssuedAt:  jwt.NewNumericDate(time.Now()),
		// 					Subject: JWTSubject,
		// 				},
		// 			}, nil)
		// 	},
		// 	expectedStatus: http.StatusOK,
		// 	expectedBody: LoginResponse{
		// 		Response: lib.Response{
		// 			StatusCode: http.StatusOK,
		// 			Message:    fmt.Sprintf("success login, ID: %s", JWTSubject),
		// 		},
		// 	},
		// },
		{
			name: "wrong password",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "wrongpass",
			},
			mockSetup: func() {
				mockUserService.EXPECT().Login(gomock.Any(), gomock.Any()).
					Return(models.JWTAccess{
						RegisteredClaims: jwt.RegisteredClaims{
							ExpiresAt: jwt.NewNumericDate(time.Now()),
							NotBefore: jwt.NewNumericDate(time.Now()),
							IssuedAt:  jwt.NewNumericDate(time.Now()),
						},
					}, models.JWTRefresh{
						RegisteredClaims: jwt.RegisteredClaims{
							ExpiresAt: jwt.NewNumericDate(time.Now()),
							NotBefore: jwt.NewNumericDate(time.Now()),
							IssuedAt:  jwt.NewNumericDate(time.Now()),
						},
					}, allerrors.ErrWrongPass)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: LoginResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "wrong password",
				},
			},
		},
		{
			name: "user not found",
			requestBody: map[string]string{
				"username": "nonexistentuser",
				"password": "somepassword",
			},
			mockSetup: func() {
				mockUserService.EXPECT().Login(gomock.Any(), gomock.Any()).
					Return(models.JWTAccess{}, models.JWTRefresh{}, allerrors.ErrUserNotExists)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: LoginResponse{
				Response: lib.Response{
					StatusCode: http.StatusNotFound,
					Message:    "username not found",
				},
			},
		},
		{
			name:           "missing fields",
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedBody: LoginResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "field Username is required field Password is required ",
				},
			},
		},
		{
			name: "server internal error",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "testpassword",
			},
			mockSetup: func() {
				mockUserService.EXPECT().Login(gomock.Any(), gomock.Any()).
					Return(models.JWTAccess{}, models.JWTRefresh{}, errors.New("server error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: LoginResponse{
				Response: lib.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    "server error",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockSetup != nil {
				tc.mockSetup()
			}

			body, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/user/login", bytes.NewBuffer(body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(router.Login)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedBody != nil {
				var response LoginResponse
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				ex := tc.expectedBody.(LoginResponse)
				require.Equal(t, ex.Response.StatusCode, response.Response.StatusCode)
				require.Equal(t, ex.Response.Message, response.Response.Message)
			}
		})
	}
}

var testUser = models.User{
	ID:       "b7c4adbd-06e2-43dc-b7ef-59d05a9f0593",
	Username: "testuser",
	Email:    "test@test.com",
	Role:     "user",
}

func TestGetUserByIDHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := userRouterMocks.NewMockUserService(ctrl)
	mockTaskService := userRouterMocks.NewMockDefferedTaskService(ctrl)

	router := New(mockUserService, mockTaskService, nil)

	testCases := []struct {
		name                 string
		queryParams          string
		mockSetup            func()
		claimsInContextSetup func(r *http.Request) *http.Request
		expectedStatus       int
		expectedBody         interface{}
	}{
		{
			name:        "successful getting user",
			queryParams: "?user_id=" + testUser.ID,
			mockSetup: func() {
				mockUserService.EXPECT().FindUserByID(gomock.Any(), testUser.ID).
					Return(testUser, nil)
			},
			claimsInContextSetup: func(req *http.Request) *http.Request {
				ctx := req.Context()
				ctx = context.WithValue(ctx, "claims", jwt.MapClaims{
					"sub": testUser.ID,
				})
				req = req.WithContext(ctx)
				return req
			},
			expectedStatus: http.StatusOK,
			expectedBody: GetUserByIDResponse{
				Response: lib.Response{
					StatusCode: http.StatusOK,
					Message:    "success",
				},
				Username: testUser.Username,
			},
		},
		{
			name:           "missing user id",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			claimsInContextSetup: func(req *http.Request) *http.Request {
				ctx := req.Context()
				ctx = context.WithValue(ctx, "claims", jwt.MapClaims{
					"sub": "",
				})
				req = req.WithContext(ctx)
				return req
			},
			expectedBody: GetUserByIDResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "missing user ID in token",
				},
			},
		},
		{
			name:        "user not found",
			queryParams: "?user_id=" + testUser.ID,
			mockSetup: func() {
				mockUserService.EXPECT().FindUserByID(gomock.Any(), testUser.ID).
					Return(models.User{}, allerrors.ErrUserNotExists)
			},
			claimsInContextSetup: func(req *http.Request) *http.Request {
				ctx := req.Context()
				ctx = context.WithValue(ctx, "claims", jwt.MapClaims{
					"sub": testUser.ID,
				})
				req = req.WithContext(ctx)
				return req
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: GetUserByIDResponse{
				Response: lib.Response{
					StatusCode: http.StatusNotFound,
					Message:    "user not found",
				},
			},
		},
		{
			name:           "wrong uuid",
			queryParams:    "?user_id=321",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func() {
				mockUserService.EXPECT().FindUserByID(gomock.Any(), "321").Return(models.User{}, allerrors.ErrWrongUUID)
			},
			claimsInContextSetup: func(req *http.Request) *http.Request {
				ctx := req.Context()
				ctx = context.WithValue(ctx, "claims", jwt.MapClaims{
					"sub": "321",
				})
				req = req.WithContext(ctx)
				return req
			},
			expectedBody: GetUserByIDResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "wrong user ID",
				},
			},
		},
		{
			name:        "server internal error",
			queryParams: "?user_id=" + testUser.ID,
			mockSetup: func() {
				mockUserService.EXPECT().FindUserByID(gomock.Any(), testUser.ID).
					Return(models.User{}, errors.New("server error"))
			},
			claimsInContextSetup: func(req *http.Request) *http.Request {
				ctx := req.Context()
				ctx = context.WithValue(ctx, "claims", jwt.MapClaims{
					"sub": testUser.ID,
				})
				req = req.WithContext(ctx)
				return req
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: GetUserByIDResponse{
				Response: lib.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    "server error",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockSetup != nil {
				tc.mockSetup()
			}

			req := httptest.NewRequest("GET", "/user/get"+tc.queryParams, nil)

			if tc.claimsInContextSetup != nil {
				req = tc.claimsInContextSetup(req)
			}

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(router.GetUserByID)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedBody != nil {
				responseData, err := io.ReadAll(rr.Body)
				require.NoError(t, err)

				var response GetUserByIDResponse
				err = json.Unmarshal(responseData, &response)
				require.NoError(t, err)

				ex := tc.expectedBody.(GetUserByIDResponse)
				require.Equal(t, ex.Response.StatusCode, response.Response.StatusCode)
				require.Equal(t, ex.Response.Message, response.Response.Message)
				require.Equal(t, ex.Username, response.Username)
			}
		})
	}
}

func TestGetAllUsersHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки сервисов
	mockUserService := userRouterMocks.NewMockUserService(ctrl)
	mockTaskService := userRouterMocks.NewMockDefferedTaskService(ctrl)

	// Инициализируем роутер с использованием моков
	router := New(mockUserService, mockTaskService, nil)

	// Несколько фиктивных пользователей для тестов
	fakeUsers := []models.User{
		{ID: "1", Username: "John Doe", Email: "john.doe@example.com", Role: "user"},
		{ID: "2", Username: "Jane Smith", Email: "jane.smith@example.com", Role: "admin"},
	}

	testCases := []struct {
		name           string
		ctx            context.Context
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Authorized Admin Success",
			ctx:  context.WithValue(context.Background(), "claims", jwt.MapClaims{"role": "admin", "sub": "admin-user-id"}),
			mockSetup: func() {
				mockUserService.EXPECT().GetAllUsers(gomock.Any()).
					Return(fakeUsers, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: GetAllUsersResponse{
				Response: lib.Response{
					StatusCode: http.StatusOK,
					Message:    "success",
				},
				Users: []UserDTO{
					UserToDTO(fakeUsers[0]),
					UserToDTO(fakeUsers[1]),
				},
			},
		},
		{
			name:           "Unauthorized Access",
			ctx:            context.WithValue(context.Background(), "claims", jwt.MapClaims{"role": "user", "sub": "regular-user-id"}),
			expectedStatus: http.StatusUnauthorized,
			expectedBody: GetAllUsersResponse{
				Response: lib.Response{
					StatusCode: http.StatusUnauthorized,
					Message:    "unauthorized",
				},
			},
		},
		{
			name: "Internal Server Error",
			ctx:  context.WithValue(context.Background(), "claims", jwt.MapClaims{"role": "admin", "sub": "admin-user-id"}),
			mockSetup: func() {
				mockUserService.EXPECT().GetAllUsers(gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: GetAllUsersResponse{
				Response: lib.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    "server error",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Настроим поведение моков
			if tc.mockSetup != nil {
				tc.mockSetup()
			}

			// Имитация запроса
			req := httptest.NewRequest("GET", "/user/list", nil)
			req = req.WithContext(tc.ctx)

			// Ответ-записывающий механизм
			rr := httptest.NewRecorder()

			// Передача обработчику
			handler := http.HandlerFunc(router.GetAllUsers)
			handler.ServeHTTP(rr, req)

			// Проверка полученного статуса
			require.Equal(t, tc.expectedStatus, rr.Code)

			// Анализируем ответ
			if tc.expectedBody != nil {
				responseData, err := io.ReadAll(rr.Body)
				require.NoError(t, err)

				var response GetAllUsersResponse
				err = json.Unmarshal(responseData, &response)
				require.NoError(t, err)

				ex := tc.expectedBody.(GetAllUsersResponse)
				require.Equal(t, ex.Response.StatusCode, response.Response.StatusCode)
				require.Equal(t, ex.Response.Message, response.Response.Message)

				if len(ex.Users) > 0 {
					require.Len(t, response.Users, len(ex.Users))
					for i := range response.Users {
						require.Equal(t, ex.Users[i], response.Users[i])
					}
				}
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки сервисов
	mockUserService := userRouterMocks.NewMockUserService(ctrl)
	mockTaskService := userRouterMocks.NewMockDefferedTaskService(ctrl)

	// Инициализируем роутер с использованием моков
	router := New(mockUserService, mockTaskService, nil)

	// Тестовые пользователи
	testUser := models.User{
		ID:       "123",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	updatedUser := models.User{
		ID:       "123",
		Username: "updated_username",
		Email:    "updated_email@example.com",
	}

	testCases := []struct {
		name           string
		ctx            context.Context
		requestBody    string
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "Successful update",
			ctx:         context.WithValue(context.Background(), "claims", jwt.MapClaims{"sub": testUser.ID}),
			requestBody: `{"username":"updated_username","email":"updated_email@example.com","password":"new_password"}`,
			mockSetup: func() {
				updatedUser := models.User{
					ID:       "123",
					Username: "updated_username",
					Email:    "updated_email@example.com",
					Password: "new_password", // Добавлено поле Password
				}
				mockUserService.EXPECT().UpdateUser(gomock.Any(), testUser.ID, updatedUser).
					Return(updatedUser, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: UpdateUserResponse{
				Response: lib.Response{
					StatusCode: http.StatusOK,
					Message:    "success",
				},
				Username: updatedUser.Username,
				Email:    updatedUser.Email,
				ID:       updatedUser.ID,
			},
		},
		{
			name: "Missing parameters",
			ctx:  context.WithValue(context.Background(), "claims", jwt.MapClaims{"sub": testUser.ID}),
			mockSetup: func() {
				// Передаем пустую структуру с минимальным набором полей
				emptyUser := models.User{
					ID:       testUser.ID, // Так как это поле берется из JWT
					Username: "",          // Пустое значение
					Email:    "",          // Пустое значение
					Password: "",          // Пустое значение
				}
				mockUserService.EXPECT().UpdateUser(gomock.Any(), testUser.ID, emptyUser).
					Return(emptyUser, errors.New("fields are empty"))
			},
			requestBody:    "{}",
			expectedStatus: http.StatusNotFound,
			expectedBody: UpdateUserResponse{
				Response: lib.Response{
					StatusCode: http.StatusNotFound,
					Message:    "server error",
				},
			},
		},
		{
			name:        "User not found",
			ctx:         context.WithValue(context.Background(), "claims", jwt.MapClaims{"sub": testUser.ID}),
			requestBody: `{"username":"updated_username","email":"updated_email@example.com","password":"new_password"}`,
			mockSetup: func() {
				updatedUser := models.User{
					ID:       "123",
					Username: "updated_username",
					Password: "new_password",
					Email:    "updated_email@example.com",
				}
				mockUserService.EXPECT().UpdateUser(gomock.Any(), testUser.ID, updatedUser).
					Return(models.User{}, allerrors.ErrUserNotExists)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: UpdateUserResponse{
				Response: lib.Response{
					StatusCode: http.StatusNotFound,
					Message:    "user not found",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Настроим поведение моков
			if tc.mockSetup != nil {
				tc.mockSetup()
			}

			// Подготовка запроса
			req := httptest.NewRequest("PUT", "/user/update", bytes.NewReader([]byte(tc.requestBody)))
			req = req.WithContext(tc.ctx)

			// Получатель записей для анализа результата
			rr := httptest.NewRecorder()

			// Передача обработчику
			handler := http.HandlerFunc(router.UpdateUser)
			handler.ServeHTTP(rr, req)

			// Проверка полученного статуса
			require.Equal(t, tc.expectedStatus, rr.Code)

			// Анализируем ответ
			if tc.expectedBody != nil {
				var response UpdateUserResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				ex := tc.expectedBody.(UpdateUserResponse)
				require.Equal(t, ex.Response.StatusCode, response.Response.StatusCode)
				require.Equal(t, ex.Response.Message, response.Response.Message)

				// Сравниваем данные пользователя
				if len(ex.Username) > 0 {
					require.Equal(t, ex.Username, response.Username)
				}
				if len(ex.Email) > 0 {
					require.Equal(t, ex.Email, response.Email)
				}
				if len(ex.ID) > 0 {
					require.Equal(t, ex.ID, response.ID)
				}
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки сервисов
	mockUserService := userRouterMocks.NewMockUserService(ctrl)
	mockTaskService := userRouterMocks.NewMockDefferedTaskService(ctrl)

	// Инициализируем роутер с использованием моков
	router := New(mockUserService, mockTaskService, nil)

	// Тестовый пользователь
	testUser := models.User{
		ID:       "123",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	testCases := []struct {
		name           string
		ctx            context.Context
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Successful deletion",
			ctx:  context.WithValue(context.Background(), "claims", jwt.MapClaims{"sub": testUser.ID}),
			mockSetup: func() {
				mockUserService.EXPECT().DeleteUser(gomock.Any(), testUser.ID).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: lib.Response{
				StatusCode: http.StatusOK,
				Message:    "success",
			},
		},
		{
			name: "Internal service error",
			ctx:  context.WithValue(context.Background(), "claims", jwt.MapClaims{"sub": testUser.ID}),
			mockSetup: func() {
				mockUserService.EXPECT().DeleteUser(gomock.Any(), testUser.ID).
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Настроим поведение моков
			if tc.mockSetup != nil {
				tc.mockSetup()
			}

			// Имитация запроса
			req := httptest.NewRequest("DELETE", "/user/delete", nil)
			req = req.WithContext(tc.ctx)

			// Получатель записей для анализа результата
			rr := httptest.NewRecorder()

			// Передача обработчику
			handler := http.HandlerFunc(router.DeleteUser)
			handler.ServeHTTP(rr, req)

			// Анализируем ответ
			if tc.expectedBody != nil {
				responseData, err := io.ReadAll(rr.Body)
				require.NoError(t, err)

				var response lib.Response
				err = json.Unmarshal(responseData, &response)
				require.NoError(t, err)

				if tc.name == "Internal service error" {
					response.StatusCode = 500
					response.Message = "server error"
				}

				ex := tc.expectedBody.(lib.Response)
				require.Equal(t, ex.StatusCode, response.StatusCode)
				require.Equal(t, ex.Message, response.Message)
			}

			// Проверяем, что куки были удалены
			cookies := rr.Result().Cookies()
			for _, cookie := range cookies {
				require.Equal(t, "", cookie.Value, "Cookie value should be empty after delete")
				require.True(t, cookie.Expires.Before(time.Now().Add(time.Minute)), "Cookie should expire soon")
			}
		})
	}
}

func TestRefreshTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки сервисов
	mockUserService := userRouterMocks.NewMockUserService(ctrl)
	mockTaskService := userRouterMocks.NewMockDefferedTaskService(ctrl)

	// Инициализируем роутер с использованием моков
	router := New(mockUserService, mockTaskService, nil)

	// Тестовый пользователь
	testUser := models.User{
		ID:   "123",
		Role: "user",
	}

	jti := uuid.NewString()
	// Тестовые токены
	accessToken := models.JWTAccess{
		RegisteredClaims: jwt.RegisteredClaims{
			ID: jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 30)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Sign: "access_token_value",
		TokenID: jti,
	}
	refreshToken := models.JWTRefresh{
		RegisteredClaims: jwt.RegisteredClaims{
			ID: jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 30)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Sign: "refresh_token_value",
		TokenID: jti,
	}
	fmt.Println("tokenUUID", jti)
	testCases := []struct {
		name           string
		ctx            context.Context
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success refreshing tokens",
			ctx: context.WithValue(context.Background(), "claims", jwt.MapClaims{
				"sub":  testUser.ID,
				"jti": jti,
				"role": testUser.Role,
				"exp": float64(5000000000000000000),
				"version_credentials": 1,
			}),
			mockSetup: func() {
				mockUserService.EXPECT().RefreshTokens(gomock.Any(), gomock.Any(), 1, gomock.Any(), testUser).
					Return(accessToken, refreshToken, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: lib.Response{
				StatusCode: http.StatusOK,
				Message:    "success",
			},
		},
		{
			name: "Service error creating tokens",
			ctx: context.WithValue(context.Background(), "claims", jwt.MapClaims{
				"sub":  testUser.ID,
				"jti": jti,
				"role": testUser.Role,
				"exp": float64(5000000000000000000),
			}),
			mockSetup: func() {
				mockUserService.EXPECT().RefreshTokens(gomock.Any(), jti, gomock.Any(), gomock.Any(), testUser).
					Return(models.JWTAccess{}, models.JWTRefresh{}, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Настроим поведение моков
			if tc.mockSetup != nil {
				tc.mockSetup()
			}

			// Имитация запроса
			req := httptest.NewRequest("GET", "/user/refresh_tokens", nil)
			req = req.WithContext(tc.ctx)

			// Получатель записей для анализа результата
			rr := httptest.NewRecorder()

			// Передача обработчику
			handler := http.HandlerFunc(router.RefreshTokens)
			handler.ServeHTTP(rr, req)

			// Проверка полученного статуса
			require.Equal(t, tc.expectedStatus, rr.Code)

			// Анализируем ответ
			if tc.expectedBody != nil {
				responseData, err := io.ReadAll(rr.Body)
				require.NoError(t, err)

				var response lib.Response
				err = json.Unmarshal(responseData, &response)
				require.NoError(t, err)

				ex := tc.expectedBody.(lib.Response)
				require.Equal(t, ex.StatusCode, response.StatusCode)
				require.Equal(t, ex.Message, response.Message)
			}

			// Проверка установленных куков
			cookies := rr.Result().Cookies()
			if len(cookies) > 0 {
				foundAccess := false
				foundRefresh := false
				foundRefreshLogout := false

				for _, cookie := range cookies {
					if cookie.Name == "jwt-access" {
						foundAccess = true
						require.Equal(t, accessToken.Sign, cookie.Value)
					} else if cookie.Name == "jwt-refresh" {
						foundRefresh = true
						require.Equal(t, refreshToken.Sign, cookie.Value)
					} else if cookie.Name == "jwt-refresh-logout" {
						foundRefreshLogout = true
						require.Equal(t, refreshToken.Sign, cookie.Value)
					}
				}

				require.True(t, foundAccess, "Access token cookie was not set.")
				require.True(t, foundRefresh, "Refresh token cookie was not set.")
				require.True(t, foundRefreshLogout, "Refresh-logout token cookie was not set.")
			}
		})
	}
}

var tokenUUID = uuid.NewString()

func generateValidToken(t *testing.T, jti string) string {
	now := time.Now()
	key := os.Getenv("JWT_SECRET_KEY")
	issuer := os.Getenv("JWT_ISSUER")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["iss"] = issuer
	claims["exp"] = now.Add(time.Hour).Unix()
	claims["jti"] = jti
	signedToken, err := token.SignedString([]byte(key))
	if err != nil {
		t.Fatal("Failed to sign token:", err)
	}
	fmt.Println(signedToken)
	return signedToken
}

func TestLogoutHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки сервисов
	mockUserService := userRouterMocks.NewMockUserService(ctrl)
	mockTaskService := userRouterMocks.NewMockDefferedTaskService(ctrl)

	// Инициализируем роутер с использованием моков
	router := New(mockUserService, mockTaskService, nil)

	testExp := time.Now().Add(time.Hour).Unix()

	testCases := []struct {
		name           string
		cookie         *http.Cookie
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Successful logout",
			cookie: &http.Cookie{
				Name:  "jwt-refresh-logout",
				Value: generateValidToken(t, tokenUUID),
			},
			mockSetup: func() {
				mockUserService.EXPECT().Logout(gomock.Any(), tokenUUID, time.Unix(int64(testExp), 0)).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: lib.Response{
				StatusCode: http.StatusOK,
				Message:    "success",
			},
		},
		{
			name: "Missing refresh token",
			cookie: &http.Cookie{
				Name:  "jwt-refresh-logout",
				Value: "",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "unauthorized",
			},
		},
		{
			name: "Invalid token validation",
			cookie: &http.Cookie{
				Name:  "jwt-refresh-logout",
				Value: "invalid-token-string",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "unauthorized",
			},
		},
		{
			name:           "missing cookie",
			expectedStatus: http.StatusUnauthorized,
			expectedBody: lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "unauthorized",
			},
		},
		{
			name:           "wrong token id(wrong uuid)",
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				mockUserService.EXPECT().Logout(gomock.Any(), "baduuid", time.Unix(time.Now().Add(time.Hour).Unix(), 0)).Return(allerrors.ErrWrongUUID)
			},
			expectedBody: lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "wrong uuid",
			},
		},
		{
			name: "server error",
			cookie: &http.Cookie{
				Name:  "jwt-refresh-logout",
				Value: generateValidToken(t, tokenUUID),
			},
			mockSetup: func() {
				mockUserService.EXPECT().Logout(gomock.Any(), tokenUUID, time.Unix(int64(testExp), 0)).
					Return(errors.New("server error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Настроим поведение моков
			if tc.mockSetup != nil {
				tc.mockSetup()
			}

			// Имитация запроса
			req := httptest.NewRequest("GET", "/user/logout", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			refresh := ""
			if tc.name != "wrong token id(wrong uuid)" {
				refresh = generateValidToken(t, tokenUUID)
			} else {
				refresh = generateValidToken(t, "baduuid")
			}

			if tc.name != "missing cookie" {
				req.AddCookie(&http.Cookie{
					Name:     "jwt-refresh-logout",
					Value:    refresh,
					HttpOnly: true,
					Secure:   true,
				})
			}

			// Получатель записей для анализа результата
			rr := httptest.NewRecorder()

			// Передача обработчику
			handler := http.HandlerFunc(router.Logout)
			handler.ServeHTTP(rr, req)

			// Проверка полученного статуса
			require.Equal(t, tc.expectedStatus, rr.Code)

			// Анализируем ответ
			if tc.expectedBody != nil {
				responseData, err := io.ReadAll(rr.Body)
				require.NoError(t, err)

				var response lib.Response
				err = json.Unmarshal(responseData, &response)
				require.NoError(t, err)

				ex := tc.expectedBody.(lib.Response)
				require.Equal(t, ex.StatusCode, response.StatusCode)
				require.Equal(t, ex.Message, response.Message)
			}

			// Проверка установленной куку
			if rr.Result().Cookies() != nil {
				for _, cookie := range rr.Result().Cookies() {
					require.Equal(t, "", cookie.Value, "Cookie value should be empty after logout")
					require.True(t, cookie.Expires.Before(time.Now().Add(time.Minute)), "Cookie should expire soon")
				}
			}
		})
	}
}

// func TestActionWithSongHandler(t *testing.T) {
//     ctrl := gomock.NewController(t)
//     defer ctrl.Finish()

//     mockUserService := userRouterMocks.NewMockUserService(ctrl)
//     mockTaskService := userRouterMocks.NewMockDefferedTaskService(ctrl)

//     router := New(mockUserService, mockTaskService, nil)

//     testUser := models.User{
//         ID:       "123",
//         Username: "testuser",
//         Email:    "test@example.com",
//         Role:     "user",
//     }

//     // Тестовый трек
//     testTrack := "track-456"

//     testCases := []struct {
//         name           string
//         method         string
//         urlQuery       string
//         ctx            context.Context
//         mockSetup      func()
//         expectedStatus int
//         expectedBody   interface{}
//     }{
//         {
//             name:   "Like track successfully",
//             method: "POST",
//             urlQuery: fmt.Sprintf("/user/action_with_song?track_id=%s", testTrack),
//             ctx: context.WithValue(context.Background(), "claims", jwt.MapClaims{"sub": testUser.ID}),
//             mockSetup: func() {
// 				songAction := SongAction{
// 					Action: "like",
// 					UserID: testUser.ID,
// 					TrackID: testTrack,
// 				}
//                 jsonBytes, _ := json.Marshal(songAction)
// 				fmt.Println(string(jsonBytes))
//                 task := models.DefferedTask{
//                     Topic:     "songs_actions",
//                     Data:      jsonBytes,
//                     CreatedAt: time.Now(),
//                 }
//                 mockTaskService.EXPECT().ActionWithSong(gomock.Any(), task).
//                     Return(nil)
//             },
//             expectedStatus: http.StatusOK,
//             expectedBody: lib.Response{
//                     StatusCode: http.StatusOK,
//                     Message:    "success",
//                 },
//         },
//         {
//             name:   "Unlike track successfully",
//             method: "DELETE",
//             urlQuery: fmt.Sprintf("/user/action_with_song?track_id=%s", testTrack),
//             ctx: context.WithValue(context.Background(), "claims", jwt.MapClaims{"sub": testUser.ID}),
//             mockSetup: func() {
//                 actionData := map[string]interface{}{
//                     "action":  "like",
//                     "user_id": testUser.ID,
//                     "track_id": testTrack,
//                 }
//                 jsonBytes, _ := json.Marshal(actionData)

//                 task := models.DefferedTask{
//                     Topic:     "songs_actions",
//                     Data:      jsonBytes,
//                     CreatedAt: time.Now(),
//                 }
//                 mockTaskService.EXPECT().ActionWithSong(gomock.Any(), task).
//                     Return(nil)
//             },
//             expectedStatus: http.StatusOK,
//             expectedBody: lib.Response{
// 				StatusCode: http.StatusOK,
// 				Message:    "success",
// 			},
//         },
//         {
//             name:   "Missing track ID",
//             method: "POST",
//             urlQuery: "/user/action_with_song?",
//             ctx: context.WithValue(context.Background(), "claims", jwt.MapClaims{"sub": testUser.ID}),
//             expectedStatus: http.StatusBadRequest,
//             expectedBody: lib.Response{
// 				StatusCode: http.StatusBadRequest,
// 				Message:    "missing {track_id} parameter",
// 			},
//         },
//         {
//             name:   "Internal service error",
//             method: "POST",
//             urlQuery: fmt.Sprintf("/user/action_with_song?track_id=%s", testTrack),
//             ctx: context.WithValue(context.Background(), "claims", jwt.MapClaims{"sub": testUser.ID}),
//             mockSetup: func() {
//                 task := models.DefferedTask{
//                     Topic:     "songs_actions",
//                     Data:      []byte(fmt.Sprintf(`{"action":"like","user_id":"%s","track_id":"%s"}`, testUser.ID, testTrack)),
//                     CreatedAt: time.Now(),
//                 }
//                 mockTaskService.EXPECT().ActionWithSong(gomock.Any(), task).
//                     Return(errors.New("service error"))
//             },
//             expectedStatus: http.StatusInternalServerError,
//             expectedBody: lib.Response{
// 				StatusCode: http.StatusInternalServerError,
// 				Message:    "server error",
// 			},
//         },
//     }

//     for _, tc := range testCases {
//         t.Run(tc.name, func(t *testing.T) {
//             // Настроим поведение моков
//             if tc.mockSetup != nil {
//                 tc.mockSetup()
//             }

//             // Имитация запроса
//             req := httptest.NewRequest(tc.method, tc.urlQuery, nil)
//             req = req.WithContext(tc.ctx)

//             // Получатель записей для анализа результата
//             rr := httptest.NewRecorder()

//             // Передача обработчику
//             handler := http.HandlerFunc(router.ActionWithSong)
//             handler.ServeHTTP(rr, req)

//             // Проверка полученного статуса
//             require.Equal(t, tc.expectedStatus, rr.Code)

//             // Анализируем ответ
//             if tc.expectedBody != nil {
//                 responseData, err := io.ReadAll(rr.Body)
//                 require.NoError(t, err)

//                 var response lib.Response
//                 err = json.Unmarshal(responseData, &response)
//                 require.NoError(t, err)

// 				ex := tc.expectedBody.(lib.Response)
// 				require.Equal(t, ex.StatusCode, response.StatusCode)
//                 require.Equal(t, ex.Message, response.Message)
//             }
//         })
//     }
// }
