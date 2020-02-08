package qry

// SetOption TODO
type SetOption string

// Set TODO
const (
	SetAllowLiteral     SetOption = "allowLiteral"
	SetDisallowLiteral  SetOption = "diallowLiteral"
	SetReplaceContainer SetOption = "replaceContainer"
	SetUpdateContainer  SetOption = "updateCoantainer"
	SetReplaceIndirect  SetOption = "replaceIndirect"
	SetUpdateIndirect   SetOption = "updateIndirect"
)

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

type levelModes map[DecodeLevel]setMode

func (lm levelModes) modifiedClone(level DecodeLevel, opts []SetOption) levelModes {
	if len(opts) < 1 {
		return lm
	}

	res := make(levelModes)
	for lvl, mode := range lm {
		if lvl == level {
			mode.modify(opts)
		}
		res[lvl] = mode
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

func (ds *decodeState) childWithSetMode(level DecodeLevel, opts []SetOption) *decodeState {
	return &decodeState{
		modes: ds.modes.modifiedClone(level, opts),
		trace: ds.trace.Child(),
	}
}
