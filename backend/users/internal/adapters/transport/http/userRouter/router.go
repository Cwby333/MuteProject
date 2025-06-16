package userrouter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Cwby333/user-microservice/internal/adapters/transport/http/lib"
	"github.com/Cwby333/user-microservice/internal/adapters/transport/http/middleware"
	allerrors "github.com/Cwby333/user-microservice/internal/allErrors"
	"github.com/Cwby333/user-microservice/internal/models"

	"github.com/go-playground/validator/v10"
	gojson "github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
)

type UserService interface {
	Register(ctx context.Context, user models.User) (models.User, error)
	Login(ctx context.Context, user models.User) (models.JWTAccess, models.JWTRefresh, error)
	Logout(ctx context.Context, tokenID string, unixTimeExpired time.Time) error

	FindUserByID(ctx context.Context, ID string) (models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)

	DeleteUser(ctx context.Context, ID string) error

	UpdateUser(ctx context.Context, ID string, newUserInfo models.User) (models.User, error)

	RefreshTokens(ctx context.Context, tokenID string, refreshVersionCredentials int, expTime time.Time, user models.User) (access models.JWTAccess, refresh models.JWTRefresh, err error)
}

type DefferedTaskService interface {
	ActionWithSong(ctx context.Context, task models.DefferedTask) error
}

type Router struct {
	Mux         *http.ServeMux
	userService UserService
	taskService DefferedTaskService
	logger      *slog.Logger
	validator   *validator.Validate
}

