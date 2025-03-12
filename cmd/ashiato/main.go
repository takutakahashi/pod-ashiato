package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/takutakahashi/pod-ashiato/pkg/controller"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfig    string
	interval      time.Duration
	oneshot       bool
	namespace     string
	podName       string
	labelSelector string
)

func init() {
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")
	}
	flag.DurationVar(&interval, "interval", 30*time.Second, "interval between pod checks")
	flag.BoolVar(&oneshot, "oneshot", false, "run only once and exit")
	flag.StringVar(&namespace, "namespace", "", "filter pods by namespace (default: all namespaces)")
	flag.StringVar(&podName, "name", "", "filter pods by name prefix")
	flag.StringVar(&labelSelector, "label", "", "filter pods by label selector (e.g. 'app=nginx,env=prod')")
}

func main() {
	flag.Parse()

	// ロガーの初期化
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Kubernetesクライアントの設定
	var config *rest.Config
	var err error

	// クラスター内で動作している場合は、サービスアカウントの認証情報を使用
	if _, err = os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			logger.Fatal("Error building in-cluster config", zap.Error(err))
		}
	} else {
		// クラスター外で動作している場合は、kubeconfigを使用
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logger.Fatal("Error building kubeconfig", zap.Error(err))
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Fatal("Error creating kubernetes clientset", zap.Error(err))
	}

	// コントローラーの初期化
	podController := controller.NewPodController(clientset, logger, interval, namespace, podName, labelSelector)

	// ログフィルター情報
	if namespace != "" || podName != "" || labelSelector != "" {
		logger.Info("Using pod filters", 
			zap.String("namespace", namespace),
			zap.String("podNamePrefix", podName),
			zap.String("labelSelector", labelSelector))
	}

	// シグナルハンドリングの設定
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalCh
		logger.Info("Received termination signal, shutting down...")
		cancel()
	}()

	// コントローラーの実行
	if oneshot {
		if err := podController.RunOnce(ctx); err != nil {
			logger.Fatal("Failed to run pod controller", zap.Error(err))
		}
	} else {
		logger.Info("Starting pod-ashiato controller", zap.Duration("interval", interval))
		if err := podController.Run(ctx); err != nil {
			logger.Fatal("Failed to run pod controller", zap.Error(err))
		}
		logger.Info("Controller shutdown complete")
	}
}