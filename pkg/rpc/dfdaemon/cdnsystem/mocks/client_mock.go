// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/rpc/cdnsystem/client/client.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	dfnet "d7y.io/dragonfly/v2/pkg/dfnet"
	base "d7y.io/dragonfly/v2/pkg/rpc/base"
	cdnsystem "d7y.io/dragonfly/v2/pkg/rpc/cdnsystem"
	client "d7y.io/dragonfly/v2/pkg/rpc/cdnsystem/client"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockCdnClient is a mock of CdnClient interface.
type MockCdnClient struct {
	ctrl     *gomock.Controller
	recorder *MockCdnClientMockRecorder
}

// MockCdnClientMockRecorder is the mock recorder for MockCdnClient.
type MockCdnClientMockRecorder struct {
	mock *MockCdnClient
}

// NewMockCdnClient creates a new mock instance.
func NewMockCdnClient(ctrl *gomock.Controller) *MockCdnClient {
	mock := &MockCdnClient{ctrl: ctrl}
	mock.recorder = &MockCdnClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCdnClient) EXPECT() *MockCdnClientMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockCdnClient) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockCdnClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockCdnClient)(nil).Close))
}

// GetPieceTasks mocks base method.
func (m *MockCdnClient) GetPieceTasks(ctx context.Context, addr dfnet.NetAddr, req *base.PieceTaskRequest, opts ...grpc.CallOption) (*base.PiecePacket, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, addr, req}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetPieceTasks", varargs...)
	ret0, _ := ret[0].(*base.PiecePacket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPieceTasks indicates an expected call of GetPieceTasks.
func (mr *MockCdnClientMockRecorder) GetPieceTasks(ctx, addr, req interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, addr, req}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPieceTasks", reflect.TypeOf((*MockCdnClient)(nil).GetPieceTasks), varargs...)
}

// ObtainSeeds mocks base method.
func (m *MockCdnClient) ObtainSeeds(ctx context.Context, sr *cdnsystem.SeedRequest, opts ...grpc.CallOption) (*client.PieceSeedStream, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, sr}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ObtainSeeds", varargs...)
	ret0, _ := ret[0].(*client.PieceSeedStream)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObtainSeeds indicates an expected call of ObtainSeeds.
func (mr *MockCdnClientMockRecorder) ObtainSeeds(ctx, sr interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, sr}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObtainSeeds", reflect.TypeOf((*MockCdnClient)(nil).ObtainSeeds), varargs...)
}

// UpdateState mocks base method.
func (m *MockCdnClient) UpdateState(addrs []dfnet.NetAddr) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UpdateState", addrs)
}

// UpdateState indicates an expected call of UpdateState.
func (mr *MockCdnClientMockRecorder) UpdateState(addrs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateState", reflect.TypeOf((*MockCdnClient)(nil).UpdateState), addrs)
}
