# Alio

Depends on `libvlc 2.X` for building (but the release should be compiled with the libvlc).

```
         _  _
     /\ | |(_)
    /  \| | _  ___
   / /\ \ || |/ _ \
  / ____ \|| | (_)
 /_/    \_\|_|\___/

Commandline Album Player!

Keybinding:
Quit: q, Ctrl-c, Esc
Move down: Ctrl-n, j
Move up: Ctrl-p, k
Play album: Enter, Tab
Pause: p (coming soon: Space)
Next Song: Right arrow, Ctrl-f
Previous Song: Left arrow, Ctrl-b

Application looks for a Music/ directory
or use -dir flag to designate directory name

  -debug
        log messages in debug.log
  -dir string
        default music collection directory (default "Music")
  -nocolor
        don't use color highlighting
```

![Alio](/screenshot.png?raw=true)


## Controls

- `Ctrl-n`/`j` + `Ctrl-p`/`k` emacs style next line prev line navigating the album collection.
- `Ctrl-c` `Esc` and `q` for quitting.
- `Tab` and `Enter` for playing an album.
- `p` pauses.

## TODO:

- [ ] Sort 1, 2 ... 10

### Audio playback

There are some functions missing from the `libvlc` c-go bindings which I need to add:

- [x] Next Song in list
- [ ] Prev Song in list
- [x] Pause songs
- [x] Get Song Position and Title from Player

### UI

- [ ] Adjust the stretch of columns
- [x] Update the progress bar
- [x] Display song name in status bar
- [x] Display keybindings on help
- [ ] Get "Space" and "Ctrl-F" keybindings
