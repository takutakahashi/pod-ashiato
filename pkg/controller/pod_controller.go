package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodController はPodの情報を収集・記録するコントローラーです
type PodController struct {
	clientset *kubernetes.Clientset
	logger    *zap.Logger
	interval  time.Duration
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
func NewPodController(clientset *kubernetes.Clientset, logger *zap.Logger, interval time.Duration) *PodController {
	return &PodController{
		clientset: clientset,
		logger:    logger,
		interval:  interval,
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

// logPodNodeInfo はすべてのPodのノード情報をログに出力します
func (c *PodController) logPodNodeInfo(ctx context.Context) error {
	pods, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range pods.Items {
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