package main

import (
	"errors"
	"github.com/glycerine/zygomys/repl"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
)

func lispPrint(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return zygo.SexpNull, zygo.WrongNargs
	}
	switch t := args[0].(type) {
	case *zygo.SexpStr:
		Global.Input = t.S
	default:
		return zygo.SexpNull, errors.New("Arg needs to be a string")
	}
	return zygo.SexpNull, nil
}

func lispRunCommand(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return zygo.SexpNull, zygo.WrongNargs
	}
	switch t := args[0].(type) {
	case *zygo.SexpStr:
		cn := StrToCmdName(t.S)
		cmd := funcnames[cn]
		if cmd != nil && cmd.Com != nil {
			cmd.Com(env)
		}
	default:
		return zygo.SexpNull, errors.New("Arg needs to be a string")
	}
	return zygo.SexpNull, nil
}

func lispSingleton(f func()) zygo.GlispUserFunction {
	return func(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
		f()
		return zygo.SexpNull, nil
	}
}

func cmdAndLispFunc(e *zygo.Glisp, cmdname, lispname string, f func()) {
	e.AddFunction(lispname, lispSingleton(f))
	DefineCommand(&CommandFunc{cmdname, func(env *zygo.Glisp) { f() }})
}

func lispBindKey(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) < 2 {
		return zygo.SexpNull, zygo.WrongNargs
	}
	var arg1 string
	switch t := args[0].(type) {
	case *zygo.SexpStr:
		arg1 = t.S
	default:
		return zygo.SexpNull, errors.New("Arg 1 needs to be a string")
	}
	var arg2 *zygo.SexpFunction
	switch t := args[1].(type) {
	case *zygo.SexpFunction:
		arg2 = t
	case *zygo.SexpStr:
		cmdname := StrToCmdName(t.S)
		cmd := funcnames[cmdname]
		if cmd == nil {
			return zygo.SexpNull, errors.New("Unknown command: " + cmdname)
		} else {
			Emacs.PutCommand(arg1, cmd)
			return zygo.SexpNull, nil
		}
	default:
		return zygo.SexpNull, errors.New("Arg 2 needs to be a string or function")
	}
	av := []zygo.Sexp{}
	if len(args) > 2 {
		av = args[2:]
	}
	Emacs.PutCommand(arg1, &CommandFunc{"lisp code", func(env *zygo.Glisp) {
		env.Apply(arg2, av)
	}})
	return zygo.SexpNull, nil
}

func lispDefineCmd(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) < 2 {
		return zygo.SexpNull, zygo.WrongNargs
	}
	var arg1 string
	switch t := args[0].(type) {
	case *zygo.SexpStr:
		arg1 = StrToCmdName(t.S)
	default:
		return zygo.SexpNull, errors.New("Arg 1 needs to be a string")
	}
	var arg2 *zygo.SexpFunction
	switch t := args[1].(type) {
	case *zygo.SexpFunction:
		arg2 = t
	default:
		return zygo.SexpNull, errors.New("Arg 2 needs to be a function")
	}
	av := []zygo.Sexp{}
	if len(args) > 2 {
		av = args[2:]
	}
	DefineCommand(&CommandFunc{arg1, func(env *zygo.Glisp) {
		env.Apply(arg2, av)
	}})
	return zygo.SexpNull, nil
}

func lispOnlyWindow(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	return zygo.GoToSexp(len(Global.Windows) == 1, env)
}

func lispSetTabStop(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return zygo.SexpNull, zygo.WrongNargs
	}
	var x int
	switch t := args[0].(type) {
	case *zygo.SexpInt:
		x = int(t.Val)
	default:
		return zygo.SexpNull, errors.New("Arg 1 needs to be an int")
	}
	Global.Tabsize = x
	return zygo.SexpNull, nil
}

