// Code generated by mockery v2.42.0. DO NOT EDIT.

package resource_provider

import (
	context "context"

	schema "github.com/raito-io/sdk/types"
	mock "github.com/stretchr/testify/mock"

	services "github.com/raito-io/sdk/services"
)

// mockAccessProviderClient is an autogenerated mock type for the accessProviderClient type
type mockAccessProviderClient struct {
	mock.Mock
}

type mockAccessProviderClient_Expecter struct {
	mock *mock.Mock
}

func (_m *mockAccessProviderClient) EXPECT() *mockAccessProviderClient_Expecter {
	return &mockAccessProviderClient_Expecter{mock: &_m.Mock}
}

// CreateAccessProvider provides a mock function with given fields: ctx, ap
func (_m *mockAccessProviderClient) CreateAccessProvider(ctx context.Context, ap schema.AccessProviderInput) (*schema.AccessProvider, error) {
	ret := _m.Called(ctx, ap)

	if len(ret) == 0 {
		panic("no return value specified for CreateAccessProvider")
	}

	var r0 *schema.AccessProvider
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, schema.AccessProviderInput) (*schema.AccessProvider, error)); ok {
		return rf(ctx, ap)
	}
	if rf, ok := ret.Get(0).(func(context.Context, schema.AccessProviderInput) *schema.AccessProvider); ok {
		r0 = rf(ctx, ap)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*schema.AccessProvider)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, schema.AccessProviderInput) error); ok {
		r1 = rf(ctx, ap)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockAccessProviderClient_CreateAccessProvider_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateAccessProvider'
type mockAccessProviderClient_CreateAccessProvider_Call struct {
	*mock.Call
}

// CreateAccessProvider is a helper method to define mock.On call
//   - ctx context.Context
//   - ap schema.AccessProviderInput
func (_e *mockAccessProviderClient_Expecter) CreateAccessProvider(ctx interface{}, ap interface{}) *mockAccessProviderClient_CreateAccessProvider_Call {
	return &mockAccessProviderClient_CreateAccessProvider_Call{Call: _e.mock.On("CreateAccessProvider", ctx, ap)}
}

func (_c *mockAccessProviderClient_CreateAccessProvider_Call) Run(run func(ctx context.Context, ap schema.AccessProviderInput)) *mockAccessProviderClient_CreateAccessProvider_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(schema.AccessProviderInput))
	})
	return _c
}

func (_c *mockAccessProviderClient_CreateAccessProvider_Call) Return(_a0 *schema.AccessProvider, _a1 error) *mockAccessProviderClient_CreateAccessProvider_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockAccessProviderClient_CreateAccessProvider_Call) RunAndReturn(run func(context.Context, schema.AccessProviderInput) (*schema.AccessProvider, error)) *mockAccessProviderClient_CreateAccessProvider_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteAccessProvider provides a mock function with given fields: ctx, id, ops
func (_m *mockAccessProviderClient) DeleteAccessProvider(ctx context.Context, id string, ops ...func(*services.UpdateAccessProviderOptions)) error {
	_va := make([]interface{}, len(ops))
	for _i := range ops {
		_va[_i] = ops[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, id)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteAccessProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...func(*services.UpdateAccessProviderOptions)) error); ok {
		r0 = rf(ctx, id, ops...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockAccessProviderClient_DeleteAccessProvider_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteAccessProvider'
type mockAccessProviderClient_DeleteAccessProvider_Call struct {
	*mock.Call
}

// DeleteAccessProvider is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - ops ...func(*services.UpdateAccessProviderOptions)
func (_e *mockAccessProviderClient_Expecter) DeleteAccessProvider(ctx interface{}, id interface{}, ops ...interface{}) *mockAccessProviderClient_DeleteAccessProvider_Call {
	return &mockAccessProviderClient_DeleteAccessProvider_Call{Call: _e.mock.On("DeleteAccessProvider",
		append([]interface{}{ctx, id}, ops...)...)}
}

func (_c *mockAccessProviderClient_DeleteAccessProvider_Call) Run(run func(ctx context.Context, id string, ops ...func(*services.UpdateAccessProviderOptions))) *mockAccessProviderClient_DeleteAccessProvider_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*services.UpdateAccessProviderOptions), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*services.UpdateAccessProviderOptions))
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *mockAccessProviderClient_DeleteAccessProvider_Call) Return(_a0 error) *mockAccessProviderClient_DeleteAccessProvider_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockAccessProviderClient_DeleteAccessProvider_Call) RunAndReturn(run func(context.Context, string, ...func(*services.UpdateAccessProviderOptions)) error) *mockAccessProviderClient_DeleteAccessProvider_Call {
	_c.Call.Return(run)
	return _c
}

