# Alio

Depends on `libvlc 2.X` for building (but the release should be compiled with the libvlc).

```
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

  -d string
        default music collection directory (default "Music")
```

![Alio](/screenshot.png?raw=true)


## Controls

- `Ctrl-n` + `Ctrl-p` emacs style next line prev line navigating the album collection.
- `Ctrl-c` `Esc` and `q` for quitting.
- `Space` and `Enter` for playing an album.

## TODO:

- [ ] Sort 1, 2 ... 10

### Audio playback

There are some functions missing from the `libvlc` c-go bindings which I need to add:

- [x] Prev and Next Song in list
- [x] Pause songs
- [ ] Get Song Position from Player

### UI

- [ ] Adjust the stretch of columns
- [ ] Update the progress bar
- [ ] Display song name in status bar
- [ ] Display keybindings on help
