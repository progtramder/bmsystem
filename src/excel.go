package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"os"
	"unicode/utf8"
)

type excel struct {
	*excelize.File
}

func trimSheetName(name string) string {
	var r []rune
	for _, v := range name {
		switch v {
		case 58, 92, 47, 63, 42, 91, 93: // replace :\/?*[]
			continue
		default:
			r = append(r, v)
		}
	}
	name = string(r)
	if utf8.RuneCountInString(name) > 31 {
		name = string([]rune(name)[0:31])
	}
	return name
}

func (self *excel) serialize(token, session string, info bminfo) {
	sheetName := trimSheetName(session)

	//make the title of each column
	if len(self.GetRows(sheetName)) == 0 {
		self.InsertRow(sheetName, 1)
		style, _ := self.NewStyle(`{"font":{"bold":true}}`)
		column := 'A'
		axis := fmt.Sprintf("%c1", column)
		self.SetCellValue(sheetName, axis, "Token")
		for _, v := range info.form {
			column++
			axis = fmt.Sprintf("%c1", column)
			self.SetCellValue(sheetName, axis, v.key)
		}
		self.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", column), style)
		self.SetColWidth(sheetName, "A", fmt.Sprintf("%c", column), 20)
	}
	//information starts from row 2
	self.InsertRow(sheetName, 2)
	column := 'A'
	axis := fmt.Sprintf("%c2", column)
	self.SetCellValue(sheetName, axis, token)
	for _, v := range info.form {
		column++
		axis = fmt.Sprintf("%c2", column)
		self.SetCellValue(sheetName, axis, v.value)
	}

	self.Save()
}

func InitReport(e Event) (*excel, error) {
	os.Mkdir("report", 0777)
	filename := fmt.Sprintf(systembasePath+"/report/%s.xlsx", e.Event)
	xlsx, err := excelize.OpenFile(filename)
	if err != nil {
		xlsx = excelize.NewFile()
		xlsx.SaveAs(filename)
		xlsx, err = excelize.OpenFile(filename)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range e.Sessions {
		sheetName := trimSheetName(v.Desc)
		xlsx.NewSheet(sheetName)
	}
	xlsx.DeleteSheet("Sheet1")
	return &excel{xlsx}, nil
}
