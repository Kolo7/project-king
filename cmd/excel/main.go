package main

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func main() {
	f, err := excelize.OpenFile("./test.xlsx")
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		_ = f.Close()
	}()
	rows, _ := f.GetRows("sheet1")
	fmt.Println(rows)
}
