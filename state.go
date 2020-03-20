package qry

import "fmt"

// SetOption TODO
type SetOption string

// Set TODO
const (
	SetAllowLiteral     SetOption = "allowLiteral"
	SetDisallowLiteral  SetOption = "disallowLiteral"
	SetReplaceContainer SetOption = "replaceContainer"
	SetUpdateContainer  SetOption = "updateCoantainer"
	SetReplaceIndirect  SetOption = "replaceIndirect"
	SetUpdateIndirect   SetOption = "updateIndirect"
)

func (so SetOption) valid() bool {
	switch so {
	case SetAllowLiteral, SetDisallowLiteral, SetReplaceIndirect, SetUpdateIndirect, SetReplaceContainer, SetUpdateContainer:
		return true
	}
	return false
}

type setMode struct{ AllowLiteral, ReplaceContainer, ReplaceIndirect bool }

func (sm *setMode) modify(opts []SetOption) {
	for _, opt := range opts {
		switch opt {
		case SetAllowLiteral:
			sm.AllowLiteral = true
		case SetDisallowLiteral:
			sm.AllowLiteral = false
		case SetReplaceContainer:
			sm.ReplaceContainer = true
		case SetUpdateContainer:
			sm.ReplaceContainer = false
		case SetReplaceIndirect:
			sm.ReplaceIndirect = true
		case SetUpdateIndirect:
			sm.ReplaceIndirect = false
		}
	}
}

// SetOptionsMap TODO
type SetOptionsMap map[DecodeLevel][]SetOption

func (som SetOptionsMap) validate() error {
	for level := range som {
		if !level.validInput() {
			return fmt.Errorf("invalid set level: %s", level)
		}
	}
	return nil
}

type levelModes map[DecodeLevel]setMode

func (lm levelModes) with(optsMap SetOptionsMap) levelModes {
	if len(optsMap) < 1 {
		return lm
	}

	res := make(levelModes)
	for level, mode := range lm {
		if opts, ok := optsMap[level]; ok {
			mode.modify(opts)
		}
		res[level] = mode
	}
	return res
}

type decodeState struct {
	modes levelModes
	trace Trace
}

func (ds *decodeState) child() *decodeState {
	return &decodeState{
		modes: ds.modes,
		trace: ds.trace.Child(),
	}
}

func (ds *decodeState) childWithSetOpts(optsMap SetOptionsMap) *decodeState {
	return &decodeState{
		modes: ds.modes.with(optsMap),
		trace: ds.trace.Child(),
	}
}
