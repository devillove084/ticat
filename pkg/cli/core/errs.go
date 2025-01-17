package core

type TolerableErr struct {
	Err    interface{}
	File   string
	Source string
	Reason string
}

type ConflictedWithSameSource map[string][]TolerableErr

type TolerableErrs struct {
	Uncatalogeds          []TolerableErr
	ConflictedWithBuiltin ConflictedWithSameSource
	Conflicteds           map[string]ConflictedWithSameSource
}

func NewTolerableErrs() *TolerableErrs {
	return &TolerableErrs{
		nil,
		ConflictedWithSameSource{},
		map[string]ConflictedWithSameSource{},
	}
}

func (self *TolerableErrs) OnErr(err interface{}, source string, file string, reason string) {
	conflicted, ok := err.(ErrConflicted)
	if !ok {
		self.Uncatalogeds = append(self.Uncatalogeds, TolerableErr{err, file, source, reason})
		return
	}

	old := conflicted.GetOldSource()
	var conflictedMap ConflictedWithSameSource
	if len(old) == 0 {
		conflictedMap = self.ConflictedWithBuiltin
	} else {
		conflictedMap, ok = self.Conflicteds[old]
		if conflictedMap == nil {
			conflictedMap = ConflictedWithSameSource{}
			self.Conflicteds[old] = conflictedMap
		}
	}

	list, _ := conflictedMap[source]
	list = append(list, TolerableErr{err, file, source, reason})
	conflictedMap[source] = list
}

type ErrConflicted interface {
	Error() string
	GetOldSource() string
	GetConflictedCmdPath() []string
}

type CmdTreeErrExecutableConflicted struct {
	Str       string
	CmdPath   []string
	OldSource string
}

func (self CmdTreeErrExecutableConflicted) Error() string {
	return self.Str
}
func (self CmdTreeErrExecutableConflicted) GetOldSource() string {
	return self.OldSource
}
func (self CmdTreeErrExecutableConflicted) GetConflictedCmdPath() []string {
	return self.CmdPath
}

type CmdTreeErrSubCmdConflicted struct {
	Str           string
	ParentCmdPath []string
	SubCmdName    string
	OldSource     string
}

func (self CmdTreeErrSubCmdConflicted) Error() string {
	return self.Str
}
func (self CmdTreeErrSubCmdConflicted) GetOldSource() string {
	return self.OldSource
}
func (self CmdTreeErrSubCmdConflicted) GetConflictedCmdPath() []string {
	return append(self.ParentCmdPath, self.SubCmdName)
}

type CmdTreeErrSubAbbrConflicted struct {
	Str           string
	ParentCmdPath []string
	Abbr          string
	ForOldCmdName string
	ForNewCmdName string
	OldSource     string
}

func (self CmdTreeErrSubAbbrConflicted) Error() string {
	return self.Str
}
func (self CmdTreeErrSubAbbrConflicted) GetOldSource() string {
	return self.OldSource
}
func (self CmdTreeErrSubAbbrConflicted) GetConflictedCmdPath() []string {
	return append(self.ParentCmdPath, self.Abbr)
}

type CmdMissedEnvValWhenRenderFlow struct {
	Str          string
	CmdPath      string
	MetaFilePath string
	Source       string
	MissedKey    string
}

func (self CmdMissedEnvValWhenRenderFlow) Error() string {
	return self.Str
}
