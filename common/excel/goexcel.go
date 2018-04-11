package excel

import (
	"github.com/Luxurioust/excelize"
	"fmt"
	"strconv"
)

func Export_Excel(filename string,title_data []string,data []map[string]interface{})bool{
	xlsx := excelize.NewFile()
	// Create a new sheet.
	xlsx.NewSheet( "Sheet1")
	// Set value of a cell.
	cols_row:=65
	for _,val :=range title_data{
		xlsx.SetCellValue("Sheet1", string(rune(cols_row))+"1", val)
		cols_row++
	}
	i:=2
	for _,val_data:=range data{
		cols_row=65
 		for _,val1:=range val_data{
			xlsx.SetCellValue("Sheet1", string(rune(cols_row))+strconv.Itoa(i), val1)
			cols_row++
		}
		i++
	}
	// Set active sheet of the workbook.
	xlsx.SetActiveSheet(2)
	// Save xlsx file by the given path.
	err := xlsx.SaveAs(filename)
	if err != nil {
		fmt.Println(err)
		return false

	}
	return true
}