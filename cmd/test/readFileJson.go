package test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

// readFileJson 从文件中读取json，并返回一个切片
func readFileJson(fname string, v any) ([]any, error) {
	// 读取文件内容
	b, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	// 获取v的反射类型
	valType := reflect.TypeOf(v)
	if valType == nil {
		return nil, fmt.Errorf("v cannot be nil")
	}
	// 尝试先按切片解析
	sliceType := reflect.SliceOf(valType)
	sliceVal := reflect.MakeSlice(sliceType, 0, 0)
	// 使用指针进行Unmarshal
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(sliceVal)

	if err := json.Unmarshal(b, slicePtr.Interface()); err == nil {
		// 解析为切片成功
		result := make([]any, slicePtr.Elem().Len())
		for i := 0; i < slicePtr.Elem().Len(); i++ {
			result[i] = slicePtr.Elem().Index(i).Interface()
		}
		return result, nil
	}
	// 解析为切片失败，尝试解析为单个对象
	elemPtr := reflect.New(valType)
	if err := json.Unmarshal(b, elemPtr.Interface()); err != nil {
		return nil, err
	}
	// 将单个对象包装成切片返回
	return []any{elemPtr.Elem().Interface()}, nil
}
