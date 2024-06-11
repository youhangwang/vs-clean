package controllers

import (
	"context"
	"time"

	vsv1 "github.com/kubernetes-csi/external-snapshotter/client/v7/apis/volumesnapshot/v1"
	"k8s.io/client-go/util/retry"
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
	logger.Info("Get VolumeSnapshot")

	vs := &vsv1.VolumeSnapshot{}
	if err := r.Client.Get(ctx, req.NamespacedName, vs); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !vs.DeletionTimestamp.IsZero() {
		logger.Info("VolumeSnapshot is under deleting")

		if time.Now().After(vs.DeletionTimestamp.Add(10 * time.Second)) {
			logger.Info("VolumeSnapshot is under deleting more than 10s")

			if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				volumeSnapshot := &vsv1.VolumeSnapshot{}
				if err := r.Client.Get(ctx, req.NamespacedName, volumeSnapshot); err != nil {
					return err
				}

				volumeSnapshot.Finalizers = []string{}
				return r.Client.Update(ctx, volumeSnapshot)
			}); err != nil {
				return ctrl.Result{RequeueAfter: 30 * time.Second}, client.IgnoreNotFound(err)
			}

		} else {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CleanVSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vsv1.VolumeSnapshot{}).
		Complete(r)
}