func lispSetSoftTab(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return zygo.SexpNull, zygo.WrongNargs
	}
	var x bool
	switch t := args[0].(type) {
	case *zygo.SexpBool:
		x = bool(t.Val)
	default:
		return zygo.SexpNull, errors.New("Arg 1 needs to be a bool")
	}
	Global.SoftTab = x
	return zygo.SexpNull, nil
}

func lispSetSyntaxOff(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return zygo.SexpNull, zygo.WrongNargs
	}
	var x bool
	switch t := args[0].(type) {
	case *zygo.SexpBool:
		x = bool(t.Val)
	default:
		return zygo.SexpNull, errors.New("Arg 1 needs to be a bool")
	}
	Global.NoSyntax = x
	return zygo.SexpNull, nil
}

func lispGetTabStr(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	return zygo.GoToSexp(getTabString(), env)
}

func lispAddDefaultMode(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) == 1 {
		var modename string
		switch t := args[0].(type) {
		case *zygo.SexpStr:
			modename = StrToCmdName(t.S)
		default:
			return zygo.SexpNull, errors.New("Arg needs to be a string")
		}
		addDefaultMode(modename)
		return zygo.SexpNull, nil
	}
	return zygo.SexpNull, zygo.WrongNargs
}

func lispRemDefaultMode(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) == 1 {
		var modename string
		switch t := args[0].(type) {
		case *zygo.SexpStr:
			modename = StrToCmdName(t.S)
		default:
			return zygo.SexpNull, errors.New("Arg needs to be a string")
		}
		remDefaultMode(modename)
		return zygo.SexpNull, nil
	}
	return zygo.SexpNull, zygo.WrongNargs
}

func lispSetMode(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) == 1 {
		var modename string
		switch t := args[0].(type) {
		case *zygo.SexpStr:
			modename = StrToCmdName(t.S)
		default:
			return zygo.SexpNull, errors.New("Arg needs to be a string")
		}
		Global.CurrentB.toggleMode(modename)
		return zygo.SexpNull, nil
	} else if len(args) == 2 {
		var modename string
		switch t := args[0].(type) {
		case *zygo.SexpStr:
			modename = StrToCmdName(t.S)
		default:
			return zygo.SexpNull, errors.New("Arg 1 needs to be a string")
		}
		var enabled bool
		switch t := args[1].(type) {
		case *zygo.SexpBool:
			enabled = bool(t.Val)
		default:
			return zygo.SexpNull, errors.New("Arg 2 needs to be a bool")
		}
		Global.CurrentB.setMode(modename, enabled)
		return zygo.SexpNull, nil
	}
	return zygo.SexpNull, zygo.WrongNargs
}

func lispHasMode(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return zygo.SexpNull, zygo.WrongNargs
	}
	switch t := args[0].(type) {
	case *zygo.SexpStr:
		return zygo.GoToSexp(Global.CurrentB.hasMode(StrToCmdName(t.S)), env)
	default:
		return zygo.SexpNull, errors.New("Arg needs to be a string")
	}
}

func lispListModes(env *zygo.Glisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	return zygo.GoToSexp(Global.CurrentB.getEnabledModes(), env)
}

