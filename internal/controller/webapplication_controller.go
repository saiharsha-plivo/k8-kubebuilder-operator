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

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	servicev1 "github.com/saiharsha-plivo/k8-kubebuilder-operator/api/v1"
)

// WebApplicationReconciler reconciles a WebApplication object
type WebApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=service.my.domain,resources=webapplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=service.my.domain,resources=webapplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=service.my.domain,resources=webapplications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WebApplication object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *WebApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// log := log.FromContext(ctx)
	log := ctrl.Log.WithName("controller").WithName("WebApplication")
	log.Info("In the controller code")
	log.Info("Reconciling WebApplication", "namespace", req.Namespace, "name", req.Name)

	webApp := &servicev1.WebApplication{} // Assuming your API group is "servicev1"
	err := r.Get(ctx, req.NamespacedName, webApp)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("WebApplication resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get WebApplication")
		return ctrl.Result{}, err
	}

	// Log WebApplication spec details
	log.Info("WebApplication Details",
		"replicas", webApp.Spec.Replica,
		"serviceType", webApp.Spec.ServiceType,
		"image", webApp.Spec.Image,
		"port", webApp.Spec.Port,
		"containerPort", webApp.Spec.ContainerPort,
		"nodePort", webApp.Spec.NodePort,
	)

	// Handle Deployment
	err = r.reconcileDeployment(ctx, webApp)
	if err != nil {
		log.Error(err, "Failed to reconcile Deployment")
		return ctrl.Result{}, err
	}

	err = r.reconcileService(ctx, webApp)
	if err != nil {
		log.Error(err, "Failed to reconcile service")
		return ctrl.Result{}, err
	}

	err = r.updateServiceEndpoint(ctx, webApp)
	if err != nil {
		log.Error(err, "Failed to reconcile service endpoint")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *WebApplicationReconciler) reconcileService(ctx context.Context, webApp *servicev1.WebApplication) error {
	log := ctrl.Log.WithName("controller").WithName("WebApplication").WithName("Service handler")

	log.Info("Reconciling the service for WebApplication",
		"name", webApp.ObjectMeta.Name,
		"namespace", webApp.ObjectMeta.Namespace,
	)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webApp.Name,
			Namespace: webApp.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": webApp.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       webApp.Spec.Port,
					TargetPort: intstr.FromInt(int(webApp.Spec.ContainerPort)),
				},
			},
		},
	}

	// Set ServiceType
	if webApp.Spec.ServiceType == "NodePort" {
		service.Spec.Type = corev1.ServiceTypeNodePort
		service.Spec.Ports[0].NodePort = int32(webApp.Spec.NodePort) // Only set NodePort if type is NodePort
	} else {
		service.Spec.Type = corev1.ServiceTypeClusterIP
	}

	// Set WebApplication as the owner of the service
	if err := controllerutil.SetControllerReference(webApp, service, r.Scheme); err != nil {
		return err
	}

	found := &corev1.Service{}

	err := r.Get(ctx, client.ObjectKey{Name: webApp.Name, Namespace: webApp.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating new Service", "Name", service.Name)
			return r.Create(ctx, service)
		}
		return err
	}

	log.Info("Updating existing Service", "Name", service.Name)
	found.Spec = service.Spec
	return r.Update(ctx, found)
}

func (r *WebApplicationReconciler) reconcileDeployment(ctx context.Context, webApp *servicev1.WebApplication) error {
	log := ctrl.Log.WithName("controller").WithName("WebApplication").WithName("deployment handler")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webApp.Name,
			Namespace: webApp.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &webApp.Spec.Replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": webApp.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": webApp.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "webapp-container",
							Image: webApp.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: webApp.Spec.ContainerPort,
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(webApp, deployment, r.Scheme); err != nil {
		return err
	}

	// Check if Deployment exists
	found := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: webApp.Name, Namespace: webApp.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		return r.Create(ctx, deployment)
	} else if err != nil {
		return err
	}

	// Update existing Deployment if needed
	found.Spec = deployment.Spec
	return r.Update(ctx, found)
}

func (r *WebApplicationReconciler) updateServiceEndpoint(ctx context.Context, webApp *servicev1.WebApplication) error {
	log := ctrl.Log.WithName("controller").WithName("WebApplication").WithName("service endpoint handler")

	service := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: webApp.Name, Namespace: webApp.Namespace}, service)
	if err != nil {
		log.Error(err, "Failed to get Service for updating endpoint")
		return err
	}

	var serviceEndpoint string

	switch service.Spec.Type {
	case corev1.ServiceTypeClusterIP:
		serviceEndpoint = fmt.Sprintf("%s:%d", service.Spec.ClusterIP, service.Spec.Ports[0].Port)

	case corev1.ServiceTypeNodePort:
		// Get any Node IP (assuming a single-node setup for local development)
		nodes := &corev1.NodeList{}
		if err := r.List(ctx, nodes); err != nil {
			log.Error(err, "Failed to list Nodes for NodePort service")
			return err
		}
		if len(nodes.Items) > 0 {
			nodeIP := nodes.Items[0].Status.Addresses[0].Address // Picking first node IP
			nodePort := service.Spec.Ports[0].NodePort
			serviceEndpoint = fmt.Sprintf("%s:%d", nodeIP, nodePort)
		}

	case corev1.ServiceTypeLoadBalancer:
		if len(service.Status.LoadBalancer.Ingress) > 0 {
			serviceEndpoint = service.Status.LoadBalancer.Ingress[0].IP
			if serviceEndpoint == "" {
				serviceEndpoint = service.Status.LoadBalancer.Ingress[0].Hostname
			}
		}
	}

	if webApp.Status.ServiceEndpoint != serviceEndpoint {
		webApp.Status.ServiceEndpoint = serviceEndpoint
		if err := r.Status().Update(ctx, webApp); err != nil {
			log.Error(err, "Failed to update WebApplication status with service endpoint")
			return err
		}
		log.Info("Updated WebApplication status with Service Endpoint", "endpoint", serviceEndpoint)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&servicev1.WebApplication{}).
		Named("webapplication").
		Complete(r)
}
