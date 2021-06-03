package main

import (
	"log"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func lfLibrary(app *app) map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"eval": func(l *lua.LState) int {
			s := l.ToString(1)
			p := newParser(strings.NewReader(s))
			for p.parse() {
				p.expr.eval(app, nil)
			}
			if p.err != nil {
				app.ui.echoerrf("%s", p.err)
			}
			return 0
		},
		"eval_exec": func(l *lua.LState) int {
			prefix := l.ToString(1)
			value := l.ToString(2)
			e := &execExpr{prefix: prefix, value: value}
			e.eval(app, nil)
			return 0
		},
		"eval_call": func(l *lua.LState) int {
			args := []string{}
			cmd := l.ToString(1)
			for i := 2; i <= l.GetTop(); i++ {
				arg := l.ToString(i)
				args = append(args, arg)
			}
			if cmd != "" {
				e := &callExpr{cmd, args, 1}
				e.eval(app, nil)
			}
			return 0
		},

		"set": func(l *lua.LState) int {
			opt := l.ToString(1)
			val := l.ToString(2)
			e := &setExpr{opt: opt, val: val}
			e.eval(app, nil)
			return 0
		},
		"unmap": func(l *lua.LState) int {
			key := l.ToString(1)
			if _, ok := gOpts.keys[key]; ok {
				delete(gOpts.keys, key)
			}
			return 0
		},
		"log": func(l *lua.LState) int {
			s := l.ToString(1)
			log.Print(s)
			return 0
		},
		"get": lfOptGet,
	}
}

var lfHelpers = `
lf = require("lf")
lf.echo = function (...) lf.eval_call("echo", ...) end
lf.echomsg = function (...) lf.eval_call("echomsg", ...) end
lf.echoerr = function (...) lf.eval_call("echoerr", ...) end

lf.shell = function (...) lf.eval_exec("$", ...) end
lf.shell_pipe = function (...) lf.eval_exec("%", ...) end
lf.shell_wait = function (...) lf.eval_exec("!", ...) end
lf.shell_async = function (...) lf.eval_exec("&", ...) end

lf.push = function (...) lf.eval_call("push", ...) end
lf.cmd = function (name, cmd) lf.eval("cmd " .. name .. " " .. cmd) end
lf.map = function (key, val) lf.eval("map " .. key .. " " .. val) end
lf.cmap = function (key, val) lf.eval("cmap " .. key .. " " .. val) end

lf.command_hooks = {}

function lf.register_command_hook (cmd, f)
	funs = lf.command_hooks[cmd]
	if funs == nil then
		lf.command_hooks[cmd] = {f}
	else
		funs[#funs+1] = f
	end
end
function run_command_hook (cmd, ...)
	funs = lf.command_hooks[cmd]
	if funs ~= nil then
		for _, f in pairs(funs) do
			f(...)
		end
	end
end
`

func LuaInit(app *app) *lua.LState {
	l := lua.NewState()
	l.PreloadModule("lf", func(l *lua.LState) int {
		mod := l.SetFuncs(l.NewTable(), lfLibrary(app))
		l.Push(mod)
		return 1
	})

	if err := l.DoString(lfHelpers); err != nil {
		app.ui.echoerr(err.Error())
	}

	return l
}

func LuaSource(app *app, file string) {
	l := app.luaState
	log.Printf("luasource: %s\n", file)
	if err := l.DoFile(file); err != nil {
		app.ui.echoerr(err.Error())
	}
}

func LuaRun(app *app, str string, args []string) {
	l := app.luaState
	argv := l.NewTable()
	for i, arg := range args {
		l.RawSetInt(argv, i+1, lua.LString(arg))
	}
	l.SetGlobal("argv", argv)
	if err := l.DoString(str); err != nil {
		app.ui.echoerr(err.Error())
	}
}

func LuaHook(app *app, cmd string, args []string) {
	l := app.luaState
	lArgs := []lua.LValue{lua.LString(cmd)}
	for _, s := range args {
		lArgs = append(lArgs, lua.LString(s))
	}
	l.CallByParam(lua.P{
		Fn:      l.GetGlobal("run_command_hook"),
		NRet:    0,
		Protect: true,
	}, lArgs...)
}

