package lib

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func CreateTable(headers []string, data [][]string) *widget.Table {
	table := widget.NewTable(
		func() (int, int) {
			return len(data) + 1, len(headers)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if i.Row == 0 {
				label.SetText(headers[i.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				label.SetText(data[i.Row-1][i.Col])
			}
		},
	)
	return table
}
