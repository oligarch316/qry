package qry

import (
	"log"
	"net/url"
	"reflect"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// Config TODO
	Config struct {
		Convert    ConvertConfig
		Separators SeparatorConfig
		LogTrace   Trace

		modes levelModes
	}

	// Option TODO
	Option func(*Config)
)

func defaultConfig() Config {
	return Config{
		Convert: ConvertConfig{
			IntegerBase: 0,
			Unescape:    url.QueryUnescape,
		},
		Separators: SeparatorConfig{
			Fields:  newSeparatorSet('&').Split,
			KeyVals: newSeparatorSet('=').Pair,
			Values:  newSeparatorSet(',').Split,
		},
		modes: levelModes{
			// indirect/container => replace | literal => disallowed
			LevelQuery:     setMode{},
			LevelField:     setMode{},
			LevelValueList: setMode{},

			// indirect/container => replace | literal => allowed
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
	cfg := c.With(opts...)
	return &Decoder{
		separators:  cfg.Separators,
		baseModes:   cfg.modes,
		logTrace:    cfg.LogTrace,
		converter:   newConverter(cfg.Convert),
		unmarshaler: newUnmarshaler(cfg.Convert.Unescape),
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

// ----- Separator options

// FieldSeparatorsAs TODO
func FieldSeparatorsAs(seps ...rune) Option {
	return func(c *Config) { c.Separators.Fields = newSeparatorSet(seps...).Split }
}

// KeyValSeparatorsAs TODO
func KeyValSeparatorsAs(seps ...rune) Option {
	return func(c *Config) { c.Separators.KeyVals = newSeparatorSet(seps...).Pair }
}

// ValueSeparatorsAs TODO
func ValueSeparatorsAs(seps ...rune) Option {
	return func(c *Config) { c.Separators.Values = newSeparatorSet(seps...).Split }
}

// ----- Set mode options

// SetOptionsAs TODO
func SetOptionsAs(level DecodeLevel, setOpts ...SetOption) Option {
	return func(c *Config) { c.modes = c.modes.modifiedClone(level, setOpts...) }
}

// SetQueryOptionsAs TODO
func SetQueryOptionsAs(setOpts ...SetOption) Option { return SetOptionsAs(LevelQuery, setOpts...) }

// SetFieldOptionsAs TODO
func SetFieldOptionsAs(setOpts ...SetOption) Option { return SetOptionsAs(LevelField, setOpts...) }

// SetKeyOptionsAs TODO
func SetKeyOptionsAs(setOpts ...SetOption) Option { return SetOptionsAs(LevelKey, setOpts...) }

// SetValueListOptionsAs TODO
func SetValueListOptionsAs(setOpts ...SetOption) Option {
	return SetOptionsAs(LevelValueList, setOpts...)
}

// SetValueOptionsAs TODO
func SetValueOptionsAs(setOpts ...SetOption) Option { return SetOptionsAs(LevelValue, setOpts...) }

// ----- LogTrace options

// MarkLogTrace TODO
func MarkLogTrace(marker func(DecodeLevel, string, reflect.Value)) Option {
	return func(c *Config) { c.LogTrace = MarkTrace(marker) }
}

// StdLogTrace TODO
func StdLogTrace(l *log.Logger) Option {
	marker := func(level DecodeLevel, input string, target reflect.Value) {
		l.Print(level.newInfo(input, target).String())
	}

	return MarkLogTrace(marker)
}

// LogrusLogTrace TODO
func LogrusLogTrace(logger *logrus.Logger, level logrus.Level) Option {
	marker := func(decodeLevel DecodeLevel, input string, target reflect.Value) {
		logger.WithFields(logrus.Fields{
			"decodeLevel": decodeLevel.String(),
			"input":       input,
			"target":      target.Type().String(),
			"targetKind":  target.Kind().String(),
		}).Log(level, "decode "+decodeLevel.String())
	}

	return MarkLogTrace(marker)
}

// ZapLogTrace TODO
func ZapLogTrace(logger *zap.Logger, level zapcore.Level) Option {
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

	return MarkLogTrace(marker)
}
