package main

import (
	"fmt"
	vlc "github.com/adrg/libvlc-go"
	"github.com/marcusolsson/tui-go"
	"io/ioutil"
	_ "log"
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
	// Collect Albums from FileSystem root Music
	albums, err := CollectAlbums(DIR)
	if err != nil {
		panic(err)
	}

	// Set Up vlc connection
	err = vlc.Init("--no-video", "--quiet")
	if err != nil {
		panic(err)
	}
	player, err := vlc.NewListPlayer()
	if err != nil {
		fmt.Printf("VLC init Error: [%s]\nAre you using libvlc 2.x?\n", err)
		panic(err)
	}
	defer func() {
		vlc.Release()
		player.Stop()
		player.Release()
	}()

	// Set up UI
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
	status.SetPermanentText("Volume")

	libTable.OnSelectionChanged(func(t *tui.Table) {
		if libTable.Selected() == 0 {
			return
		}
		progress.SetCurrent(libTable.Selected() * 5)
		list.RemoveItems()

		for _, v := range albums[libTable.Selected()-1].Songs {
			list.AddItems(v)
		}
	})

	//go func() {
	// if player.IsPlaying()
	// TODO: Update progress.SetCurrent()
	//}()

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

	// move the cursor to the first album
	libTable.Select(1)

	// Key bindings
	ui.SetKeybinding("q", func() { ui.Quit() })
	ui.SetKeybinding("Esc", func() { ui.Quit() })
	ui.SetKeybinding("Ctrl+c", func() { ui.Quit() })
	ui.SetKeybinding("Ctrl+n", func() {
		if libTable.Selected() == len(albums) {
			return
		}
		libTable.Select(libTable.Selected() + 1)
	})
	ui.SetKeybinding("Ctrl+p", func() {
		if libTable.Selected() == 1 {
			return
		}
		libTable.Select(libTable.Selected() - 1)
	})

	play := func() {
		s := libTable.Selected() - 1
		if s == -1 {
			return
		}

		err = playAlbum(player, albums[s])
		if err != nil {
			panic(err)
		}
		// fmt.Println("Playing: ", albums[s].Title)
	}
	// TODO: Add bindings for next and prev song.
	ui.SetKeybinding("Enter", play)
	ui.SetKeybinding("Space", play)

	if err := ui.Run(); err != nil {
		panic(err)
	}
	fmt.Println("Adios from Alio Music Player!")
}

func playAlbum(p *vlc.ListPlayer, a *Album) (err error) {
	if p.IsPlaying() {
		p.Stop()
	}
	list, err := vlc.NewMediaList()
	if err != nil {
		return err
	}

	for _, path := range a.Paths {
		media, err := vlc.NewMediaFromPath(path)
		if err != nil {
			return err
		}

		err = list.AddMedia(media)
		if err != nil {
			return err
		}
	}

	err = p.SetMediaList(list)
	if err != nil {
		return err
	}

	err = p.Play()
	return err
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
			album.Title = filepath.Base(val)
			album.Count = len(songs)
			albums = append(albums, album)
		}
	}

	return albums, nil
}
