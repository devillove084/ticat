package mod_meta

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/ticat/pkg/cli/core"
	"github.com/pingcap/ticat/pkg/proto/meta_file"
)

func RegMod(
	cc *core.Cli,
	metaPath string,
	executablePath string,
	isDir bool,
	isFlow bool,
	cmdPath []string,
	abbrsSep string,
	envPathSep string,
	source string) {

	defer func() {
		if err := recover(); err != nil {
			cc.TolerableErrs.OnErr(err, source, metaPath, "module loading failed")
		}
	}()

	mod := cc.Cmds.GetOrAddSub(cmdPath...)
	meta := meta_file.NewMetaFile(metaPath)

	cmd := regMod(meta, mod, executablePath, isDir)
	cmd.SetSource(source).SetMetaFile(metaPath)

	// Reg by isFlow, not 'cmd.Type()'
	if isFlow {
		regFlowAbbrs(meta, cc.Cmds, cmdPath)
	} else {
		regModAbbrs(meta, mod)
	}

	regArgs(meta, cmd, abbrsSep)
	regDeps(meta, cmd)
	regEnvOps(cc.EnvAbbrs, meta, cmd, abbrsSep, envPathSep)
	regVal2Env(cc.EnvAbbrs, meta, cmd, abbrsSep, envPathSep)
	regArg2Env(cc.EnvAbbrs, meta, cmd, abbrsSep, envPathSep)
}

func regMod(
	meta *meta_file.MetaFile,
	mod *core.CmdTree,
	executablePath string,
	isDir bool) *core.Cmd {

	cmdPath := mod.DisplayPath()

	globalSection := meta.GetGlobalSection()
	flow := globalSection.GetMultiLineVal("flow", false)

	// 'cmd' should be a relative path base on this file when 'isDir'
	cmdLine := meta.Get("cmd")

	help := meta.Get("help")
	// If has executable file, it need to have help string, a flow can have not
	if len(help) == 0 && (!isDir && len(flow) == 0 || len(cmdLine) != 0) {
		panic(fmt.Errorf("[regMod] cmd '%s' has no help string in '%s'",
			cmdPath, meta.Path()))
	}

	// Even if 'isFlow' is true, if it does not have 'flow' content, it can't reg as flow
	if len(flow) != 0 {
		return mod.RegFlowCmd(flow, help)
	}

	if len(executablePath) == 0 {
		return mod.RegEmptyCmd(help)
	}

	// Adjust 'executablePath'
	if len(cmdLine) != 0 {
		if !isDir {
			executablePath = filepath.Dir(executablePath)
		}
		var err error
		executablePath, err = filepath.Abs(filepath.Join(executablePath, cmdLine))
		if err != nil {
			panic(fmt.Errorf("[regMod] cmd '%s' get abs path of '%s' failed",
				cmdPath, executablePath))
		}
		if !fileExists(executablePath) {
			panic(fmt.Errorf("[regMod] cmd '%s' point to a not existed file '%s'",
				cmdPath, executablePath))
		}
	}

	if isDir {
		if len(cmdLine) != 0 {
			return mod.RegDirWithCmd(executablePath, help)
		} else {
			return mod.RegEmptyDirCmd(executablePath, help)
		}
	} else {
		return mod.RegFileCmd(executablePath, help)
	}
}

func regModAbbrs(meta *meta_file.MetaFile, mod *core.CmdTree) {
	abbrs := meta.Get("abbrs")
	if len(abbrs) == 0 {
		abbrs = meta.Get("abbr")
	}
	if len(abbrs) == 0 {
		return
	}
	abbrsSep := mod.Strs.AbbrsSep
	mod.AddAbbrs(strings.Split(abbrs, abbrsSep)...)
}

func regFlowAbbrs(meta *meta_file.MetaFile, cmds *core.CmdTree, cmdPath []string) {
	abbrsStr := meta.Get("abbrs")
	if len(abbrsStr) == 0 {
		abbrsStr = meta.Get("abbr")
	}
	if len(abbrsStr) == 0 {
		return
	}

	pathSep := cmds.Strs.PathSep
	abbrsSep := cmds.Strs.AbbrsSep

	var abbrs [][]string
	for _, abbrSeg := range strings.Split(abbrsStr, pathSep) {
		abbrList := strings.Split(abbrSeg, abbrsSep)
		abbrs = append(abbrs, abbrList)
	}

	mod := cmds
	for i, cmd := range cmdPath {
		mod = mod.GetOrAddSub(cmd)
		if i < len(abbrs) {
			mod.AddAbbrs(abbrs[i]...)
		}
	}
}

