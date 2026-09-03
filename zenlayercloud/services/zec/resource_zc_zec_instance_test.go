package zec

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestZecInstanceUserData(t *testing.T) {
	resourceSchema := ResourceZenlayerCloudZecInstance().Schema

	tests := []struct {
		name     string
		raw      map[string]interface{}
		expected *string
	}{
		{
			name:     "absent",
			raw:      map[string]interface{}{},
			expected: nil,
		},
		{
			name:     "plain text is sent as is",
			raw:      map[string]interface{}{"user_data": "#cloud-config\nhostname: node-01\n"},
			expected: stringPtr("#cloud-config\nhostname: node-01\n"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceSchema, test.raw)
			userData := zecInstanceUserData(d)

			if test.expected == nil {
				if userData != nil {
					t.Fatalf("expected no user data, got %q", *userData)
				}
				return
			}
			if userData == nil {
				t.Fatalf("expected user data %q, got none", *test.expected)
			}
			if *userData != *test.expected {
				t.Fatalf("expected user data %q, got %q", *test.expected, *userData)
			}
		})
	}
}

func TestValidateZecInstanceUserData(t *testing.T) {
	if _, errs := validateZecInstanceUserData("#cloud-config\n", "user_data"); len(errs) > 0 {
		t.Fatalf("expected no error, got %v", errs)
	}

	maxSize := strings.Repeat("a", zecInstanceUserDataMaxSize)
	if _, errs := validateZecInstanceUserData(maxSize, "user_data"); len(errs) > 0 {
		t.Fatalf("expected no error at the maximum size, got %v", errs)
	}

	tooLarge := strings.Repeat("a", zecInstanceUserDataMaxSize+1)
	if _, errs := validateZecInstanceUserData(tooLarge, "user_data"); len(errs) == 0 {
		t.Fatal("expected an error for user data which exceeds the maximum size")
	}
}

func stringPtr(value string) *string {
	return &value
}
