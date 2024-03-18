package zenlayercloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net"
)

// validateCIDRNetworkAddress ensures that the string value is a valid CIDR that
// represents a network address - it adds an error otherwise
func validateCIDRNetworkAddress(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, ipnet, err := net.ParseCIDR(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q must contain a valid CIDR, got error parsing: %s", k, err))
		return
	}
	if ipnet == nil || value != ipnet.String() {
		errors = append(errors, fmt.Errorf("%q must contain a valid network CIDR, expected %q, got %q", k, ipnet, value))
	}
	return
}

func validateSizeEqual(size int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		value := i.([]interface{})

		length := len(value)
		if length != size {
			errors = append(errors, fmt.Errorf("expected length of %s to be %d, got %s", k, size, length))
		}
		return warnings, errors
	}
}

func validateSizeAtLeast(size int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		value := i.([]interface{})

		length := len(value)
		if length < size {
			errors = append(errors, fmt.Errorf("expected length of %s to be greather or equal than %d, got %s", k, size, length))
		}
		return warnings, errors
	}
}
