package controllers

import (
	"context"
	"time"

	vgsv1alphfa1 "github.com/kubernetes-csi/external-snapshotter/client/v7/apis/volumegroupsnapshot/v1alpha1"
	vsv1 "github.com/kubernetes-csi/external-snapshotter/client/v7/apis/volumesnapshot/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CleanVGSReconciler reconciles a DRCluster object
type CleanVGSReconciler struct {
	client.Client
}

// +kubebuilder:rbac:groups=groupsnapshot.storage.k8s.io,resources=volumegroupsnapshots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=snapshot.storage.k8s.io,resources=volumesnapshots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=list;watch

func (r *CleanVGSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("reconcile enter")

	rgs := &vgsv1alphfa1.VolumeGroupSnapshot{}
	if err := r.Client.Get(ctx, req.NamespacedName, rgs); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, rs := range rgs.Status.VolumeSnapshotRefList {
		volumeSnapshot := &vsv1.VolumeSnapshot{}
		if err := r.Client.Get(ctx, types.NamespacedName{Name: rs.Name, Namespace: rs.Namespace}, volumeSnapshot); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		if !volumeSnapshot.DeletionTimestamp.IsZero() {
			if time.Now().After(volumeSnapshot.DeletionTimestamp.Add(10 * time.Second)) {
				if err := r.Client.Delete(ctx, volumeSnapshot); err != nil {
					return ctrl.Result{}, client.IgnoreNotFound(err)
				}
			} else {
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}
		}
	}

	if !rgs.DeletionTimestamp.IsZero() {
		if time.Now().After(rgs.DeletionTimestamp.Add(10 * time.Second)) {
			if err := r.Client.Delete(ctx, rgs); err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
		} else {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CleanVGSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vgsv1alphfa1.VolumeGroupSnapshot{}).
		Complete(r)
}
