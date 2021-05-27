package main

import (
	"log"
	"strings"

	"github.com/Shopify/go-lua"
)

func lfLibrary(app *app) []lua.RegistryFunction {
	return []lua.RegistryFunction{
		{"eval", func(l *lua.State) int {
			s, _ := l.ToString(1)
			p := newParser(strings.NewReader(s))
			for p.parse() {
				p.expr.eval(app, nil)
			}
			if p.err != nil {
				app.ui.echoerrf("%s", p.err)
			}
			return 0
		}},
		{"eval_exec", func(l *lua.State) int {
			prefix, _ := l.ToString(1)
			value, _ := l.ToString(2)
			e := &execExpr{prefix: prefix, value: value}
			e.eval(app, nil)
			return 0
		}},
		{"eval_call", func(l *lua.State) int {
			args := []string{}
			cmd, _ := l.ToString(1)
			for i := 2; i <= l.Top(); i++ {
				arg, _ := l.ToString(i)
				args = append(args, arg)
			}
			if cmd != "" {
				e := &callExpr{cmd, args, 1}
				e.eval(app, nil)
			}
			return 0
		}},

		{"set", func(l *lua.State) int {
			opt, _ := l.ToString(1)
			val, _ := l.ToString(2)
			e := &setExpr{opt: opt, val: val}
			e.eval(app, nil)
			return 0
		}},
		{"unmap", func(l *lua.State) int {
			key, _ := l.ToString(1)
			if _, ok := gOpts.keys[key]; ok {
				delete(gOpts.keys, key)
			}
			return 0
		}},
		{"log", func(l *lua.State) int {
			s, _ := l.ToString(1)
			log.Print(s)
			return 0
		}},
		{"get", lfOptGet},
	}
}

var lfHelpers = `
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

func LuaInit(app *app) *lua.State {
	l := lua.NewState()
	lua.OpenLibraries(l)
	lua.Require(l, "lf", func(l *lua.State) int {
		lua.NewLibrary(l, lfLibrary(app))
		return 1
	}, true)
	l.Pop(1)

	if err := lua.DoString(l, lfHelpers); err != nil {
		app.ui.echoerr(err.Error())
	}

	l.Register("stringsplit", func(l *lua.State) int {
		s, _ := l.ToString(1)
		sep, _ := l.ToString(2)
		tokens := strings.Split(s, sep)
		for _, tok := range tokens {
			l.PushString(tok)
			log.Print(tok)
		}
		return len(tokens)
	})

	return l
}

func LuaSource(app *app, file string) {
	log.Printf("luasource: %s\n", file)
	if err := lua.DoFile(app.luaState, file); err != nil {
		app.ui.echoerr(err.Error())
	}
}

func LuaRun(app *app, str string, args []string) {
	l := app.luaState
	l.CreateTable(len(args), 0)
	for i, arg := range args {
		l.PushInteger(i + 1)
		l.PushString(arg)
		l.SetTable(-3)
	}
	l.SetGlobal("argv")
	if err := lua.DoString(app.luaState, str); err != nil {
		app.ui.echoerr(err.Error())
	}
}

func LuaHook(app *app, cmd string, args []string) {
	l := app.luaState
	l.Global("run_command_hook")
	l.PushString(cmd)
	for _, s := range args {
		l.PushString(s)
	}
	l.Call(len(args)+1, 0)
}

// var gOpts struct {
// 	keys           map[string]expr
// 	cmdkeys        map[string]expr
// 	cmds           map[string]expr
// 	sortType       sortType
// }

func lfOptGet(l *lua.State) int {
	opt, _ := l.ToString(1)
	switch opt {
	case "anchorfind":
		l.PushBoolean(gOpts.anchorfind)
		return 1
	case "dircounts":
		l.PushBoolean(gOpts.dircounts)
		return 1
	case "drawbox":
		l.PushBoolean(gOpts.drawbox)
		return 1
	case "globsearch":
		l.PushBoolean(gOpts.globsearch)
		return 1
	case "icons":
		l.PushBoolean(gOpts.icons)
		return 1
	case "ignorecase":
		l.PushBoolean(gOpts.ignorecase)
		return 1
	case "ignoredia":
		l.PushBoolean(gOpts.ignoredia)
		return 1
	case "incsearch":
		l.PushBoolean(gOpts.incsearch)
		return 1
	case "mouse":
		l.PushBoolean(gOpts.mouse)
		return 1
	case "number":
		l.PushBoolean(gOpts.number)
		return 1
	case "preview":
		l.PushBoolean(gOpts.preview)
		return 1
	case "relativenumber":
		l.PushBoolean(gOpts.relativenumber)
		return 1
	case "smartcase":
		l.PushBoolean(gOpts.smartcase)
		return 1
	case "smartdia":
		l.PushBoolean(gOpts.smartdia)
		return 1
	case "waitmsg":
		l.PushString(gOpts.waitmsg)
		return 1
	case "wrapscan":
		l.PushBoolean(gOpts.wrapscan)
		return 1
	case "wrapscroll":
		l.PushBoolean(gOpts.wrapscroll)
		return 1
	case "findlen":
		l.PushInteger(gOpts.findlen)
		return 1
	case "period":
		l.PushInteger(gOpts.period)
		return 1
	case "scrolloff":
		l.PushInteger(gOpts.scrolloff)
		return 1
	case "tabstop":
		l.PushInteger(gOpts.tabstop)
		return 1
	case "errorfmt":
		l.PushString(gOpts.errorfmt)
		return 1
	case "filesep":
		l.PushString(gOpts.filesep)
		return 1
	case "ifs":
		l.PushString(gOpts.ifs)
		return 1
	case "previewer":
		l.PushString(gOpts.previewer)
		return 1
	case "cleaner":
		l.PushString(gOpts.cleaner)
		return 1
	case "promptfmt":
		l.PushString(gOpts.promptfmt)
		return 1
	case "shell":
		l.PushString(gOpts.shell)
		return 1
	case "shellflag":
		l.PushString(gOpts.shellflag)
		return 1
	case "timefmt":
		l.PushString(gOpts.timefmt)
		return 1
	case "truncatechar":
		l.PushString(gOpts.truncatechar)
		return 1
	case "ratios":
		for _, el := range gOpts.ratios {
			l.PushInteger(el)
		}
		return len(gOpts.ratios)
	case "hiddenfiles":
		for _, el := range gOpts.hiddenfiles {
			l.PushString(el)
		}
		return len(gOpts.hiddenfiles)
	case "info":
		for _, el := range gOpts.info {
			l.PushString(el)
		}
		return len(gOpts.info)
	case "shellopts":
		for _, el := range gOpts.shellopts {
			l.PushString(el)
		}
		return len(gOpts.shellopts)
	default:
		return -1
	}
}