func New(userService UserService, taskService DefferedTaskService, logger *slog.Logger) Router {
	return Router{
		Mux:         http.NewServeMux(),
		userService: userService,
		taskService: taskService,
		logger:      logger,
		validator:   validator.New(validator.WithRequiredStructEnabled()),
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (r *Router) Handle(pattern string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) {
	for i := 0; i < len(middlewares); i++ {
		handler = middlewares[i](handler)
	}

	r.Mux.Handle(pattern, handler)
}

func (router *Router) Run() {
	router.Handle("POST /user/register", http.HandlerFunc(router.Register), CORS, middleware.Recover, middleware.Logging)
	router.Handle("POST /user/login", http.HandlerFunc(router.Login), CORS, middleware.Recover, middleware.Logging)
	router.Handle("POST /user/logout", http.HandlerFunc(router.Logout), CORS, middleware.Recover, middleware.Logging)
	router.Handle("POST /user/refresh", http.HandlerFunc(router.RefreshTokens), CORS, middleware.Recover, middleware.Logging, middleware.RefreshJWT)

	router.Handle("OPTIONS /user/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	}), middleware.Recover, middleware.Logging)
	router.Handle("GET /user/get", http.HandlerFunc(router.GetUserByID), CORS, middleware.Recover, middleware.Logging, middleware.AccessJWT)
	router.Handle("GET /user/all", http.HandlerFunc(router.GetAllUsers), CORS, middleware.Recover, middleware.Logging, middleware.AccessJWT)

	router.Handle("DELETE /user/delete", http.HandlerFunc(router.DeleteUser), CORS, middleware.Recover, middleware.Logging, middleware.AccessJWT)
	router.Handle("PUT /user/update", http.HandlerFunc(router.UpdateUser), CORS, middleware.Recover, middleware.Logging, middleware.AccessJWT)

	// Handle OPTIONS requests for tracks/favorite separately to allow preflight without JWT
	router.Handle("OPTIONS /user/track/favorite", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	}), middleware.Recover, middleware.Logging)

	// Regular non-OPTIONS requests still need JWT auth
	router.Handle("POST /user/track/favorite", http.HandlerFunc(router.ActionWithSong), CORS, middleware.Recover, middleware.Logging, middleware.AccessJWT)
	router.Handle("DELETE /user/track/favorite", http.HandlerFunc(router.ActionWithSong), CORS, middleware.Recover, middleware.Logging, middleware.AccessJWT)

	// Добавляем обработку OPTIONS запросов для нового эндпоинта
	router.Handle("OPTIONS /user/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	}), middleware.Recover, middleware.Logging)

	// Добавляем сам эндпоинт
	router.Handle("GET /user/{id}", http.HandlerFunc(router.GetUserByIDParam), CORS, middleware.Recover, middleware.Logging)
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegisterResponse struct {
	Response lib.Response
	Username string `json:"username" omitempty:"true"`
	Email    string `json:"email" omitempty:"true"`
	ID       string `json:"id" omitempty:"true"`
}

func (router *Router) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	fmt.Println("Received method:", r.Method)

	data, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Info("registerHandler read body", slog.String("error", err.Error()))

		resp := RegisterResponse{
			Response: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		}
		data, err = gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	var req RegisterRequest
	err = gojson.Unmarshal(data, &req)
	if err != nil {
		slog.Info("gojson unmarshal", slog.String("error", err.Error()))

		resp := lib.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "server error",
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	err = router.validator.Struct(req)
	if err != nil {
		errorsValidate := err.(validator.ValidationErrors)

		errors := lib.Validate(errorsValidate)

		errForResp := ""
		for i := range errors {
			errForResp += errors[i] + " "
		}
		errForResp = strings.TrimSpace(errForResp)

		slog.Info("validation", slog.String("error", errForResp))

		resp := RegisterResponse{
			Response: lib.Response{
				StatusCode: http.StatusBadRequest,
				Message:    errForResp,
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			http.Error(w, fmt.Sprintf("bad request: %s", errForResp), http.StatusBadRequest)
			return
		}

		http.Error(w, string(data), http.StatusBadRequest)
		return
	}

	userDTO := UserDTO{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	user := DTOToUser(userDTO)

	// isAdmin := r.Context().Value("isAdmin").(bool)
	// if isAdmin {
	// 	user.Role = "admin"
	// }

	user, err = router.userService.Register(r.Context(), user)
	if err != nil {
		slog.Info("register handler", slog.String("error", err.Error()))

		if errors.Is(err, allerrors.ErrUsernameExists) {
			resp := RegisterResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "username already exists",
				},
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "username already exists", http.StatusBadRequest)
				return
			}

			http.Error(w, string(data), http.StatusBadRequest)
			return
		}

		if errors.Is(err, allerrors.ErrEmailExists) {
			resp := RegisterResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "email already exists",
				},
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "email already exists", http.StatusBadRequest)
				return
			}

			http.Error(w, string(data), http.StatusBadRequest)
			return
		}

		if errors.Is(err, allerrors.ErrPasswordSmall) {
			resp := RegisterResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "password to small",
				},
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "password to small", http.StatusBadRequest)
				return
			}

			http.Error(w, string(data), http.StatusBadRequest)
			return
		}

		if errors.Is(err, allerrors.ErrPasswordBig) {
			resp := RegisterResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "password to big",
				},
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "password to big", http.StatusBadRequest)
				return
			}

			http.Error(w, string(data), http.StatusBadRequest)
			return
		}

		resp := RegisterResponse{
			Response: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	userDTO = UserToDTO(user)

	response := RegisterResponse{
		Response: lib.Response{
			StatusCode: http.StatusOK,
			Message:    "success register",
		},
		Username: userDTO.Username,
		Email:    userDTO.Email,
		ID:       userDTO.ID,
	}
	data, err = gojson.Marshal(response)
	if err != nil {
		slog.Info("gojson unmarshall", slog.String("error", err.Error()))

		strResponse := fmt.Sprintf("success register, username: %s email: %s ID: %s", userDTO.Username, userDTO.Email, userDTO.ID)
		w.Write([]byte(strResponse))
		return
	}

	w.Write(data)
	return
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Response lib.Response `json:"response"`
	Token    string       `json:"token,omitempty"`
	ID       string       `json:"id,omitempty"`
}

