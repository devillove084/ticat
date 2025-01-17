package display

import (
	"sort"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
)

func DumpFlow(
	cc *core.Cli,
	env *core.Env,
	flow []core.ParsedCmd,
	args *DumpFlowArgs) {

	if len(flow) == 0 {
		return
	}

	env = env.Clone()
	maxDepth := env.GetInt("display.flow.depth")

	PrintTipTitle(cc.Screen, env, "flow executing description:")
	cc.Screen.Print("--->>>\n")
	dumpFlow(cc, env, flow, args, maxDepth, 0)
	cc.Screen.Print("<<<---\n")
}

func dumpFlow(
	cc *core.Cli,
	env *core.Env,
	flow []core.ParsedCmd,
	args *DumpFlowArgs,
	maxDepth int,
	indentAdjust int) {

	metFlows := map[string]bool{}
	for _, cmd := range flow {
		if !cmd.IsEmpty() {
			dumpFlowCmd(cc, cc.Screen, env, cmd, args,
				maxDepth, indentAdjust, metFlows)
		}
	}
}

func dumpFlowCmd(
	cc *core.Cli,
	screen core.Screen,
	env *core.Env,
	parsedCmd core.ParsedCmd,
	args *DumpFlowArgs,
	maxDepth int,
	indentAdjust int,
	metFlows map[string]bool) {

	cmd := parsedCmd.Last().Matched.Cmd
	if cmd == nil {
		return
	}

	sep := cmd.Strs.PathSep
	envOpSep := " " + cmd.Strs.EnvOpSep + " "

	prt := func(indentLvl int, msg string) {
		indentLvl += indentAdjust
		padding := rpt(" ", args.IndentSize*indentLvl)
		msg = autoPadNewLine(padding, msg)
		screen.Print(padding + msg + "\n")
	}

	cic := cmd.Cmd()
	if cic == nil {
		return
	}
	var name string
	if args.Skeleton {
		name = strings.Join(parsedCmd.Path(), sep)
	} else {
		name = parsedCmd.DisplayPath(sep, true)
	}
	prt(0, "["+name+"]")
	if len(cic.Help()) != 0 {
		prt(1, " '"+cic.Help()+"'")
	}

	cmdEnv, argv := parsedCmd.GenEnvAndArgv(env, cc.Cmds.Strs.EnvValDelAllMark, sep)

	if !args.Skeleton {
		args := parsedCmd.Args()
		argLines := DumpArgs(&args, argv, true)
		if len(argLines) != 0 {
			prt(1, "- args:")
		}
		for _, line := range argLines {
			prt(2, line)
		}
	}

	if !args.Skeleton {
		// TODO BUG: missed kvs in GlobalEnv
		cmdEssEnv := parsedCmd.GenEnv(core.NewEnv(), cc.Cmds.Strs.EnvValDelAllMark)
		flatten := cmdEssEnv.Flatten(false, nil, true)
		if len(flatten) != 0 {
			prt(1, "- env-values:")
			var keys []string
			for k, _ := range flatten {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				prt(2, k+" = "+flatten[k])
			}
		}
	}

	if !args.Skeleton {
		val2env := cic.GetVal2Env()
		if len(val2env.EnvKeys()) != 0 {
			prt(1, "- env-direct-write:")
		}
		for _, k := range val2env.EnvKeys() {
			prt(2, k+" = "+mayQuoteStr(val2env.Val(k)))
		}

		arg2env := cic.GetArg2Env()
		if len(arg2env.EnvKeys()) != 0 {
			prt(1, "- env-from-argv:")
		}
		for _, k := range arg2env.EnvKeys() {
			prt(2, k+" <- "+mayQuoteStr(arg2env.GetArgName(k)))
		}

		envOps := cic.EnvOps()
		envOpKeys := envOps.EnvKeys()
		if len(envOpKeys) != 0 {
			prt(1, "- env-ops:")
		}
		for _, k := range envOpKeys {
			prt(2, k+" = "+dumpEnvOps(envOps.Ops(k), envOpSep))
		}
	}

	if !args.Simple && !args.Skeleton {
		line := string(cic.Type())
		if cic.IsQuiet() {
			line += " (quiet)"
		}
		if cic.IsPriority() {
			line += " (priority)"
		}
		prt(1, "- cmd-type:")
		prt(2, line)

		if len(cic.Source()) != 0 && !strings.HasPrefix(cic.CmdLine(), cic.Source()) {
			prt(1, "- from:")
			prt(2, cic.Source())
		}
	}

	if (len(cic.CmdLine()) != 0 || len(cic.FlowStrs()) != 0) &&
		cic.Type() != core.CmdTypeNormal && cic.Type() != core.CmdTypePower {
		metFlow := false
		if cic.Type() == core.CmdTypeFlow {
			flowStrs, _ := cic.RenderedFlowStrs(cmdEnv, true)
			flowStr := strings.Join(flowStrs, " ")
			metFlow = metFlows[flowStr]
			if metFlow {
				prt(1, "- flow (duplicated):")
			} else {
				metFlows[flowStr] = true
				prt(1, "- flow:")
			}
			for _, flowStr := range flowStrs {
				prt(2, flowStr)
			}
		} else if !args.Simple && !args.Skeleton {
			if cic.Type() == core.CmdTypeEmptyDir {
				prt(1, "- dir:")
				prt(2, cic.CmdLine())
			} else {
				prt(1, "- executable:")
				prt(2, cic.CmdLine())
			}
			if len(cic.MetaFile()) != 0 {
				prt(1, "- meta:")
				prt(2, cic.MetaFile())
			}
		}
		if cic.Type() == core.CmdTypeFlow && maxDepth > 1 {
			subFlow, rendered := cic.Flow(cmdEnv, true)
			if rendered && len(subFlow) != 0 {
				if !metFlow {
					prt(2, "--->>>")
					parsedFlow := cc.Parser.Parse(cc.Cmds, cc.EnvAbbrs, subFlow...)
					err := parsedFlow.FirstErr()
					if err != nil {
						panic(err.Error)
					}
					dumpFlow(cc, env, parsedFlow.Cmds, args, maxDepth-1, indentAdjust+2)
					prt(2, "<<<---")
				}
			}
		}
	}
}

type DumpFlowArgs struct {
	Simple     bool
	Skeleton   bool
	IndentSize int
}

func NewDumpFlowArgs() *DumpFlowArgs {
	return &DumpFlowArgs{false, false, 4}
}

func (self *DumpFlowArgs) SetSimple() *DumpFlowArgs {
	self.Simple = true
	return self
}

func (self *DumpFlowArgs) SetSkeleton() *DumpFlowArgs {
	self.Simple = true
	self.Skeleton = true
	return self
}
