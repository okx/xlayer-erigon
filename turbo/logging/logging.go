package logging

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ledgerwatch/log/v3"
	"github.com/spf13/cobra"
	"github.com/urfave/cli/v2"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ledgerwatch/erigon-lib/common/metrics"
)

const timeFormat = "2006-01-02T15:04:05-0700"
const errorKey = "LOG15_ERROR"

// Determine the log dir path based on the given urfave context
func LogDirPath(ctx *cli.Context) string {
	dirPath := ""
	if !ctx.Bool(LogDirDisableFlag.Name) {
		dirPath = ctx.String(LogDirPathFlag.Name)
		if dirPath == "" {
			datadir := ctx.String("datadir")
			if datadir != "" {
				dirPath = filepath.Join(datadir, "logs")
			}
		}
	}
	return dirPath
}

// SetupLoggerCtx performs the logging setup according to the parameters
// containted in the given urfave context. It returns either root logger,
// if rootHandler argument is set to true, or a newly created logger.
// This is to ensure gradual transition to the use of non-root logger thoughout
// the erigon code without a huge change at once.
// This function which is used in Erigon itself.
// Note: urfave and cobra are two CLI frameworks/libraries for the same functionalities
// and it would make sense to choose one over another
func SetupLoggerCtx(filePrefix string, ctx *cli.Context,
	consoleDefaultLevel log.Lvl, dirDefaultLevel log.Lvl, rootHandler bool) log.Logger {
	var consoleJson = ctx.Bool(LogJsonFlag.Name) || ctx.Bool(LogConsoleJsonFlag.Name)
	var dirJson = ctx.Bool(LogDirJsonFlag.Name)

	metrics.DelayLoggingEnabled = ctx.Bool(LogBlockDelayFlag.Name)

	consoleLevel, lErr := tryGetLogLevel(ctx.String(LogConsoleVerbosityFlag.Name))
	if lErr != nil {
		// try verbosity flag
		consoleLevel, lErr = tryGetLogLevel(ctx.String(LogVerbosityFlag.Name))
		if lErr != nil {
			consoleLevel = consoleDefaultLevel
		}
	}

	dirLevel, dErr := tryGetLogLevel(ctx.String(LogDirVerbosityFlag.Name))
	if dErr != nil {
		dirLevel = dirDefaultLevel
	}

	dirPath := ""
	if !ctx.Bool(LogDirDisableFlag.Name) && dirPath != "/dev/null" {
		dirPath = ctx.String(LogDirPathFlag.Name)
		if dirPath == "" {
			datadir := ctx.String("datadir")
			if datadir != "" {
				dirPath = filepath.Join(datadir, "logs")
			}
		}
		if logDirPrefix := ctx.String(LogDirPrefixFlag.Name); len(logDirPrefix) > 0 {
			filePrefix = logDirPrefix
		}
	}

	var logger log.Logger
	if rootHandler {
		logger = log.Root()
	} else {
		logger = log.New()
	}

	initSeparatedLogging(logger, filePrefix, dirPath, consoleLevel, dirLevel, consoleJson, dirJson)
	return logger
}

// SetupLoggerCmd perform the logging for a cobra command, and sets it to the root logger
// This is the function which is NOT used by Erigon itself, but instead by some cobra-based commands,
// for example, rpcdaemon or integration.
// Note: urfave and cobra are two CLI frameworks/libraries for the same functionalities
// and it would make sense to choose one over another
func SetupLoggerCmd(filePrefix string, cmd *cobra.Command) log.Logger {

	logJsonVal, ljerr := cmd.Flags().GetBool(LogJsonFlag.Name)
	if ljerr != nil {
		logJsonVal = false
	}

	logConsoleJsonVal, lcjerr := cmd.Flags().GetBool(LogConsoleJsonFlag.Name)
	if lcjerr != nil {
		logConsoleJsonVal = false
	}

	var consoleJson = logJsonVal || logConsoleJsonVal
	dirJson, djerr := cmd.Flags().GetBool(LogDirJsonFlag.Name)
	if djerr != nil {
		dirJson = false
	}

	consoleLevel, lErr := tryGetLogLevel(cmd.Flags().Lookup(LogConsoleVerbosityFlag.Name).Value.String())
	if lErr != nil {
		// try verbosity flag
		consoleLevel, lErr = tryGetLogLevel(cmd.Flags().Lookup(LogVerbosityFlag.Name).Value.String())
		if lErr != nil {
			consoleLevel = log.LvlInfo
		}
	}

	dirLevel, dErr := tryGetLogLevel(cmd.Flags().Lookup(LogDirVerbosityFlag.Name).Value.String())
	if dErr != nil {
		dirLevel = log.LvlInfo
	}

	dirPath := ""
	disableFileLogging, err := cmd.Flags().GetBool(LogDirDisableFlag.Name)
	if err != nil {
		disableFileLogging = false
	}
	if !disableFileLogging && dirPath != "/dev/null" {
		dirPath = cmd.Flags().Lookup(LogDirPathFlag.Name).Value.String()
		if dirPath == "" {
			datadirFlag := cmd.Flags().Lookup("datadir")
			if datadirFlag != nil {
				datadir := datadirFlag.Value.String()
				if datadir != "" {
					dirPath = filepath.Join(datadir, "logs")
				}
			}
		}
		if logDirPrefix := cmd.Flags().Lookup(LogDirPrefixFlag.Name).Value.String(); len(logDirPrefix) > 0 {
			filePrefix = logDirPrefix
		}
	}

	initSeparatedLogging(log.Root(), filePrefix, dirPath, consoleLevel, dirLevel, consoleJson, dirJson)
	return log.Root()
}

