package builtin

import (
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func GlobalHelpMoreInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return globalHelpLessMoreInfo(argv, cc, env, flow, currCmdIdx, false)
}

func GlobalHelpLessInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return globalHelpLessMoreInfo(argv, cc, env, flow, currCmdIdx, true)
}

func globalHelpLessMoreInfo(
	argv core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	skeleton bool) (int, bool) {

	findStrs := getFindStrsFromArgv(argv)

	for _, cmd := range flow.Cmds {
		if cmd.ParseResult.Error == nil {
			continue
		}
		findStrs = append(cmd.ParseResult.Input, findStrs...)
		cmdPathStr := ""
		cic := cc.Cmds
		if !cmd.IsEmpty() {
			cic = cmd.Last().Matched.Cmd
			cmdPathStr = cmd.DisplayPath(cc.Cmds.Strs.PathSep, true)
		}
		return dumpMoreLessFindResult(flow, cc.Screen, env, cmdPathStr, cic, skeleton, findStrs...)
	}

	if len(flow.Cmds) >= 2 {
		cmdPathStr := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
		cmd := cc.Cmds.GetSub(strings.Split(cmdPathStr, cc.Cmds.Strs.PathSep)...)
		if cmd == nil {
			panic("[globalHelpLessMoreInfo] should never happen")
		}
		cmdPathStr = flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, true)
		if len(findStrs) != 0 {
			return dumpMoreLessFindResult(flow, cc.Screen, env, cmdPathStr, cmd, skeleton, findStrs...)
		}
		if len(flow.Cmds) > 2 ||
			cmd.Cmd() != nil && cmd.Cmd().Type() == core.CmdTypeFlow {
			if skeleton {
				return DumpFlowSkeleton(argv, cc, env, flow, currCmdIdx)
			} else {
				return DumpFlowAllSimple(argv, cc, env, flow, currCmdIdx)
			}
		}
		input := flow.Last().ParseResult.Input
		if len(input) > 1 {
			return dumpMoreLessFindResult(flow, cc.Screen, env, "", cc.Cmds, skeleton, input...)
		}
		return dumpMoreLessFindResult(flow, cc.Screen, env, cmdPathStr, cmd, skeleton)
	}

	return dumpMoreLessFindResult(flow, cc.Screen, env, "", cc.Cmds, skeleton, findStrs...)
}

func DumpTailCmdInfo(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	cmdPath := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
	dumpArgs := display.NewDumpCmdArgs().NoFlatten().NoRecursive()
	dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	return clearFlow(flow)
}

func DumpTailCmdSub(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	if len(flow.Cmds) < 2 {
		cmdPath := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
		dumpArgs := display.NewDumpCmdArgs().NoFlatten().NoRecursive()
		dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	} else {
		cmdPath := flow.Last().DisplayPath(cc.Cmds.Strs.PathSep, false)
		dumpArgs := display.NewDumpCmdArgs().SetSkeleton()
		dumpCmdsByPath(cc, env, dumpArgs, cmdPath)
	}
	return clearFlow(flow)
}

func FindAny(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	findStrs := getFindStrsFromArgv(argv)
	if len(findStrs) == 0 {
		return true
	}
	display.DumpEnvFlattenVals(cc.Screen, env, findStrs...)
	dumpArgs := display.NewDumpCmdArgs().AddFindStrs(findStrs...)
	display.DumpCmdsWithTips(cc.Cmds, cc.Screen, env, dumpArgs, "", false)
	return true
}

func GlobalHelp(_ core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	display.PrintGlobalHelp(cc.Screen, env)
	return true
}

func dumpMoreLessFindResult(
	flow *core.ParsedCmds,
	screen core.Screen,
	env *core.Env,
	cmdPathStr string,
	cmd *core.CmdTree,
	skeleton bool,
	findStrs ...string) (int, bool) {

	printer := display.NewCacheScreen()
	dumpArgs := display.NewDumpCmdArgs().AddFindStrs(findStrs...)
	dumpArgs.Skeleton = skeleton
	display.DumpCmdsWithTips(cmd, printer, env, dumpArgs, cmdPathStr, true)
	printer.WriteTo(screen)
	return clearFlow(flow)
}
