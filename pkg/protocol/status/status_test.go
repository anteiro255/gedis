package status_test

import (
	"strconv"
	"testing"

	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func TestStatus_ErrorStrings(t *testing.T) {
	tests := []struct {
		code status.Status
		want string
	}{
		{status.OK, "OK"},
		{status.KeyAlreadyExists, "Such key already exists"},
		{status.NoSuchKey, "No such key"},
		{status.NoTTL, "The key doesn't have TTL"},
		{status.WrongInput, "Wrong input"},
		{status.InternalError, "Internal error"},
		{status.DeadlineExceeded, "Deadline exceeded"},
		{status.Status(99), "Unknown status code: 99"},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(int(tt.code)), func(t *testing.T) {
			if got := tt.code.Error(); got != tt.want {
				t.Errorf("Status(%d).Error() = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}
