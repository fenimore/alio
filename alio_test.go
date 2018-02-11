package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	code := m.Run()
	os.Exit(code)
}

func TestCollectAlbum(t *testing.T) {
	albums, err := CollectAlbums("test-data")
	ok(t, err)

	expected := []*Album{
		&Album{
			"album_one", []string{
				"example_one",
				"example_two",
			},
			[]string{
				"test-data/album_one/example_one.mp3",
				"test-data/album_one/example_two.mp3",
			},
			"test-data/album_one/cover_art.png", 2, 1,
		},
		&Album{
			"album_two", []string{"example_two"},
			[]string{"test-data/album_two/example_two.mp3"},
			"", 1, 2,
		},
	}
	for idx, _ := range albums {
		equals(t, expected[idx], albums[idx])
	}
}

func TestTimeStamp(t *testing.T) {
	equals(t, timestamp(0.11881344, 121078), `  00:00:14 - 00:02:01`)
	equals(t, timestamp(0.22867697, 53107), `  00:00:12 - 00:00:53`)

}

//////////////////////
// Helper functions //
//////////////////////

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
