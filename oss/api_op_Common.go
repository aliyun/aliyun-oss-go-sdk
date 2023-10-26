package oss

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

func (c *Client) InvokeOperation(ctx context.Context, input *OperationInput, optFns ...func(*Options)) (output *OperationOutput, err error) {
	return c.invokeOperation(ctx, input, optFns)
}

func (c *Client) MarshalInput(request interface{}, input *OperationInput) error {
	val := reflect.ValueOf(request)
	switch val.Kind() {
	case reflect.Pointer, reflect.Interface:
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct || input == nil {
		return nil
	}

	t := val.Type()
	for k := 0; k < t.NumField(); k++ {
		if v := val.Field(k); !isEmptyValue(v) {
			if tag := t.Field(k).Tag.Get("input"); tag != "" {
				tokens := strings.Split(tag, ",")
				if len(tokens) < 2 {
					continue
				}
				switch tokens[0] {
				case "query":
					if input.Parameters == nil {
						input.Parameters = map[string]string{}
					}
					input.Parameters[tokens[1]] = fmt.Sprintf("%v", v.Interface())
				case "header":
					if input.Headers == nil {
						input.Headers = map[string]string{}
					}
					input.Headers[tokens[1]] = fmt.Sprintf("%v", v.Interface())
				}
			}
		}
	}
	return nil
}
