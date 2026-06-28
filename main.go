package main

import (
	_ "embed"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed font/Roboto_Condensed-ExtraBold.ttf
var fontData []byte

type ClickProfile struct {
	ButtonType string
	Interval   int
}

var profiles = map[string]*ClickProfile{
	"プロファイル A": {ButtonType: "左クリック", Interval: 100},
	"プロファイル B": {ButtonType: "左クリック", Interval: 50},
	"プロファイル C": {ButtonType: "右クリック", Interval: 200},
}

var currentProfile = "プロファイル A"

type PureWhiteTheme struct{}

func (m PureWhiteTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	pureWhite := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	vscodeLightGray := color.RGBA{R: 243, G: 243, B: 243, A: 255}
	vscodeDarkText := color.RGBA{R: 51, G: 51, B: 51, A: 255}
	vscodeBlue := color.RGBA{R: 0, G: 122, B: 204, A: 255}
	vscodeBorder := color.RGBA{R: 206, G: 206, B: 206, A: 255}

	switch name {
	case theme.ColorNameBackground, theme.ColorNameMenuBackground:
		return pureWhite
	case theme.ColorNameButton, theme.ColorNameInputBackground, theme.ColorNameOverlayBackground:
		return vscodeLightGray
	case theme.ColorNameForeground, theme.ColorNamePlaceHolder:
		return vscodeDarkText
	case theme.ColorNameSeparator, theme.ColorNameScrollBar:
		return vscodeBorder
	case theme.ColorNamePrimary:
		return vscodeBlue
	}
	return theme.DefaultTheme().Color(name, theme.VariantLight)
}

func (m PureWhiteTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(name) }
func (m PureWhiteTheme) Font(style fyne.TextStyle) fyne.Resource {
	if len(fontData) > 0 {
		return fyne.NewStaticResource("customFont.ttf", fontData)
	}
	return theme.DefaultTheme().Font(style)
}

func (m PureWhiteTheme) Size(name fyne.ThemeSizeName) float32 {
	n := string(name)
	if n == "text" || n == "heading" || n == "subHeading" { return theme.DefaultTheme().Size(name) * 1.3 }
	if n == "caption" { return 14 }
	if n == "lineSpacing" { return 24 }
	if n == "radius" || n == "inputRadius" || n == "selectionRadius" || n == "padding" || n == "spacer" { return 0 }
	if n == "innerPadding" { return 14 }
	return theme.DefaultTheme().Size(name)
}

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(PureWhiteTheme{})

	myWindow := myApp.NewWindow("elphadeal - Auto Clicker")
	myWindow.Resize(fyne.NewSize(700, 450))

	title := widget.NewLabelWithStyle("elphadeal 連打ツール", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	clickTypeLabel := widget.NewLabel("クリックボタン:")
	clickTypeSelect := widget.NewSelect([]string{"左クリック", "右クリック"}, func(selected string) {
		profiles[currentProfile].ButtonType = selected
	})

	intervalLabel := widget.NewLabel("連打間隔: 100 ミリ秒")
	slider := widget.NewSlider(1, 1000)
	slider.SetValue(100)

	statusLabel := widget.NewLabelWithStyle("ステータス: 停止中", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	
	startKeyStr := "F1"
	stopKeyStr := "F2"

	startKeyEntry := widget.NewEntry()
	startKeyEntry.SetText(startKeyStr)
	startKeyEntry.OnChanged = func(text string) { startKeyStr = text }

	stopKeyEntry := widget.NewEntry()
	stopKeyEntry.SetText(stopKeyStr)
	stopKeyEntry.OnChanged = func(text string) { stopKeyStr = text }

	var startBtn *widget.Button
	var stopBtn *widget.Button

	// バックグラウンドスレッドから呼ばれても安全にGUIを書き換えるためのハンドラ
	handleStart := func() {
		myWindow.QueueEvent(func() {
			statusLabel.SetText("ステータス: 連打中...")
			startBtn.Disable()
			stopBtn.Enable()
		})
		p := profiles[currentProfile]
		StartClicking(p.ButtonType, p.Interval)
	}

	handleStop := func() {
		myWindow.QueueEvent(func() {
			statusLabel.SetText("ステータス: 停止中")
			stopBtn.Disable()
			startBtn.Enable()
		})
		StopClicking()
	}

	startBtn = widget.NewButton("連打開始", func() { handleStart() })
	stopBtn = widget.NewButton("停止", func() { handleStop() })
	stopBtn.Disable()

	slider.OnChanged = func(value float64) {
		ms := int(value)
		intervalLabel.SetText("連打間隔: " + strconv.Itoa(ms) + " ミリ秒")
		profiles[currentProfile].Interval = ms
		UpdateSpeed(ms)
	}

	SetupCustomKeybinds(myWindow, &startKeyStr, &stopKeyStr, handleStart, handleStop)

	rightLayout := container.NewVBox(
		title,
		widget.NewSeparator(),
		clickTypeLabel,
		clickTypeSelect,
		widget.NewSeparator(),
		intervalLabel,
		slider,
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("開始キー設定:"), startKeyEntry),
			container.NewVBox(widget.NewLabel("停止キー設定:"), stopKeyEntry),
		),
		widget.NewSeparator(),
		statusLabel,
		container.NewGridWithColumns(2, startBtn, stopBtn),
	)

	options := []string{"プロファイル A", "プロファイル B", "プロファイル C"}
	dropdown := widget.NewSelect(options, func(selected string) {
		currentProfile = selected
		p := profiles[selected]
		clickTypeSelect.SetSelected(p.ButtonType)
		slider.SetValue(float64(p.Interval))
		intervalLabel.SetText("連打間隔: " + strconv.Itoa(p.Interval) + " ミリ秒")
	})
	dropdown.SetSelected("プロファイル A")

	leftLayout := container.NewVBox(
		widget.NewLabelWithStyle("  PROFILE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		dropdown,
	)

	leftContainer := container.NewStack(leftLayout)
	borderLayout := container.NewBorder(nil, nil, leftContainer, nil, rightLayout)
	leftLayout.Resize(fyne.NewSize(200, 450))

	myWindow.SetContent(borderLayout)
	myWindow.ShowAndRun()
}