// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/adapters/transport/http/userRouter/router.go

// Package userRouterMocks is a generated GoMock package.
package userRouterMocks

import (
	context "context"
	reflect "reflect"
	time "time"

	models "github.com/Cwby333/user-microservice/internal/models"
	gomock "github.com/golang/mock/gomock"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// DeleteUser mocks base method.
func (m *MockUserService) DeleteUser(ctx context.Context, ID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", ctx, ID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockUserServiceMockRecorder) DeleteUser(ctx, ID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockUserService)(nil).DeleteUser), ctx, ID)
}

// FindUserByID mocks base method.
func (m *MockUserService) FindUserByID(ctx context.Context, ID string) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindUserByID", ctx, ID)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindUserByID indicates an expected call of FindUserByID.
func (mr *MockUserServiceMockRecorder) FindUserByID(ctx, ID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindUserByID", reflect.TypeOf((*MockUserService)(nil).FindUserByID), ctx, ID)
}

// GetAllUsers mocks base method.
func (m *MockUserService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllUsers", ctx)
	ret0, _ := ret[0].([]models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllUsers indicates an expected call of GetAllUsers.
func (mr *MockUserServiceMockRecorder) GetAllUsers(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllUsers", reflect.TypeOf((*MockUserService)(nil).GetAllUsers), ctx)
}

// Login mocks base method.
func (m *MockUserService) Login(ctx context.Context, user models.User) (models.JWTAccess, models.JWTRefresh, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", ctx, user)
	ret0, _ := ret[0].(models.JWTAccess)
	ret1, _ := ret[1].(models.JWTRefresh)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Login indicates an expected call of Login.
func (mr *MockUserServiceMockRecorder) Login(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockUserService)(nil).Login), ctx, user)
}

// Logout mocks base method.
func (m *MockUserService) Logout(ctx context.Context, tokenID string, unixTimeExpired time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Logout", ctx, tokenID, unixTimeExpired)
	ret0, _ := ret[0].(error)
	return ret0
}

// Logout indicates an expected call of Logout.
func (mr *MockUserServiceMockRecorder) Logout(ctx, tokenID, unixTimeExpired interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logout", reflect.TypeOf((*MockUserService)(nil).Logout), ctx, tokenID, unixTimeExpired)
}

// RefreshTokens mocks base method.
func (m *MockUserService) RefreshTokens(ctx context.Context, tokenID string, refreshVersionCredentials int, expTime time.Time, user models.User) (models.JWTAccess, models.JWTRefresh, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshTokens", ctx, tokenID, refreshVersionCredentials, expTime, user)
	ret0, _ := ret[0].(models.JWTAccess)
	ret1, _ := ret[1].(models.JWTRefresh)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// RefreshTokens indicates an expected call of RefreshTokens.
func (mr *MockUserServiceMockRecorder) RefreshTokens(ctx, tokenID, refreshVersionCredentials, expTime, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshTokens", reflect.TypeOf((*MockUserService)(nil).RefreshTokens), ctx, tokenID, refreshVersionCredentials, expTime, user)
}

// Register mocks base method.
func (m *MockUserService) Register(ctx context.Context, user models.User) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", ctx, user)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Register indicates an expected call of Register.
func (mr *MockUserServiceMockRecorder) Register(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockUserService)(nil).Register), ctx, user)
}

// UpdateUser mocks base method.
func (m *MockUserService) UpdateUser(ctx context.Context, ID string, newUserInfo models.User) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUser", ctx, ID, newUserInfo)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateUser indicates an expected call of UpdateUser.
func (mr *MockUserServiceMockRecorder) UpdateUser(ctx, ID, newUserInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockUserService)(nil).UpdateUser), ctx, ID, newUserInfo)
}

// MockDefferedTaskService is a mock of DefferedTaskService interface.
type MockDefferedTaskService struct {
	ctrl     *gomock.Controller
	recorder *MockDefferedTaskServiceMockRecorder
}

// MockDefferedTaskServiceMockRecorder is the mock recorder for MockDefferedTaskService.
type MockDefferedTaskServiceMockRecorder struct {
	mock *MockDefferedTaskService
}

// NewMockDefferedTaskService creates a new mock instance.
func NewMockDefferedTaskService(ctrl *gomock.Controller) *MockDefferedTaskService {
	mock := &MockDefferedTaskService{ctrl: ctrl}
	mock.recorder = &MockDefferedTaskServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDefferedTaskService) EXPECT() *MockDefferedTaskServiceMockRecorder {
	return m.recorder
}

// ActionWithSong mocks base method.
func (m *MockDefferedTaskService) ActionWithSong(ctx context.Context, task models.DefferedTask) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ActionWithSong", ctx, task)
	ret0, _ := ret[0].(error)
	return ret0
}

// ActionWithSong indicates an expected call of ActionWithSong.
func (mr *MockDefferedTaskServiceMockRecorder) ActionWithSong(ctx, task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ActionWithSong", reflect.TypeOf((*MockDefferedTaskService)(nil).ActionWithSong), ctx, task)
}
