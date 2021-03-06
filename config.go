package qry

// TODO: Possibly move logrus and zap dependencies to a separate package?
// Slightly annoying that this query decoding lib carries heavy structured
// logging dependencies with it.

import (
	"log"
	"net/url"
	"reflect"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ===== Config =====

const (
	configDefaultBaseTagName  = "qry"
	configDefaultSetTagSuffix = "Set"
	configDefaultSetTagName   = configDefaultBaseTagName + configDefaultSetTagSuffix
)

var configDefaultLevelModes = levelModes{
	// indirect/container => update | literal => disallowed
	LevelQuery:     setMode{},
	LevelField:     setMode{},
	LevelValueList: setMode{},

	// indirect/container => update | literal => allowed
	LevelKey:   setMode{AllowLiteral: true},
	LevelValue: setMode{AllowLiteral: true},
}

// Config TODO
type Config struct {
	Convert           ConfigConvert
	IgnoreInvalidKeys bool
	LogTrace          Trace
	Separators        ConfigSeparate
	SetModes          SetOptionsMap
	StructParse       ConfigStructParse
}

func defaultConfig() Config {
	return Config{
		Convert: ConfigConvert{
			IntegerBase: 0,
			Unescape:    url.QueryUnescape,
		},
		IgnoreInvalidKeys: false,
		LogTrace:          nil,
		Separators: ConfigSeparate{
			Fields:   newSeparatorSet('&').Split, // TODO: Add ';' to default Fields separator set? Check RFC
			KeyVals:  newSeparatorSet('=').Pair,
			KeyChain: separateNoopSplit,
			Values:   newSeparatorSet(',').Split,
		},
		SetModes: nil,
		StructParse: ConfigStructParse{
			BaseTagName: configDefaultBaseTagName,
			SetTagName:  configDefaultSetTagName,
		},
	}
}

// NewConfig TODO
func NewConfig(opts ...Option) Config { return defaultConfig().With(opts...) }

// With TODO
func (c Config) With(opts ...Option) Config {
	// NOTE: Added complexity to the Config type may require added complexity here
	res := c

	for _, opt := range opts {
		opt(&res)
	}
	return res
}

// NewDecoder TODO
func (c Config) NewDecoder(opts ...Option) (*Decoder, error) {
	cfg := c.With(opts...)

	if err := cfg.SetModes.validate(); err != nil {
		return nil, err
	}

	var (
		converter    = newConverter(cfg.Convert)
		unmarshaler  = newUnmarshaler(converter.Unescape)
		structParser = newStructParser(cfg.StructParse, unmarshaler.check)
	)

	return &Decoder{
		baseModes:         configDefaultLevelModes.with(cfg.SetModes),
		ignoreInvalidKeys: cfg.IgnoreInvalidKeys,
		logTrace:          cfg.LogTrace,
		separators:        cfg.Separators,

		converter:    converter,
		structParser: structParser,
		unmarshaler:  unmarshaler,
	}, nil
}

// ===== Options =====

// Option TODO
type Option func(*Config)

// ----- Convert options

// ConvertIntegerBaseAs TODO
func ConvertIntegerBaseAs(base int) Option {
	return func(c *Config) { c.Convert.IntegerBase = base }
}

// ConvertUnescapeAs TODO
func ConvertUnescapeAs(unescape func(string) (string, error)) Option {
	return func(c *Config) { c.Convert.Unescape = unescape }
}

// ----- Ignore invalid keys option

// IgnoreInvalidKeys TODO
func IgnoreInvalidKeys(b bool) Option {
	return func(c *Config) { c.IgnoreInvalidKeys = b }
}

// ----- Separator options

// SeparateFieldsBy TODO
func SeparateFieldsBy(seps ...rune) Option {
	return func(c *Config) { c.Separators.Fields = newSeparatorSet(seps...).Split }
}

// SeparateKeyValsBy TODO
func SeparateKeyValsBy(seps ...rune) Option {
	return func(c *Config) { c.Separators.KeyVals = newSeparatorSet(seps...).Pair }
}

// SeparateKeyChainBy TODO
func SeparateKeyChainBy(seps ...rune) Option {
	return func(c *Config) { c.Separators.KeyChain = newSeparatorSet(seps...).Split }
}

// SeparateValuesBy TODO
func SeparateValuesBy(seps ...rune) Option {
	return func(c *Config) { c.Separators.Values = newSeparatorSet(seps...).Split }
}

// ----- Set mode options

// SetVia TODO
func SetVia(optsMap SetOptionsMap) Option {
	return func(c *Config) {
		if c.SetModes == nil {
			c.SetModes = make(SetOptionsMap)
		}

		for level, opts := range optsMap {
			c.SetModes[level] = append(c.SetModes[level], opts...)
		}
	}
}

// SetLevelVia TODO
func SetLevelVia(level DecodeLevel, setOpts ...SetOption) Option {
	return SetVia(SetOptionsMap{level: setOpts})
}

// SetAllLevelsVia TODO
func SetAllLevelsVia(setOpts ...SetOption) Option {
	optsMap := make(SetOptionsMap)
	for i := LevelQuery; i <= LevelValue; i++ {
		optsMap[i] = setOpts
	}

	return SetVia(optsMap)
}

// SetQueryVia TODO
func SetQueryVia(setOpts ...SetOption) Option { return SetLevelVia(LevelQuery, setOpts...) }

// SetFieldVia TODO
func SetFieldVia(setOpts ...SetOption) Option { return SetLevelVia(LevelField, setOpts...) }

// SetKeyVia TODO
func SetKeyVia(setOpts ...SetOption) Option { return SetLevelVia(LevelKey, setOpts...) }

// SetValueListVia TODO
func SetValueListVia(setOpts ...SetOption) Option { return SetLevelVia(LevelValueList, setOpts...) }

// SetValueVia TODO
func SetValueVia(setOpts ...SetOption) Option { return SetLevelVia(LevelValue, setOpts...) }

// ----- Log trace options

// LogToMarker TODO
func LogToMarker(marker func(DecodeLevel, string, reflect.Value)) Option {
	return func(c *Config) { c.LogTrace = TraceMarker(marker) }
}

// LogToStd TODO
func LogToStd(l *log.Logger) Option {
	marker := func(level DecodeLevel, input string, target reflect.Value) {
		l.Print(level.newInfo(input, target).String())
	}

	return LogToMarker(marker)
}

// LogToLogrus TODO
func LogToLogrus(logger *logrus.Logger, level logrus.Level) Option {
	marker := func(decodeLevel DecodeLevel, input string, target reflect.Value) {
		logger.WithFields(logrus.Fields{
			"decodeLevel": decodeLevel.String(),
			"input":       input,
			"target":      target.Type().String(),
			"targetKind":  target.Kind().String(),
		}).Log(level, "decode "+decodeLevel.String())
	}

	return LogToMarker(marker)
}

// LogToZap TODO
func LogToZap(logger *zap.Logger, level zapcore.Level) Option {
	var loggerFunc func(string, ...zap.Field)

	switch level {
	case zapcore.DebugLevel:
		loggerFunc = logger.Debug
	case zapcore.InfoLevel:
		loggerFunc = logger.Info
	case zapcore.WarnLevel:
		loggerFunc = logger.Warn
	case zapcore.ErrorLevel:
		loggerFunc = logger.Error
	case zapcore.DPanicLevel:
		loggerFunc = logger.DPanic
	case zapcore.PanicLevel:
		loggerFunc = logger.Panic
	case zapcore.FatalLevel:
		loggerFunc = logger.Fatal
	}

	marker := func(decodeLevel DecodeLevel, input string, target reflect.Value) {
		loggerFunc("decode "+decodeLevel.String(),
			zap.String("decodeLevel", decodeLevel.String()),
			zap.String("input", input),
			zap.String("target", target.Type().String()),
			zap.String("targetKind", target.Kind().String()),
		)
	}

	return LogToMarker(marker)
}

// ----- StructParse options

// StructTagNameAs TODO
func StructTagNameAs(name string) Option {
	return func(c *Config) {
		c.StructParse.BaseTagName = name
		c.StructParse.SetTagName = name + configDefaultSetTagSuffix
	}
}
