//go:build windows && rtmidi
// +build windows,rtmidi

package game

import (
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)
