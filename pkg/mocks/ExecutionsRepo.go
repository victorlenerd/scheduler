// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	models "scheduler0/pkg/models"

	mock "github.com/stretchr/testify/mock"
)

// ExecutionsRepo is an autogenerated mock type for the ExecutionsRepo type
type ExecutionsRepo struct {
	mock.Mock
}

// BatchInsert provides a mock function with given fields: jobs, nodeId, state, jobQueueVersion, executionVersions
func (_m *ExecutionsRepo) BatchInsert(jobs []models.Job, nodeId uint64, state models.JobExecutionLogState, jobQueueVersion uint64, executionVersions map[uint64]uint64) {
	_m.Called(jobs, nodeId, state, jobQueueVersion, executionVersions)
}

// CountExecutionLogs provides a mock function with given fields: committed
func (_m *ExecutionsRepo) CountExecutionLogs(committed bool) uint64 {
	ret := _m.Called(committed)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(bool) uint64); ok {
		r0 = rf(committed)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// CountLastFailedExecutionLogs provides a mock function with given fields: jobId, nodeId, executionVersion
func (_m *ExecutionsRepo) CountLastFailedExecutionLogs(jobId uint64, nodeId uint64, executionVersion uint64) uint64 {
	ret := _m.Called(jobId, nodeId, executionVersion)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(uint64, uint64, uint64) uint64); ok {
		r0 = rf(jobId, nodeId, executionVersion)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetLastExecutionLogForJobIds provides a mock function with given fields: jobIds
func (_m *ExecutionsRepo) GetLastExecutionLogForJobIds(jobIds []uint64) map[uint64]models.JobExecutionLog {
	ret := _m.Called(jobIds)

	var r0 map[uint64]models.JobExecutionLog
	if rf, ok := ret.Get(0).(func([]uint64) map[uint64]models.JobExecutionLog); ok {
		r0 = rf(jobIds)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[uint64]models.JobExecutionLog)
		}
	}

	return r0
}

// GetUncommittedExecutionsLogForNode provides a mock function with given fields: nodeId
func (_m *ExecutionsRepo) GetUncommittedExecutionsLogForNode(nodeId uint64) []models.JobExecutionLog {
	ret := _m.Called(nodeId)

	var r0 []models.JobExecutionLog
	if rf, ok := ret.Get(0).(func(uint64) []models.JobExecutionLog); ok {
		r0 = rf(nodeId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.JobExecutionLog)
		}
	}

	return r0
}

type mockConstructorTestingTNewExecutionsRepo interface {
	mock.TestingT
	Cleanup(func())
}

// NewExecutionsRepo creates a new instance of ExecutionsRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewExecutionsRepo(t mockConstructorTestingTNewExecutionsRepo) *ExecutionsRepo {
	mock := &ExecutionsRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