func (router *Router) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	fmt.Println("Received method:", r.Method)
	var req LoginRequest

	data, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Info("registerHandler read body", slog.String("error", err.Error()))

		resp := RegisterResponse{
			Response: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "serve error",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	err = gojson.Unmarshal(data, &req)
	if err != nil {
		slog.Info("gojson unmarshal", slog.String("error", err.Error()))

		resp := lib.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "server error",
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	err = router.validator.Struct(req)
	if err != nil {
		slog.Info("validator error", slog.String("error", err.Error()))

		errorsValidate := err.(validator.ValidationErrors)

		errors := lib.Validate(errorsValidate)

		errForResp := ""
		for i := range errors {
			errForResp += errors[i] + " "
		}

		resp := RegisterResponse{
			Response: lib.Response{
				StatusCode: http.StatusBadRequest,
				Message:    errForResp,
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, fmt.Sprintf("bad request: %s", errForResp), http.StatusBadRequest)
			return
		}

		http.Error(w, string(data), http.StatusBadRequest)
		return
	}

	userDTO := UserDTO{
		Username: req.Username,
		Password: req.Password,
	}

	user := DTOToUser(userDTO)

	access, refresh, err := router.userService.Login(r.Context(), user)
	if err != nil {
		slog.Info("login handler", slog.String("error", err.Error()))

		if errors.Is(err, allerrors.ErrUserNotExists) {
			slog.Info("wrong username")
			resp := LoginResponse{
				Response: lib.Response{
					StatusCode: http.StatusNotFound,
					Message:    "username not found",
				},
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "username not found", http.StatusNotFound)
				return
			}

			http.Error(w, string(data), http.StatusNotFound)
			return
		}
		if errors.Is(err, allerrors.ErrWrongPass) {
			resp := LoginResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "wrong password",
				},
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "wrong password", http.StatusBadRequest)
				return
			}

			http.Error(w, string(data), http.StatusBadRequest)
			return
		}

		slog.Error("server error")
		resp := LoginResponse{
			Response: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		Response: lib.Response{
			StatusCode: http.StatusOK,
			Message:    fmt.Sprintf("success login, ID: %s", access.Subject),
		},
		Token: access.Sign,
		ID:    access.Subject,
	}
	fmt.Println("resp: ", resp)
	data, err = gojson.Marshal(resp)
	if err != nil {
		slog.Info("gojson marshal", slog.String("error", err.Error()))

		_, err := w.Write([]byte(fmt.Sprintf("success login, ID: %s", access.Subject)))
		if err != nil {
			slog.Info("response write", slog.String("error", err.Error()))
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "jwt-access",
			Value:    access.Sign,
			HttpOnly: true,
			Secure:   true,
			Expires:  access.ExpiresAt.Time,
		})
		http.SetCookie(w, &http.Cookie{
			Name:    "jwt-refresh",
			Value:   refresh.Sign,
			Secure:  true,
			Expires: refresh.ExpiresAt.Time,
			Path:    "/user/refresh",
		})
		http.SetCookie(w, &http.Cookie{
			Name:    "jwt-refresh-logout",
			Value:   refresh.Sign,
			Secure:  true,
			Expires: refresh.ExpiresAt.Time,
			Path:    "/user/logout",
		})
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt-access",
		Value:    access.Sign,
		HttpOnly: true,
		Secure:   true,
		Expires:  access.ExpiresAt.Time,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt-refresh",
		Value:   refresh.Sign,
		Secure:  true,
		Expires: refresh.ExpiresAt.Time,
		Path:    "/user/refresh",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt-refresh-logout",
		Value:   refresh.Sign,
		Secure:  true,
		Expires: refresh.ExpiresAt.Time,
		Path:    "/user/logout",
	})

	_, err = w.Write(data)
	if err != nil {
		slog.Error("response write", slog.String("error", err.Error()))
	}

	// Write the JSON response with user ID
	w.Write(data)
	return
}

