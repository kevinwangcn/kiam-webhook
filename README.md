# KIAM Mutating Admission Webhook

## Prerequisites

kiam server and agent running on the node.

More information: https://github.com/uswitch/kiam

## Deploy KIAM WebHook

1. Create the kiam webhook base configuration

    ```
    oc project kiam
    
    oc apply -f contrib/openshift/kiam-config.yaml
    ```

2. Process Mutating WebHook Template.
   
   The template is going to create the following resources:
    * kiam-webhook-psp PodSecurityPolicy
    * kiam-webhook-clusterrole ClusterRole
    * kiam-webhook ServiceAccount
    * kiam-webhook-rolebinding ClusterRoleBinding
    * kiam-webhook Service
    * kiam-webhook Deployment
    * kiam-webhook MutatingWebhookConfiguration
    
   2.1 Retrieve service-ca.crt from one pod

    ```
    pod=$(oc get pods -lapp=kiam --no-headers -o custom-columns=NAME:.metadata.name)
    export CA_BUNDLE=$(oc exec $pod -- cat /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt | base64 | tr -d '\n')
    ```

   2.2 Process the webhook-template

    ```
    oc process -f contrib/openshift/webhook-template.yaml -p CA_BUNDLE=${CA_BUNDLE} | oc apply -f -
    ```

    |     PARAMETER   |  DEFAULT           |  DESCRIPTION                                                              |
    |-----------------|--------------------|---------------------------------------------------------------------------|
    | CA_BUNDLE       |                    |    CA used by kubernetes to trust the webhook                             |
    | KIAM_NAMESPACE  |    kiam            |    KIAM  Namespace                                                        |
    | GIN_MODE        |    release         |    Http server startup mode [gin-gonic](https://github.com/gin-gonic/gin) |
    | LOG_LEVEL       |    INFO            |    Log level from [logrus](https://github.com/sirupsen/logrus)            |

## Verify Config Injection

1. Label the target project where you want the webhook to inject the vault agent sidecar container.

    ```
    oc label namespace app kiam-webhook=enabled
    ```

2. **Not yet implemented** | Add the *kiam.amazonaws.com/inject* annotation with value true to the pod template spec to enable injection.


    ```
    oc patch dc/thorntail-example -p '{
                                     "spec": {
                                       "template": {
                                         "metadata": {
                                           "annotations": {
                                             "kiam.amazonaws.com/inject": "true"
                                           }
                                         }
                                       }
                                     }
                                   }'
    ```
3. The kiam agent webhook will:
    * Check if kiam-webhook label is present . 
    * Verify base kiam-webhook-config. 
    * Verify kiam-webhook-config on the target project.
    * Inject kiam configuration on the target pod.

# References

* https://github.com/openlab-red/mutating-webhook-vault-agent
