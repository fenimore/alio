# Alio

Alio is a commandline album player! By default it looks for a `Music/` directory
and adds any directory with audio files to the music library. From there you can
play your music from the comfort of the commandline, with familiar **emacs** style
keybindings.

If you don't store your music locally, this isn't the right player for you :)

Alio is written in **Go** but relies on C bindings for the `libvlc v2.x` library. Check out the releases tap to get the latest executable _dependency free_ (currently only compiled for Linux). Otherwise building it should be as simple as (if you have libvlc installed):

    go build alio.go
    # to install:
    go install github.com/fenimore/alio


And to run it,

    alio -dir=/home/USER/Music


![Alio](/screenshot.png?raw=true)

## Flags

- `-h       ` for help
- `-dir     ` music collection directory (default "Music")
- `-nocolor ` remove color highlighting

## Keybindings (Emacs with some Vim bonuses)

```
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
```

## Dependencies:

Alio depends on `libvlc 2.X` for _build_, but the **releases** are dependency free).

- [libvlc-go](https://github.com/adrg/libvlc-go) - MIT
- [tui-go](https://github.com/marcusolsson/tui-go/) - MIT

## License

```
Alio commandline music player
Copyright (C) 2017 - 2018  Fenimore Love

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
```

## UI todos

- [ ] Scrollable song list
- [ ] Theme options
- [ ] Skip `n` rows

## The name Alio

Alio was formed from a typo when developing. The original name was `Adio`, which itself
was an intentional typo for Audio. I figured `alio` makes sense because it is intended to
play **al**bums.
