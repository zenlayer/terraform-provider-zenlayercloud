package zenlayercloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/pkg/errors"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	"log"
	"strings"
)

const (
	ServiceNotAvailable = "SERVICE_TEMPORARY_UNAVAILABLE"
	InternalServerError = "INTERNAL_SERVER_ERROR"
	ReadTimedOut        = "REQUEST_TIMED_OUT"
	ResourceNotFound    = "OPERATION_FAILED_RESOURCE_NOT_FOUND"
)

var retryableErrorCode = []string{
	// client
	ServiceNotAvailable,
}

func Error(msg string, args ...interface{}) error {
	return fmt.Errorf(msg, args...)
}

func retryError(ctx context.Context, err error, additionRetryableError ...string) *resource.RetryError {
	switch realErr := errors.Cause(err).(type) {
	case *common.ZenlayerCloudSdkError:
		if isExpectError(realErr, retryableErrorCode) {
			tflog.Info(ctx, "Retryable defined error:", map[string]interface{}{
				"err": err,
			})
			return resource.RetryableError(err)
		}

		if len(additionRetryableError) > 0 {
			if isExpectError(realErr, additionRetryableError) {
				tflog.Info(ctx, "Retryable addition error:", map[string]interface{}{
					"err": err,
				})
				return resource.RetryableError(err)
			}
		}
	default:
	}

	log.Printf("[CRITAL] NonRetryable error: %v", err)
	return resource.NonRetryableError(err)
}

func isExpectError(err error, expectError []string) bool {
	e, ok := err.(*common.ZenlayerCloudSdkError)
	if !ok {
		return false
	}

	longCode := e.Code
	if IsContains(expectError, longCode) {
		return true
	}

	if strings.Contains(longCode, ".") {
		shortCode := strings.Split(longCode, ".")[0]
		if IsContains(expectError, shortCode) {
			return true
		}
	}

	return false
}
