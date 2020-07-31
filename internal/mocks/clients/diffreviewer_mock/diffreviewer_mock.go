// Code generated by MockGen. DO NOT EDIT.
// Source: ../../clients/diffreviewer/diff_reviewer.go

// Package diffreviewer_mock is a generated GoMock package.
package diffreviewer_mock

import (
	gomock "github.com/golang/mock/gomock"
	diffreviewer "github.com/nightfallai/jenkins_test/internal/clients/diffreviewer"
	logger "github.com/nightfallai/jenkins_test/internal/clients/logger"
	nightfallconfig "github.com/nightfallai/jenkins_test/internal/nightfallconfig"
	reflect "reflect"
)

// DiffReviewer is a mock of DiffReviewer interface
type DiffReviewer struct {
	ctrl     *gomock.Controller
	recorder *DiffReviewerMockRecorder
}

// DiffReviewerMockRecorder is the mock recorder for DiffReviewer
type DiffReviewerMockRecorder struct {
	mock *DiffReviewer
}

// NewDiffReviewer creates a new mock instance
func NewDiffReviewer(ctrl *gomock.Controller) *DiffReviewer {
	mock := &DiffReviewer{ctrl: ctrl}
	mock.recorder = &DiffReviewerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *DiffReviewer) EXPECT() *DiffReviewerMockRecorder {
	return m.recorder
}

// LoadConfig mocks base method
func (m *DiffReviewer) LoadConfig(nightfallConfigFileName string) (*nightfallconfig.Config, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadConfig", nightfallConfigFileName)
	ret0, _ := ret[0].(*nightfallconfig.Config)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoadConfig indicates an expected call of LoadConfig
func (mr *DiffReviewerMockRecorder) LoadConfig(nightfallConfigFileName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadConfig", reflect.TypeOf((*DiffReviewer)(nil).LoadConfig), nightfallConfigFileName)
}

// GetDiff mocks base method
func (m *DiffReviewer) GetDiff() ([]*diffreviewer.FileDiff, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDiff")
	ret0, _ := ret[0].([]*diffreviewer.FileDiff)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDiff indicates an expected call of GetDiff
func (mr *DiffReviewerMockRecorder) GetDiff() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDiff", reflect.TypeOf((*DiffReviewer)(nil).GetDiff))
}

// WriteComments mocks base method
func (m *DiffReviewer) WriteComments(comments []*diffreviewer.Comment) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteComments", comments)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteComments indicates an expected call of WriteComments
func (mr *DiffReviewerMockRecorder) WriteComments(comments interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteComments", reflect.TypeOf((*DiffReviewer)(nil).WriteComments), comments)
}

// GetLogger mocks base method
func (m *DiffReviewer) GetLogger() logger.Logger {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLogger")
	ret0, _ := ret[0].(logger.Logger)
	return ret0
}

// GetLogger indicates an expected call of GetLogger
func (mr *DiffReviewerMockRecorder) GetLogger() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLogger", reflect.TypeOf((*DiffReviewer)(nil).GetLogger))
}
