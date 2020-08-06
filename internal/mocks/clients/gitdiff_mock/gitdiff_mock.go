// Code generated by MockGen. DO NOT EDIT.
// Source: ../gitdiffintf/gitdiff.go

// Package gitdiff_mock is a generated GoMock package.
package gitdiff_mock

import (
	gomock "github.com/golang/mock/gomock"
	diffreviewer "github.com/nightfallai/jenkins_test/internal/clients/diffreviewer"
	gitdiff "github.com/nightfallai/jenkins_test/internal/clients/gitdiff"
	reflect "reflect"
)

// GitDiff is a mock of GitDiff interface
type GitDiff struct {
	ctrl     *gomock.Controller
	recorder *GitDiffMockRecorder
}

// GitDiffMockRecorder is the mock recorder for GitDiff
type GitDiffMockRecorder struct {
	mock *GitDiff
}

// NewGitDiff creates a new mock instance
func NewGitDiff(ctrl *gomock.Controller) *GitDiff {
	mock := &GitDiff{ctrl: ctrl}
	mock.recorder = &GitDiffMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *GitDiff) EXPECT() *GitDiffMockRecorder {
	return m.recorder
}

// GetDiff mocks base method
func (m *GitDiff) GetDiff(baseRev, headRev, repoURL string, diffOpts *gitdiff.DiffOptions) ([]*diffreviewer.FileDiff, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDiff", baseRev, headRev, repoURL, diffOpts)
	ret0, _ := ret[0].([]*diffreviewer.FileDiff)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDiff indicates an expected call of GetDiff
func (mr *GitDiffMockRecorder) GetDiff(baseRev, headRev, repoURL, diffOpts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDiff", reflect.TypeOf((*GitDiff)(nil).GetDiff), baseRev, headRev, repoURL, diffOpts)
}