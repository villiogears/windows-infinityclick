package main

import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

var (
	isClicking = false
	stopChan   = make(chan struct{})
	speedMu    sync.RWMutex
	currentMs  int
)

// Windows API の定義
var (
	moduser32       = syscall.NewLazyDLL("user32.dll")
	procSendInput   = moduser32.NewProc("SendInput")
)

type mouseInput struct {
	dx          int32
	dy          int32
	mouseData   uint32
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type input struct {
	inputType uint32
	mi        mouseInput
	padding   uint64 // 64bit環境用のアライメント調整
}

const (
	inputMouse   = 0
	mouseEventLfDown = 0x0002
	mouseEventLfUp   = 0x0004
	mouseEventRtDown = 0x0008
	mouseEventRtUp   = 0x0010
)

func windowsClick(isRightClick bool) {
	var downFlags, upFlags uint32 = mouseEventLfDown, mouseEventLfUp
	if isRightClick {
		downFlags, upFlags = mouseEventRtDown, mouseEventRtUp
	}

	// 1. マウスダウンイベント
	var inputs [2]input
	inputs[0].inputType = inputMouse
	inputs[0].mi = mouseInput{dwFlags: downFlags}
	
	// 2. マウスアップイベント
	inputs[1].inputType = inputMouse
	inputs[1].mi = mouseInput{dwFlags: upFlags}

	// Windowsにイベントを送信
	procSendInput.Call(
		uintptr(2),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)
}

func StartClicking(buttonType string, intervalMs int) {
	if isClicking {
		return
	}
	isClicking = true

	speedMu.Lock()
	currentMs = intervalMs
	speedMu.Unlock()

	isRight := buttonType == "右クリック"

	fmt.Println("[Log] 連打ループを開始しました...")

	go func() {
		for {
			speedMu.RLock()
			ms := currentMs
			speedMu.RUnlock()

			select {
			case <-time.After(time.Duration(ms) * time.Millisecond):
				fmt.Println("[Log] カチッ")
				windowsClick(isRight)
			case <-stopChan:
				fmt.Println("[Log] 連打ループを停止しました。")
				return
			}
		}
	}()
}

func UpdateSpeed(intervalMs int) {
	speedMu.Lock()
	currentMs = intervalMs
	speedMu.Unlock()
}

func StopClicking() {
	if !isClicking {
		return
	}
	isClicking = false
	select {
	case stopChan <- struct{}{}:
	default:
	}
}