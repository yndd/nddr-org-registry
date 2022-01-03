package handler

import (
	"context"

	"github.com/yndd/ndd-runtime/pkg/logging"
	orgv1alpha1 "github.com/yndd/nddr-org-registry/apis/org/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Option can be used to manipulate Options.
type Option func(Handler)

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(log logging.Logger) Option {
	return func(s Handler) {
		s.WithLogger(log)
	}
}

func WithClient(c client.Client) Option {
	return func(s Handler) {
		s.WithClient(c)
	}
}

type Handler interface {
	WithLogger(log logging.Logger)
	WithClient(client.Client)
	Init(string)
	Delete(string)
	ResetSpeedy(string)
	GetSpeedy(crName string) int
	IncrementSpeedy(crName string)
	CreateOrganizationNamespace(ctx context.Context, cr orgv1alpha1.Org) error
	DeleteOrganizationNamespace(ctx context.Context, cr orgv1alpha1.Org) error
	CreateDeploymentNamespace(ctx context.Context, cr orgv1alpha1.Dp) error
	DeleteDeploymentNamespace(ctx context.Context, cr orgv1alpha1.Dp) error
}
