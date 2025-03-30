/*
Copyright 2025.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WebApplicationSpec defines the desired state of WebApplication.
type WebApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Foo is an example field of WebApplication. Edit webapplication_types.go to remove/update
	Replica       int32  `json:"replica"`
	ServiceType   string `json:"servicetype"` // Can be set to clusterip / nodeport
	Image         string `json:"image"`
	Port          int32  `json:"port"`
	ContainerPort int32  `json:"containerport"`
	NodePort      int32  `json:"nodeport,omitempty"`
}

// WebApplicationStatus defines the observed state of WebApplication.
type WebApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ServiceEndpoint string `json:"serviceendpoint,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// WebApplication is the Schema for the webapplications API.
type WebApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebApplicationSpec   `json:"spec,omitempty"`
	Status WebApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WebApplicationList contains a list of WebApplication.
type WebApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WebApplication{}, &WebApplicationList{})
}
