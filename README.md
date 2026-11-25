# WarDogs Keybind Editor

The in-game keybind menu in **WarDogs** won't let you put one key on two actions. This tool edits your settings file directly so you can, for example Crouch and Crouch Prone both on `C`, or a second key for Slow Walk.

![The editor with Crouch and Crouch Prone stacked on C](docs/screenshot.png)

## Download

Get **WD-Keybind-Editor.exe** from the [Releases](../../releases) page.

## How to use

1. Close WarDogs.
2. Run `WD-Keybind-Editor.exe`. A small console window opens and the editor opens in your web browser. Keep the console window open while you edit.
3. Your settings load automatically. Click any key to rebind it: press a key, click a mouse button, or scroll. The same key can go on as many actions as you like.
4. Click **Save to game**, then start WarDogs.
5. Close the console window when you're done.

## What it does (and doesn't do)

- Lists all 105 keyboard & mouse actions from the game's keybind menu, with their defaults, a reset button for each, and **Reset all to default**.
- Hold / Toggle settings for crouch, prone, sprint, ADS and the rest.
- Only ever touches one file: `%LOCALAPPDATA%\Wardogs\Saved\Config\WindowsClient\GameUserSettings.ini`.
- Makes a backup before every save (`GameUserSettings.backup-<date>.ini`), and keeps your very first file as `GameUserSettings.original-backup.ini`. To undo everything, rename that one back to `GameUserSettings.ini`.
- Runs only on your PC (`127.0.0.1`). It doesn't connect to the internet or send anything anywhere.
- Doesn't touch the game itself, its files or memory, just the settings file the game's own menu writes.
- Keyboard & mouse only for now. HOTAS and gamepad binds are left alone.

## Notes

- **"Windows protected your PC":** the exe isn't code-signed, so Windows warns about it. Click **More info → Run anyway**. You can check the code here or build it yourself (below).
- If you change binds in the game's own keybind menu afterwards, the game may undo your stacked keys. Just save again from the editor.
- Defaults were recorded from the game's keybind menu on Nov 24, 2026. A game update could add or rename actions.

## Build it yourself

Needs [Go](https://go.dev/dl/) 1.24 or newer.

```
go build -ldflags="-s -w" -o WD-Keybind-Editor.exe .
```

`main.go` is the small local server that reads and saves the file. `index.html` is the whole editor.

---

Fan-made tool, not affiliated with the WarDogs developers.
