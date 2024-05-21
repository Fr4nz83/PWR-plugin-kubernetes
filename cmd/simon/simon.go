package simon

import (
	goflag "flag"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/hkust-adsl/kubernetes-scheduler-simulator/cmd/apply"
	"github.com/hkust-adsl/kubernetes-scheduler-simulator/cmd/doc"
	"github.com/hkust-adsl/kubernetes-scheduler-simulator/cmd/version"
)

const (
	EnvLogLevel = "LOGLEVEL"
	LogPanic    = "PANIC"
	LogFatal    = "FATAL"
	LogError    = "ERROR"
	LogWarn     = "WARN"
	LogInfo     = "INFO"
	LogDebug    = "DEBUG"
	LogTrace    = "TRACE"
)

func NewSimonCommand() *cobra.Command {

	// NOTE: cobra.Command{...} instantiates a struct of cobra.Command. The "&" operator makes Go (1) instantiate the struct on the heap rather than the stack and (2)
	// returns a pointer to the instantiated struct.
	simonCmd := &cobra.Command{
		Use:   "simon",
		Short: "Simon is a simulator, which will simulate a cluster and simulate workload scheduling.",
	}

	// Add the cobra commands defined in the other packages within "./cmd".
	// NOTE: the "." operator in go automatically dereferences a pointer, if it is used with a pointer.
	simonCmd.AddCommand(version.VersionCmd,
		apply.ApplyCmd,
		doc.GenDoc.DocCmd)

	// NOTE: cliflag.WordSepNormalizeFunc is a function provided by the cliflag packagethat normalizes flag names from camelCase to words separated by dashes.
	// When users run, e.g., simon apply --num-iterations 5, the normalization function ensures that numIterations is correctly recognized as the flag numIterations.
	simonCmd.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc)
	// Note: This line adds all Go flags from the goflag.CommandLine flag set to the flags of the simonCmd command.
	simonCmd.Flags().AddGoFlagSet(goflag.CommandLine)
	simonCmd.DisableAutoGenTag = true

	return simonCmd
}

func init() {
	logLevel := os.Getenv(EnvLogLevel)
	switch logLevel {
	case LogPanic:
		log.SetLevel(log.PanicLevel)
	case LogFatal:
		log.SetLevel(log.FatalLevel)
	case LogError:
		log.SetLevel(log.ErrorLevel)
	case LogWarn:
		log.SetLevel(log.WarnLevel)
	case LogInfo:
		log.SetLevel(log.InfoLevel)
	case LogDebug:
		log.SetLevel(log.DebugLevel)
	case LogTrace:
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}