// SetupLoggerCmd perform the logging using parametrs specifying by `flag` package, and sets it to the root logger
// This is the function which is NOT used by Erigon itself, but instead by utility commans
func SetupLogger(filePrefix string) log.Logger {
	var logConsoleVerbosity = flag.String(LogConsoleVerbosityFlag.Name, "", LogConsoleVerbosityFlag.Usage)
	var logDirVerbosity = flag.String(LogDirVerbosityFlag.Name, "", LogDirVerbosityFlag.Usage)
	var logDirPath = flag.String(LogDirPathFlag.Name, "", LogDirPathFlag.Usage)
	var logDirPrefix = flag.String(LogDirPrefixFlag.Name, "", LogDirPrefixFlag.Usage)
	var logVerbosity = flag.String(LogVerbosityFlag.Name, "", LogVerbosityFlag.Usage)
	var logConsoleJson = flag.Bool(LogConsoleJsonFlag.Name, false, LogConsoleJsonFlag.Usage)
	var logJson = flag.Bool(LogJsonFlag.Name, false, LogJsonFlag.Usage)
	var logDirJson = flag.Bool(LogDirJsonFlag.Name, false, LogDirJsonFlag.Usage)
	flag.Parse()

	var consoleJson = *logJson || *logConsoleJson
	var dirJson = logDirJson

	consoleLevel, lErr := tryGetLogLevel(*logConsoleVerbosity)
	if lErr != nil {
		// try verbosity flag
		consoleLevel, lErr = tryGetLogLevel(*logVerbosity)
		if lErr != nil {
			consoleLevel = log.LvlInfo
		}
	}

	dirLevel, dErr := tryGetLogLevel(*logDirVerbosity)
	if dErr != nil {
		dirLevel = log.LvlInfo
	}

	if logDirPrefix != nil && len(*logDirPrefix) > 0 {
		filePrefix = *logDirPrefix
	}

	initSeparatedLogging(log.Root(), filePrefix, *logDirPath, consoleLevel, dirLevel, consoleJson, *dirJson)
	return log.Root()
}

// initSeparatedLogging construct a log handler accrosing to the configuration parameters passed to it
// and sets the constructed handler to be the handler of the given logger. It then uses that logger
// to report the status of this initialisation
func initSeparatedLogging(
	logger log.Logger,
	filePrefix string,
	dirPath string,
	consoleLevel log.Lvl,
	dirLevel log.Lvl,
	consoleJson bool,
	dirJson bool) {

	var consoleHandler log.Handler

	if consoleJson {
		consoleHandler = log.LvlFilterHandler(consoleLevel, log.StreamHandler(os.Stderr, JsonFormatEx(true, true)))
	} else {
		consoleHandler = log.LvlFilterHandler(consoleLevel, log.StderrHandler)
	}
	logger.SetHandler(consoleHandler)

	if len(dirPath) == 0 {
		logger.Info("console logging only")
		return
	}

	err := os.MkdirAll(dirPath, 0764)
	if err != nil {
		logger.Warn("failed to create log dir, console logging only")
		return
	}
	dirFormat := log.TerminalFormatNoColor()
	if dirJson {
		dirFormat = JsonFormatEx(true, true)
	}

	lumberjack := &lumberjack.Logger{
		Filename:   filepath.Join(dirPath, filePrefix+".log"),
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
	}
	userLog := log.StreamHandler(lumberjack, dirFormat)

	mux := log.MultiHandler(consoleHandler, log.LvlFilterHandler(dirLevel, userLog))
	logger.SetHandler(mux)
	logger.Info("logging to file system", "log dir", dirPath, "file prefix", filePrefix, "log level", dirLevel, "json", dirJson)
}

func JsonFormatEx(pretty, lineSeparated bool) log.Format {
	jsonMarshal := json.Marshal
	if pretty {
		jsonMarshal = func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "    ")
		}
	}

	return log.FormatFunc(func(r *log.Record) []byte {

		r.KeyNames = log.RecordKeyNames{
			Time: "time",
			Msg:  "content",
			Lvl:  "level",
		}

		props := make(map[string]interface{})

		props[r.KeyNames.Time] = r.Time
		props[r.KeyNames.Lvl] = strings.ToUpper(r.Lvl.String())
		props[r.KeyNames.Msg] = r.Msg

		for i := 0; i < len(r.Ctx); i += 2 {
			k, ok := r.Ctx[i].(string)
			if !ok {
				props[errorKey] = fmt.Sprintf("%+v is not a string key", r.Ctx[i])
			}
			props[k] = formatJSONValue(r.Ctx[i+1])
		}

		b, err := jsonMarshal(props)
		if err != nil {
			b, _ = jsonMarshal(map[string]string{
				errorKey: err.Error(),
			})
			return b
		}

		if lineSeparated {
			b = append(b, '\n')
		}

		return b
	})
}

func formatJSONValue(value interface{}) interface{} {
	value = formatShared(value)

	switch value.(type) {
	case int, int8, int16, int32, int64, float32, float64, uint, uint8, uint16, uint32, uint64, string:
		return value
	case interface{}, map[string]interface{}, []interface{}:
		return value
	default:
		return fmt.Sprintf("%+v", value)
	}
}

func formatShared(value interface{}) (result interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
				result = "nil"
			} else {
				panic(err)
			}
		}
	}()

	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)

	case error:
		return v.Error()

	case fmt.Stringer:
		return v.String()

	default:
		return v
	}
}

func tryGetLogLevel(s string) (log.Lvl, error) {
	lvl, err := log.LvlFromString(s)
	if err != nil {
		l, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		return log.Lvl(l), nil
	}
	return lvl, nil
}
