package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	_ "log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	vlc "github.com/adrg/libvlc-go"
	tui "github.com/fenimore/tui-go"
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

Keybinding (Emacs with some Vim bonuses):

Quit:         Ctrl-c | q | Esc

Move down:    Ctrl-n | j
Move up:      Ctrl-p | k
Scroll:       Up / Down
Page down:    Ctrl-v
Page up:      Alt-v

Focus cursor: Ctrl-l
Pause:        p
Play album:   Enter | Tab
Next:         Right | Ctrl-f | l
Previous:     Left  | Ctrl-b | h

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

	var currentAlbum int
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
	for idx := range albums {
		libTable.AppendRow(
			tui.NewLabel(albums[idx].Title),
		)
	}
	aLabel := tui.NewLabel("Albums")
	sLabel := tui.NewLabel("Songs")
	aLabel.SetStyleName("album")
	sLabel.SetStyleName("album")
	wrap := tui.NewScrollArea(libTable)
	wrap.SetSizePolicy(tui.Expanding, tui.Expanding)
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

	progress := tui.NewProgress(1000)
	progress.SetCurrent(0)

	status := tui.NewStatusBar(`		คɭเ๏ :: {} ς๏ɭɭєςՇเ๏ภร`)
	status.SetPermanentText(`run: alio -h for help`)

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

	ui, err = tui.New(root)
	if err != nil {
		panic("Ui didn't build?")
	}

	// move the cursor to the first album
	libTable.Select(1)

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
			currentAlbum = s
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
			if err != nil {
				log.Printf("Error playing %s", err)
			}
		}()

	}

	// navigation
	pause := func() {
		lock.Lock()
		err = player.TogglePause()
		lock.Unlock()
		if err != nil {
			log.Printf("Pause err: %s", err)
		}
	}
	exit := func() {
		ui.Quit()
		log.Printf("Done [q]: ui.Quit")
		err = player.Release()
		log.Printf("Done [q]: Release")
		if err != nil {
			log.Printf("VLC Release err: %s", err)
		}
	}
	// Quitting
	ui.SetKeybinding("q", exit)
	ui.SetKeybinding("Esc", exit)
	ui.SetKeybinding("Ctrl+c", exit)
	scrolledDown := 0
	scrolledUp := 0
	down := func(scroll bool) {
		if libTable.Selected() == len(albums) {
			return
		} else if libTable.Selected()-scrolledDown >= wrap.Size().Y {
			scroll = true
		}
		if scroll {
			scrolledUp--
			scrolledDown++
			wrap.Scroll(0, 1)
		}
		libTable.Select(libTable.Selected() + 1)
	}
	up := func(scroll bool) {
		if libTable.Selected() == 1 {
			return
		} else if scrolledDown == 0 {
			scroll = false
		} else if scrolledUp+libTable.Selected() == 1 {
			scroll = true
		}
		if scroll {
			scrolledUp++
			scrolledDown--
			wrap.Scroll(0, -1)
		}
		libTable.Select(libTable.Selected() - 1)
	}
	pageDown := func() {
		for i := 1; i <= wrap.Size().Y/2; i++ {
			down(false)
		}
	}
	pageUp := func() {
		for i := 1; i <= wrap.Size().Y/2; i++ {
			up(false)
		}
	}
	// Album selection
	ui.SetKeybinding("Ctrl+n", func() { down(false) })
	ui.SetKeybinding("Ctrl+p", func() { up(false) })
	ui.SetKeybinding("Up", func() { up(true) })
	ui.SetKeybinding("Down", func() { down(true) })
	ui.SetKeybinding("k", func() { up(true) })
	ui.SetKeybinding("j", func() { down(true) })

	ui.SetKeybinding("Ctrl+v", func() { pageDown() })
	ui.SetKeybinding("Alt+v", func() { pageUp() })

	ui.SetKeybinding("Ctrl+l", func() {
		lowerBound := 0 - scrolledUp
		upperBound := wrap.Size().Y - scrolledUp
		position := currentAlbum
		if position < lowerBound {
			for i := 0; i < lowerBound-position; i++ {
				up(true)
			}
		} else if position > upperBound {
			for i := 0; i < lowerBound-position; i++ {
				down(true)
			}
		}
		libTable.Select(currentAlbum + 1)
	})
	// Playback
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
					// position is a float between 0 and 1
					progress.SetCurrent(int(position * 1000))
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
		ticker := time.NewTicker(time.Millisecond * 50)
		defer ticker.Stop()
	PlaybackLoop:
		for status != vlc.MediaEnded {
			select {
			case <-next:
				log.Println("Recv on Next")
				if current.next != nil {
					current = current.next
					break PlaybackLoop
				}
			case <-prev:
				log.Println("Recv on Prev")
				if current.prev != nil {
					current = current.prev
				}
				break PlaybackLoop
			case <-done:
				log.Println("Return Done")
				return err
			case <-ticker.C:
				lock.Lock()
				status, err = p.MediaState()
				lock.Unlock()
				if err != nil {
					log.Printf("Failure Getting Media State: %s", err)
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