// TODO: change get refresh token by logout handler (check login, refreshTokens handlers, in cookie)
// Or remove logic store invalid refresh tokens in redis
func (router *Router) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	fmt.Println("Received method:", r.Method)
	refresh, err := r.Cookie("jwt-refresh-logout")
	if err != nil {
		slog.Info("missing refresh token for logout")

		resp := lib.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "unauthorized",
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		http.Error(w, string(data), http.StatusUnauthorized)
		return
	}

	token := refresh.Value
	claims, err := lib.ValidateJWT(token)
	if err != nil {
		slog.Info("validate jwt", slog.String("error", err.Error()))

		resp := lib.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "unauthorized",
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		http.Error(w, string(data), http.StatusUnauthorized)
		return
	}

	jti := claims["jti"].(string)
	_ = jti
	exp := claims["exp"].(float64)
	expiredTime := time.Unix(int64(exp), 0)

	err = router.userService.Logout(r.Context(), jti, expiredTime)
	if err != nil {
		if errors.Is(err, allerrors.ErrWrongUUID) {
			slog.Info("logout handler", slog.String("error", err.Error()))

			resp := lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "wrong uuid",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "wrong uuid", http.StatusUnauthorized)
				return
			}

			http.Error(w, string(data), http.StatusUnauthorized)
			return
		}

		slog.Info("logout handler", slog.String("error", err.Error()))

		resp := lib.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "server error",
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt-access",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(time.Second * 3),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt-refresh",
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		Expires:  time.Now().Add(time.Second * 3),
		Path:     "/user/refresh",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt-refresh-logout",
		Value:   "",
		Secure:  true,
		Expires: time.Now().Add(time.Second * 3),
		Path:    "/user/logout",
	})

	resp := lib.Response{
		StatusCode: http.StatusOK,
		Message:    "success",
	}
	data, err := gojson.Marshal(resp)
	if err != nil {
		slog.Info("gojson marshal", slog.String("error", err.Error()))

		_, err = w.Write([]byte("success"))
		if err != nil {
			slog.Error("response write", slog.String("error", err.Error()))
		}

		return
	}

	slog.Info("success logout")

	_, err = w.Write(data)
	if err != nil {
		slog.Error("response write", slog.String("error", err.Error()))
	}
}

func (router *Router) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	fmt.Println("Received method:", r.Method)
	claims := r.Context().Value("claims").(jwt.MapClaims)
	userID, ok := claims["sub"].(string)
	if !ok {
		slog.Info("refreshTokens sub missed")
	}
	role, ok := claims["role"].(string)
	if !ok {
		slog.Info("refreshTokens role missed")
	}
	jti, ok := claims["jti"].(string)
	if !ok {
		slog.Info("refreshTokens jti missed")
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		slog.Info("refreshTokens exp missed")
	}
	versionCredentials, ok := claims["version_credentials"].(int)
	if !ok {
		slog.Info("refreshTokens version_credentials missed")
	}
	unixExp := time.Unix(int64(exp), 0)

	user := models.User{
		ID:   userID,
		Role: role,
	}

	access, refresh, err := router.userService.RefreshTokens(r.Context(), jti, versionCredentials, unixExp, user)
	if err != nil {
		if errors.Is(err, allerrors.ErrTokenInBlackList) {
			slog.Info("token in blacklist")

			resp := lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message: "token in blacklist, please, change your credentials",
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				_, err = w.Write([]byte("token in blacklist"))
				if err != nil {
					slog.Info("response write", slog.String("error", err.Error()))
				}
				return
			}

			_, err = w.Write(data)
			if err != nil {
				slog.Info("response write", slog.String("error", err.Error()))
			}
			return
		}

		slog.Info("create tokens", slog.String("error", err.Error()))

		resp := lib.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "server error",
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	slog.Info("success refresh")

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt-access",
		Value:    access.Sign,
		HttpOnly: false,
		Secure:   true,
		Expires:  access.ExpiresAt.Time,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt-refresh",
		Value:   refresh.Sign,
		Secure:  true,
		Expires: refresh.ExpiresAt.Time,
		Path:    "/user/refresh",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt-refresh-logout",
		Value:   refresh.Sign,
		Secure:  true,
		Expires: refresh.ExpiresAt.Time,
		Path:    "/user/logout",
	})

	resp := lib.Response{
		StatusCode: http.StatusOK,
		Message:    "success",
	}
	data, err := gojson.Marshal(resp)
	if err != nil {
		slog.Info("gojson marshal", slog.String("error", err.Error()))

		w.Write([]byte("success"))
		return
	}

	_, err = w.Write(data)
	if err != nil {
		slog.Info("response write", slog.String("error", err.Error()))
	}
}

