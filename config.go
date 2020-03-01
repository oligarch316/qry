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

// Config TODO
type Config struct {
	Convert           ConfigConvert
	IgnoreInvalidKeys bool
	LogTrace          Trace
	Separators        ConfigSeparate
	StructParse       ConfigStructParse

	modes levelModes
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
		StructParse: ConfigStructParse{TagName: "qry"},

		modes: levelModes{
			// indirect/container => update | literal => disallowed
			LevelQuery:     setMode{},
			LevelField:     setMode{},
			LevelValueList: setMode{},

			// indirect/container => update | literal => allowed
			LevelKey:   setMode{AllowLiteral: true},
			LevelValue: setMode{AllowLiteral: true},
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
func (c Config) NewDecoder(opts ...Option) *Decoder {
	var (
		cfg          = c.With(opts...)
		converter    = newConverter(cfg.Convert)
		unmarshaler  = newUnmarshaler(converter.Unescape)
		structParser = newStructParser(cfg.StructParse, unmarshaler.check)
	)

	return &Decoder{
		baseModes:         cfg.modes,
		ignoreInvalidKeys: cfg.IgnoreInvalidKeys,
		logTrace:          cfg.LogTrace,
		separators:        cfg.Separators,

		converter:    converter,
		structParser: structParser,
		unmarshaler:  unmarshaler,
	}
}

// ===== Options =====

// Option TODO
type Option func(*Config)

func mergeOptions(opts []Option) Option {
	return func(c *Config) {
		for _, opt := range opts {
			opt(c)
		}
	}
}

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

// SetLevelVia TODO
func SetLevelVia(level DecodeLevel, setOpts ...SetOption) Option {
	return func(c *Config) { c.modes = c.modes.modifiedClone(level, setOpts) }
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

// SetAllLevelsVia TODO
func SetAllLevelsVia(setOpts ...SetOption) Option {
	var opts []Option
	for _, level := range []DecodeLevel{LevelQuery, LevelField, LevelKey, LevelValueList, LevelValue} {
		opts = append(opts, SetLevelVia(level, setOpts...))
	}

	return mergeOptions(opts)
}

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
	return func(c *Config) { c.StructParse.TagName = name }
}
