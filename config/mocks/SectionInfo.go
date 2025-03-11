// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	config "github.com/creativeprojects/resticprofile/config"
	mock "github.com/stretchr/testify/mock"

	restic "github.com/creativeprojects/resticprofile/restic"
)

// SectionInfo is an autogenerated mock type for the SectionInfo type
type SectionInfo struct {
	mock.Mock
}

type SectionInfo_Expecter struct {
	mock *mock.Mock
}

func (_m *SectionInfo) EXPECT() *SectionInfo_Expecter {
	return &SectionInfo_Expecter{mock: &_m.Mock}
}

// Command provides a mock function with no fields
func (_m *SectionInfo) Command() restic.CommandIf {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Command")
	}

	var r0 restic.CommandIf
	if rf, ok := ret.Get(0).(func() restic.CommandIf); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(restic.CommandIf)
		}
	}

	return r0
}

// SectionInfo_Command_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Command'
type SectionInfo_Command_Call struct {
	*mock.Call
}

// Command is a helper method to define mock.On call
func (_e *SectionInfo_Expecter) Command() *SectionInfo_Command_Call {
	return &SectionInfo_Command_Call{Call: _e.mock.On("Command")}
}

func (_c *SectionInfo_Command_Call) Run(run func()) *SectionInfo_Command_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SectionInfo_Command_Call) Return(_a0 restic.CommandIf) *SectionInfo_Command_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SectionInfo_Command_Call) RunAndReturn(run func() restic.CommandIf) *SectionInfo_Command_Call {
	_c.Call.Return(run)
	return _c
}

// Description provides a mock function with no fields
func (_m *SectionInfo) Description() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Description")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SectionInfo_Description_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Description'
type SectionInfo_Description_Call struct {
	*mock.Call
}

// Description is a helper method to define mock.On call
func (_e *SectionInfo_Expecter) Description() *SectionInfo_Description_Call {
	return &SectionInfo_Description_Call{Call: _e.mock.On("Description")}
}

func (_c *SectionInfo_Description_Call) Run(run func()) *SectionInfo_Description_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SectionInfo_Description_Call) Return(_a0 string) *SectionInfo_Description_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SectionInfo_Description_Call) RunAndReturn(run func() string) *SectionInfo_Description_Call {
	_c.Call.Return(run)
	return _c
}

// IsAllOptions provides a mock function with no fields
func (_m *SectionInfo) IsAllOptions() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsAllOptions")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// SectionInfo_IsAllOptions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsAllOptions'
type SectionInfo_IsAllOptions_Call struct {
	*mock.Call
}

// IsAllOptions is a helper method to define mock.On call
func (_e *SectionInfo_Expecter) IsAllOptions() *SectionInfo_IsAllOptions_Call {
	return &SectionInfo_IsAllOptions_Call{Call: _e.mock.On("IsAllOptions")}
}

func (_c *SectionInfo_IsAllOptions_Call) Run(run func()) *SectionInfo_IsAllOptions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SectionInfo_IsAllOptions_Call) Return(_a0 bool) *SectionInfo_IsAllOptions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SectionInfo_IsAllOptions_Call) RunAndReturn(run func() bool) *SectionInfo_IsAllOptions_Call {
	_c.Call.Return(run)
	return _c
}

// IsClosed provides a mock function with no fields
func (_m *SectionInfo) IsClosed() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsClosed")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// SectionInfo_IsClosed_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsClosed'
type SectionInfo_IsClosed_Call struct {
	*mock.Call
}

// IsClosed is a helper method to define mock.On call
func (_e *SectionInfo_Expecter) IsClosed() *SectionInfo_IsClosed_Call {
	return &SectionInfo_IsClosed_Call{Call: _e.mock.On("IsClosed")}
}

func (_c *SectionInfo_IsClosed_Call) Run(run func()) *SectionInfo_IsClosed_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SectionInfo_IsClosed_Call) Return(_a0 bool) *SectionInfo_IsClosed_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SectionInfo_IsClosed_Call) RunAndReturn(run func() bool) *SectionInfo_IsClosed_Call {
	_c.Call.Return(run)
	return _c
}

// IsCommandSection provides a mock function with no fields
func (_m *SectionInfo) IsCommandSection() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsCommandSection")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// SectionInfo_IsCommandSection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsCommandSection'
type SectionInfo_IsCommandSection_Call struct {
	*mock.Call
}

// IsCommandSection is a helper method to define mock.On call
func (_e *SectionInfo_Expecter) IsCommandSection() *SectionInfo_IsCommandSection_Call {
	return &SectionInfo_IsCommandSection_Call{Call: _e.mock.On("IsCommandSection")}
}

