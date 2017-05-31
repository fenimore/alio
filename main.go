package main

import (
	"fmt"
	vlc "github.com/fenimore/libvlc-go"
	"github.com/nsf/termbox-go"
	"io/ioutil"
	"log"
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
	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	// Keybindings
	events := make(chan termbox.Event)
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	tbprint(0, 0, termbox.ColorMagenta, termbox.ColorDefault, "Wooto")
	tbprint(0, 1, termbox.ColorMagenta, termbox.ColorDefault, "Wooto")
	// render
	for _, _ = range albums {

	}
	for {
		ev := <-events
		if ev.Key == termbox.KeyCtrlN {
			log.Println("Events")
		} else if ev.Key == termbox.KeyEsc {
			return
		}
	}

}

// Function tbprint draws a string.
func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func playAlbum(p *vlc.ListPlayer, a *Album) error {
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
