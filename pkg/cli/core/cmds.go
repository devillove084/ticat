package core

import (
	"fmt"
	"strings"
)

// TODO: share some code with EnvAbbrs

type CmdTreeStrs struct {
	RootDisplayName          string
	BuiltinDisplayName       string
	PathSep                  string
	PathAlterSeps            string
	AbbrsSep                 string
	EnvOpSep                 string
	EnvValDelAllMark         string
	EnvKeyValSep             string
	EnvPathSep               string
	ProtoSep                 string
	FlowTemplateBracketLeft  string
	FlowTemplateBracketRight string
}

type CmdTree struct {
	Strs            *CmdTreeStrs
	name            string
	parent          *CmdTree
	subs            map[string]*CmdTree
	subOrderedNames []string
	cmd             *Cmd
	subAbbrs        map[string][]string
	subAbbrsRevIdx  map[string]string
	hidden          bool
}

func NewCmdTree(strs *CmdTreeStrs) *CmdTree {
	return &CmdTree{
		strs,
		"",
		nil,
		map[string]*CmdTree{},
		[]string{},
		nil,
		map[string][]string{},
		map[string]string{},
		false,
	}
}

func (self *CmdTree) Execute(
	argv ArgVals,
	cc *Cli,
	env *Env,
	flow *ParsedCmds,
	currCmdIdx int) (int, bool) {

	if self.cmd == nil {
		return currCmdIdx, true
	} else {
		return self.cmd.Execute(argv, cc, env, flow, currCmdIdx)
	}
}

func (self *CmdTree) cmdConflictCheck(help string, funName string) {
	if self.cmd == nil {
		return
	}
	// Allow to overwrite empty dir cmd
	if self.cmd.Type() == CmdTypeEmptyDir {
		return
	}
	err := CmdTreeErrExecutableConflicted{
		fmt.Sprintf("reg-cmd conflicted. old-help '%s', new-help '%s'",
			strings.Split(self.cmd.Help(), "\n")[0],
			strings.Split(help, "\n")[0]),
		self.Path(),
		self.cmd.Source(),
	}
	panic(err)
}

func (self *CmdTree) SetHidden() *CmdTree {
	self.hidden = true
	return self
}

func (self *CmdTree) IsHidden() bool {
	return self.hidden
}

func (self *CmdTree) RegCmd(cmd NormalCmd, help string) *Cmd {
	self.cmdConflictCheck(help, "RegCmd")
	self.cmd = NewCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) RegFileCmd(cmd string, help string) *Cmd {
	self.cmdConflictCheck(help, "RegFileCmd")
	self.cmd = NewFileCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) RegDirWithCmd(cmd string, help string) *Cmd {
	self.cmdConflictCheck(help, "RegDirWithCmd")
	self.cmd = NewDirWithCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) RegEmptyDirCmd(dir string, help string) *Cmd {
	//dirnore empty dir cmd register
	if self.cmd != nil {
		return self.cmd
	}
	self.cmd = NewEmptyDirCmd(self, help, dir)
	return self.cmd
}

func (self *CmdTree) RegEmptyCmd(help string) *Cmd {
	self.cmdConflictCheck(help, "RegEmptyCmd")
	self.cmd = NewEmptyCmd(self, help)
	return self.cmd
}

func (self *CmdTree) RegFlowCmd(flow []string, help string) *Cmd {
	self.cmdConflictCheck(help, "RegFlowCmd")
	self.cmd = NewFlowCmd(self, help, flow)
	return self.cmd
}

func (self *CmdTree) RegPowerCmd(cmd PowerCmd, help string) *Cmd {
	self.cmdConflictCheck(help, "RegPowerCmd")
	self.cmd = NewPowerCmd(self, help, cmd)
	return self.cmd
}

func (self *CmdTree) AddSub(name string, abbrs ...string) *CmdTree {
	if old, ok := self.subs[name]; ok && old.name != name {
		err := CmdTreeErrSubCmdConflicted{
			fmt.Sprintf("sub-cmd name conflicted: %s", name),
			self.Path(),
			name,
			old.cmd.Source(),
		}
		panic(err)
	}
	sub := NewCmdTree(self.Strs)
	sub.name = name
	sub.parent = self
	self.subs[name] = sub
	self.subOrderedNames = append(self.subOrderedNames, name)
	self.addSubAbbrs(name, abbrs...)
	self.subAbbrsRevIdx[name] = name
	return sub
}

func (self *CmdTree) AddAbbrs(abbrs ...string) {
	if self.parent == nil {
		panic(fmt.Errorf("[CmdTree.AddAbbrs] can't add abbrs %v to root", abbrs))
	}
	self.parent.addSubAbbrs(self.name, abbrs...)
}

func (self *CmdTree) GetOrAddSub(path ...string) *CmdTree {
	return self.getSub(true, path...)
}