func loadLispFunctions(env *zygo.Glisp) {
	env.AddFunction("emacsprint", lispPrint)
	cmdAndLispFunc(env, "save-buffers-kill-emacs", "emacsquit", EditorQuit)
	env.AddFunction("emacsbindkey", lispBindKey)
	env.AddFunction("emacsonlywindow", lispOnlyWindow)
	env.AddFunction("settabstop", lispSetTabStop)
	env.AddFunction("gettabstr", lispGetTabStr)
	env.AddFunction("setsofttab", lispSetSoftTab)
	env.AddFunction("disablesyntax", lispSetSyntaxOff)
	env.AddFunction("unbindall", lispSingleton(func() { Emacs.UnbindAll() }))
	env.AddFunction("emacsdefinecmd", lispDefineCmd)
	env.AddFunction("runemacscmd", lispRunCommand)
	env.AddFunction("setmode", lispSetMode)
	env.AddFunction("hasmode", lispHasMode)
	env.AddFunction("listmodes", lispListModes)
	env.AddFunction("adddefaultmode", lispAddDefaultMode)
	env.AddFunction("remdefaultmode", lispRemDefaultMode)
	DefineCommand(&CommandFunc{"describe-key-briefly", func(env *zygo.Glisp) { DescribeKeyBriefly() }})
	DefineCommand(&CommandFunc{"run-command", RunCommand})
	DefineCommand(&CommandFunc{"redo", func(env *zygo.Glisp) { editorRedoAction() }})
	DefineCommand(&CommandFunc{"suspend-emacs", func(env *zygo.Glisp) { suspend() }})
	DefineCommand(&CommandFunc{"move-end-of-line", func(env *zygo.Glisp) { MoveCursorToEol() }})
	DefineCommand(&CommandFunc{"move-beginning-of-line", func(env *zygo.Glisp) { MoveCursorToBol() }})
	DefineCommand(&CommandFunc{"scroll-up-command", func(env *zygo.Glisp) { MoveCursorBackPage() }})
	DefineCommand(&CommandFunc{"scroll-down-command", func(env *zygo.Glisp) { MoveCursorForthPage() }})
	DefineCommand(&CommandFunc{"save-buffer", func(env *zygo.Glisp) { EditorSave() }})
	DefineCommand(&CommandFunc{"delete-char", func(env *zygo.Glisp) { MoveCursor(1, 0); editorDelChar() }})
	DefineCommand(&CommandFunc{"delete-backward-char", func(env *zygo.Glisp) { editorDelChar() }})
	DefineCommand(&CommandFunc{"find-file", func(env *zygo.Glisp) { editorFindFile() }})
	DefineCommand(&CommandFunc{"insert-newline", func(env *zygo.Glisp) { editorInsertNewline() }})
	DefineCommand(&CommandFunc{"isearch", func(env *zygo.Glisp) { editorFind() }})
	DefineCommand(&CommandFunc{"buffers-list", func(env *zygo.Glisp) { editorSwitchBuffer() }})
	DefineCommand(&CommandFunc{"end-of-buffer", func(env *zygo.Glisp) { Global.CurrentB.cy = Global.CurrentB.NumRows }})
	DefineCommand(&CommandFunc{"beginning-of-buffer", func(env *zygo.Glisp) { Global.CurrentB.cy = 0 }})
	DefineCommand(&CommandFunc{"undo", func(env *zygo.Glisp) { editorUndoAction() }})
	DefineCommand(&CommandFunc{"indent", func(env *zygo.Glisp) { editorInsertStr(getTabString()) }})
	DefineCommand(&CommandFunc{"other-window", func(env *zygo.Glisp) { switchWindow() }})
	DefineCommand(&CommandFunc{"delete-window", func(env *zygo.Glisp) { closeThisWindow() }})
	DefineCommand(&CommandFunc{"delete-other-windows", func(env *zygo.Glisp) { closeOtherWindows() }})
	DefineCommand(&CommandFunc{"split-window", func(env *zygo.Glisp) { splitWindows() }})
	DefineCommand(&CommandFunc{"find-file-other-window", func(env *zygo.Glisp) { callFunOtherWindow(editorFindFile) }})
	DefineCommand(&CommandFunc{"switch-buffer-other-window", func(env *zygo.Glisp) { callFunOtherWindow(editorSwitchBuffer) }})
	DefineCommand(&CommandFunc{"set-mark", func(env *zygo.Glisp) { setMark(Global.CurrentB) }})
	DefineCommand(&CommandFunc{"kill-region", func(env *zygo.Glisp) { doKillRegion() }})
	DefineCommand(&CommandFunc{"yank-region", func(env *zygo.Glisp) { doYankRegion() }})
	DefineCommand(&CommandFunc{"copy-region", func(env *zygo.Glisp) { doCopyRegion() }})
	DefineCommand(&CommandFunc{"forward-word", func(env *zygo.Glisp) { moveForwardWord() }})
	DefineCommand(&CommandFunc{"backward-word", func(env *zygo.Glisp) { moveBackWord() }})
	DefineCommand(&CommandFunc{"backward-kill-word", func(env *zygo.Glisp) { delBackWord() }})
	DefineCommand(&CommandFunc{"kill-word", func(env *zygo.Glisp) { delForwardWord() }})
	DefineCommand(&CommandFunc{"recenter-top-bottom", func(env *zygo.Glisp) { editorCentreView() }})
	DefineCommand(&CommandFunc{"kill-buffer", func(env *zygo.Glisp) { killBuffer() }})
	DefineCommand(&CommandFunc{"kill-line", func(env *zygo.Glisp) { killToEol() }})
	DefineCommand(&CommandFunc{"downcase-region", func(*zygo.Glisp) { doLCRegion() }})
	DefineCommand(&CommandFunc{"upcase-region", func(*zygo.Glisp) { doUCRegion() }})
	DefineCommand(&CommandFunc{"upcase-word", func(*zygo.Glisp) { upcaseWord() }})
	DefineCommand(&CommandFunc{"downcase-word", func(*zygo.Glisp) { downcaseWord() }})
	DefineCommand(&CommandFunc{"toggle-mode", func(*zygo.Glisp) {
		mode := editorPrompt("Which mode?", nil)
		Global.CurrentB.toggleMode(StrToCmdName(mode))
	}})
	DefineCommand(&CommandFunc{"show-modes", func(*zygo.Glisp) { showModes() }})
	DefineCommand(&CommandFunc{"indent-mode", func(*zygo.Glisp) { doToggleMode("indent-mode") }})
	DefineCommand(&CommandFunc{"line-number-mode", func(*zygo.Glisp) { doToggleMode("line-number-mode") }})
	DefineCommand(&CommandFunc{"forward-char", func(*zygo.Glisp) { MoveCursor(1, 0) }})
	DefineCommand(&CommandFunc{"backward-char", func(*zygo.Glisp) { MoveCursor(-1, 0) }})
	DefineCommand(&CommandFunc{"next-line", func(*zygo.Glisp) { MoveCursor(0, 1) }})
	DefineCommand(&CommandFunc{"previous-line", func(*zygo.Glisp) { MoveCursor(0, -1) }})
	DefineCommand(&CommandFunc{"describe-bindings", func(*zygo.Glisp) { showMessages(WalkCommandTree(Emacs, "")) }})
	DefineCommand(&CommandFunc{"quick-help", func(*zygo.Glisp) {
		showMessages(`Welcome to Gomacs - Go-powered emacs!

If you've not edited your rc file (~/.gomacs.lisp), here are some emergency
commands that should help you out. C-n means hold Ctrl and press n, M-n means
hold Meta (Alt on modern keyboards) and press n.

- C-x C-c - Save all buffers and quit emacs
- C-x C-s - Save currently selected buffer
- C-x C-f - Open a file (prompt)
- C-@ (control-space) - Set mark to current cursor position
- C-w - Kill (cut) the region (the space between the mark and cursor)
- M-w - Copy the region
- C-y - Yank (paste) the last thing you killed or copied.

Current key bindings:
`, WalkCommandTree(Emacs, ""))
	}})
}

