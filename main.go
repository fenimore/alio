package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/marcusolsson/tui-go"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
		fmt.Println(err)
	}

	libTable := tui.NewTable(0, 0)
	libTable.SetColumnStretch(0, 0)

	library := tui.NewVBox(
		tui.NewLabel("Albums"),
		libTable,
		tui.NewSpacer(),
	)
	library.SetBorder(true)

	list := tui.NewList()
	for _, v := range albums[0].Songs {
		list.AddItems(v)
	}

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

	libTable.OnSelectionChanged(func(t *tui.Table) {
		list.RemoveItems()
		for _, v := range albums[libTable.Selected()].Songs {
			list.AddItems(v)
		}
		list.SetFocused(true)
	})

	progress := tui.NewProgress(100)
	progress.SetCurrent(0)

	status := tui.NewStatusBar("Song Title + Time")
	status.SetPermanentText("Volume")

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
	ui.SetKeybinding(tui.KeyEsc, func() { ui.Quit() })
	ui.SetKeybinding('q', func() { ui.Quit() })
	ui.SetKeybinding(tui.KeySpace, func() { status.SetText("Play") })
	// ui.SetKeybinding(tui.KeyArrowRight, func() { list.SetFocused(true); library.SetFocused(false) })
	// ui.SetKeybinding(tui.KeyArrowLeft, func() { list.SetFocused(false); library.SetFocused(true) })
	if err := ui.Run(); err != nil {
		panic(err)
	}
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
