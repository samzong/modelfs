package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	datasetv1alpha1 "github.com/BaizeAI/dataset/api/dataset/v1alpha1"
	modelv1 "github.com/samzong/modelfs/api/v1"
	"github.com/samzong/modelfs/controllers"
	versionpkg "github.com/samzong/modelfs/internal/version"
	//+kubebuilder:scaffold:imports
)

//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=create;get;list;update

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	version  = "dev"
	commit   = "unknown"
	date     = "unknown"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(modelv1.AddToScheme(scheme))
	utilruntime.Must(datasetv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

type managerConfig struct {
	metricsAddr          string
	enableLeaderElection bool
	probeAddr            string
	showVersion          bool
	zapOpts              zap.Options
}

func parseManagerConfig(fs *flag.FlagSet, args []string) (managerConfig, error) {
	cfg := managerConfig{
		zapOpts: zap.Options{
			Development: true,
		},
	}
	fs.StringVar(&cfg.metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	fs.StringVar(&cfg.probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.BoolVar(&cfg.enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.BoolVar(&cfg.showVersion, "version", false, "Print version information and exit.")
	cfg.zapOpts.BindFlags(fs)
	if err := fs.Parse(args); err != nil {
		return managerConfig{}, err
	}
	return cfg, nil
}

func runManager(
	args []string,
	stdout io.Writer,
	fs *flag.FlagSet,
	readBuildInfo func() (*debug.BuildInfo, bool),
	start func(managerConfig) error,
) error {
	cfg, err := parseManagerConfig(fs, args)
	if err != nil {
		return err
	}
	if cfg.showVersion {
		_, err := fmt.Fprintln(stdout, versionpkg.Info{Version: version, Commit: commit, Date: date}.Format(readBuildInfo))
		return err
	}
	return start(cfg)
}

func startManager(cfg managerConfig) error {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&cfg.zapOpts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: cfg.metricsAddr},
		HealthProbeBindAddress: cfg.probeAddr,
		LeaderElection:         cfg.enableLeaderElection,
		LeaderElectionID:       "modelfs.samzong.dev",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	if err = (&controllers.ModelReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Model")
		return err
	}

	if err = (&controllers.ModelSourceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ModelSource")
		return err
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		return err
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}
	return nil
}

func main() {
	if err := runManager(os.Args[1:], os.Stdout, flag.CommandLine, nil, startManager); err != nil {
		os.Exit(1)
	}
}
