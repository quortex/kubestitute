/*


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

	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
	"quortex.io/kubestitute/controllers"
	_ "quortex.io/kubestitute/metrics"
	"quortex.io/kubestitute/utils/supervisor"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = corev1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var clusterAutoscalerStatusNamespace string
	var clusterAutoscalerStatusName string
	var enableDevLogs bool
	var logVerbosity int
	var asgPollInterval int
	var evictionGlobalTimeout int

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&clusterAutoscalerStatusNamespace, "clusterautoscaler-status-namespace", "kube-system", "The namespace the clusterautoscaler status configmap belongs to.")
	flag.StringVar(&clusterAutoscalerStatusName, "clusterautoscaler-status-name", "cluster-autoscaler-status", "The name of the clusterautoscaler status configmap.")
	flag.BoolVar(&enableDevLogs, "dev", false, "Enable dev mode for logging.")
	flag.IntVar(&logVerbosity, "v", 3, "Logs verbosity. 0 => panic, 1 => error, 2 => warning, 3 => info, 4 => debug")
	flag.IntVar(&asgPollInterval, "asg-poll-interval", 30, "AutoScaling Groups polling interval (used to generate custom metrics about ASGs).")
	flag.IntVar(&evictionGlobalTimeout, "eviction-timeout", 300, "The timeout in seconds for pods eviction on Instance deletion.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(enableDevLogs), zap.Level(zapcore.Level(int8(zapcore.DPanicLevel)-int8(logVerbosity)))))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "9cbc928d.kubestitute.quortex.io",
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
		Log:    ctrl.Log.WithName("controllers").WithName("Instance"),
		Scheme: mgr.GetScheme(),
		Configuration: controllers.InstanceReconcilerConfiguration{
			EvictionGlobalTimeout: evictionGlobalTimeout,
		},
		Kubernetes: kubeClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Instance")
		os.Exit(1)
	}
	if err = (&controllers.SchedulerReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Scheduler"),
		Scheme: mgr.GetScheme(),
		Conf: controllers.SchedulerReconcilerConfiguration{
			ClusterAutoscalerStatusNamespace: clusterAutoscalerStatusNamespace,
			ClusterAutoscalerStatusName:      clusterAutoscalerStatusName,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Scheduler")
		os.Exit(1)
	}
	if err = (&controllers.SupervisionReconciler{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("Supervision"),
		Scheme:     mgr.GetScheme(),
		Supervisor: supervisor.New(time.Second*time.Duration(asgPollInterval), ctrl.Log.WithName("supervisor")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Supervision")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