type GetUserByIDResponse struct {
	Response lib.Response `json:"response"`
	Username string       `json:"username" omitempty:"true"`
	ID       string       `json:"id" omitempty:"true"`
}

func (router *Router) GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	fmt.Println("Received method:", r.Method)

	// Extract user ID from JWT claims in the context
	claims := r.Context().Value("claims").(jwt.MapClaims)
	ID := claims["sub"].(string)

	if ID == "" {
		slog.Info("getUserByIDHandler", slog.String("error", "missing ID in token"))

		resp := GetUserByIDResponse{
			Response: lib.Response{
				StatusCode: http.StatusBadRequest,
				Message:    "missing user ID in token",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "missing {user_id} in query parameter", http.StatusBadRequest)
			return
		}

		http.Error(w, string(data), http.StatusBadRequest)
		return
	}

	user, err := router.userService.FindUserByID(r.Context(), ID)
	if err != nil {
		slog.Info("getUserByIdHandler", slog.String("err", err.Error()))

		if errors.Is(err, allerrors.ErrWrongUUID) {
			resp := GetUserByIDResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "wrong user ID",
				},
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "wrong user ID", http.StatusBadRequest)
				return
			}

			http.Error(w, string(data), http.StatusBadRequest)
			return
		}
		if errors.Is(err, allerrors.ErrUserNotExists) {
			resp := GetUserByIDResponse{
				Response: lib.Response{
					StatusCode: http.StatusNotFound,
					Message:    "user not found",
				},
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "user not found", http.StatusNotFound)
				return
			}

			http.Error(w, string(data), http.StatusNotFound)
			return
		}

		resp := GetUserByIDResponse{
			Response: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	resp := GetUserByIDResponse{
		Response: lib.Response{
			StatusCode: http.StatusOK,
			Message:    "success",
		},
		Username: user.Username,
		ID:       user.ID,
	}
	data, err := gojson.Marshal(resp)
	if err != nil {
		slog.Info("gojson marshal", slog.String("error", err.Error()))

		strOut := fmt.Sprintf("username: {%s}, ID: {%s}", user.Username, user.ID)
		w.Write([]byte(strOut))
		return
	}

	slog.Info("getUserByIDHandler success")

	_, err = w.Write(data)
	if err != nil {
		slog.Error("response write", slog.String("error", err.Error()))
	}
}

type GetAllUsersResponse struct {
	Response lib.Response `json:"response"`
	Users    []UserDTO    `json:"users" omitempty:"true"`
}

