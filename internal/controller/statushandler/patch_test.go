package statushandler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/api"
	akov2 "github.com/mongodb/mongodb-atlas-kubernetes/v2/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/api/v1/status"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/controller/workflow"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/kube"
)

func Test_PatchUpdateStatus(t *testing.T) {
	existingProject := &akov2.AtlasProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-project",
			Namespace: "test-ns",
		},
		Status: status.AtlasProjectStatus{
			Common: api.Common{Conditions: []api.Condition{{
				Type:   api.IPAccessListReadyType,
				Status: corev1.ConditionFalse,
			}}},
		},
	}
	// Fake client
	scheme := runtime.NewScheme()
	utilruntime.Must(akov2.AddToScheme(scheme))
	// Subresources need to be explicitly set now since controller-runtime 1.15
	// https://github.com/kubernetes-sigs/controller-runtime/issues/2362#issuecomment-1698194188
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingProject).
		WithStatusSubresource(existingProject).Build()

	// Patch the existing project via workflow context
	ctx := &workflow.Context{}
	ctx.SetConditionTrue(api.IPAccessListReadyType)
	existingProject.Status.ID = "theId"
	assert.NoError(t, patchUpdateStatus(ctx, fakeClient, existingProject))

	projectAfterPatch := &akov2.AtlasProject{}
	assert.NoError(t, fakeClient.Get(context.Background(), kube.ObjectKeyFromObject(existingProject), projectAfterPatch))
	// ignore last transition time
	projectAfterPatch.Status.Common.Conditions[0].LastTransitionTime = metav1.Time{}
	assert.Equal(t, []api.Condition{{Type: api.IPAccessListReadyType, Status: corev1.ConditionTrue}}, projectAfterPatch.Status.Common.Conditions)
	assert.Equal(t, "theId", projectAfterPatch.Status.ID)
}
