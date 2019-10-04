package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"

	"github.com/rivo/tview"
)

const logo = `
 __________________________________________
 ___/\/\/\/\/\____/\/\/\/\/\____/\/\/\/\/\_
 _/\/\__________/\/\__________/\/\_________ 
 _/\/\__________/\/\__/\/\/\__/\/\_________  
 _/\/\__________/\/\____/\/\__/\/\_________   
 ___/\/\/\/\/\____/\/\/\/\/\____/\/\/\/\/\_    
 __________________________________________    
`

type HMI func(*tview.Pages) (title string, content tview.Primitive)

var app = tview.NewApplication()
var table *tview.Table

func updateScheduler() {
	refreshInterval := time.Tick(500 * time.Millisecond)
	for range refreshInterval {
		cellData := rand.Intn(10)
		tableCell := tview.NewTableCell(strconv.Itoa(cellData)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignCenter).
			SetSelectable(true)
		table.SetCell(1, 1, tableCell)
		app.Draw()
	}
}

func main() {

	hmis := []HMI{
		Splash,
		Overview,
	}

	pages := tview.NewPages()

	for _, hmi := range hmis {
		title, primative := hmi(pages)
		pages.AddPage(title, primative, true, title == "Splash")
	}

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true)

	//go updateScheduler()
	if err := app.SetRoot(layout, true).Run(); err != nil {
		panic(err)
	}
}

func Splash(pages *tview.Pages) (title string, content tview.Primitive) {
	lines := strings.Split(logo, "\n")
	logoWidth := 0
	logoHeight := len(lines)
	for _, line := range lines {
		if len(line) > logoWidth {
			logoWidth = len(line)
		}
	}
	logoBox := tview.NewTextView().
		SetTextColor(tcell.ColorBlue).
		SetDoneFunc(func(key tcell.Key) {
			pages.ShowPage("Overview")
		})

	fmt.Fprint(logoBox, logo)

	frame := tview.NewFrame(tview.NewBox()).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText("Composite Grid Controller v0.1", true, tview.AlignCenter, tcell.ColorWhite).
		AddText("", true, tview.AlignCenter, tcell.ColorWhite).
		AddText("press enter", true, tview.AlignCenter, tcell.ColorDarkMagenta)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), 0, 5, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(logoBox, logoWidth, 1, true).
			AddItem(tview.NewBox(), 0, 1, false), logoHeight, 1, true).
		AddItem(frame, 0, 10, false)

	return "Splash", flex
}

func Overview(pages *tview.Pages) (title string, content tview.Primitive) {

	rootNode := "Bus A"
	root := tview.NewTreeNode(rootNode).
		SetColor(tcell.ColorBlue)
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	add := 
}

/*
const tableData = `Machine|kW|kVAR|Frequency|Voltage|Online|Gridforming
ESS|0|0|60|480|False|False
Grid|10|3|60|480|True|True
Feeder|10|3|60|480|True|False
`

func Overview(pages *tview.Pages) (title string, content tview.Primitive) {

	table = tview.NewTable().
		SetFixed(1, 1)

	for row, line := range strings.Split(tableData, "\n") {
		for column, cell := range strings.Split(line, "|") {
			color := tcell.ColorWhite
			if row == 0 {
				color = tcell.ColorYellow
			} else if column == 0 {
				color = tcell.ColorDarkCyan
			}
			align := tview.AlignLeft
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(align).
				SetSelectable(row != 0)
			table.SetCell(row, column, tableCell)
		}
	}
	table.SetBorder(true).SetTitle(" Assets ")

	table.SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ').
		SetSelectedFunc()

	app.SetFocus(table)

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(table, 0, 1, true), 0, 1, true)

	return "Overview", flex
}

*/