func (_c *SectionInfo_IsCommandSection_Call) Run(run func()) *SectionInfo_IsCommandSection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SectionInfo_IsCommandSection_Call) Return(_a0 bool) *SectionInfo_IsCommandSection_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SectionInfo_IsCommandSection_Call) RunAndReturn(run func() bool) *SectionInfo_IsCommandSection_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with no fields
func (_m *SectionInfo) Name() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Name")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SectionInfo_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type SectionInfo_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *SectionInfo_Expecter) Name() *SectionInfo_Name_Call {
	return &SectionInfo_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *SectionInfo_Name_Call) Run(run func()) *SectionInfo_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SectionInfo_Name_Call) Return(_a0 string) *SectionInfo_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SectionInfo_Name_Call) RunAndReturn(run func() string) *SectionInfo_Name_Call {
	_c.Call.Return(run)
	return _c
}

// OtherPropertyInfo provides a mock function with no fields
func (_m *SectionInfo) OtherPropertyInfo() config.PropertyInfo {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for OtherPropertyInfo")
	}

	var r0 config.PropertyInfo
	if rf, ok := ret.Get(0).(func() config.PropertyInfo); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(config.PropertyInfo)
		}
	}

	return r0
}

// SectionInfo_OtherPropertyInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OtherPropertyInfo'
type SectionInfo_OtherPropertyInfo_Call struct {
	*mock.Call
}

// OtherPropertyInfo is a helper method to define mock.On call
func (_e *SectionInfo_Expecter) OtherPropertyInfo() *SectionInfo_OtherPropertyInfo_Call {
	return &SectionInfo_OtherPropertyInfo_Call{Call: _e.mock.On("OtherPropertyInfo")}
}

func (_c *SectionInfo_OtherPropertyInfo_Call) Run(run func()) *SectionInfo_OtherPropertyInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SectionInfo_OtherPropertyInfo_Call) Return(_a0 config.PropertyInfo) *SectionInfo_OtherPropertyInfo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SectionInfo_OtherPropertyInfo_Call) RunAndReturn(run func() config.PropertyInfo) *SectionInfo_OtherPropertyInfo_Call {
	_c.Call.Return(run)
	return _c
}

// Properties provides a mock function with no fields
func (_m *SectionInfo) Properties() []string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Properties")
	}

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// SectionInfo_Properties_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Properties'
type SectionInfo_Properties_Call struct {
	*mock.Call
}

// Properties is a helper method to define mock.On call
func (_e *SectionInfo_Expecter) Properties() *SectionInfo_Properties_Call {
	return &SectionInfo_Properties_Call{Call: _e.mock.On("Properties")}
}

func (_c *SectionInfo_Properties_Call) Run(run func()) *SectionInfo_Properties_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SectionInfo_Properties_Call) Return(_a0 []string) *SectionInfo_Properties_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SectionInfo_Properties_Call) RunAndReturn(run func() []string) *SectionInfo_Properties_Call {
	_c.Call.Return(run)
	return _c
}

// PropertyInfo provides a mock function with given fields: name
func (_m *SectionInfo) PropertyInfo(name string) config.PropertyInfo {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for PropertyInfo")
	}

	var r0 config.PropertyInfo
	if rf, ok := ret.Get(0).(func(string) config.PropertyInfo); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(config.PropertyInfo)
		}
	}

	return r0
}

// SectionInfo_PropertyInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PropertyInfo'
type SectionInfo_PropertyInfo_Call struct {
	*mock.Call
}

// PropertyInfo is a helper method to define mock.On call
//   - name string
func (_e *SectionInfo_Expecter) PropertyInfo(name interface{}) *SectionInfo_PropertyInfo_Call {
	return &SectionInfo_PropertyInfo_Call{Call: _e.mock.On("PropertyInfo", name)}
}

func (_c *SectionInfo_PropertyInfo_Call) Run(run func(name string)) *SectionInfo_PropertyInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *SectionInfo_PropertyInfo_Call) Return(_a0 config.PropertyInfo) *SectionInfo_PropertyInfo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SectionInfo_PropertyInfo_Call) RunAndReturn(run func(string) config.PropertyInfo) *SectionInfo_PropertyInfo_Call {
	_c.Call.Return(run)
	return _c
}

// TypeName provides a mock function with no fields
func (_m *SectionInfo) TypeName() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for TypeName")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SectionInfo_TypeName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TypeName'
type SectionInfo_TypeName_Call struct {
	*mock.Call
}

// TypeName is a helper method to define mock.On call
func (_e *SectionInfo_Expecter) TypeName() *SectionInfo_TypeName_Call {
	return &SectionInfo_TypeName_Call{Call: _e.mock.On("TypeName")}
}

func (_c *SectionInfo_TypeName_Call) Run(run func()) *SectionInfo_TypeName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SectionInfo_TypeName_Call) Return(_a0 string) *SectionInfo_TypeName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SectionInfo_TypeName_Call) RunAndReturn(run func() string) *SectionInfo_TypeName_Call {
	_c.Call.Return(run)
	return _c
}

// NewSectionInfo creates a new instance of SectionInfo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSectionInfo(t interface {
	mock.TestingT
	Cleanup(func())
}) *SectionInfo {
	mock := &SectionInfo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
