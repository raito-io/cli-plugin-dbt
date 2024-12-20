// Code generated by mockery v2.50.0. DO NOT EDIT.

package resource_provider

import (
	context "context"

	schema "github.com/raito-io/sdk-go/types"
	mock "github.com/stretchr/testify/mock"
)

// MockUserRepo is an autogenerated mock type for the UserRepo type
type MockUserRepo struct {
	mock.Mock
}

type MockUserRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUserRepo) EXPECT() *MockUserRepo_Expecter {
	return &MockUserRepo_Expecter{mock: &_m.Mock}
}

// GetUserByEmail provides a mock function with given fields: ctx, email
func (_m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (*schema.User, error) {
	ret := _m.Called(ctx, email)

	if len(ret) == 0 {
		panic("no return value specified for GetUserByEmail")
	}

	var r0 *schema.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*schema.User, error)); ok {
		return rf(ctx, email)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *schema.User); ok {
		r0 = rf(ctx, email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*schema.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserRepo_GetUserByEmail_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUserByEmail'
type MockUserRepo_GetUserByEmail_Call struct {
	*mock.Call
}

// GetUserByEmail is a helper method to define mock.On call
//   - ctx context.Context
//   - email string
func (_e *MockUserRepo_Expecter) GetUserByEmail(ctx interface{}, email interface{}) *MockUserRepo_GetUserByEmail_Call {
	return &MockUserRepo_GetUserByEmail_Call{Call: _e.mock.On("GetUserByEmail", ctx, email)}
}

func (_c *MockUserRepo_GetUserByEmail_Call) Run(run func(ctx context.Context, email string)) *MockUserRepo_GetUserByEmail_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockUserRepo_GetUserByEmail_Call) Return(_a0 *schema.User, _a1 error) *MockUserRepo_GetUserByEmail_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserRepo_GetUserByEmail_Call) RunAndReturn(run func(context.Context, string) (*schema.User, error)) *MockUserRepo_GetUserByEmail_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockUserRepo creates a new instance of MockUserRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUserRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserRepo {
	mock := &MockUserRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
