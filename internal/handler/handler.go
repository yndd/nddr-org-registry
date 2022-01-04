package handler

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/nddo-runtime/pkg/resource"
	orgv1alpha1 "github.com/yndd/nddr-org-registry/apis/org/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errApplyNamespace  = "cannot apply namespace"
	errDeleteNamespace = "cannot delete namespace"
)

func New(opts ...Option) (Handler, error) {
	//rgfn := func() niregv1alpha1.Rg { return &niregv1alpha1.Registry{} }
	s := &handler{
		speedy: make(map[string]int),
		//newRegistry: rgfn,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func (r *handler) WithLogger(log logging.Logger) {
	r.log = log
}

func (r *handler) WithClient(c client.Client) {
	r.client = resource.ClientApplicator{
		Client:     c,
		Applicator: resource.NewAPIPatchingApplicator(c),
	}
}

type handler struct {
	log logging.Logger
	// kubernetes
	client resource.ClientApplicator

	speedyMutex sync.Mutex
	speedy      map[string]int
}

func (r *handler) Init(crName string) {
	r.speedyMutex.Lock()
	defer r.speedyMutex.Unlock()
	if _, ok := r.speedy[crName]; !ok {
		r.speedy[crName] = 0
	}
}

func (r *handler) Delete(crName string) {
	r.speedyMutex.Lock()
	defer r.speedyMutex.Unlock()
	delete(r.speedy, crName)
}

func (r *handler) ResetSpeedy(crName string) {
	r.speedyMutex.Lock()
	defer r.speedyMutex.Unlock()
	if _, ok := r.speedy[crName]; ok {
		r.speedy[crName] = 0
	}
}

func (r *handler) GetSpeedy(crName string) int {
	r.speedyMutex.Lock()
	defer r.speedyMutex.Unlock()
	if _, ok := r.speedy[crName]; ok {
		return r.speedy[crName]
	}
	return 9999
}

func (r *handler) IncrementSpeedy(crName string) {
	r.speedyMutex.Lock()
	defer r.speedyMutex.Unlock()
	if _, ok := r.speedy[crName]; ok {
		r.speedy[crName]++
	}
}

func (r *handler) CreateOrganizationNamespace(ctx context.Context, cr orgv1alpha1.Org) error {
	ns := r.buildOrganizationNamespace(cr)
	if err := r.client.Apply(ctx, ns); err != nil {
		return errors.Wrap(err, errApplyNamespace)
	}
	return nil
}

func (r *handler) DeleteOrganizationNamespace(ctx context.Context, cr orgv1alpha1.Org) error {
	ns := r.buildOrganizationNamespace(cr)
	if err := r.client.Delete(ctx, ns); err != nil {
		return errors.Wrap(err, errApplyNamespace)
	}
	return nil
}

func (r *handler) buildOrganizationNamespace(cr orgv1alpha1.Org) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cr.GetName(),
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(cr, orgv1alpha1.OrganizationGroupVersionKind))},
		},
	}
}

func (r *handler) CreateDeploymentNamespace(ctx context.Context, cr orgv1alpha1.Dp) error {
	ns := r.buildDeploymentNamespace(cr)
	if err := r.client.Apply(ctx, ns); err != nil {
		return errors.Wrap(err, errApplyNamespace)
	}
	return nil
}

func (r *handler) DeleteDeploymentNamespace(ctx context.Context, cr orgv1alpha1.Dp) error {
	ns := r.buildDeploymentNamespace(cr)
	if err := r.client.Delete(ctx, ns); err != nil {
		return errors.Wrap(err, errApplyNamespace)
	}
	return nil
}

func (r *handler) buildDeploymentNamespace(cr orgv1alpha1.Dp) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:            strings.Join([]string{cr.GetNamespace(), cr.GetName()}, "-"),
			OwnerReferences: []metav1.OwnerReference{meta.AsController(meta.TypedReferenceTo(cr, orgv1alpha1.DeploymentGroupVersionKind))},
		},
	}
}
