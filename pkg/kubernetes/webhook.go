package kubernetes

import (
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/kubernetes/pkg/apis/core/v1"
	"net/http"
	"github.com/sirupsen/logrus"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	alwaysValidFunc = func(value string) error {
		return nil
	}

	annotationRegistry = []*registeredAnnotation{
		{"kiam.amazonaws.com/inject", alwaysValidFunc},
		{"kiam.amazonaws.com/status", alwaysValidFunc},
	}

	annotationPolicy = annotationRegistry[0]
	annotationStatus        = annotationRegistry[1]

	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
)

func (wk *WebHook) Mutate(context *gin.Context) {

	ar := v1beta1.AdmissionReview{}

	if err := context.ShouldBindJSON(&ar); err == nil {
		admissionResponse := wk.admit(ar)
		admissionReview := v1beta1.AdmissionReview{}
		if admissionResponse != nil {
			admissionReview.Response = admissionResponse
			if ar.Request != nil {
				admissionReview.Response.UID = ar.Request.UID
			}
		}
		context.JSON(http.StatusOK, admissionReview)
	} else {
		context.AbortWithStatusJSON(http.StatusBadRequest, ToAdmissionResponse(err))
	}

}

func (wk *WebHook) admit(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	pod := corev1.Pod{}
	var err error
	var name string

	if err = Pod(req.Object.Raw, &pod); err != nil {
		return ToAdmissionResponse(err)
	}

	pod.Name = PotentialPodName(&pod.ObjectMeta)
	pod.Namespace = PotentialNamespace(req, &pod)

	log.WithFields(logrus.Fields{
		"Kind":           req.Kind,
		"Namespace":      req.Namespace,
		"Name":           pod.Name,
		"UID":            req.UID,
		"PatchOperation": req.Operation,
		"UserInfo":       req.UserInfo,
	}).Infoln("AdmissionReview for")

	if !injectRequired(ignoredNamespaces, &pod) {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	name, err = GetDeploymentName(pod.OwnerReferences[0].Name)
	if err != nil {
		return ToAdmissionResponse(err)
	}

	data := KiamData{
		Name:          name,
		Container:     pod.Spec.Containers[0],
	}

	// TODO verify namespace config map to inject
	//configMap, err = retrieveConfigMap(pod, wk, &data)
	//if err != nil {
	//	return ToAdmissionResponse(err)
	//}

	wk.KiamConfig, err = injectData(&data, wk.Config)
	if err != nil {
		return ToAdmissionResponse(err)
	}
	annotations := map[string]string{annotationStatus.name: "injected"}

	//patch
	patches, err := createPatch(&pod, wk.KiamConfig, annotations)
	if err != nil {
		return ToAdmissionResponse(err)
	}

	log.Debugf("AdmissionResponse: patch=%v\n", string(patches))

	log.WithFields(logrus.Fields{
		"Kind":           req.Kind,
		"Namespace":      req.Namespace,
		"Name":           pod.Name,
		"UID":            req.UID,
		"PatchOperation": req.Operation,
		"UserInfo":       req.UserInfo,
	}).Infoln("AdmissionResponse Allowed for")

	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patches,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	_ = v1.AddToScheme(runtimeScheme)
}