func (router *Router) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	fmt.Println("Received method:", r.Method)
	claims := r.Context().Value("claims").(jwt.MapClaims)
	role := claims["role"].(string)

	if role != "admin" {
		slog.Info("get all users", slog.String("error", fmt.Sprintf("not a admin, ID: %s", claims["sub"].(string))))

		resp := GetAllUsersResponse{
			Response: lib.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "unauthorized",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		http.Error(w, string(data), http.StatusUnauthorized)
		return
	}

	slice, err := router.userService.GetAllUsers(r.Context())
	if err != nil {
		slog.Info("getAllUsersHandler", slog.String("error", err.Error()))

		resp := GetAllUsersResponse{
			Response: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	sliceDTO := make([]UserDTO, 0, len(slice))
	for i := range slice {
		sliceDTO = append(sliceDTO, UserToDTO(slice[i]))
	}

	resp := GetAllUsersResponse{
		Response: lib.Response{
			StatusCode: http.StatusOK,
			Message:    "success",
		},
		Users: sliceDTO,
	}
	data, err := gojson.Marshal(resp)
	if err != nil {
		slog.Info("gojson marshal", slog.String("error", err.Error()))

		strOut := ""
		for i := range sliceDTO {
			str := fmt.Sprintf("ID: {%s}, username: {%s}, email: {%s}, password: {%s}, role: {%s}", sliceDTO[i].ID, sliceDTO[i].Username, sliceDTO[i].Email, sliceDTO[i].Password, sliceDTO[i].Role)

			strOut += str + " "
		}

		w.Write([]byte(strOut))
		return
	}

	slog.Info("getAllUsersHandler success")

	_, err = w.Write(data)
	if err != nil {
		slog.Error("response write", slog.String("error", err.Error()))
	}
}

type UpdateUserRequest struct {
	NewUsername string `json:"username"`
	NewEmail    string `json:"email" `
	NewPassword string `json:"password"`
}

type UpdateUserResponse struct {
	Response lib.Response `yaml:"response"`
	Username string       `json:"username" omitempty:"true"`
	Email    string       `json:"email" omitempty:"true"`
	ID       string       `json:"id" omitempty:"true"`
}

func (router *Router) UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	fmt.Println("Received method:", r.Method)
	claims := r.Context().Value("claims").(jwt.MapClaims)

	var req UpdateUserRequest

	data, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Info("updateHandler read body", slog.String("error", err.Error()))

		resp := RegisterResponse{
			Response: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	err = gojson.Unmarshal(data, &req)
	if err != nil {
		slog.Info("gojson unmarshal", slog.String("error", err.Error()))

		resp := UpdateUserResponse{
			Response: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:       claims["sub"].(string),
		Username: req.NewUsername,
		Email:    req.NewEmail,
		Password: req.NewPassword,
	}

	user, err = router.userService.UpdateUser(r.Context(), user.ID, user)
	if err != nil {
		if errors.Is(err, allerrors.ErrUserNotExists) {
			slog.Info("updateUser handler: user not exists")

			resp := UpdateUserResponse{
				Response: lib.Response{
					StatusCode: http.StatusNotFound,
					Message:    "user not found",
				},
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				slog.Info("gojson marshal", slog.String("error", err.Error()))

				http.Error(w, "user not found", http.StatusNotFound)
				return
			}

			http.Error(w, string(data), http.StatusNotFound)
			return
		}

		slog.Info("updateUser handler", slog.String("error", err.Error()))

		resp := UpdateUserResponse{
			Response: lib.Response{
				StatusCode: http.StatusNotFound,
				Message:    "server error",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshal", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusNotFound)
			return
		}

		http.Error(w, string(data), http.StatusNotFound)
		return
	}

	resp := UpdateUserResponse{
		Response: lib.Response{
			StatusCode: http.StatusOK,
			Message:    "success",
		},
		Username: user.Username,
		Email:    user.Email,
		ID:       user.ID,
	}
	data, err = gojson.Marshal(resp)
	if err != nil {
		slog.Info("gojson marshal", slog.String("error", err.Error()))

		w.Write([]byte("success"))
		return
	}

	_, err = w.Write(data)
	if err != nil {
		slog.Info("gojson marshal", slog.String("error", err.Error()))
	}
}

func (router *Router) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	fmt.Println("Received method:", r.Method)
	claims := r.Context().Value("claims").(jwt.MapClaims)

	userID := claims["sub"].(string)

	err := router.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		slog.Info("deleteUser handler", slog.String("error", err.Error()))

		resp := UpdateUserResponse{
			Response: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			slog.Info("gojson marshall", slog.String("error", err.Error()))

			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	resp := lib.Response{
		StatusCode: http.StatusOK,
		Message:    "success",
	}
	data, err := gojson.Marshal(resp)
	if err != nil {
		slog.Info("gojson marshal", slog.String("error", err.Error()))

		w.Write([]byte("success"))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt-access",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(time.Second * 3),
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt-refresh",
		Value:   "",
		Secure:  true,
		Expires: time.Now().Add(time.Second * 3),
		Path:    "/api/users/refresh",
	})

	_, err = w.Write(data)
	if err != nil {
		slog.Info("gojson marshal", slog.String("error", err.Error()))
	}
}

type SongAction struct {
	Action  string `json:"action"`
	UserID  string `json:"user_id"`
	TrackID string `json:"track_id"`
}

func (router *Router) ActionWithSong(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	fmt.Println("CALLED!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 1) Извлекаем user_id из JWT
	claims := r.Context().Value("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)

	// 2) Забираем track_id из query-параметров
	songID := r.URL.Query().Get("track_id")
	fmt.Printf("SONG ID!!!!!!!!!: %v\n", songID)
	if songID == "" {
		resp := lib.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "missing {track_id} parameter",
		}
		data, _ := gojson.Marshal(resp)
		http.Error(w, string(data), http.StatusBadRequest)
		return
	}

	// 3) Определяем действие: like или unlike
	action := "like"
	if r.Method == http.MethodDelete {
		action = "dislike"
	}

	// 4) Формируем задачу
	songAction := SongAction{
		Action:  action,
		UserID:  userID,
		TrackID: songID,
	}

	// 5) Сериализуем в JSON и логируем
	rawData, err := gojson.Marshal(songAction)
	if err != nil {
		slog.Info("gojson marshal", slog.String("error", err.Error()))
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	slog.Info("enqueue song action", slog.String("raw data", string(rawData)))

	// 6) Пушим в таблицу deferred_tasks
	task := models.DefferedTask{
		Topic:     "songs_actions",
		Data:      rawData,
		CreatedAt: time.Now(),
	}
	if err := router.taskService.ActionWithSong(r.Context(), task); err != nil {
		slog.Info("actionWithSong handler", slog.String("error", err.Error()))
		resp := lib.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "server error",
		}
		data, _ := gojson.Marshal(resp)
		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	// 7) Отвечаем success
	resp := lib.Response{
		StatusCode: http.StatusOK,
		Message:    "success",
	}
	out, _ := gojson.Marshal(resp)
	w.Write(out)
}

// Определение типа ответа для GetUserByIDParam
type GetUserByIDParamResponse struct {
	Response lib.Response `json:"response"`
	User     UserDTO      `json:"user,omitempty"`
}

// Новый обработчик для получения пользователя по ID из URL-параметра
func (router *Router) GetUserByIDParam(w http.ResponseWriter, r *http.Request) {
	// Извлечение ID из URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		resp := GetUserByIDParamResponse{
			Response: lib.Response{
				StatusCode: http.StatusBadRequest,
				Message:    "invalid URL, missing user ID",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			http.Error(w, "invalid URL, missing user ID", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}

	userID := parts[2]

	// Получение пользователя из сервиса
	user, err := router.userService.FindUserByID(r.Context(), userID)
	if err != nil {
		slog.Info("get user by id param handler", slog.String("error", err.Error()))

		if errors.Is(err, allerrors.ErrUserNotExists) {
			resp := GetUserByIDParamResponse{
				Response: lib.Response{
					StatusCode: http.StatusNotFound,
					Message:    "user not found",
				},
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(data)
			return
		}

		if errors.Is(err, allerrors.ErrWrongUUID) {
			resp := GetUserByIDParamResponse{
				Response: lib.Response{
					StatusCode: http.StatusBadRequest,
					Message:    "invalid user ID format",
				},
			}
			data, err := gojson.Marshal(resp)
			if err != nil {
				http.Error(w, "invalid user ID format", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		resp := GetUserByIDParamResponse{
			Response: lib.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    "server error",
			},
		}
		data, err := gojson.Marshal(resp)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(data)
		return
	}

	// Создаем DTO без пароля
	userDTO := UserToDTO(user)
	userDTO.Password = "" // Не отправляем пароль клиенту

	// Успешный ответ
	resp := GetUserByIDParamResponse{
		Response: lib.Response{
			StatusCode: http.StatusOK,
			Message:    "success",
		},
		User: userDTO,
	}

	data, err := gojson.Marshal(resp)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
