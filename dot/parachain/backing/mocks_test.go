// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ChainSafe/gossamer/dot/parachain/backing (interfaces: Table,ImplicitView)
//
// Generated by this command:
//
//	mockgen -destination=mocks_test.go -package=backing . Table,ImplicitView
//

// Package backing is a generated GoMock package.
package backing

import (
	reflect "reflect"

	parachaintypes "github.com/ChainSafe/gossamer/dot/parachain/types"
	common "github.com/ChainSafe/gossamer/lib/common"
	gomock "go.uber.org/mock/gomock"
)

// MockTable is a mock of Table interface.
type MockTable struct {
	ctrl     *gomock.Controller
	recorder *MockTableMockRecorder
}

// MockTableMockRecorder is the mock recorder for MockTable.
type MockTableMockRecorder struct {
	mock *MockTable
}

// NewMockTable creates a new mock instance.
func NewMockTable(ctrl *gomock.Controller) *MockTable {
	mock := &MockTable{ctrl: ctrl}
	mock.recorder = &MockTableMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTable) EXPECT() *MockTableMockRecorder {
	return m.recorder
}

// attestedCandidate mocks base method.
func (m *MockTable) attestedCandidate(arg0 parachaintypes.CandidateHash, arg1 *TableContext, arg2 uint32) (*attestedCandidate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "attestedCandidate", arg0, arg1, arg2)
	ret0, _ := ret[0].(*attestedCandidate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// attestedCandidate indicates an expected call of attestedCandidate.
func (mr *MockTableMockRecorder) attestedCandidate(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "attestedCandidate", reflect.TypeOf((*MockTable)(nil).attestedCandidate), arg0, arg1, arg2)
}

// drainMisbehaviors mocks base method.
func (m *MockTable) drainMisbehaviors() []parachaintypes.ProvisionableDataMisbehaviorReport {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "drainMisbehaviors")
	ret0, _ := ret[0].([]parachaintypes.ProvisionableDataMisbehaviorReport)
	return ret0
}

// drainMisbehaviors indicates an expected call of drainMisbehaviors.
func (mr *MockTableMockRecorder) drainMisbehaviors() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "drainMisbehaviors", reflect.TypeOf((*MockTable)(nil).drainMisbehaviors))
}

// getCandidate mocks base method.
func (m *MockTable) getCandidate(arg0 parachaintypes.CandidateHash) (parachaintypes.CommittedCandidateReceipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "getCandidate", arg0)
	ret0, _ := ret[0].(parachaintypes.CommittedCandidateReceipt)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// getCandidate indicates an expected call of getCandidate.
func (mr *MockTableMockRecorder) getCandidate(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "getCandidate", reflect.TypeOf((*MockTable)(nil).getCandidate), arg0)
}

// importStatement mocks base method.
func (m *MockTable) importStatement(arg0 *TableContext, arg1 parachaintypes.SignedFullStatementWithPVD) (*Summary, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "importStatement", arg0, arg1)
	ret0, _ := ret[0].(*Summary)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// importStatement indicates an expected call of importStatement.
func (mr *MockTableMockRecorder) importStatement(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "importStatement", reflect.TypeOf((*MockTable)(nil).importStatement), arg0, arg1)
}

// MockImplicitView is a mock of ImplicitView interface.
type MockImplicitView struct {
	ctrl     *gomock.Controller
	recorder *MockImplicitViewMockRecorder
}

// MockImplicitViewMockRecorder is the mock recorder for MockImplicitView.
type MockImplicitViewMockRecorder struct {
	mock *MockImplicitView
}

// NewMockImplicitView creates a new mock instance.
func NewMockImplicitView(ctrl *gomock.Controller) *MockImplicitView {
	mock := &MockImplicitView{ctrl: ctrl}
	mock.recorder = &MockImplicitViewMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockImplicitView) EXPECT() *MockImplicitViewMockRecorder {
	return m.recorder
}

// activeLeaf mocks base method.
func (m *MockImplicitView) activeLeaf(arg0 common.Hash) ([]parachaintypes.ParaID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "activeLeaf", arg0)
	ret0, _ := ret[0].([]parachaintypes.ParaID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// activeLeaf indicates an expected call of activeLeaf.
func (mr *MockImplicitViewMockRecorder) activeLeaf(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "activeLeaf", reflect.TypeOf((*MockImplicitView)(nil).activeLeaf), arg0)
}

// allAllowedRelayParents mocks base method.
func (m *MockImplicitView) allAllowedRelayParents() []common.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "allAllowedRelayParents")
	ret0, _ := ret[0].([]common.Hash)
	return ret0
}

// allAllowedRelayParents indicates an expected call of allAllowedRelayParents.
func (mr *MockImplicitViewMockRecorder) allAllowedRelayParents() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "allAllowedRelayParents", reflect.TypeOf((*MockImplicitView)(nil).allAllowedRelayParents))
}

// deactivateLeaf mocks base method.
func (m *MockImplicitView) deactivateLeaf(arg0 common.Hash) []common.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "deactivateLeaf", arg0)
	ret0, _ := ret[0].([]common.Hash)
	return ret0
}

// deactivateLeaf indicates an expected call of deactivateLeaf.
func (mr *MockImplicitViewMockRecorder) deactivateLeaf(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "deactivateLeaf", reflect.TypeOf((*MockImplicitView)(nil).deactivateLeaf), arg0)
}

// knownAllowedRelayParentsUnder mocks base method.
func (m *MockImplicitView) knownAllowedRelayParentsUnder(arg0 common.Hash, arg1 *parachaintypes.ParaID) []common.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "knownAllowedRelayParentsUnder", arg0, arg1)
	ret0, _ := ret[0].([]common.Hash)
	return ret0
}

// knownAllowedRelayParentsUnder indicates an expected call of knownAllowedRelayParentsUnder.
func (mr *MockImplicitViewMockRecorder) knownAllowedRelayParentsUnder(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "knownAllowedRelayParentsUnder", reflect.TypeOf((*MockImplicitView)(nil).knownAllowedRelayParentsUnder), arg0, arg1)
}
