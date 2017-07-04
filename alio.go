package main

import (
	"flag"
	"fmt"
	vlc "github.com/adrg/libvlc-go"
	"github.com/marcusolsson/tui-go"
	"io/ioutil"
	_ "log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var DIR = "Music"
var MusicExts = ".mp3 .ogg .m4a .flac"
var PhotoExts = ".jpg .jpeg .png"

var ui tui.UI
var wg *sync.WaitGroup

type Task int

const (
	Pause Task = iota
	Play
	Next
	Prev
)

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
	flag.Usage = func() {
		fmt.Println(`
	   _ _
     /\   | (_)
    /  \  | |_  ___
   / /\ \ | | |/ _ \
  / ____ \| | | (_) |
 /_/    \_\_|_|\___/

Commandline album player!

Keybinding:
Quit: q, Ctrl-c, Esc
Move down: Ctrl-n, j
Move up: Ctrl-p, k
Play album: Enter, Tab
Pause: p (coming soon: Space)
Next Song: Right arrow, Ctrl-f
Previous Song: Left arrow, Ctrl-b

Launch application in pwd of a Music/ directory
or use -d flag to designate directory name
`)
		flag.PrintDefaults()
		os.Exit(0)
	}
	DIR = *flag.String("d", "Music", "default music collection directory")
	flag.Parse()
	// Collect Albums from FileSystem root Music
	albums, err := CollectAlbums(DIR)
	if err != nil {
		panic(err)
	}

	wg = new(sync.WaitGroup)

	if len(albums) < 1 {
		flag.Usage()
	}

	// Set Up vlc connection
	err = vlc.Init("--no-video", "--quiet")
	if err != nil {
		panic(err)
	}
	player, err := vlc.NewPlayer()
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

	progress := tui.NewProgress(1000)
	progress.SetCurrent(0)

	status := tui.NewStatusBar(`		คɭเ๏ :: {} ς๏ɭɭєςՇเ๏ภร`)
	status.SetPermanentText(`run alio -h for keybindings`)

	libTable.OnSelectionChanged(func(t *tui.Table) {
		if libTable.Selected() == 0 {
			return
		}
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
		tui.NewSpacer(),
		progress,
		status,
		tui.NewSpacer(),
	)

	ui = tui.New(root)

	// move the cursor to the first album
	libTable.Select(1)

	// navigation
	down := func() {
		if libTable.Selected() == len(albums) {
			return
		}
		libTable.Select(libTable.Selected() + 1)
	}
	up := func() {
		if libTable.Selected() == 1 {
			return
		}
		libTable.Select(libTable.Selected() - 1)
	}
	//ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{}, 128)
	forward := make(chan struct{}, 128)
	previous := make(chan struct{}, 128)
	play := func() {

		s := libTable.Selected() - 1
		if s == -1 {
			return
		}

		if player.IsPlaying() {
			err = player.Stop()
			// TODO: handle error?
			// if err != nil {
			//	panic("Stop" + err.Error())
			// }
			done <- struct{}{}
		}
		go func() {
			wg.Wait()
			wg.Add(1)
			err = playAlbum(
				done,
				player,
				*albums[s],
				list,
				status,
				forward,
				previous,
				0,
			)
			wg.Done()
			// do nothing with the error
		}()

	}

	pause := func() {
		err = player.TogglePause()
		if err != nil {
			// panic("Pause " + err.Error())
		}
	}
	// Key bindings
	ui.SetKeybinding("q", func() { ui.Quit() })
	ui.SetKeybinding("Esc", func() { ui.Quit() })
	ui.SetKeybinding("Ctrl+c", func() { ui.Quit() })

	ui.SetKeybinding("Ctrl+n", down)
	ui.SetKeybinding("j", down)
	ui.SetKeybinding("Ctrl+p", up)
	ui.SetKeybinding("k", func() {
		up()
		fmt.Println("K")
	})

	ui.SetKeybinding("Enter", play)
	ui.SetKeybinding("Tab", play)
	ui.SetKeybinding("p", pause)

	ui.SetKeybinding("Right", func() { forward <- struct{}{} })
	ui.SetKeybinding("Ctrl-F", func() { forward <- struct{}{} })
	ui.SetKeybinding("Ctrl-f", func() { forward <- struct{}{} })
	ui.SetKeybinding("Left", func() { previous <- struct{}{} })
	ui.SetKeybinding("Ctrl-b", func() { previous <- struct{}{} })

	// update goroutine // TODO: must end?
	go func() {
		for {
			time.Sleep(40 * time.Millisecond)
			if player.IsPlaying() {
				length, err := player.MediaLength()
				if err != nil {
					continue
				}

				position, err := player.MediaPosition()
				if err != nil {
					continue
				}
				ui.Update(func() {
					progress.SetCurrent(int(position * 1000))
					// position is a float between 0 and 1
					time.Sleep(time.Millisecond * 50)
					status.SetText(timestamp(position, length))
				})
			}
		}
	}()

	if err := ui.Run(); err != nil {
		panic("Run " + err.Error())
	}
	fmt.Println("Adios from Alio Music Player!")
}

// in its own goroutine...
func playAlbum(done chan struct{}, p *vlc.Player, a Album, l *tui.List, s *tui.StatusBar, next, prev chan struct{}, first int) (err error) {
	playlist := make([]*vlc.Media, 0)
	for _, path := range a.Paths {
		media, err := vlc.NewMediaFromPath(path)
		if err != nil {
			return err
		}
		playlist = append(playlist, media)
	}

	for idx := range playlist[first:] {
		p.SetMedia(playlist[idx])
		err = p.Play()
		if err != nil {
			return err
		}

		status, err := p.MediaState()
		if err != nil {
			return err
		}
	PlaybackLoop:
		for status != vlc.MediaEnded {
			status, err = p.MediaState()
			// handle err?
			ui.Update(func() {
				song := songStatus(a, l.Selected())
				if song != "" {
					s.SetPermanentText(song)
				}
				l.SetSelected(idx)
			})
			select {
			case <-done:
				return err
			case <-next:
				break PlaybackLoop
			case <-prev:
				if idx < 1 {
					continue
				}
				return playAlbum(done, p, a, l, s, next, prev, idx-1)
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}

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
				song := strings.TrimSuffix(songs[i], filepath.Ext(songs[i]))
				album.Songs = append(album.Songs, song)
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

func timestamp(pos float32, length int) string {
	r := float32(length) * pos
	d := int(r / 1000)
	seconds := d % 60
	minutes := int(d/60) % 60
	hours := int(minutes / 60)
	duration := fmt.Sprintf(
		"%02d:%02d:%02d",
		hours,
		minutes,
		seconds,
	)
	d = int(length / 1000)
	seconds = d % 60
	minutes = int(d/60) % 60
	hours = int(minutes / 60)

	total := fmt.Sprintf(
		"%02d:%02d:%02d",
		hours,
		minutes,
		seconds,
	)
	return fmt.Sprintf(`  %s - %s`,
		duration,
		total,
	)
}

// song index and album index, for UI
func songStatus(a Album, idx int) string {
	if idx == -1 {
		return ""
	}

	return fmt.Sprintf("%s / %s",
		a.Title,
		a.Songs[idx],
	)
}
