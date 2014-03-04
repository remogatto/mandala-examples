// +build android

package main

// #include <android/native_activity.h>
// #include "admob_android.h"
import "C"
import "unsafe"

func ShowAdPopup(activity unsafe.Pointer) {
	C.showAdPopup((*C.ANativeActivity)(activity))
}