func LuaComplete(app *app, tokens []string) (matches []string, longest string) {
	l := app.luaState
	lArgs := []lua.LValue{}
	for _, s := range tokens {
		lArgs = append(lArgs, lua.LString(s))
	}
	l.CallByParam(lua.P{
		Fn:      l.GetGlobal("complete"),
		NRet:    lua.MultRet,
		Protect: true,
	}, lArgs...)

	if str, ok := l.Get(-1).(lua.LString); ok {
		longest = str.String()
	}
	l.Pop(1)
	t := l.GetTop()
	for i := 0; i < t; i = i + 1 {
		if str, ok := l.Get(-1).(lua.LString); ok {
			matches = append(matches, str.String())
		}
		l.Pop(1)
	}

	return
}

// var gOpts struct {
// 	keys           map[string]expr
// 	cmdkeys        map[string]expr
// 	cmds           map[string]expr
// 	sortType       sortType
// }

func lfOptGet(l *lua.LState) int {
	opt := l.ToString(1)
	switch opt {
	case "anchorfind":
		l.Push(lua.LBool(gOpts.anchorfind))
		return 1
	case "dircounts":
		l.Push(lua.LBool(gOpts.dircounts))
		return 1
	case "drawbox":
		l.Push(lua.LBool(gOpts.drawbox))
		return 1
	case "globsearch":
		l.Push(lua.LBool(gOpts.globsearch))
		return 1
	case "icons":
		l.Push(lua.LBool(gOpts.icons))
		return 1
	case "ignorecase":
		l.Push(lua.LBool(gOpts.ignorecase))
		return 1
	case "ignoredia":
		l.Push(lua.LBool(gOpts.ignoredia))
		return 1
	case "incsearch":
		l.Push(lua.LBool(gOpts.incsearch))
		return 1
	case "mouse":
		l.Push(lua.LBool(gOpts.mouse))
		return 1
	case "number":
		l.Push(lua.LBool(gOpts.number))
		return 1
	case "preview":
		l.Push(lua.LBool(gOpts.preview))
		return 1
	case "relativenumber":
		l.Push(lua.LBool(gOpts.relativenumber))
		return 1
	case "smartcase":
		l.Push(lua.LBool(gOpts.smartcase))
		return 1
	case "smartdia":
		l.Push(lua.LBool(gOpts.smartdia))
		return 1
	case "waitmsg":
		l.Push(lua.LString(gOpts.waitmsg))
		return 1
	case "wrapscan":
		l.Push(lua.LBool(gOpts.wrapscan))
		return 1
	case "wrapscroll":
		l.Push(lua.LBool(gOpts.wrapscroll))
		return 1
	case "findlen":
		l.Push(lua.LNumber(gOpts.findlen))
		return 1
	case "period":
		l.Push(lua.LNumber(gOpts.period))
		return 1
	case "scrolloff":
		l.Push(lua.LNumber(gOpts.scrolloff))
		return 1
	case "tabstop":
		l.Push(lua.LNumber(gOpts.tabstop))
		return 1
	case "errorfmt":
		l.Push(lua.LString(gOpts.errorfmt))
		return 1
	case "filesep":
		l.Push(lua.LString(gOpts.filesep))
		return 1
	case "ifs":
		l.Push(lua.LString(gOpts.ifs))
		return 1
	case "previewer":
		l.Push(lua.LString(gOpts.previewer))
		return 1
	case "cleaner":
		l.Push(lua.LString(gOpts.cleaner))
		return 1
	case "promptfmt":
		l.Push(lua.LString(gOpts.promptfmt))
		return 1
	case "shell":
		l.Push(lua.LString(gOpts.shell))
		return 1
	case "shellflag":
		l.Push(lua.LString(gOpts.shellflag))
		return 1
	case "timefmt":
		l.Push(lua.LString(gOpts.timefmt))
		return 1
	case "truncatechar":
		l.Push(lua.LString(gOpts.truncatechar))
		return 1
	case "ratios":
		for _, el := range gOpts.ratios {
			l.Push(lua.LNumber(el))
		}
		return len(gOpts.ratios)
	case "hiddenfiles":
		for _, el := range gOpts.hiddenfiles {
			l.Push(lua.LString(el))
		}
		return len(gOpts.hiddenfiles)
	case "info":
		for _, el := range gOpts.info {
			l.Push(lua.LString(el))
		}
		return len(gOpts.info)
	case "shellopts":
		for _, el := range gOpts.shellopts {
			l.Push(lua.LString(el))
		}
		return len(gOpts.shellopts)
	default:
		return -1
	}
}
