package builtin

import (
	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/cli/display"
)

func SetDumpFlowDepth(argv core.ArgVals, cc *core.Cli, env *core.Env, _ core.ParsedCmd) bool {
	depth := argv.GetInt("depth")
	env.GetLayer(core.EnvLayerSession).SetInt("display.flow.depth", depth)
	return true
}

func DumpFlow(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpFlowArgs()
	display.DumpFlow(cc, env, flow.Cmds[currCmdIdx+1:], dumpArgs)
	return clearFlow(flow)
}

func DumpFlowSimple(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpFlowArgs().SetSimple()
	display.DumpFlow(cc, env, flow.Cmds[currCmdIdx+1:], dumpArgs)
	return clearFlow(flow)
}

func DumpFlowSkeleton(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	dumpArgs := display.NewDumpFlowArgs().SetSkeleton()
	display.DumpFlow(cc, env, flow.Cmds[currCmdIdx+1:], dumpArgs)
	return clearFlow(flow)
}

func DumpFlowDepends(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	deps := display.Depends{}
	display.CollectDepends(cc, flow.Cmds[currCmdIdx+1:], deps)

	if len(deps) != 0 {
		display.DumpDepends(cc, env, deps)
	} else {
		cc.Screen.Print("no depended os commands\n")
	}
	return clearFlow(flow)
}

func DumpFlowEnvOpsCheckResult(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	checker := &core.EnvOpsChecker{}
	result := []core.EnvOpsCheckResult{}
	core.CheckEnvOps(cc, flow, env, checker, false, &result)

	if len(result) != 0 {
		display.DumpEnvOpsCheckResult(cc.Screen, env, result, cc.Cmds.Strs.PathSep)
	} else {
		cc.Screen.Print("all env-ops are satisfied, can directly run\n")
	}

	return clearFlow(flow)
}

func DumpFlowAllSimple(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return dumpFlowAll(cc, env, flow, currCmdIdx, true)
}

func DumpFlowAll(
	_ core.ArgVals,
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int) (int, bool) {

	return dumpFlowAll(cc, env, flow, currCmdIdx, false)
}

func dumpFlowAll(
	cc *core.Cli,
	env *core.Env,
	flow *core.ParsedCmds,
	currCmdIdx int,
	simple bool) (int, bool) {

	cmds := flow.Cmds[currCmdIdx+1:]

	dumpArgs := display.NewDumpFlowArgs()
	dumpArgs.Simple = simple
	display.DumpFlow(cc, env, cmds, dumpArgs)

	deps := display.Depends{}
	display.CollectDepends(cc, flow.Cmds[currCmdIdx+1:], deps)

	if len(deps) != 0 {
		cc.Screen.Print("\n")
		display.DumpDepends(cc, env, deps)
	}

	checker := &core.EnvOpsChecker{}
	result := []core.EnvOpsCheckResult{}
	core.CheckEnvOps(cc, flow, env, checker, false, &result)

	if len(result) != 0 {
		cc.Screen.Print("\n")
		display.DumpEnvOpsCheckResult(cc.Screen, env, result, cc.Cmds.Strs.PathSep)
	}

	return clearFlow(flow)
}