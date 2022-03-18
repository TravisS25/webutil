package webutiltest

import "testing"

func TestValidateObjectSlice(t *testing.T) {
	mockTestLog := &MockTestLog{}
	mockTestLog.On("Helper").Return(nil)

}
