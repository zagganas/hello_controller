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

package v1

import (
	dockerparser "github.com/novln/docker-parser"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var hellojoblog = logf.Log.WithName("hellojob-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *HelloJob) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-batch-kostis-test-eu-v1-hellojob,mutating=true,failurePolicy=fail,sideEffects=None,groups=batch.kostis.test.eu,resources=hellojobs,verbs=create;update,versions=v1,name=mhellojob.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &HelloJob{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *HelloJob) Default() {
	hellojoblog.Info("default", "name", r.Name)

	if r.Spec.DelaySeconds == nil {
		r.Spec.DelaySeconds = new(int)
		*r.Spec.DelaySeconds = 0
	}

	if r.Spec.Image == "" {
		r.Spec.Image = "alpine:latest"
	}

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-batch-kostis-test-eu-v1-hellojob,mutating=false,failurePolicy=fail,sideEffects=None,groups=batch.kostis.test.eu,resources=hellojobs,verbs=create;update,versions=v1,name=vhellojob.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &HelloJob{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HelloJob) ValidateCreate() (admission.Warnings, error) {
	hellojoblog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, r.validateHelloJob()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HelloJob) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	hellojoblog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, r.validateHelloJob()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HelloJob) ValidateDelete() (admission.Warnings, error) {
	hellojoblog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *HelloJob) validateHelloJob() error {
	var allErrs field.ErrorList

	if err := r.validateHelloJobName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateHelloJobSpec(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "batch.kostis.test.eu", Kind: "HelloJob"}, r.Name, allErrs)
}

func (r *HelloJob) validateHelloJobSpec() *field.Error {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.
	return validateImageFormat(
		r.Spec.Image,
		field.NewPath("spec").Child("image"))
}

func validateImageFormat(image string, fldPath *field.Path) *field.Error {
	if _, err := dockerparser.Parse(image); err != nil {
		return field.Invalid(fldPath, image, err.Error())
	}
	return nil
}

func (r *HelloJob) validateHelloJobName() *field.Error {
	if len(r.ObjectMeta.Name) > validation.DNS1035LabelMaxLength-11 {
		// The job name length is 63 character like all Kubernetes objects
		// (which must fit in a DNS subdomain). The cronjob controller appends
		// a 11-character suffix to the cronjob (`-$TIMESTAMP`) when creating
		// a job. The job name length limit is 63 characters. Therefore cronjob
		// names must have length <= 63-11=52. If we don't validate this here,
		// then job creation will fail later.
		return field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "must be no more than 52 characters")
	}
	return nil
}
