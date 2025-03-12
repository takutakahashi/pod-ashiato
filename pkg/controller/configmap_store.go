package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ConfigMapStore は Pod の情報を ConfigMap に保存するための機能を提供します
type ConfigMapStore struct {
	clientset     *kubernetes.Clientset
	logger        *zap.Logger
	namespace     string
	configMapName string
	data          map[string]string
	lastUpdateCM  time.Time
}

// NewConfigMapStore は新しい ConfigMapStore を作成します
func NewConfigMapStore(clientset *kubernetes.Clientset, logger *zap.Logger, namespace string) *ConfigMapStore {
	return &ConfigMapStore{
		clientset: clientset,
		logger:    logger,
		namespace: namespace,
		data:      make(map[string]string),
	}
}

// updateConfigMapName は現在の時刻に基づいて ConfigMap の名前を更新します
// 名前は "pod-ashiato-YYYYMMDDHH" の形式です
func (s *ConfigMapStore) updateConfigMapName() {
	now := time.Now()
	s.configMapName = fmt.Sprintf("pod-ashiato-%s", now.Format("2006010215"))
}

// Store は与えられた Pod 情報を ConfigMap に保存します
func (s *ConfigMapStore) Store(ctx context.Context, podName, nodeName string) error {
	// 新しい時間帯に入ったかチェック (1時間ごと)
	now := time.Now()
	currentHour := now.Format("2006010215")
	lastHour := s.lastUpdateCM.Format("2006010215")
	
	if s.configMapName == "" || currentHour != lastHour {
		// 新しい ConfigMap の名前を設定
		s.updateConfigMapName()
		// データをリセット
		s.data = make(map[string]string)
		s.lastUpdateCM = now
		
		s.logger.Info("Creating new ConfigMap for the current hour", 
			zap.String("configMapName", s.configMapName),
			zap.String("hour", currentHour))
	}

	// ConfigMapのキーに使用できるように形式変換（/は使用できないため_に置換）
	cmKey := strings.ReplaceAll(podName, "/", "_")

	// データを内部マップに追加
	s.data[cmKey] = nodeName
	
	// ConfigMap を作成または更新
	return s.saveToConfigMap(ctx)
}

// saveToConfigMap は現在のデータを ConfigMap に保存します
func (s *ConfigMapStore) saveToConfigMap(ctx context.Context) error {
	// ConfigMap が存在するかチェック
	_, err := s.clientset.CoreV1().ConfigMaps(s.namespace).Get(ctx, s.configMapName, metav1.GetOptions{})
	
	if err != nil {
		if errors.IsNotFound(err) {
			// ConfigMap が存在しない場合は新規作成
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      s.configMapName,
					Namespace: s.namespace,
					Labels: map[string]string{
						"app": "pod-ashiato",
						"type": "pod-node-mapping",
					},
				},
				Data: s.data,
			}
			
			_, err = s.clientset.CoreV1().ConfigMaps(s.namespace).Create(ctx, cm, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create ConfigMap: %w", err)
			}
			
			s.logger.Info("Created new ConfigMap", zap.String("name", s.configMapName))
			return nil
		}
		
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}
	
	// ConfigMap が存在する場合は更新
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.configMapName,
			Namespace: s.namespace,
			Labels: map[string]string{
				"app": "pod-ashiato",
				"type": "pod-node-mapping",
			},
		},
		Data: s.data,
	}
	
	_, err = s.clientset.CoreV1().ConfigMaps(s.namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap: %w", err)
	}
	
	return nil
}