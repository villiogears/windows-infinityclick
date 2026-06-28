package main

import (
	"strings"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
)

var (
	nativeStartCode int = 0x70 // F1
	nativeStopCode  int = 0x71 // F2
	globalOnStart   func()
	globalOnStop    func()
)

var procGetAsyncState = syscall.NewLazyDLL("user32.dll").NewProc("GetAsyncKeyState")

// Windows の仮想キーコード (Virtual-Key Codes) にマッピング
func parseToWinKeyCode(input string) int {
	clean := strings.TrimSpace(strings.ToUpper(input))
	switch clean {
	case "F1": return 0x70
	case "F2": return 0x71
	case "F3": return 0x72
	case "F4": return 0x73
	case "F5": return 0x74
	case "F6": return 0x75
	case "F7": return 0x76
	case "F8": return 0x77
	case "F9": return 0x78
	case "F10": return 0x79
	case "F11": return 0x7A
	case "F12": return 0x7B
	case "SPACE": return 0x20
	case "ENTER": return 0x0D
	default:
		if len(clean) == 1 {
			r := clean[0]
			if r >= 'A' && r <= 'Z' {
				return int(r) // A-Z はそのままアスキーコードが仮想キーコード
			}
		}
		return 0x70 // デフォルト F1
	}
}

func SetupCustomKeybinds(w fyne.Window, startKeyTarget, stopKeyTarget *string, onStart func(), onStop func()) {
	globalOnStart = onStart
	globalOnStop = onStop

	// キーバインド文字列を定期的に反映するループ
	go func() {
		for {
			if startKeyTarget != nil && stopKeyTarget != nil {
				nativeStartCode = parseToWinKeyCode(*startKeyTarget)
				nativeStopCode = parseToWinKeyCode(*stopKeyTarget)
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	// グローバルキー入力を監視するループ (Windows用ポーリング)
	go func() {
		for {
			// GetAsyncKeyState の最上位ビットが1の場合、キーが押されている
			startState, _, _ := procGetAsyncState.Call(uintptr(nativeStartCode))
			if (startState & 0x8000) != 0 {
				if !isClicking && globalOnStart != nil {
					// GUIスレッドのブロックを防ぐため別ゴルーチンで実行
					go globalOnStart()
				}
			}

			stopState, _, _ := procGetAsyncState.Call(uintptr(nativeStopCode))
			if (stopState & 0x8000) != 0 {
				if isClicking && globalOnStop != nil {
					go globalOnStop()
				}
			}
			
			// CPU使用率が跳ね上がらないように 50ms 待機
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

func parseKeyName(input string) fyne.KeyName {
	return fyne.KeyF1
}