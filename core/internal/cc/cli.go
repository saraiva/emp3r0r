package cc

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/bettercap/readline"
	"github.com/fatih/color"
	"github.com/jm33-m0/emp3r0r/core/internal/agent"
)

var (
	// CliCompleter holds all command completions
	CliCompleter = readline.NewPrefixCompleter()

	// CmdCompls completions for readline
	CmdCompls []readline.PrefixCompleterInterface

	// EmpReadLine : our commandline
	EmpReadLine *readline.Instance

	// EmpPrompt : the prompt string
	EmpPrompt = color.HiCyanString("emp3r0r > ")

	err error
)

// CliMain launches the commandline UI
func CliMain() {
	// completer
	CmdCompls = []readline.PrefixCompleterInterface{
		readline.PcItem("set",
			readline.PcItemDynamic(listOptions(),
				readline.PcItemDynamic(listValChoices()))),

		readline.PcItem("use",
			readline.PcItemDynamic(listMods())),

		readline.PcItem(HELP,
			readline.PcItemDynamic(listMods())),

		readline.PcItem("target",
			readline.PcItemDynamic(listTargetIndex())),
	}

	for cmd := range Commands {
		if cmd == "set" ||
			cmd == "use" ||
			cmd == "target" ||
			cmd == HELP {
			continue
		}
		CmdCompls = append(CmdCompls, readline.PcItem(cmd))
	}
	CmdCompls = append(CmdCompls, readline.PcItemDynamic(listFiles("./")))
	CliCompleter.SetChildren(CmdCompls)

	// prompt setup
	filterInput := func(r rune) (rune, bool) {
		switch r {
		// block CtrlZ feature
		case readline.CharCtrlZ:
			return r, false
		}
		return r, true
	}

	// set up readline instance
	EmpReadLine, err = readline.NewEx(&readline.Config{
		Prompt:          EmpPrompt,
		HistoryFile:     "./emp3r0r.history",
		AutoComplete:    CliCompleter,
		InterruptPrompt: "^C\nExiting...\n",
		EOFPrompt:       "^D\nEOF caught\nExiting...\n",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer EmpReadLine.Close()
	log.SetOutput(EmpReadLine.Stderr())

start:
	for {
		line, err := EmpReadLine.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			CliPrintError("EOF")
			os.Exit(0)
		}

		line = strings.TrimSpace(line)
		// readline-related commands
		switch line {
		case "commands":
			CliListCmds(EmpReadLine.Stderr())
		case "exit":
			os.Exit(0)

		// process other commands
		default:
			err = CmdHandler(line)
			if err != nil {
				color.Red(err.Error())
			}
		}
		fmt.Printf("\n")
	}

	// ask the user if they really want to leave
	if CliYesNo("Are you sure you want to leave") {
		os.Exit(0)
	}

	fmt.Printf("\n")
	goto start
}

// CliPrintInfo print log in blue
func CliPrintInfo(format string, a ...interface{}) {
	if DebugLevel == 0 {
		log.Println(color.BlueString(format, a...))
	}
}

// CliPrintWarning print log in yellow
func CliPrintWarning(format string, a ...interface{}) {
	if DebugLevel <= 1 {
		log.Println(color.YellowString(format, a...))
	}
}

// CliPrintSuccess print log in green
func CliPrintSuccess(format string, a ...interface{}) {
	log.Println(color.HiGreenString(format, a...))
}

// CliPrintError print log in red
func CliPrintError(format string, a ...interface{}) {
	log.Println(color.HiRedString(format, a...))
}

// CliYesNo prompt for a y/n answer from user
func CliYesNo(prompt string) bool {
	EmpReadLine.SetPrompt(color.CyanString(prompt + "? [y/N] "))
	EmpReadLine.Config.EOFPrompt = ""
	EmpReadLine.Config.InterruptPrompt = ""

	defer EmpReadLine.SetPrompt(EmpPrompt)

	answer, err := EmpReadLine.Readline()
	if err != nil {
		if err == readline.ErrInterrupt || err == io.EOF {
			return false
		}
		color.Red(err.Error())
	}

	answer = strings.ToLower(answer)
	return answer == "y"
}

// CliListOptions list currently available options for `set`
func CliListOptions() {
	opts := make(map[string]string)
	opts["module"] = CurrentMod
	tc, exist := Targets[CurrentTarget]
	if exist {
		opts["target"] = strconv.Itoa(tc.Index)
	} else {
		opts["target"] = "<blank>"
	}

	for k, v := range Options {
		opts[k] = v.Val
	}
	CliPrettyPrint("Option", "Value", &opts)
}

// CliListCmds list all commands in tree format
func CliListCmds(w io.Writer) {
	_, err := io.WriteString(w, "Commands:\n")
	if err != nil {
		return
	}
	_, err = io.WriteString(w, CliCompleter.Tree("    "))
	if err != nil {
		return
	}
}

// CliBanner prints banner
func CliBanner() error {
	data, err := base64.StdEncoding.DecodeString(cliBannerB64)
	if err != nil {
		return errors.New("Failed to print banner: " + err.Error())
	}

	color.Cyan(string(data))
	color.Cyan("version: %s\n\n", agent.Version)
	return nil
}

// CliPrettyPrint prints two-column help info
func CliPrettyPrint(header1, header2 string, map2write *map[string]string) {
	cnt := 10
	sep := strings.Repeat(" ", cnt)
	color.Cyan("%s%s%s\n", header1, sep, header2)

	color.Cyan("%s%s%s\n", strings.Repeat("=", len(header1)), sep, strings.Repeat("=", len(header2)))
	fmt.Println("")

	for c1, c2 := range *map2write {
		cnt = len(header1) + 10 - len(c1)
		sep = strings.Repeat(" ", cnt)
		color.Cyan("%s%s%s\n", c1, sep, c2)
	}
}

// encoded logo of emp3r0r
const cliBannerB64 string = `
CuKWkeKWkeKWkeKWkeKWkeKWkeKWkSDilpHilpHilpEgICAg4paR4paR4paRIOKWkeKWkeKWkeKW
keKWkeKWkSAg4paR4paR4paR4paR4paR4paRICDilpHilpHilpHilpHilpHilpEgICDilpHilpHi
lpHilpHilpHilpEgIOKWkeKWkeKWkeKWkeKWkeKWkQrilpLilpIgICAgICDilpLilpLilpLilpIg
IOKWkuKWkuKWkuKWkiDilpLilpIgICDilpLilpIgICAgICDilpLilpIg4paS4paSICAg4paS4paS
IOKWkuKWkiAg4paS4paS4paS4paSIOKWkuKWkiAgIOKWkuKWkgrilpLilpLilpLilpLilpIgICDi
lpLilpIg4paS4paS4paS4paSIOKWkuKWkiDilpLilpLilpLilpLilpLilpIgICDilpLilpLilpLi
lpLilpIgIOKWkuKWkuKWkuKWkuKWkuKWkiAg4paS4paSIOKWkuKWkiDilpLilpIg4paS4paS4paS
4paS4paS4paSCuKWk+KWkyAgICAgIOKWk+KWkyAg4paT4paTICDilpPilpMg4paT4paTICAgICAg
ICAgICDilpPilpMg4paT4paTICAg4paT4paTIOKWk+KWk+KWk+KWkyAg4paT4paTIOKWk+KWkyAg
IOKWk+KWkwrilojilojilojilojilojilojilogg4paI4paIICAgICAg4paI4paIIOKWiOKWiCAg
ICAgIOKWiOKWiOKWiOKWiOKWiOKWiCAg4paI4paIICAg4paI4paIICDilojilojilojilojiloji
loggIOKWiOKWiCAgIOKWiOKWiAoKCmEgbGludXggcG9zdC1leHBsb2l0YXRpb24gZnJhbWV3b3Jr
IG1hZGUgYnkgbGludXggdXNlcgoKYnkgam0zMy1uZwoKaHR0cHM6Ly9naXRodWIuY29tL2ptMzMt
bTAvZW1wM3IwcgoKCg==
`

// autocomplete module options
func listValChoices() func(string) []string {
	return func(line string) []string {
		switch CurrentMod {
		case agent.ModCMD_EXEC:
			return Options["cmd_to_exec"].Vals
		case agent.ModCLEAN_LOG:
			return Options["keyword"].Vals
		case agent.ModLPE_SUGGEST:
			return Options["lpe_helper"].Vals
		case agent.ModPERSISTENCE:
			return Options["method"].Vals
		case agent.ModPROXY:
			return append(Options["status"].Vals, Options["port"].Vals...)
		case agent.ModINJECTOR:
			return append(Options["pid"].Vals, Options["method"].Vals...)
		case agent.ModPORT_FWD:
			ret := append(Options["listen_port"].Vals, Options["to"].Vals...)
			ret = append(ret, Options["switch"].Vals...)
			return ret
		}

		return nil
	}
}

// autocomplete modules names
func listMods() func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		for mod := range ModuleHelpers {
			names = append(names, mod)
		}
		return names
	}
}

// autocomplete target index
func listTargetIndex() func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		for _, c := range Targets {
			idx := c.Index
			names = append(names, strconv.Itoa(idx))
		}
		return names
	}
}

// autocomplete option names
func listOptions() func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)

		for opt := range Options {
			names = append(names, opt)
		}
		return names
	}
}

// Function constructor - constructs new function for listing given directory
func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}
