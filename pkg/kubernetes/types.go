package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
)

type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type WebHook struct {
	Config     *WebHookConfig
	KiamConfig *KiamInject
}

type WebHookConfig struct {
	Template   string `json:"template"`
	KiamConfig string `json:"kiam-config"`
}

type KiamData struct {
	Name          string
	Container     corev1.Container
}

type KiamInject struct {
	Annotations map[string]string
}

type registeredAnnotation struct {
	name      string
	validator annotationValidationFunc
}

type annotationValidationFunc func(value string) error
