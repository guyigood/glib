package excel

import (
	"github.com/Luxurioust/excelize"
	"fmt"
	"strconv"
)

func Export_Excel(filename string,title_data []string,data []map[string]interface{},title []string)bool{
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
	//var title =[]string{"bh","zh_name","duty_no","address","bank_no","skr","shr","kpr","memo","spname","ggxh","jldw","quantity","price","ws_price","sl","ssbm","is_kp"}
	for _,val_data:=range data{
		cols_row=65
		for j:=0;j<len(title);j++{
			//fmt.Println(title[j],val_data[title[j]])
			xlsx.SetCellValue("Sheet1", string(rune(cols_row))+strconv.Itoa(i), val_data[title[j]])
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



func Import_Excel(filename string){
	xlsx, err := excelize.OpenFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Get value from cell by given worksheet name and axis.
	cell := xlsx.GetCellValue("Sheet1", "B2")
	fmt.Println(cell)
	// Get all the rows in the Sheet1.
	rows := xlsx.GetRows("Sheet1")
	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
}