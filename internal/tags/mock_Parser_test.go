// Code generated by mockery v2.50.0. DO NOT EDIT.

package tags

import (
	manifest "github.com/raito-io/cli-plugin-dbt/internal/manifest"
	mock "github.com/stretchr/testify/mock"
)

// MockParser is an autogenerated mock type for the Parser type
type MockParser struct {
	mock.Mock
}

type MockParser_Expecter struct {
	mock *mock.Mock
}

func (_m *MockParser) EXPECT() *MockParser_Expecter {
	return &MockParser_Expecter{mock: &_m.Mock}
}

// LoadManifest provides a mock function with given fields: path
func (_m *MockParser) LoadManifest(path string) (*manifest.Manifest, error) {
	ret := _m.Called(path)

	if len(ret) == 0 {
		panic("no return value specified for LoadManifest")
	}

	var r0 *manifest.Manifest
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*manifest.Manifest, error)); ok {
		return rf(path)
	}
	if rf, ok := ret.Get(0).(func(string) *manifest.Manifest); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*manifest.Manifest)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockParser_LoadManifest_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LoadManifest'
type MockParser_LoadManifest_Call struct {
	*mock.Call
}

// LoadManifest is a helper method to define mock.On call
//   - path string
func (_e *MockParser_Expecter) LoadManifest(path interface{}) *MockParser_LoadManifest_Call {
	return &MockParser_LoadManifest_Call{Call: _e.mock.On("LoadManifest", path)}
}

func (_c *MockParser_LoadManifest_Call) Run(run func(path string)) *MockParser_LoadManifest_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockParser_LoadManifest_Call) Return(_a0 *manifest.Manifest, _a1 error) *MockParser_LoadManifest_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockParser_LoadManifest_Call) RunAndReturn(run func(string) (*manifest.Manifest, error)) *MockParser_LoadManifest_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockParser creates a new instance of MockParser. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockParser(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockParser {
	mock := &MockParser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
