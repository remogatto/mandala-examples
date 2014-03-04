// +build !android

package main

import "unsafe"

func ShowAdPopup(activity unsafe.Pointer) {
	// do nothing on desktop
}
