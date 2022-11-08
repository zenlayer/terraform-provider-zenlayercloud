package zenlayercloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"hash/crc32"
	"io/ioutil"
	"os"
	"os/user"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	PROVIDER_WRITE_RETRY_TIMEOUT = "ZENLAYERCLOUD_WRITE_RETRY_TIMEOUT"
	PROVIDER_READ_RETRY_TIMEOUT  = "ZENLAYERCLOUD_READ_RETRY_TIMEOUT"
	PROVIDER_BMC_CREATE_TIMEOUT  = "ZENLAYERCLOUD_BMC_CREATE_TIMEOUT"
	PROVIDER_BMC_UPDATE_TIMEOUT  = "ZENLAYERCLOUD_BMC_UPDATE_TIMEOUT"
)

var writeRetry = getEnvDefault(PROVIDER_WRITE_RETRY_TIMEOUT, 5)
var writeRetryTimeout = time.Duration(writeRetry) * time.Minute

var readRetry = getEnvDefault(PROVIDER_READ_RETRY_TIMEOUT, 3)
var readRetryTimeout = time.Duration(readRetry) * time.Minute

var bmcCreateTimeout = time.Duration(getEnvDefault(PROVIDER_BMC_CREATE_TIMEOUT, 90)) * time.Minute
var bmcUpdateTimeout = time.Duration(getEnvDefault(PROVIDER_BMC_UPDATE_TIMEOUT, 90)) * time.Minute

func getEnvDefault(key string, defVal int) int {
	val, ex := os.LookupEnv(key)
	if !ex {
		return defVal
	}
	int, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Sprintf("%s must be int.", key))
	}
	return int
}

func toJsonString(data interface{}) string {
	if data == nil {
		return ""
	}
	b, _ := json.Marshal(data)
	return string(b)
}

func logElapsed(ctx context.Context, mark ...string) func() {
	startAt := time.Now()
	return func() {
		tflog.Debug(ctx, "[ELAPSED] function elapsed", map[string]interface{}{
			"mark":   strings.Join(mark, " "),
			"timeMs": int64(time.Since(startAt) / time.Millisecond),
		})
	}
}

func String(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

func dataResourceIdHash(ids []string) string {
	var buf bytes.Buffer

	for _, id := range ids {
		buf.WriteString(fmt.Sprintf("%s-", id))
	}

	return fmt.Sprintf("%d", String(buf.String()))
}

func writeToFile(filePath string, data interface{}) error {
	if strings.HasPrefix(filePath, "~") {
		usr, err := user.Current()
		if err != nil {
			return fmt.Errorf("get current user fail,reason %s", err.Error())
		}
		if usr.HomeDir != "" {
			filePath = strings.Replace(filePath, "~", usr.HomeDir, 1)
		}
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat old file error,reason %s", err.Error())
	}

	if !os.IsNotExist(err) {
		if fileInfo.IsDir() {
			return fmt.Errorf("old filepath is a dir,can not delete")
		}
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("delete old file error,reason %s", err.Error())
		}
	}

	jsonStr, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return fmt.Errorf("json decode error,reason %s", err.Error())
	}

	return ioutil.WriteFile(filePath, jsonStr, 0422)
}

func IsContains(array interface{}, value interface{}) bool {
	vv := reflect.ValueOf(array)
	if vv.Kind() == reflect.Ptr || vv.Kind() == reflect.Interface {
		if vv.IsNil() {
			return false
		}
		vv = vv.Elem()
	}

	switch vv.Kind() {
	case reflect.Invalid:
		return false
	case reflect.Slice:
		for i := 0; i < vv.Len(); i++ {
			if reflect.DeepEqual(value, vv.Index(i).Interface()) {
				return true
			}
		}
		return false
	case reflect.Map:
		s := vv.MapKeys()
		for i := 0; i < len(s); i++ {
			if reflect.DeepEqual(value, s[i].Interface()) {
				return true
			}
		}
		return false
	case reflect.String:
		ss := reflect.ValueOf(value)
		switch ss.Kind() {
		case reflect.String:
			return strings.Contains(vv.String(), ss.String())
		}
		return false
	default:
		return reflect.DeepEqual(array, value)
	}
}

func NewGoRoutine(num int) *GoRoutineLimit {
	return &GoRoutineLimit{
		Count: num,
		Chan:  make(chan struct{}, num),
	}
}

type GoRoutineLimit struct {
	Count int
	Chan  chan struct{}
}

func (g *GoRoutineLimit) Run(f func()) {
	g.Chan <- struct{}{}
	go func() {
		f()
		<-g.Chan
	}()
}

func logApiRequest(ctx context.Context, action string, request interface{}, response interface{}, err error) {
	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("Call api [%s]", action), map[string]interface{}{
			"request": toJsonString(request),
			"err":     err.Error(),
		})
	} else {
		tflog.Debug(ctx, fmt.Sprintf("Call api [%s]", action), map[string]interface{}{
			"request":  toJsonString(request),
			"response": toJsonString(response),
		})
	}
}

func toStringList(value []interface{}) []string {
	list := make([]string, 0)
	for _, v := range value {
		vStr := v.(string)
		list = append(list, vStr)
	}
	return list
}
func toIntList(value []interface{}) []int {
	res := make([]int, 0, len(value))
	for _, v := range value {
		if e, ok := v.(int); ok {
			res = append(res, e)
		} else {
			panic(fmt.Errorf("value %s shoud be int", v))
		}
	}
	return res
}

func ParseResourceId(id string, length int) (parts []string, err error) {
	parts = strings.Split(id, ":")

	if len(parts) != length {
		err = fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, length, len(parts))
	}
	return parts, err
}
