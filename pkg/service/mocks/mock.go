// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package mock_service is a generated GoMock package.
package mock_service

import (
	AVITO "avito/elements"
	cache "avito/pkg/cache"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockUsers is a mock of Users interface.
type MockUsers struct {
	ctrl     *gomock.Controller
	recorder *MockUsersMockRecorder
}

// MockUsersMockRecorder is the mock recorder for MockUsers.
type MockUsersMockRecorder struct {
	mock *MockUsers
}

// NewMockUsers creates a new mock instance.
func NewMockUsers(ctrl *gomock.Controller) *MockUsers {
	mock := &MockUsers{ctrl: ctrl}
	mock.recorder = &MockUsersMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUsers) EXPECT() *MockUsersMockRecorder {
	return m.recorder
}

// CheckUserToken mocks base method.
func (m *MockUsers) CheckUserToken(token string) (bool, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckUserToken", token)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CheckUserToken indicates an expected call of CheckUserToken.
func (mr *MockUsersMockRecorder) CheckUserToken(token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckUserToken", reflect.TypeOf((*MockUsers)(nil).CheckUserToken), token)
}

// MockBanners is a mock of Banners interface.
type MockBanners struct {
	ctrl     *gomock.Controller
	recorder *MockBannersMockRecorder
}

// MockBannersMockRecorder is the mock recorder for MockBanners.
type MockBannersMockRecorder struct {
	mock *MockBanners
}

// NewMockBanners creates a new mock instance.
func NewMockBanners(ctrl *gomock.Controller) *MockBanners {
	mock := &MockBanners{ctrl: ctrl}
	mock.recorder = &MockBannersMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBanners) EXPECT() *MockBannersMockRecorder {
	return m.recorder
}

// AdminGetBanners mocks base method.
func (m *MockBanners) AdminGetBanners(tag, feature, limit, offset int) ([]AVITO.Banner, [][]int, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AdminGetBanners", tag, feature, limit, offset)
	ret0, _ := ret[0].([]AVITO.Banner)
	ret1, _ := ret[1].([][]int)
	ret2, _ := ret[2].(int)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// AdminGetBanners indicates an expected call of AdminGetBanners.
func (mr *MockBannersMockRecorder) AdminGetBanners(tag, feature, limit, offset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AdminGetBanners", reflect.TypeOf((*MockBanners)(nil).AdminGetBanners), tag, feature, limit, offset)
}

// ChangeBanner mocks base method.
func (m *MockBanners) ChangeBanner(banner_id int, tags []int, feature int, content AVITO.Banner, redisCache cache.CacheImages) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChangeBanner", banner_id, tags, feature, content, redisCache)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChangeBanner indicates an expected call of ChangeBanner.
func (mr *MockBannersMockRecorder) ChangeBanner(banner_id, tags, feature, content, redisCache interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChangeBanner", reflect.TypeOf((*MockBanners)(nil).ChangeBanner), banner_id, tags, feature, content, redisCache)
}

// CreateNewBanner mocks base method.
func (m *MockBanners) CreateNewBanner(tags []int, feature int, content AVITO.Banner, is_active bool, redisCache cache.CacheImages) (int, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateNewBanner", tags, feature, content, is_active, redisCache)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateNewBanner indicates an expected call of CreateNewBanner.
func (mr *MockBannersMockRecorder) CreateNewBanner(tags, feature, content, is_active, redisCache interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateNewBanner", reflect.TypeOf((*MockBanners)(nil).CreateNewBanner), tags, feature, content, is_active, redisCache)
}

// DeleteBanner mocks base method.
func (m *MockBanners) DeleteBanner(banner_id int, redisCache cache.CacheImages) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteBanner", banner_id, redisCache)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteBanner indicates an expected call of DeleteBanner.
func (mr *MockBannersMockRecorder) DeleteBanner(banner_id, redisCache interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBanner", reflect.TypeOf((*MockBanners)(nil).DeleteBanner), banner_id, redisCache)
}

// GetBanner mocks base method.
func (m *MockBanners) GetBanner(tag, feature int, useLastRevision, isAdmin bool, redisCache cache.CacheImages) (string, string, string, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBanner", tag, feature, useLastRevision, isAdmin, redisCache)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(string)
	ret3, _ := ret[3].(int)
	ret4, _ := ret[4].(error)
	return ret0, ret1, ret2, ret3, ret4
}

// GetBanner indicates an expected call of GetBanner.
func (mr *MockBannersMockRecorder) GetBanner(tag, feature, useLastRevision, isAdmin, redisCache interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBanner", reflect.TypeOf((*MockBanners)(nil).GetBanner), tag, feature, useLastRevision, isAdmin, redisCache)
}
