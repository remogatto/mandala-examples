// +build gotask

package tasks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/jingweno/gotask/tasking"
)

var (
	// The name of the library
	LibName = "chipmunk"

	// The domain without the last part
	Domain = "net.mandala"

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

	// KeyStore and KeyAlias, needed to release the app for Android platform
	KeyStore = "/home/andrea/backup/my-release-key.keystore"
	KeyAlias = "remogatto"

	buildFun = map[string]func(*tasking.T, ...bool){
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

// NAME
//    release - Build the application in 'release mode'
//
// DESCRIPTION
//    Build the application for Android in 'release mode'.
//
// OPTIONS
//    --flags=<FLAGS>
//        pass FLAGS to the compiler
//    --verbose, -v
//        run in verbose mode
func TaskRelease(t *tasking.T) {
	// Build app in 'release mode'
	buildAndroid(t, true)
	// Sign and 'zipalign' app
	signAndroid(t)
	// Check task
	if t.Failed() {
		t.Fatalf("%-20s %s\n", status(t.Failed()), "Release the application for Android.")
	}
	t.Logf("%-20s %s\n", status(t.Failed()), "Release the application for Android.")
}

func buildXorg(t *tasking.T, buildMode ...bool) {
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

func buildAndroid(t *tasking.T, buildMode ...bool) {
	// Build mode for ant:
	// buildMode not specified or false => ant debug
	// buildMode true => ant release
	antBuildParam := "debug"
	if len(buildMode) > 0 && buildMode[0] == true {
		antBuildParam = "release"
	}

	// Generate AdmobActivity using the UnitId in admob.json
	err := generateAdmobActivity()
	if err != nil {
		t.Error(err)
	}

	os.MkdirAll("android/libs/armeabi-v7a", 0777)
	os.MkdirAll("android/obj/local/armeabi-v7a", 0777)

	err = t.Exec(`sh -c "`,
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
		err = t.Exec("ant -f android/build.xml clean", antBuildParam)
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
			"adb shell am start -a android.intent.action.MAIN -n %s.%s/.AdmobActivity",
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

// Sign and zipalign Android application.
func signAndroid(t *tasking.T) {
	unsignedAppPath := fmt.Sprintf("android/bin/%s-release-unsigned.apk", LibName)
	// Sign app
	cmdJarsigner := "jarsigner -verbose -sigalg SHA1withRSA -digestalg SHA1 -keystore %s %s %s"
	if err := t.Exec(fmt.Sprintf(cmdJarsigner, KeyStore, unsignedAppPath, KeyAlias)); err != nil {
		t.Error(err)
	}
	// Verify sign
	cmdJarsignerVerify := "jarsigner -verify -verbose -certs %s"
	if err := t.Exec(fmt.Sprintf(cmdJarsignerVerify, unsignedAppPath)); err != nil {
		t.Error(err)
	}
	// Align app
	cmdZipAlign := "zipalign -v 4 %s android/bin/%s.apk"
	if err := t.Exec(fmt.Sprintf(cmdZipAlign, unsignedAppPath, LibName)); err != nil {
		t.Error(err)
	}
}

func generateAdmobActivity() error {
	var admob map[string]string
	jsonBuffer, err := ioutil.ReadFile("admob.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonBuffer, &admob)
	if err != nil {
		return err
	}

	// Open and read the template
	srcBuffer, err := ioutil.ReadFile("android/src/net/mandala/chipmunk/AdmobActivity.java.template")
	if err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.Create("android/src/net/mandala/chipmunk/AdmobActivity.java")
	if err != nil {
		return err
	}

	defer dstFile.Close()

	tmpl, err := template.New("AdmobActivity template").Parse(string(srcBuffer))
	if err != nil {
		return err
	}

	err = tmpl.Execute(dstFile, admob)
	if err != nil {
		return err
	}
	return nil
}

func cp(t *tasking.T, src, dest string) error {
	return t.Exec("cp", src, dest)
}

func rm_rf(t *tasking.T, path string) error {
	return t.Exec("rm -rf", path)
}
