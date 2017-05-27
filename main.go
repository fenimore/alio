package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/marcusolsson/tui-go"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var DIR = "Music"
var MusicExts = ".mp3 .ogg .m4a"
var PhotoExts = ".jpg .jpeg .png"

type Album struct {
	Title string
	Songs []string
	Paths []string
	Cover string
	Count int
}

func (a *Album) String() string {
	return fmt.Sprintf("%s | %d", a.Title, a.Count)
}

func main() {
	fmt.Println("Welcome")
	albums, err := CollectAlbums(DIR)
	if err != nil {
		panic(err)
	}

	libTable := tui.NewTable(0, 1)
	libTable.SetColumnStretch(0, 1)

	library := tui.NewVBox(
		tui.NewLabel("Albums"),
		libTable,
		tui.NewSpacer(),
	)
	library.SetBorder(true)

	list := tui.NewList()
	songList := tui.NewVBox(
		tui.NewLabel("Songs"),
		list,
		tui.NewSpacer(),
	)
	songList.SetBorder(true)

	for _, album := range albums {
		libTable.AppendRow(
			tui.NewLabel(album.Title),
		)
	}

	progress := tui.NewProgress(100)
	progress.SetCurrent(0)

	status := tui.NewStatusBar("Song Title + Time")
	status.SetText(strconv.Itoa(libTable.Selected()))
	status.SetPermanentText("Volume")

	libTable.OnSelectionChanged(func(t *tui.Table) {
		if libTable.Selected() == 0 {
			return
		}
		progress.SetCurrent(libTable.Selected() * 10)
		status.SetText(strconv.Itoa(libTable.Selected()))
		list.RemoveItems()

		for _, v := range albums[libTable.Selected()-1].Songs {
			list.AddItems(v)
		}
	})

	selection := tui.NewHBox(
		library,
		songList,
		tui.NewSpacer(),
	)

	root := tui.NewVBox(
		selection,
		progress,
		status,
		tui.NewSpacer(),
	)

	ui := tui.New(root)

	// Key bindings
	ui.SetKeybinding(tui.KeyEsc, func() { ui.Quit() })
	ui.SetKeybinding('q', func() { ui.Quit() })
	ui.SetKeybinding(tui.KeySpace, func() { status.SetText("Play") })
	ui.SetKeybinding(tui.KeyEnter, func() {
		s := libTable.Selected() - 1
		if s == -1 {
			return
		}

		go playAlbum(albums[s])
	})

	if err := ui.Run(); err != nil {
		panic(err)
	}
}

func playAlbum(album *Album) {

}

func CollectAlbums(root string) ([]*Album, error) {
	dirs := make([]string, 0)
	// Never returns an Error?
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if path == root {
			return nil
		}
		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})

	albums := make([]*Album, 0)
	for _, val := range dirs {
		album := new(Album)
		songs := make([]string, 0)
		files, err := ioutil.ReadDir(val)
		if err != nil {
			return albums, err
		}

		for _, info := range files {
			//fmt.Println(info.Name())
			if info.IsDir() || info.Name()[0] == '.' {
				continue
			}
			if strings.Contains(MusicExts, path.Ext(info.Name())) {
				songs = append(songs, info.Name())
				continue
			}

			if strings.Contains(PhotoExts, path.Ext(info.Name())) {
				album.Cover = path.Join(val, info.Name())
			}
		}

		if len(songs) > 0 {
			for i, v := range songs {
				album.Songs = append(album.Songs, songs[i])
				album.Paths = append(album.Paths, path.Join(val, v))
			}
		}

		if len(album.Songs) > 0 {
			album.Title = val
			album.Count = len(songs)
			albums = append(albums, album)
		}
	}

	return albums, nil
}

// WatchFiles watches for changes to filesystem in directory.
func WatchFiles(dir string, errChan chan<- error) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return watcher, err
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				if ev.Op == fsnotify.Write { // FIXME: use Create instead?
					// do something
				}
			case _ = <-watcher.Errors:
				errChan <- err
			}
		}
	}()
	err = watcher.Add(dir)
	if err != nil {
		return watcher, err
	}

	return watcher, nil
}