func NewLispInterp() *zygo.Glisp {
	ret := zygo.NewGlisp()
	loadLispFunctions(ret)
	LoadDefaultConfig(ret)
	LoadUserConfig(ret)
	return ret
}

func LoadUserConfig(env *zygo.Glisp) {
	usr, ue := homedir.Dir()
	if ue != nil {
		Global.Input = "Error getting current user's home directory: " + ue.Error()
		return
	}
	rc, err := ioutil.ReadFile(usr + "/.gomacs.lisp")
	if err != nil {
		Global.Input = "Error loading rc file: " + err.Error()
		return
	}
	err = env.LoadString(string(rc))
	if err != nil {
		Global.Input = "Error parsing rc file: " + err.Error()
		return
	}
	_, err = env.Run()
	if err != nil {
		Global.Input = "Error executing rc file: " + err.Error()
		return
	}
}

func LoadDefaultConfig(env *zygo.Glisp) {
	env.LoadString(`
(emacsbindkey "C-s" "isearch")
(emacsbindkey "C-x C-c" "save-buffers-kill-emacs")
(emacsbindkey "C-x C-s" "save-buffer")
(emacsbindkey "LEFT" "backward-char")
(emacsbindkey "C-b" "backward-char")
(emacsbindkey "RIGHT" "forward-char")
(emacsbindkey "C-f" "forward-char")
(emacsbindkey "DOWN" "next-line")
(emacsbindkey "C-n" "next-line")
(emacsbindkey "UP" "previous-line")
(emacsbindkey "C-p" "previous-line")
(emacsbindkey "Home" "move-beginning-of-line")
(emacsbindkey "End" "move-end-of-line")
(emacsbindkey "C-a" "move-beginning-of-line")
(emacsbindkey "C-e" "move-end-of-line")
(emacsbindkey "C-v" "scroll-down-command")
(emacsbindkey "M-v" "scroll-up-command")
(emacsbindkey "next" "scroll-down-command")
(emacsbindkey "prior" "scroll-up-command")
(emacsbindkey "DEL" "delete-backward-char")
(emacsbindkey "deletechar" "delete-char")
(emacsbindkey "C-d" "delete-char")
(emacsbindkey "RET" "insert-newline")
(emacsbindkey "C-x C-f" "find-file")
(emacsbindkey "C-x b" "buffers-list")
(emacsbindkey "M-<" "beginning-of-buffer")
(emacsbindkey "M->" "end-of-buffer")
(emacsbindkey "C-_" "undo")
(emacsbindkey "TAB" "indent")
(emacsbindkey "C-x o" "other-window")
(emacsbindkey "C-x 0" "delete-window")
(emacsbindkey "C-x 1" "delete-other-windows")
(emacsbindkey "C-x 2" "split-window")
(emacsbindkey "C-x 4 C-f" "find-file-other-window")
(emacsbindkey "C-x 4 f" "find-file-other-window")
(emacsbindkey "C-x 4 b" "switch-buffer-other-window")
(emacsbindkey "C-@" "set-mark")
(emacsbindkey "C-w" "kill-region")
(emacsbindkey "M-w" "copy-region")
(emacsbindkey "C-y" "yank-region")
(emacsbindkey "M-f" "forward-word")
(emacsbindkey "M-d" "kill-word")
(emacsbindkey "M-b" "backward-word")
(emacsbindkey "M-D" "backward-kill-word")
(emacsbindkey "M-DEL" "backward-kill-word")
(emacsbindkey "C-l" "recenter-top-bottom")
(emacsbindkey "C-x k" "kill-buffer")
(emacsbindkey "C-k" "kill-line")
(emacsbindkey "C-x C-_" "redo")
(emacsbindkey "C-z" "suspend-emacs")
(emacsbindkey "C-h c" "describe-key-briefly")
(emacsbindkey "M-x" "run-command")
(emacsbindkey "C-x C-u" "upcase-region")
(emacsbindkey "C-x C-l" "downcase-region")
(emacsbindkey "M-u" "upcase-word")
(emacsbindkey "M-l" "downcase-word")
(emacsbindkey "C-h m" "show-modes")
(emacsbindkey "C-h b" "describe-bindings")
(emacsbindkey "f1" "quick-help")
`)
	env.Run()
}
