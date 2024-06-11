package controllers

import (
	"context"
	"time"

	vsv1 "github.com/kubernetes-csi/external-snapshotter/client/v7/apis/volumesnapshot/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CleanVSReconciler reconciles a DRCluster object
type CleanVSReconciler struct {
	client.Client
}

// +kubebuilder:rbac:groups=groupsnapshot.storage.k8s.io,resources=volumegroupsnapshots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=snapshot.storage.k8s.io,resources=volumesnapshots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=list;watch

func (r *CleanVSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("reconcile enter")

	vs := &vsv1.VolumeSnapshot{}
	if err := r.Client.Get(ctx, req.NamespacedName, vs); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !vs.DeletionTimestamp.IsZero() {
		if time.Now().After(vs.DeletionTimestamp.Add(10 * time.Second)) {
			if err := r.Client.Delete(ctx, vs); err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
		} else {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CleanVSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vsv1.VolumeSnapshot{}).
		Complete(r)
}