// ListAccessProviders provides a mock function with given fields: ctx, ops
func (_m *mockAccessProviderClient) ListAccessProviders(ctx context.Context, ops ...func(*services.AccessProviderListOptions)) <-chan schema.ListItem[schema.AccessProvider] {
	_va := make([]interface{}, len(ops))
	for _i := range ops {
		_va[_i] = ops[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListAccessProviders")
	}

	var r0 <-chan schema.ListItem[schema.AccessProvider]
	if rf, ok := ret.Get(0).(func(context.Context, ...func(*services.AccessProviderListOptions)) <-chan schema.ListItem[schema.AccessProvider]); ok {
		r0 = rf(ctx, ops...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan schema.ListItem[schema.AccessProvider])
		}
	}

	return r0
}

// mockAccessProviderClient_ListAccessProviders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListAccessProviders'
type mockAccessProviderClient_ListAccessProviders_Call struct {
	*mock.Call
}

// ListAccessProviders is a helper method to define mock.On call
//   - ctx context.Context
//   - ops ...func(*services.AccessProviderListOptions)
func (_e *mockAccessProviderClient_Expecter) ListAccessProviders(ctx interface{}, ops ...interface{}) *mockAccessProviderClient_ListAccessProviders_Call {
	return &mockAccessProviderClient_ListAccessProviders_Call{Call: _e.mock.On("ListAccessProviders",
		append([]interface{}{ctx}, ops...)...)}
}

func (_c *mockAccessProviderClient_ListAccessProviders_Call) Run(run func(ctx context.Context, ops ...func(*services.AccessProviderListOptions))) *mockAccessProviderClient_ListAccessProviders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*services.AccessProviderListOptions), len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(func(*services.AccessProviderListOptions))
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *mockAccessProviderClient_ListAccessProviders_Call) Return(_a0 <-chan schema.ListItem[schema.AccessProvider]) *mockAccessProviderClient_ListAccessProviders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockAccessProviderClient_ListAccessProviders_Call) RunAndReturn(run func(context.Context, ...func(*services.AccessProviderListOptions)) <-chan schema.ListItem[schema.AccessProvider]) *mockAccessProviderClient_ListAccessProviders_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateAccessProvider provides a mock function with given fields: ctx, id, ap, ops
func (_m *mockAccessProviderClient) UpdateAccessProvider(ctx context.Context, id string, ap schema.AccessProviderInput, ops ...func(*services.UpdateAccessProviderOptions)) (*schema.AccessProvider, error) {
	_va := make([]interface{}, len(ops))
	for _i := range ops {
		_va[_i] = ops[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, id, ap)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpdateAccessProvider")
	}

	var r0 *schema.AccessProvider
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, schema.AccessProviderInput, ...func(*services.UpdateAccessProviderOptions)) (*schema.AccessProvider, error)); ok {
		return rf(ctx, id, ap, ops...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, schema.AccessProviderInput, ...func(*services.UpdateAccessProviderOptions)) *schema.AccessProvider); ok {
		r0 = rf(ctx, id, ap, ops...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*schema.AccessProvider)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, schema.AccessProviderInput, ...func(*services.UpdateAccessProviderOptions)) error); ok {
		r1 = rf(ctx, id, ap, ops...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockAccessProviderClient_UpdateAccessProvider_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateAccessProvider'
type mockAccessProviderClient_UpdateAccessProvider_Call struct {
	*mock.Call
}

// UpdateAccessProvider is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - ap schema.AccessProviderInput
//   - ops ...func(*services.UpdateAccessProviderOptions)
func (_e *mockAccessProviderClient_Expecter) UpdateAccessProvider(ctx interface{}, id interface{}, ap interface{}, ops ...interface{}) *mockAccessProviderClient_UpdateAccessProvider_Call {
	return &mockAccessProviderClient_UpdateAccessProvider_Call{Call: _e.mock.On("UpdateAccessProvider",
		append([]interface{}{ctx, id, ap}, ops...)...)}
}

func (_c *mockAccessProviderClient_UpdateAccessProvider_Call) Run(run func(ctx context.Context, id string, ap schema.AccessProviderInput, ops ...func(*services.UpdateAccessProviderOptions))) *mockAccessProviderClient_UpdateAccessProvider_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*services.UpdateAccessProviderOptions), len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(func(*services.UpdateAccessProviderOptions))
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(schema.AccessProviderInput), variadicArgs...)
	})
	return _c
}

func (_c *mockAccessProviderClient_UpdateAccessProvider_Call) Return(_a0 *schema.AccessProvider, _a1 error) *mockAccessProviderClient_UpdateAccessProvider_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockAccessProviderClient_UpdateAccessProvider_Call) RunAndReturn(run func(context.Context, string, schema.AccessProviderInput, ...func(*services.UpdateAccessProviderOptions)) (*schema.AccessProvider, error)) *mockAccessProviderClient_UpdateAccessProvider_Call {
	_c.Call.Return(run)
	return _c
}

// newMockAccessProviderClient creates a new instance of mockAccessProviderClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockAccessProviderClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockAccessProviderClient {
	mock := &mockAccessProviderClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}