/*
Copyright 2024.

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
	"errors"
	"fmt"
	"strconv"

	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	batchv1 "github.com/zagganas/hello_controller/api/v1"
)

// HelloJobReconciler reconciles a HelloJob object
type HelloJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func checkEmpty(s string) error {
	if len(s) == 0 {
		return errors.New("empty field")
	}
	return nil
}

//+kubebuilder:rbac:groups=batch.kostis.test.eu,resources=hellojobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.kostis.test.eu,resources=hellojobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.kostis.test.eu,resources=hellojobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HelloJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *HelloJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var helloJob batchv1.HelloJob
	if err := r.Get(ctx, req.NamespacedName, &helloJob); err != nil {
		log.Error(err, "hello job was not found")
	}

	if err := checkEmpty(helloJob.Spec.Message); err != nil {
		log.Error(err, "message cannot be empty")
	}

	if err := checkEmpty(helloJob.Spec.Image); err != nil {
		log.Error(err, "image cannot be empty")
	}
	childJobName := fmt.Sprintf("%s-job", helloJob.Name)

	var childJobs kbatch.JobList
	if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err == nil {
		// If job already exists, don't do anything else
		// and end reconciliation here
		return ctrl.Result{}, nil
	}

	containerCommand := []string{"echo", helloJob.Spec.Message}

	containerSpec := corev1.PodSpec{}

	if *helloJob.Spec.DelaySeconds != 0 {
		containerSpec.InitContainers = []corev1.Container{
			{
				Name:    "delay-container",
				Image:   "alpine:latest",
				Command: []string{"sleep", strconv.Itoa(*helloJob.Spec.DelaySeconds)},
			},
		}
	}

	containerSpec.Containers = []corev1.Container{
		{
			Name:    "echo-message",
			Image:   helloJob.Spec.Image,
			Command: containerCommand,
		},
	}

	createHelloJob := func(helloJob *batchv1.HelloJob) (*kbatch.Job, error) {
		name := childJobName
		labels := map[string]string{"k8s-app": name, "job-type": "hello-job"}
		job := &kbatch.Job{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      labels,
				Annotations: labels,
				Name:        name,
				Namespace:   helloJob.Namespace,
			},
			Spec: kbatch.JobSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "hello-job-" + name + "-",
					},
					Spec: containerSpec,
				},
			},
		}
		return job, nil
	}

	job, err := createHelloJob(&helloJob)
	if err != nil {
		log.Error(err, "unable to construct job from template")
		return ctrl.Result{}, nil
	}
	if err := r.Create(ctx, job); err != nil {
		log.Error(err, "unable to create Job for HelloJob", job)
		return ctrl.Result{}, err
	}

	log.V(1).Info("created Job for HelloJob", helloJob.Name, "job", job)

	return ctrl.Result{}, nil
}

var (
	jobOwnerKey = ".metadata.controller"
	apiGVStr    = batchv1.GroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.
func (r *HelloJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kbatch.Job{}, jobOwnerKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*kbatch.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...make sure it's a CronJob...
		if owner.APIVersion != apiGVStr || owner.Kind != "HelloJob" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.HelloJob{}).
		Owns(&kbatch.Job{}).
		Complete(r)
}
