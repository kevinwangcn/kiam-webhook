package kubernetes

import (
	"strings"
	"bytes"
	"github.com/sirupsen/logrus"
	"text/template"
	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"encoding/json"
)

const (
	KiamConfig = "kiam-config"
)

func injectData(data *KiamData, config *WebHookConfig) (*KiamInject, error) {

	sic := KiamInject{}

	tmpl, err := executeTemplate(config.Template, data)
	if err != nil {
		return nil, err
	}

	err = unmarshalTemplate(tmpl, &sic)
	if err != nil {
		return nil, err
	}

	log.Debugln("KiamInject: ", sic)
	return &sic, nil
}

func injectRequired(ignored []string, pod *corev1.Pod) bool {
	var status, inject string
	required := false
	metadata := pod.ObjectMeta

	// skip special kubernetes system namespaces
	for _, namespace := range ignored {
		if metadata.Namespace == namespace {
			return false
		}
	}

	annotations := metadata.GetAnnotations()
	log.Debugf("Annotations: %v", annotations)

	if annotations != nil {
		status = annotations[annotationStatus.name]

		log.Debugln(status)
		if strings.ToLower(status) == "injected" {
			required = false
		} else {
			inject = annotations[annotationPolicy.name]
			log.Debugln(inject)
			// per namespace
			required = true
			//switch strings.ToLower(inject) {
			//default:
			//	required = false
			//case "y", "yes", "true", "on":
			//	required = true
			//}
		}
	}

	log.WithFields(logrus.Fields{
		"name":      metadata.Name,
		"namespace": metadata.Namespace,
		"status":    status,
		"inject":    inject,
		"required":  required,
	}).Infoln("Mutation policy")

	return required
}

func retrieveConfigMap(pod corev1.Pod, wk *WebHook, kiamData *KiamData) (*corev1.ConfigMap, error) {
	client := Client()
	configMaps := client.CoreV1().ConfigMaps(pod.Namespace)
	currentConfigMap, err := configMaps.Get(KiamConfig+"-"+kiamData.Name, metav1.GetOptions{})
	return currentConfigMap, err

}

func executeTemplate(source string, data interface{}) (*bytes.Buffer, error) {
	var tmpl bytes.Buffer

	funcMap := template.FuncMap{
		"valueOrDefault": valueOrDefault,
		"toJSON":         toJSON,
	}

	temp := template.New("inject")
	t := template.Must(temp.Funcs(funcMap).Parse(source))

	if err := t.Execute(&tmpl, &data); err != nil {
		log.Errorf("Failed to execute template %v %s", err, source)
		return nil, err
	}

	return &tmpl, nil
}

func unmarshalTemplate(tmpl *bytes.Buffer, target interface{}) (error) {
	log.Debugf("Template executed, %s", string(tmpl.Bytes()))

	if err := yaml.Unmarshal(tmpl.Bytes(), &target); err != nil {
		log.Errorf("Failed to unmarshal template %v %s", err, string(tmpl.Bytes()))
		return err
	}

	return nil
}

func valueOrDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func toJSON(m map[string]string) string {
	if m == nil {
		return "{}"
	}

	ba, err := json.Marshal(m)
	if err != nil {
		log.Warnf("Unable to marshal %v", m)
		return "{}"
	}

	return string(ba)
}
