package main

import (
	"bytes"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/afero"
)

var goodConf = []byte(`
---
  name: echo
  directory: /bin
  command: "echo foobar"
  git: false
`)

var badConf = []byte(`
---
  nam: echio
  cmd: "echo foobar
  git: false
`)

var confTests = []struct {
	name string
	conf []byte
	app  *App
}{
	{"good_conf", goodConf, &App{Name: "echo", Dir: "/bin", Cmd: "echo foobar"}},
	{"bad_conf", badConf, &App{}},
}

func TestConfigure(t *testing.T) {
	for _, test := range confTests {
		writeFile(test.name, test.conf)
		app := newApp(test.name)
		if !app.eql(test.app) {
			t.Errorf("%q: Unmarshall'd config did not produce expected object!", test.name)
			spew.Dump(app)
		}
	}

}

func writeFile(name string, content []byte) afero.File {
	file, err := fs.Create(name)
	if err != nil {
		return file
	}
	fs.Chmod(file.Name(), 0666)
	b := bytes.NewBuffer(content)
	b.WriteTo(file)
	return file
}

func init() {
	fs = &afero.MemMapFs{}
}
