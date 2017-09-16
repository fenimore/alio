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
Scroll: Up, Down Arrow
Play album: Enter, Tab
Pause: p
Next Song: Right arrow, Ctrl-f, l
Previous Song: Left arrow, Ctrl-b, h

looks for a Music/ directory
    or use -dir flag to designate directory name

  -debug
        log messages in debug.log
  -dir string
        music collection directory (default "Music")
  -nocolor
        don't use color highlighting
```

![Alio](/screenshot.png?raw=true)


### Controls (Hybrid Emacs and Vim)

- `Ctrl-n`/`j` + `Ctrl-p`/`k` previous and next album.
- `Ctrl-b`/`h` + `Ctrl-f`/`l` forward and back song.
- `Ctrl-c` `Esc` and `q` to quit.
-  `Up` and `Down` arrow for scrolling library
- `Tab` and `Enter` to play an album.
- `p` to pause.

### UI todos

- [x] Scrollable view (mvp)
- [ ] theme options
