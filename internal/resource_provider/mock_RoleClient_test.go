// Code generated by mockery v2.50.0. DO NOT EDIT.

package resource_provider

import (
	context "context"

	schema "github.com/raito-io/sdk-go/types"
	mock "github.com/stretchr/testify/mock"
)

// MockRoleClient is an autogenerated mock type for the RoleClient type
type MockRoleClient struct {
	mock.Mock
}

type MockRoleClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRoleClient) EXPECT() *MockRoleClient_Expecter {
	return &MockRoleClient_Expecter{mock: &_m.Mock}
}

// UpdateRoleAssigneesOnAccessProvider provides a mock function with given fields: ctx, accessProviderId, roleId, assignees
func (_m *MockRoleClient) UpdateRoleAssigneesOnAccessProvider(ctx context.Context, accessProviderId string, roleId string, assignees ...string) (*schema.Role, error) {
	_va := make([]interface{}, len(assignees))
	for _i := range assignees {
		_va[_i] = assignees[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, accessProviderId, roleId)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpdateRoleAssigneesOnAccessProvider")
	}

	var r0 *schema.Role
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...string) (*schema.Role, error)); ok {
		return rf(ctx, accessProviderId, roleId, assignees...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...string) *schema.Role); ok {
		r0 = rf(ctx, accessProviderId, roleId, assignees...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*schema.Role)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, ...string) error); ok {
		r1 = rf(ctx, accessProviderId, roleId, assignees...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockRoleClient_UpdateRoleAssigneesOnAccessProvider_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateRoleAssigneesOnAccessProvider'
type MockRoleClient_UpdateRoleAssigneesOnAccessProvider_Call struct {
	*mock.Call
}

// UpdateRoleAssigneesOnAccessProvider is a helper method to define mock.On call
//   - ctx context.Context
//   - accessProviderId string
//   - roleId string
//   - assignees ...string
func (_e *MockRoleClient_Expecter) UpdateRoleAssigneesOnAccessProvider(ctx interface{}, accessProviderId interface{}, roleId interface{}, assignees ...interface{}) *MockRoleClient_UpdateRoleAssigneesOnAccessProvider_Call {
	return &MockRoleClient_UpdateRoleAssigneesOnAccessProvider_Call{Call: _e.mock.On("UpdateRoleAssigneesOnAccessProvider",
		append([]interface{}{ctx, accessProviderId, roleId}, assignees...)...)}
}

func (_c *MockRoleClient_UpdateRoleAssigneesOnAccessProvider_Call) Run(run func(ctx context.Context, accessProviderId string, roleId string, assignees ...string)) *MockRoleClient_UpdateRoleAssigneesOnAccessProvider_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(string), variadicArgs...)
	})
	return _c
}

func (_c *MockRoleClient_UpdateRoleAssigneesOnAccessProvider_Call) Return(_a0 *schema.Role, _a1 error) *MockRoleClient_UpdateRoleAssigneesOnAccessProvider_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockRoleClient_UpdateRoleAssigneesOnAccessProvider_Call) RunAndReturn(run func(context.Context, string, string, ...string) (*schema.Role, error)) *MockRoleClient_UpdateRoleAssigneesOnAccessProvider_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockRoleClient creates a new instance of MockRoleClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRoleClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRoleClient {
	mock := &MockRoleClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
