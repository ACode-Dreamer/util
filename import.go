package util

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"reflect"
	"strconv"
)

// ParseExcel 解析Excel文件并返回结构体切片
func ParseExcel[T any](file *excelize.File, sheetName string) ([]T, error) {
	var result []T
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	// 获取列的映射关系
	columnMap := make(map[string]int)
	headers := rows[0]
	for i, header := range headers {
		columnMap[header] = i
	}

	// 遍历行并创建结构体实例
	for _, row := range rows[1:] {
		instance := new(T)
		value := reflect.ValueOf(instance).Elem() // 获取结构体实例的值
		typeOfT := value.Type()                   // 获取结构体类型信息

		for i := 0; i < value.NumField(); i++ {
			field := typeOfT.Field(i)         // 获取结构体字段信息
			tag := field.Tag.Get("xlsx")      // 获取字段的xlsx标签值
			columnIndex, ok := columnMap[tag] // 从列映射中获取标签对应的列索引
			if ok {
				cellValue := row[columnIndex] // 获取单元格的值
				fieldValue := value.Field(i)  // 获取字段的值

				switch fieldValue.Kind() {
				case reflect.String:
					fieldValue.SetString(cellValue)
				case reflect.Int, reflect.Int64:
					intValue, err := strconv.Atoi(cellValue)
					if err != nil {
						return nil, err
					}
					fieldValue.SetInt(int64(intValue))
				case reflect.Float64:
					floatValue, err := strconv.ParseFloat(cellValue, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to convert '%s' to float64: %v", cellValue, err)
					}
					fieldValue.SetFloat(floatValue)
				default:
					fieldValue.SetZero()
				}
			}
		}
		result = append(result, *instance)
	}

	return result, nil
}
