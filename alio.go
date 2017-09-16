package main

import (
	"flag"
	"fmt"
	vlc "github.com/adrg/libvlc-go"
	"github.com/marcusolsson/tui-go"
	"io/ioutil"
	"log"
	_ "log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var MusicExts = ".mp3 .ogg .m4a .flac"
var PhotoExts = ".jpg .jpeg .png"

var ui tui.UI
var wg *sync.WaitGroup // Wrapped with two waitgroups?
var pg *sync.WaitGroup // A single playback goroutine
var lock *sync.Mutex

type Album struct {
	Title string
	Songs []string
	Paths []string
	Cover string
	Count int
	Index int
}

func (a *Album) String() string {
	return fmt.Sprintf("%s | %d", a.Title, a.Count)
}

var help = `
	 _   _
     /\ | | (_)
    /  \| |  _  ___
   / /\ \ | | |/ _ \
  / ____ \| | | (_)
 /_/    \_\ |_|\___/

Commandline Album Player!

Keybinding (Hybrid Emacs + Vim):
Quit: q, Ctrl-c, Esc
Move down: Ctrl-n, j
Move up: Ctrl-p, k
Scroll: Up, Down Arrow
Play album: Enter, Tab
Pause: p (coming soon: Space)
Next Song: Right arrow, Ctrl-f, l
Previous Song: Left arrow, Ctrl-b, h

looks for a Music/ directory
    or use -dir flag to designate directory name
`

func main() {
	flag.Usage = func() {
		fmt.Printf("%s\n\n%s\n", help, "-h for help")
		flag.PrintDefaults()
		os.Exit(0)
	}
	DIR := flag.String("dir", "Music",
		"music collection directory")
	DEBUG := flag.Bool("debug", false,
		"log messages in debug.log")
	NOTHEME := flag.Bool("nocolor", false,
		"don't use color highlighting")
	flag.Parse()

	if *DEBUG {
		logfile, err := os.Create("debug.log")
		if err == nil {
			log.SetOutput(logfile)
		}
	} else {
		log.SetOutput(ioutil.Discard)
	}
	log.Printf("Flags: debug: %t | dir: %s | nocolor: %t",
		*DEBUG, *DIR, *NOTHEME)

	// Collect Albums from FileSystem root Music
	albums, err := CollectAlbums(*DIR)
	if err != nil {
		panic(err)
	}

	wg = new(sync.WaitGroup)
	pg = new(sync.WaitGroup)
	lock = new(sync.Mutex)

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
	libTable.SetColumnStretch(0, 0)

	aLabel := tui.NewLabel("Albums")
	sLabel := tui.NewLabel("Songs")
	aLabel.SetStyleName("album")
	sLabel.SetStyleName("album")
	wrap := tui.NewScrollArea(libTable)
	library := tui.NewVBox(
		aLabel,
		wrap,
		tui.NewSpacer(),
	)
	library.SetBorder(false)

	list := tui.NewList()
	songList := tui.NewVBox(
		sLabel,
		list,
		tui.NewSpacer(),
	)
	songList.SetBorder(false)
	//	albums = albums[10:]
	for idx := range albums {
		libTable.AppendRow(
			tui.NewLabel(albums[idx].Title),
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
		tui.NewPadder(1, 0, songList),
	)

	root := tui.NewVBox(
		selection,
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
	ui.SetKeybinding("Up", func() {
		wrap.Scroll(0, -1)
		up()
	})
	ui.SetKeybinding("Down", func() {
		wrap.Scroll(0, 1)
		down()
	})
	// TODO: focus command to go to current playing album
	// controls
	done := make(chan struct{}, 1)
	forward := make(chan struct{}, 1)
	previous := make(chan struct{}, 1)
	play := func() {

		s := libTable.Selected() - 1
		if s == -1 {
			return
		}
		lock.Lock()
		state, err := player.MediaState()
		if err != nil {
			log.Printf("State err: %s in inline play function", err)
			return
		}
		log.Printf("State: %d", state)
		lock.Unlock()
		if state == vlc.MediaPaused ||
			state == vlc.MediaPlaying {
			lock.Lock()
			err = player.Stop()
			lock.Unlock()
			if err != nil {
				log.Printf("Stop err: %s", err)
			}
			// TODO: handle error?
			done <- struct{}{}
		}
		go func() {
			wg.Wait()
			wg.Add(1)
			err = playAlbum(
				player,
				*albums[s],
				list,
				libTable,
				status,
				done,
				forward,
				previous,
			)
			wg.Done()
			// do nothing with the error
		}()

	}

	pause := func() {
		lock.Lock()
		err = player.TogglePause()
		lock.Unlock()
		if err != nil {
			log.Printf("Pause err: %s", err)
		}
	}
	// Key bindings
	ui.SetKeybinding("q", func() {
		ui.Quit()
		log.Printf("Done [q]: ui.Quit")
		err = player.Release()
		log.Printf("Done [q]: Release")
		if err != nil {
			log.Printf("VLC Release err: %s", err)
		}
	})
	ui.SetKeybinding("Esc", func() { ui.Quit() })
	ui.SetKeybinding("Ctrl+c", func() { ui.Quit() })

	ui.SetKeybinding("Ctrl+n", down)
	ui.SetKeybinding("Ctrl+p", up)

	ui.SetKeybinding("Enter", play)
	ui.SetKeybinding("Tab", play)
	ui.SetKeybinding("p", pause)
	ui.SetKeybinding("Space", pause)

	back := func() {
		select {
		case previous <- struct{}{}:
		default:
			log.Print("Forward Press Default")
		}
	}
	next := func() {
		select {
		case forward <- struct{}{}:
		default:
			log.Print("Forward Press Default")
		}
	}
	ui.SetKeybinding("Right", next)
	ui.SetKeybinding("Ctrl+f", next)
	ui.SetKeybinding("l", next)
	ui.SetKeybinding("Left", back)
	ui.SetKeybinding("h", back)
	ui.SetKeybinding("Ctrl+b", back)

	// update goroutine // TODO: must end?
	go func() {
		for {
			lock.Lock()
			playing := player.IsPlaying()
			lock.Unlock()
			if playing {
				lock.Lock()
				length, err := player.MediaLength()
				lock.Unlock()
				if err != nil {
					continue
				}
				lock.Lock()
				position, err := player.MediaPosition()
				lock.Unlock()
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

	// Theme
	theme := tui.NewTheme()
	theme.SetStyle("table.cell", tui.Style{
		Fg: tui.ColorCyan,
	})
	theme.SetStyle("table.cell.selected", tui.Style{
		Bg: tui.ColorCyan, Fg: tui.ColorWhite,
	})
	theme.SetStyle("list.item", tui.Style{
		Fg: tui.ColorGreen,
	})
	theme.SetStyle("list.item.selected", tui.Style{
		Bg: tui.ColorGreen, Fg: tui.ColorWhite,
	})
	theme.SetStyle("label.album", tui.Style{
		Bg: tui.ColorDefault, Fg: tui.ColorYellow,
	})
	if !*NOTHEME {
		ui.SetTheme(theme)
	}
	log.Println(libTable.MinSizeHint())
	log.Println(libTable.Selected())
	log.Println(libTable.Size())
	log.Println(libTable.SizeHint())

	if err := ui.Run(); err != nil {
		panic("Run " + err.Error())
	}
	log.Printf("Ending Program")
	fmt.Println("Adios from Alio Music Player!")
}

// in its own goroutine...
func playAlbum(p *vlc.Player, a Album, l *tui.List, t *tui.Table, s *tui.StatusBar, done, next, prev chan struct{}) (err error) {
	log.Printf("PlayAlbum %s", a.Title)
	pg.Wait()
	pg.Add(1)
	defer pg.Done()

	playlist := new(PlayList)
	for _, path := range a.Paths {
		media, err := vlc.NewMediaFromPath(path)
		if err != nil {
			return err
		}
		song := &Song{media, path, playlist, nil, nil, 0}
		playlist.append(song)
	}

	current := playlist.head
	for current != nil {
		lock.Lock()
		p.SetMedia(current.media)
		err = p.Play()
		lock.Unlock()
		if err != nil {
			return err
		}

		lock.Lock()
		status, err := p.MediaState()
		lock.Unlock()
		if err != nil {
			return err
		}
		log.Printf("Playback start %d/%d",
			current.index, playlist.size)
	PlaybackLoop:
		for status != vlc.MediaEnded {
			lock.Lock()
			status, err = p.MediaState()
			lock.Unlock()
			if err != nil {
				log.Println("Failure", err)
				return err
			}
			song := songStatus(a, current.index)
			if song != "" {
				s.SetPermanentText(song)
			}

			if t.Selected() == a.Index {
				l.SetSelected(current.index)
			} else {
				l.SetSelected(-1)
			}
			select {
			case <-next:
				log.Println("Recv on Next")
				current = current.next
				break PlaybackLoop
			case <-prev:
				log.Println("Recv on Prev")
				if current.prev != nil {
					current = current.prev
				}
				break PlaybackLoop
			case <-done:
				log.Println("Return Done")
				return err
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
		if status == vlc.MediaEnded {
			current = current.next
		}
		log.Printf("Playback end")
	}
	log.Println("End of playlist loop")
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
			album.Index = len(albums) + 1
			albums = append(albums, album)
		}
	}

	log.Printf("Album Lenght %d", len(albums))
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

/////////////////////////////////////////////
// Linked list for iterating over an album //
/////////////////////////////////////////////
type Song struct {
	media *vlc.Media
	path  string
	list  *PlayList
	next  *Song
	prev  *Song
	index int
}

type PlayList struct {
	head *Song
	last *Song
	size int
	lock *sync.Mutex
}

func (l *PlayList) append(s *Song) {
	if l.head == nil {
		s.index = 0
		l.head, l.last = s, s
	} else {
		s.index = l.last.index + 1
		l.last.next = s
		s.prev = l.last
		l.last = s
	}
	l.size++
}
