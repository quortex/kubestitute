/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
	"quortex.io/kubestitute/controllers"
	"quortex.io/kubestitute/utils/supervisor"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(corev1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var enableDevLogs bool
	var logVerbosity int

	var clusterAutoscalerNamespace string
	var clusterAutoscalerStatusName string
	var clusterAutoscalerStatusLegacyFormat bool
	var clusterAutoscalerPEConfigMapName string
	var priorityExpanderNamespace string
	var priorityExpanderName string
	var priorityExpanderEnabled bool

	var asgPollInterval int
	var evictionGlobalTimeout int
	var instancesMaxConcurrentReconciles int

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableDevLogs, "dev", false, "Enable dev mode for logging.")
	flag.IntVar(&logVerbosity, "v", 3, "Logs verbosity. 0 => panic, 1 => error, 2 => warning, 3 => info, 4 => debug")

	flag.StringVar(&clusterAutoscalerNamespace, "clusterautoscaler-namespace", "kube-system", "The namespace the clusterautoscaler belongs to.")
	flag.StringVar(&clusterAutoscalerStatusName, "clusterautoscaler-status-name", "cluster-autoscaler-status", "The name of the clusterautoscaler status configmap.")
	flag.BoolVar(&clusterAutoscalerStatusLegacyFormat, "clusterautoscaler-status-legacy-format", false, "Set if the clusterautoscaler status configmap is formatted the legacy readable format, used by cluster-autoscaler up to version 1.29.")
	flag.StringVar(&clusterAutoscalerPEConfigMapName, "cluster-autoscaler-priority-expander-config-map", "cluster-autoscaler-priority-expander", "The name of the clusterautoscaler priority expander config map.")
	flag.StringVar(&priorityExpanderNamespace, "priority-expander-namespace", "kubestitute-system", "The namespace the _unique_ priority expander object belongs to.")
	flag.StringVar(&priorityExpanderName, "priority-expander-name", "priority-expander-default", "The only accepted name for the priority expander object.")
	flag.BoolVar(&priorityExpanderEnabled, "priority-expander-enabled", false, "Is the controller enabled.")

	flag.IntVar(&asgPollInterval, "asg-poll-interval", 30, "AutoScaling Groups polling interval (used to generate custom metrics about ASGs).")
	flag.IntVar(&evictionGlobalTimeout, "eviction-timeout", 300, "The timeout in seconds for pods eviction on Instance deletion.")
	flag.IntVar(&instancesMaxConcurrentReconciles, "instances-max-concurrent-reconciles", 10, "The maximum number of concurrent Reconciles which can be run for Instances.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(enableDevLogs), zap.Level(zapcore.Level(int8(zapcore.DPanicLevel)-int8(logVerbosity)))))

	// AWS clients initialization with a default session.
	// Session is created from SDK defaults, config files, environment, and user
	// provided config files.
	setupLog.Info("initializing aws session")
	session, err := session.NewSession()
	if err != nil {
		setupLog.Error(err, "unable to init aws session")
		os.Exit(1)
	}
	autoscaling := autoscaling.New(session)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		WebhookServer:          webhook.NewServer(webhook.Options{}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "9cbc928d.kubestitute.quortex.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		setupLog.Error(err, "unable to instantiate kubernetes client")
		os.Exit(1)
	}

	if err = (&controllers.InstanceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Configuration: controllers.InstanceReconcilerConfiguration{
			EvictionGlobalTimeout:   evictionGlobalTimeout,
			MaxConcurrentReconciles: instancesMaxConcurrentReconciles,
		},
		Kubernetes:  kubeClient,
		Autoscaling: autoscaling,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Instance")
		os.Exit(1)
	}
	if err = (&controllers.SchedulerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Configuration: controllers.SchedulerReconcilerConfiguration{
			ClusterAutoscalerNamespace:          clusterAutoscalerNamespace,
			ClusterAutoscalerStatusName:         clusterAutoscalerStatusName,
			ClusterAutoscalerStatusLegacyFormat: clusterAutoscalerStatusLegacyFormat,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Scheduler")
		os.Exit(1)
	}
	if err = (&controllers.SupervisionReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		Supervisor: supervisor.New(autoscaling, time.Second*time.Duration(asgPollInterval), ctrl.Log.WithName("supervisor")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Supervision")
		os.Exit(1)
	}
	if priorityExpanderEnabled {
		if err = (&controllers.PriorityExpanderReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
			Configuration: controllers.PriorityExpanderReconcilerConfiguration{
				ClusterAutoscalerNamespace:          clusterAutoscalerNamespace,
				ClusterAutoscalerStatusName:         clusterAutoscalerStatusName,
				ClusterAutoscalerStatusLegacyFormat: clusterAutoscalerStatusLegacyFormat,
				ClusterAutoscalerPEConfigMapName:    clusterAutoscalerPEConfigMapName,
				PriorityExpanderNamespace:           priorityExpanderNamespace,
				PriorityExpanderName:                priorityExpanderName,
			},
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "PriorityExpander")
			os.Exit(1)
		}
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
