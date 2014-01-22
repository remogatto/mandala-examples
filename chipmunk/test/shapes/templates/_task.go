// +build gotask

package tasks

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jingweno/gotask/tasking"
)

var (
	// The name of the library
	LibName = "{{.LibName}}"

	// The domain without the last part
	Domain = "{{.Domain}}"

	// The path for the ARM binary. The binary is then copied on
	// each of SharedLibraryPaths
	ARMBinaryPath = "bin/linux_arm"

	// The path for shared libraries.
	SharedLibraryPaths = []string{
		"android/obj/local/armeabi-v7a/",
		"android/libs/armeabi-v7a/",
	}

	// Android path
	AndroidPath = "android"

	buildFun = map[string]func(*tasking.T){
		"xorg":    buildXorg,
		"android": buildAndroid,
	}

	runFun = map[string]func(*tasking.T){
		"xorg":    runXorg,
		"android": runAndroid,
	}
)

// NAME
//    build - Build the application
//
// DESCRIPTION
//    Build the application for the given platforms.
//
// OPTIONS
//    --flags=<FLAGS>
//        pass FLAGS to the compiler
//    --verbose, -v
//        run in verbose mode
func TaskBuild(t *tasking.T) {
	for _, platform := range t.Args {
		buildFun[platform](t)
	}
	if t.Failed() {
		t.Fatalf("%-20s %s\n", status(t.Failed()), "Build the application for the given platforms.")
	}
	t.Logf("%-20s %s\n", status(t.Failed()), "Build the application for the given platforms.")
}

// NAME
//    run - Run the application
//
// DESCRIPTION
//    Build and run the application on the given platforms.
//
// OPTIONS
//    --flags=<FLAGS>
//        pass the flags to the executable
//    --logcat=Mandala:* stdout:* stderr:* *:S
//        show logcat output (android only)
//    --verbose, -v
//        run in verbose mode
func TaskRun(t *tasking.T) {
	TaskBuild(t)
	for _, platform := range t.Args {
		runFun[platform](t)
	}
	if t.Failed() {
		t.Fatalf("%-20s %s\n", status(t.Failed()), "Run the application on the given platforms.")
	}
	t.Logf("%-20s %s\n", status(t.Failed()), "Run the application on the given platforms.")
}

// NAME
//    deploy - Deploy the application
//
// DESCRIPTION
//    Build and deploy the application on the device via ant.
//
// OPTIONS
//    --verbose, -v
//        run in verbose mode
func TaskDeploy(t *tasking.T) {
	deployAndroid(t)
	if t.Failed() {
		t.Fatalf("%-20s %s\n", status(t.Failed()), "Build and deploy the application on the device via ant.")
	}
	t.Logf("%-20s %s\n", status(t.Failed()), "Build and deploy the application on the device via ant.")
}

// NAME
//    clean - Clean all generated files
//
// DESCRIPTION
//    Clean all generated files and paths.
//
func TaskClean(t *tasking.T) {
	var paths []string

	paths = append(
		paths,
		ARMBinaryPath,
		"pkg",
		filepath.Join("bin"),
		filepath.Join(AndroidPath, "bin"),
		filepath.Join(AndroidPath, "gen"),
		filepath.Join(AndroidPath, "libs"),
		filepath.Join(AndroidPath, "obj"),
	)

	// Actually remove files using rm
	for _, path := range paths {
		err := rm_rf(t, path)
		if err != nil {
			t.Error(err)
		}
	}
	if t.Failed() {
		t.Fatalf("%-20s %s\n", status(t.Failed()), "Clean all generated files and paths.")
	}
	t.Logf("%-20s %s\n", status(t.Failed()), "Clean all generated files and paths.")
}

func buildXorg(t *tasking.T) {
	err := t.Exec(
		`sh -c "`,
		"GOPATH=`pwd`:$GOPATH",
		`go get`, t.Flags.String("flags"),
		LibName, `"`,
	)
	if err != nil {
		t.Error(err)
	}
}

func buildAndroid(t *tasking.T) {
	os.MkdirAll("android/libs/armeabi-v7a", 0777)
	os.MkdirAll("android/obj/local/armeabi-v7a", 0777)

	err := t.Exec(`sh -c "`,
		`CC="$NDK_ROOT/bin/arm-linux-androideabi-gcc"`,
		"GOPATH=`pwd`:$GOPATH",
		`GOROOT=""`,
		"GOOS=linux",
		"GOARCH=arm",
		"GOARM=7",
		"CGO_ENABLED=1",
		"$GOANDROID/go get", t.Flags.String("flags"),
		"$GOFLAGS",
		`-ldflags=\"-android -shared -extld $NDK_ROOT/bin/arm-linux-androideabi-gcc -extldflags '-march=armv7-a -mfloat-abi=softfp -mfpu=vfpv3-d16'\"`,
		"-tags android",
		LibName, `"`,
	)

	if err != nil {
		t.Error(err)
	}

	for _, path := range SharedLibraryPaths {
		err := t.Exec(
			"cp",
			filepath.Join(ARMBinaryPath, LibName),
			filepath.Join(path, "lib"+LibName+".so"),
		)

		if err != nil {
			t.Error(err)
		}
	}

	if err == nil {
		err = t.Exec("ant -f android/build.xml clean debug")
		if err != nil {
			t.Error(err)
		}
	}

}

func runXorg(t *tasking.T) {
	err := t.Exec(
		filepath.Join("bin", LibName),
		t.Flags.String("flags"),
	)
	if err != nil {
		t.Error(err)
	}
}

func runAndroid(t *tasking.T) {
	deployAndroid(t)
	err := t.Exec(
		fmt.Sprintf(
			"adb shell am start -a android.intent.action.MAIN -n %s.%s/android.app.NativeActivity",
			Domain,
			LibName,
		))
	if err != nil {
		t.Error(err)
	}
	if tags := t.Flags.String("logcat"); tags != "" {
		err = t.Exec("adb", "shell", "logcat", tags)
		if err != nil {
			t.Error(err)
		}
	}
}

func deployAndroid(t *tasking.T) {
	err := t.Exec(fmt.Sprintf("adb install -r android/bin/%s-debug.apk", LibName))
	if err != nil {
		t.Error(err)
	}
}

func cp(t *tasking.T, src, dest string) error {
	return t.Exec("cp", src, dest)
}

func rm_rf(t *tasking.T, path string) error {
	return t.Exec("rm -rf", path)
}
