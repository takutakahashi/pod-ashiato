package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodController はPodの情報を収集・記録するコントローラーです
type PodController struct {
	clientset  *kubernetes.Clientset
	logger     *zap.Logger
	interval   time.Duration
	namespace  string
	podName    string
	labelSelector string
}

// PodInfo はPodのノード情報を含む構造体です
type PodInfo struct {
	Namespace  string    `json:"namespace"`
	PodName    string    `json:"pod_name"`
	NodeName   string    `json:"node_name"`
	PodIP      string    `json:"pod_ip"`
	Phase      string    `json:"phase"`
	Timestamp  time.Time `json:"timestamp"`
	Conditions []PodCondition `json:"conditions,omitempty"`
}

// PodCondition はPodのコンディション情報を含む構造体です
type PodCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"last_transition_time"`
}

// NewPodController は新しいPodControllerを作成します
func NewPodController(clientset *kubernetes.Clientset, logger *zap.Logger, interval time.Duration, namespace string, podName string, labelSelector string) *PodController {
	return &PodController{
		clientset:     clientset,
		logger:        logger,
		interval:      interval,
		namespace:     namespace,
		podName:       podName,
		labelSelector: labelSelector,
	}
}

// Run はコントローラーの実行を開始します
func (c *PodController) Run(ctx context.Context) error {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := c.logPodNodeInfo(ctx); err != nil {
				c.logger.Error("Failed to log pod node info", zap.Error(err))
			}
		}
	}
}

// RunOnce は1回だけPod情報を出力します
func (c *PodController) RunOnce(ctx context.Context) error {
	return c.logPodNodeInfo(ctx)
}

// logPodNodeInfo はフィルター条件に一致するPodのノード情報をログに出力します
func (c *PodController) logPodNodeInfo(ctx context.Context) error {
	// ListOptionsの構築
	listOptions := metav1.ListOptions{
		LabelSelector: c.labelSelector,
	}

	// Podのリストを取得
	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, listOptions)
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	filteredPods := []corev1.Pod{}
	
	// フィルタリング処理（podNameはプレフィックスマッチ）
	if c.podName != "" {
		for _, pod := range pods.Items {
			if strings.HasPrefix(pod.Name, c.podName) {
				filteredPods = append(filteredPods, pod)
			}
		}
	} else {
		filteredPods = pods.Items
	}

	if len(filteredPods) == 0 {
		c.logger.Info("No pods found matching the specified filters",
			zap.String("namespace", c.namespace),
			zap.String("podNamePrefix", c.podName),
			zap.String("labelSelector", c.labelSelector))
		return nil
	}

	for _, pod := range filteredPods {
		info := c.createPodInfo(&pod)
		
		// 構造化ログ出力
		jsonData, err := json.Marshal(info)
		if err != nil {
			c.logger.Error("Failed to marshal pod info", zap.Error(err))
			continue
		}
		
		fmt.Println(string(jsonData))
	}

	return nil
}

// createPodInfo はPodからPodInfo構造体を作成します
func (c *PodController) createPodInfo(pod *corev1.Pod) PodInfo {
	conditions := make([]PodCondition, 0, len(pod.Status.Conditions))
	for _, cond := range pod.Status.Conditions {
		conditions = append(conditions, PodCondition{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			LastTransitionTime: cond.LastTransitionTime.Time,
		})
	}
	
	return PodInfo{
		Namespace:  pod.Namespace,
		PodName:    pod.Name,
		NodeName:   pod.Spec.NodeName,
		PodIP:      pod.Status.PodIP,
		Phase:      string(pod.Status.Phase),
		Timestamp:  time.Now(),
		Conditions: conditions,
	}
}