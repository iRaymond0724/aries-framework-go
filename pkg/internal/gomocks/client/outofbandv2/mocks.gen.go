// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/hyperledger/aries-framework-go/pkg/client/outofbandv2 (interfaces: Provider,OobService)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	outofbandv2 "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofbandv2"
	kms "github.com/hyperledger/aries-framework-go/pkg/kms"
)

// MockProvider is a mock of Provider interface.
type MockProvider struct {
	ctrl     *gomock.Controller
	recorder *MockProviderMockRecorder
}

// MockProviderMockRecorder is the mock recorder for MockProvider.
type MockProviderMockRecorder struct {
	mock *MockProvider
}

// NewMockProvider creates a new mock instance.
func NewMockProvider(ctrl *gomock.Controller) *MockProvider {
	mock := &MockProvider{ctrl: ctrl}
	mock.recorder = &MockProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProvider) EXPECT() *MockProviderMockRecorder {
	return m.recorder
}

// KMS mocks base method.
func (m *MockProvider) KMS() kms.KeyManager {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KMS")
	ret0, _ := ret[0].(kms.KeyManager)
	return ret0
}

// KMS indicates an expected call of KMS.
func (mr *MockProviderMockRecorder) KMS() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KMS", reflect.TypeOf((*MockProvider)(nil).KMS))
}

// KeyAgreementType mocks base method.
func (m *MockProvider) KeyAgreementType() kms.KeyType {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KeyAgreementType")
	ret0, _ := ret[0].(kms.KeyType)
	return ret0
}

// KeyAgreementType indicates an expected call of KeyAgreementType.
func (mr *MockProviderMockRecorder) KeyAgreementType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KeyAgreementType", reflect.TypeOf((*MockProvider)(nil).KeyAgreementType))
}

// KeyType mocks base method.
func (m *MockProvider) KeyType() kms.KeyType {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KeyType")
	ret0, _ := ret[0].(kms.KeyType)
	return ret0
}

// KeyType indicates an expected call of KeyType.
func (mr *MockProviderMockRecorder) KeyType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KeyType", reflect.TypeOf((*MockProvider)(nil).KeyType))
}

// MediaTypeProfiles mocks base method.
func (m *MockProvider) MediaTypeProfiles() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MediaTypeProfiles")
	ret0, _ := ret[0].([]string)
	return ret0
}

// MediaTypeProfiles indicates an expected call of MediaTypeProfiles.
func (mr *MockProviderMockRecorder) MediaTypeProfiles() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MediaTypeProfiles", reflect.TypeOf((*MockProvider)(nil).MediaTypeProfiles))
}

// Service mocks base method.
func (m *MockProvider) Service(arg0 string) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Service", arg0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Service indicates an expected call of Service.
func (mr *MockProviderMockRecorder) Service(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Service", reflect.TypeOf((*MockProvider)(nil).Service), arg0)
}

// ServiceEndpoint mocks base method.
func (m *MockProvider) ServiceEndpoint() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServiceEndpoint")
	ret0, _ := ret[0].(string)
	return ret0
}

// ServiceEndpoint indicates an expected call of ServiceEndpoint.
func (mr *MockProviderMockRecorder) ServiceEndpoint() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServiceEndpoint", reflect.TypeOf((*MockProvider)(nil).ServiceEndpoint))
}

// MockOobService is a mock of OobService interface.
type MockOobService struct {
	ctrl     *gomock.Controller
	recorder *MockOobServiceMockRecorder
}

// MockOobServiceMockRecorder is the mock recorder for MockOobService.
type MockOobServiceMockRecorder struct {
	mock *MockOobService
}

// NewMockOobService creates a new mock instance.
func NewMockOobService(ctrl *gomock.Controller) *MockOobService {
	mock := &MockOobService{ctrl: ctrl}
	mock.recorder = &MockOobServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOobService) EXPECT() *MockOobServiceMockRecorder {
	return m.recorder
}

// AcceptInvitation mocks base method.
func (m *MockOobService) AcceptInvitation(arg0 *outofbandv2.Invitation) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AcceptInvitation", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AcceptInvitation indicates an expected call of AcceptInvitation.
func (mr *MockOobServiceMockRecorder) AcceptInvitation(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AcceptInvitation", reflect.TypeOf((*MockOobService)(nil).AcceptInvitation), arg0)
}

// SaveInvitation mocks base method.
func (m *MockOobService) SaveInvitation(arg0 *outofbandv2.Invitation) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveInvitation", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveInvitation indicates an expected call of SaveInvitation.
func (mr *MockOobServiceMockRecorder) SaveInvitation(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveInvitation", reflect.TypeOf((*MockOobService)(nil).SaveInvitation), arg0)
}