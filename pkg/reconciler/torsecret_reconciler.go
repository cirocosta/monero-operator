package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cirocosta/monero-operator/pkg/tor"
)

type TorSecretsReconciler struct {
	client.Client
	Log logr.Logger
}

func (r *TorSecretsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	secret, err := r.GetSecret(ctx, req.Name, req.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			return EmptyResult(), nil
		}

		return EmptyResult(), fmt.Errorf("get moneronodeset: %w", err)
	}

	err = r.ReconcileSecret(ctx, secret)
	if err != nil {
		return EmptyResult(), fmt.Errorf("reconcile moneronodeset: %w", err)
	}

	return EmptyResult(), nil
}

func (r *TorSecretsReconciler) ReconcileSecret(
	ctx context.Context,
	secret *corev1.Secret,
) error {
	if r.SecretAlreadyFilled(secret) {
		return nil
	}

	if err := r.FillSecret(secret); err != nil {
		return fmt.Errorf("fill secret: %w", err)
	}

	if err := r.Client.Update(ctx, secret); err != nil {
		return fmt.Errorf("update secret: %w", err)
	}

	return nil
}

func (r *TorSecretsReconciler) FillSecret(secret *corev1.Secret) error {
	creds, err := tor.GenerateCredentials()
	if err != nil {
		return fmt.Errorf("generate credentials: %w", err)
	}

	secret.Data = map[string][]byte(creds)
	return nil
}

func (r *TorSecretsReconciler) SecretAlreadyFilled(
	secret *corev1.Secret,
) bool {
	for _, field := range []string{
		tor.FilenameHostname,
		tor.FilenamePublicKey,
		tor.FilenameSecretKey,
	} {
		v, found := secret.Data[field]
		if !found || len(v) == 0 {
			return false
		}
	}

	return true
}

func (r *TorSecretsReconciler) GetSecret(
	ctx context.Context,
	name, namespace string,
) (*corev1.Secret, error) {
	obj := &corev1.Secret{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, obj); err != nil {
		return nil, fmt.Errorf("get %s/%s: %w", namespace, name, err)
	}

	return obj, nil
}
