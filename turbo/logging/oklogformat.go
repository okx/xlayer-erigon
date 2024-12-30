package logging

import (
	"encoding/json"
	"fmt"
	"github.com/ledgerwatch/log/v3"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

func OkLogV1Format(r *log.Record) []byte {
	r.KeyNames = log.RecordKeyNames{
		Time: "t",
		Msg:  "content",
		Lvl:  "level",
	}
	props := make(map[string]interface{})

	callFrame := r.Call.Frame()

	props["time"] = r.Time.UnixMilli()
	props[r.KeyNames.Time] = r.Time
	props[r.KeyNames.Lvl] = strings.ToUpper(r.Lvl.String())
	props[r.KeyNames.Msg] = r.Msg
	props["line_num"] = callFrame.Line
	props["class_name"] = filepath.Base(callFrame.File)

	for i := 0; i < len(r.Ctx); i += 2 {
		k, ok := r.Ctx[i].(string)
		if !ok {
			props[errorKey] = fmt.Sprintf("%+v is not a string key", r.Ctx[i])
		}
		props[k] = formatJSONValue(r.Ctx[i+1])
	}

	b, err := json.Marshal(props)
	if err != nil {
		b, _ = json.Marshal(map[string]string{
			errorKey: err.Error(),
		})
		return b
	}

	b = append(b, '\n')

	return b
}

func formatJSONValue(value interface{}) interface{} {
	value = formatShared(value)

	switch value.(type) {
	case int, int8, int16, int32, int64, float32, float64, uint, uint8, uint16, uint32, uint64, string:
		return value
	case interface{}, map[string]interface{}, []interface{}:
		return value
	default:
		return fmt.Sprintf("%+v", value)
	}
}

func formatShared(value interface{}) (result interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
				result = "nil"
			} else {
				panic(err)
			}
		}
	}()

	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)

	case error:
		return v.Error()

	case fmt.Stringer:
		return v.String()

	default:
		return v
	}
}
