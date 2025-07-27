package soft

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

type ContextKey string

var ErrInvalidParams = errors.New("invalid parameters")

// Parse 自动解析 context → struct，并填充默认值
func Parse[T any](ctx context.Context) (*T, error) {
	// 1. 填充默认值
	var t T
	if err := defaults.Set(&t); err != nil {
		return &t, err
	}

	// 2. 构建 ctx map
	ctxMap := make(map[string]interface{})
	typ := reflect.TypeOf(t)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		key := field.Tag.Get("ctx")
		if key == "" {
			continue
		}

		v := ctx.Value(ContextKey(key))
		if v != nil {
			ctxMap[key] = v
		}
	}

	// 3. 用 mapstructure 解码到 struct
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &t,
		TagName: "ctx",
	})
	if err != nil {
		return &t, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(ctxMap); err != nil {
		return &t, fmt.Errorf("failed to decode context to struct: %w", err)
	}

	return &t, nil
}
