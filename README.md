# Alio

## Commandline album player

Depends on `libvlc 2.X` for _build_ (but the **releases** are dependency free).

```
         _  _
     /\ | |(_)
    /  \| | _  ___
   / /\ \ || |/ _ \
  / ____ \|| | (_)
 /_/    \_\|_|\___/

Commandline Album Player!

Keybinding (Hybrid Emacs + Vim):
Quit: q, Ctrl-c, Esc
Move down: Ctrl-n, j
Move up: Ctrl-p, k
Play album: Enter, Tab
Pause: p (coming soon: Space)
Next Song: Right arrow, Ctrl-f, l
Previous Song: Left arrow, Ctrl-b, h

looks for a Music/ directory
    or use -dir flag to designate directory name

  -debug
        log messages in debug.log
  -dir string
        default music collection directory (default "Music")
  -nocolor
        don't use color highlighting
```

![Alio](/screenshot.png?raw=true)


### Controls (Hybrid Emacs and Vim)

- [x] `Ctrl-n`/`j` + `Ctrl-p`/`k` emacs style next line prev line navigating the album collection.
- [x] `Ctrl-c` `Esc` and `q` for quitting.
- [x] `Tab` and `Enter` for playing an album.
- [x] `p` pauses.
- [ ] `Ctrl-f Ctrl-b`.
- [ ] `SPC`.

### Audio playback

There are some functions missing from the `libvlc` c-go bindings which I need to add:

- [x] Next Song in list
- [ ] Prev Song in list
- [x] Pause songs
- [x] Get Song Position and Title from Player

### UI

- [ ] Adjust the stretch of columns
- [ ] Scrollable view
- [x] Update the progress bar
- [x] Display song name in status bar
- [x] Display keybindings on help
- [ ] Get "Space" and "Ctrl-F" keybindings
