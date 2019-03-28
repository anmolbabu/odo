package labels

import (
	applabels "github.com/openshift/odo/pkg/application/labels"
)

// ComponentLabel is a label key used to identify the component name
const ComponentLabel = "app.kubernetes.io/component-name"

// ComponentTypeLabel is Kubernetes label that identifies the type of a component being used
const ComponentTypeLabel = "app.kubernetes.io/component-type"

// ComponentTypeVersion is a Kubernetes label that identifies the component version
const ComponentTypeVersion = "app.kubernetes.io/component-version"

// ComponentSettingsHash is a label that indicates the version of component settings applied on component
// This field will be used to detect if config changed
const ComponentSettingsHash = "app.kubernetes.io/componentSettingsHash"

// GetLabels return labels that should be applied to every object for given component in active application
// additional labels are used only for creating object
// if you are creating something use additional=true
// if you need labels to filter component that use additional=false
func GetLabels(componentName string, applicationName string, additional bool) map[string]string {
	labels := applabels.GetLabels(applicationName, additional)
	labels[ComponentLabel] = componentName
	return labels
}
