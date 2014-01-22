// +build gotask

package tasks

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jingweno/gotask/tasking"
)

var (
	jsonPath      = "app.json"
	templatePaths = []string{
		"templates/README.md",
		"templates/_task.go",
		"templates/android/AndroidManifest.xml",
		"templates/android/build.xml",
		"templates/android/res/values/strings.xml",
		"templates/src/_app/main.go",
	}
	labelFAIL = red("F")
	labelPASS = green("OK")
)

type Application struct {
	Domain  string
	LibName string
	AppName string
}

func (p *Application) GetFullDomain() string {
	return p.Domain + "." + p.LibName
}

// NAME
//   init - Initialize a new Mandala application
//
// DESCRIPTION
//    Initialize a new Mandala application based on application.json
//
func TaskInit(t *tasking.T) {
	var err error

	// read application.json
	app, err := readJSON(jsonPath)
	if err != nil {
		t.Error(err)
	}

	var paths []string

	// execute templates and copy the files
	err = filepath.Walk(
		"templates",
		func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				paths = append(paths, path)
			}
			return nil
		},
	)
	if err != nil {
		t.Error(err)
	}

	for _, path := range paths {
		var dstPath string
		splits := strings.Split(path, "/")
		if len(splits) > 1 {
			dstPath = filepath.Join(splits[1:]...)
		} else {
			dstPath = filepath.Base(path)
		}
		if err = copyFile(path, dstPath, app); err != nil {
			t.Error(err)
		}
	}

	// Rename paths accordly to app.LibName
	if err = os.Rename("_task.go", strings.ToLower(app.LibName)+"_task.go"); err != nil {
		t.Error(err)
	}
	if err = os.Rename("src/_app", filepath.Join("src", strings.ToLower(app.LibName))); err != nil {
		t.Error(err)
	}

	if t.Failed() {
		t.Fatalf("%-20s %s\n", status(t.Failed()), "Initialize a new Mandala application")
	}
	t.Logf("%-20s %s\n", status(t.Failed()), "Initialize a new Mandala application")
}

func readJSON(filename string) (*Application, error) {
	app := new(Application)

	jsonBuffer, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonBuffer, app)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// Copy file from src to dest. Execute a template if file is in
// templatePaths.
func copyFile(src, dst string, app *Application) (err error) {
	if dstFile, err := os.Open(dst); err == nil {
		// file already exists, close and exit
		dstFile.Close()
		return nil
	} else {

		defer dstFile.Close()

		// take directory name from dest path
		dir := filepath.Dir(dst)

		// create the directory along with any necessary
		// parents
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}

		// Open and read the source file
		srcBuffer, err := ioutil.ReadFile(src)
		if err != nil {
			return err
		}

		// Create destination file and write source data into
		// it
		dstFile, err = os.Create(dst)

		if err != nil {
			return err
		}

		// Execute template if source file is in the template
		// path
		for _, tplPath := range templatePaths {
			if src == tplPath {
				tmpl, err := template.New(tplPath).Parse(string(srcBuffer))
				if err != nil {
					return err
				}
				err = tmpl.Execute(dstFile, app)
				if err != nil {
					return err
				}
				// We're done, exit.
				return nil
			}
		}
		_, err = dstFile.Write(srcBuffer)
		if err != nil {
			return err
		}
	}
	return nil
}

func green(text string) string {
	return "\033[32m" + text + "\033[0m"
}

func red(text string) string {
	return "\033[31m" + text + "\033[0m"
}

func yellow(text string) string {
	return "\033[33m" + text + "\033[0m"
}

func status(failed bool) string {
	if failed {
		return red("FAIL")
	}
	return green("OK")
}