func (self *CmdTree) HasSub() bool {
	return len(self.subs) != 0 && !self.IsHidden()
}

func (self *CmdTree) GetSub(path ...string) *CmdTree {
	return self.getSub(false, path...)
}

func (self *CmdTree) IsQuiet() bool {
	return self.cmd != nil && self.cmd.IsQuiet()
}

func (self *CmdTree) IsNoExecutableCmd() bool {
	if self.cmd == nil {
		return true
	}
	return self.cmd.IsNoExecutableCmd()
}

func (self *CmdTree) IsPowerCmd() bool {
	return self.cmd != nil && self.cmd.IsPowerCmd()
}

func (self *CmdTree) Parent() *CmdTree {
	return self.parent
}

func (self *CmdTree) IsRoot() bool {
	return self.parent == nil
}

func (self *CmdTree) Name() string {
	return self.name
}

func (self *CmdTree) DisplayName() string {
	if len(self.name) == 0 {
		return self.Strs.RootDisplayName
	}
	return self.name
}

func (self *CmdTree) Cmd() *Cmd {
	return self.cmd
}

func (self *CmdTree) Args() (args Args) {
	if self.cmd == nil {
		return
	}
	return self.cmd.Args()
}

func (self *CmdTree) Path() []string {
	if self.parent == nil {
		return nil
	}
	return append(self.parent.Path(), self.name)
}

func (self *CmdTree) AbbrsPath() []string {
	if self.parent == nil {
		return nil
	}
	abbrs := self.parent.SubAbbrs(self.name)
	if len(abbrs) == 0 {
		return nil
	}
	return append(self.parent.AbbrsPath(), strings.Join(abbrs, self.Strs.AbbrsSep))
}

func (self *CmdTree) Depth() int {
	if self.parent == nil {
		return 0
	} else {
		return self.parent.Depth() + 1
	}
}

func (self *CmdTree) MatchFind(findStrs ...string) bool {
	for _, str := range findStrs {
		if !self.matchFind(str) {
			return false
		}
	}
	return true
}

func (self *CmdTree) matchFind(findStr string) bool {
	if len(findStr) == 0 {
		return true
	}
	if strings.Index(self.name, findStr) >= 0 {
		return true
	}
	if self.cmd != nil && self.cmd.MatchFind(findStr) {
		return true
	}
	if self.parent != nil {
		for _, abbr := range self.parent.SubAbbrs(self.name) {
			if strings.Index(abbr, findStr) >= 0 {
				return true
			}
		}
	}
	return false
}

func (self *CmdTree) DisplayPath() string {
	path := self.Path()
	if len(path) == 0 {
		return self.Strs.RootDisplayName
	} else {
		return strings.Join(path, self.Strs.PathSep)
	}
}

func (self *CmdTree) DisplayAbbrsPath() string {
	path := self.AbbrsPath()
	if len(path) == 0 {
		return ""
	} else {
		return strings.Join(path, self.Strs.PathSep)
	}
}

func (self *CmdTree) Realname(nameOrAbbr string) (realname string) {
	if self.parent == nil {
		return
	}
	realname, _ = self.parent.subAbbrsRevIdx[nameOrAbbr]
	return
}

func (self *CmdTree) SubNames() []string {
	return self.subOrderedNames
}

func (self *CmdTree) SubAbbrs(name string) (abbrs []string) {
	abbrs, _ = self.subAbbrs[name]
	return
}

func (self *CmdTree) Abbrs() (abbrs []string) {
	if self.parent == nil {
		return
	}
	return self.parent.SubAbbrs(self.name)
}

func (self *CmdTree) addSubAbbrs(name string, abbrs ...string) {
	for _, abbr := range append([]string{name}, abbrs...) {
		if len(abbr) == 0 {
			continue
		}
		old, ok := self.subAbbrsRevIdx[abbr]
		if old == name {
			continue
		}
		if ok {
			err := CmdTreeErrSubAbbrConflicted{
				fmt.Sprintf("%s: sub command abbr name '%s' conflicted, "+
					"old for '%s', new for '%s'",
					self.DisplayPath(), abbr, old, name),
				self.Path(),
				abbr,
				old,
				name,
				self.GetSub(old).cmd.Source(),
			}
			panic(err)
		}
		self.subAbbrsRevIdx[abbr] = name
		olds, _ := self.subAbbrs[name]
		self.subAbbrs[name] = append(olds, abbr)
	}
}

func (self *CmdTree) getSub(addIfNotExists bool, path ...string) *CmdTree {
	if len(path) == 0 {
		return self
	}
	name := path[0]
	if realName, ok := self.subAbbrsRevIdx[name]; ok {
		name = realName
	}
	sub, ok := self.subs[name]
	if !ok {
		if !addIfNotExists {
			return nil
		}
		sub = self.AddSub(name)
	}
	return sub.getSub(addIfNotExists, path[1:]...)
}
