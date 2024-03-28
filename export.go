package util

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"reflect"
)

// ExportExcel 将结构体数组导出到Excel
func ExportExcel(c *gin.Context, xlsxName string, arr interface{}) error {
	f := excelize.NewFile()

	sheetName := "Sheet1"
	index, _ := f.GetSheetIndex(sheetName)
	if index == 0 {
		// 如果工作表不存在，则创建一个新的工作表
		_, err := f.NewSheet(sheetName)
		if err != nil {
			return err
		}
	}

	// 写标题行
	headers, err := headersFromStruct(arr)
	if err != nil {
		return err
	}
	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return err
		}
		if err = f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	// 写入数据
	if err := writeDataToExcel(f, sheetName, arr); err != nil {
		return err
	}

	// 将Excel文件写入HTTP响应体
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xlsx", xlsxName))
	if err := f.Write(c.Writer); err != nil {
		return err
	}

	return nil
}

// headersFromStruct 从结构体中获取Excel表头
func headersFromStruct(arr interface{}) ([]string, error) {
	headers := make([]string, 0)
	val := reflect.ValueOf(arr)
	if val.Kind() != reflect.Array && val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("请传入数组或切片")
	}
	if val.Len() > 0 { // 取出一个结构体 获取其字段 编写表头
		elem := val.Index(0)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Struct {
			return nil, fmt.Errorf("元素必须是结构体")
		}
		elemType := elem.Type()
		for j := 0; j < elem.NumField(); j++ {
			fieldType := elemType.Field(j)
			xlsxTag := fieldType.Tag.Get("xlsx")
			headers = append(headers, xlsxTag)
		}
	}
	return headers, nil
}

// writeDataToExcel 将数据写入Excel
func writeDataToExcel(f *excelize.File, sheetName string, arr interface{}) error {
	val := reflect.ValueOf(arr)
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i) // 逐一取出结构体进行输出
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		for j := 0; j < elem.NumField(); j++ {
			fieldVal := elem.Field(j) // 取出每个结构体的字段值输出
			cell, err := excelize.CoordinatesToCellName(j+1, i+2)
			if err != nil {
				return err
			}
			if err = f.SetCellValue(sheetName, cell, fieldVal.Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}