func regArgs(meta *meta_file.MetaFile, cmd *core.Cmd, abbrsSep string) {
	args := meta.GetSection("args")
	if args == nil {
		args = meta.GetSection("arg")
	}
	if args == nil {
		return
	}
	for _, names := range args.Keys() {
		defVal := args.Get(names)
		nameAndAbbrs := strings.Split(names, abbrsSep)
		name := strings.TrimSpace(nameAndAbbrs[0])
		var argAbbrs []string
		for _, abbr := range nameAndAbbrs[1:] {
			argAbbrs = append(argAbbrs, strings.TrimSpace(abbr))
		}
		cmd.AddArg(name, defVal, argAbbrs...)
	}
}

func regDeps(meta *meta_file.MetaFile, cmd *core.Cmd) {
	deps := meta.GetSection("deps")
	if deps == nil {
		deps = meta.GetSection("dep")
	}
	if deps != nil {
		for _, dep := range deps.Keys() {
			reason := deps.Get(dep)
			cmd.AddDepend(dep, reason)
		}
	}
}

func regEnvOps(
	envAbbrs *core.EnvAbbrs,
	meta *meta_file.MetaFile,
	cmd *core.Cmd,
	abbrsSep string,
	envPathSep string) {

	envOps := meta.GetSection("env")
	if envOps == nil {
		return
	}

	for _, envKey := range envOps.Keys() {
		op := envOps.Get(envKey)
		key := regEnvKeyAbbrs(envAbbrs, envKey, abbrsSep, envPathSep)
		opFields := strings.Split(op, abbrsSep)
		if len(opFields) == 1 {
			opFields = strings.Split(op, ":")
		}
		for _, it := range opFields {
			regEnvOp(cmd, key, it)
		}
	}
}

func regVal2Env(
	envAbbrs *core.EnvAbbrs,
	meta *meta_file.MetaFile,
	cmd *core.Cmd,
	abbrsSep string,
	envPathSep string) {

	writes := meta.GetSection("env.write")
	if writes == nil {
		writes = meta.GetSection("val2env")
	}
	if writes == nil {
		return
	}

	for _, envKey := range writes.Keys() {
		val := writes.Get(envKey)
		key := regEnvKeyAbbrs(envAbbrs, envKey, abbrsSep, envPathSep)
		cmd.AddVal2Env(key, val)
	}
}

func regArg2Env(
	envAbbrs *core.EnvAbbrs,
	meta *meta_file.MetaFile,
	cmd *core.Cmd,
	abbrsSep string,
	envPathSep string) {

	writes := meta.GetSection("env.from-arg")
	if writes == nil {
		writes = meta.GetSection("env.arg")
	}
	if writes == nil {
		writes = meta.GetSection("arg2env")
	}
	if writes == nil {
		return
	}

	for _, envKey := range writes.Keys() {
		argName := writes.Get(envKey)
		key := regEnvKeyAbbrs(envAbbrs, envKey, abbrsSep, envPathSep)
		cmd.AddArg2Env(key, argName)
	}
}

func regEnvKeyAbbrs(
	envAbbrs *core.EnvAbbrs,
	envKeyWithAbbrs string,
	abbrsSep string,
	envPathSep string) (key string) {

	var path []string
	segs := strings.Split(envKeyWithAbbrs, envPathSep)
	for _, seg := range segs {
		var abbrs []string
		fields := strings.Split(seg, abbrsSep)
		if len(fields) == 1 {
			fields = strings.Split(seg, ":")
		}
		for _, abbr := range fields {
			abbrs = append(abbrs, strings.TrimSpace(abbr))
		}
		name := abbrs[0]
		abbrs = abbrs[1:]
		subEnvAbbrs := envAbbrs.GetOrAddSub(name)
		if len(abbrs) > 0 {
			envAbbrs.AddSubAbbrs(name, abbrs...)
		}
		envAbbrs = subEnvAbbrs
		path = append(path, name)
	}

	return strings.Join(path, envPathSep)
}

func regEnvOp(cmd *core.Cmd, key string, opOrigin string) {
	op := strings.ToLower(opOrigin)
	may := strings.Index(op, "may") >= 0 || strings.Index(op, "opt") >= 0
	write := strings.Index(op, "w") >= 0
	read := strings.Index(op, "rd") >= 0 ||
		strings.Index(op, "read") >= 0 || op == "r"
	if write && read {
		panic(fmt.Errorf("[LoadLocalMods.regEnvOp] "+
			"parse env r|w definition failed: %v", opOrigin))
	}
	if write {
		if may {
			cmd.AddEnvOp(key, core.EnvOpTypeMayWrite)
		} else {
			cmd.AddEnvOp(key, core.EnvOpTypeWrite)
		}
	}
	if read {
		if may {
			cmd.AddEnvOp(key, core.EnvOpTypeMayRead)
		} else {
			cmd.AddEnvOp(key, core.EnvOpTypeRead)
		}
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && !info.IsDir()
}
