package clusterautoscaler

import (
	"testing"
	"time"

	"github.com/go-test/deep"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var lastProbingTime = metav1.NewTime(time.Date(2020, time.November, 25, 8, 19, 44, 88071148, time.UTC))

const yamlStatus = `
time: 2020-11-25 08:19:44.090873082 +0000 UTC
autoscalerStatus: Running
clusterWide:
  health:
    status: Healthy
    nodeCounts:
      registered:
        total: 5
        ready: 4
        notStarted: 1
        unready:
          total: 2
          resourceUnready: 0
      longUnregistered: 5
      unregistered: 6
    lastProbeTime: "2020-11-25T08:19:44.088071148Z"
    lastTransitionTime: "2020-11-25T07:46:04.409158551Z"
  scaleUp:
    status: InProgress
    lastProbeTime: "2020-11-25T08:19:44.088071148Z"
    lastTransitionTime: "2020-11-25T08:18:33.613103712Z"
  scaleDown:
    status: CandidatesPresent
    candidates: 1
    lastProbeTime: "2020-11-25T08:19:44.088071148Z"
    lastTransitionTime: "2020-11-25T08:19:34.073648791Z"
nodeGroups:
- name: foo
  health:
    status: Healthy
    nodeCounts:
      registered:
        total: 5
        ready: 1
        notStarted: 3
        unready:
          total: 2
          resourceUnready: 0
      longUnregistered: 6
      unregistered: 7
    cloudProviderTarget: 2
    minSize: 1
    maxSize: 3
    lastProbeTime: "2020-11-25T08:19:44.088071148Z"
    lastTransitionTime: "2020-11-25T07:46:04.409158551Z"
  scaleUp:
    status: InProgress
    lastProbeTime: "2020-11-25T08:19:44.088071148Z"
    lastTransitionTime: "2020-11-25T08:18:33.613103712Z"
  scaleDown:
    status: CandidatesPresent
    candidates: 1
    lastProbeTime: "2020-11-25T08:19:44.088071148Z"
    lastTransitionTime: "2020-11-25T08:19:34.073648791Z"
- name: bar
  health:
    status: Healthy
    nodeCounts:
      registered:
        total: 2
        ready: 2
        notStarted: 2
        unready:
          total: 1
          resourceUnready: 0
      longUnregistered: 4
      unregistered: 0
    cloudProviderTarget: 2
    minSize: 0
    maxSize: 3
    lastProbeTime: "2020-11-25T08:19:44.088071148Z"
    lastTransitionTime: "0001-01-01T00:00:00Z"
  scaleUp:
    status: NoActivity
    lastProbeTime: "2020-11-25T08:19:44.088071148Z"
    lastTransitionTime: "2020-11-25T08:14:42.467240558Z"
  scaleDown:
    status: NoCandidates
    lastProbeTime: "2020-11-25T08:19:44.088071148Z"
    lastTransitionTime: "2020-11-25T08:14:52.480583803Z"
`

func TestParseYamlStatus(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *ClusterAutoscalerStatus
		wantErr bool
	}{
		{
			name: "a fully functional status",
			args: args{
				s: yamlStatus,
			},
			want: &ClusterAutoscalerStatus{
				Time:             "2020-11-25 08:19:44.090873082 +0000 UTC",
				AutoscalerStatus: ClusterAutoscalerRunning,
				ClusterWide: ClusterWideStatus{
					Health: ClusterHealthCondition{
						Status: ClusterAutoscalerHealthy,
						NodeCounts: NodeCount{
							Registered: RegisteredNodeCount{
								Total: 5,
								Ready: 4,
								Unready: RegisteredUnreadyNodeCount{
									Total:           2,
									ResourceUnready: 0,
								},
								NotStarted:   1,
								BeingDeleted: 0,
							},
							LongUnregistered: 5,
							Unregistered:     6,
						},
						LastProbeTime:      lastProbingTime,
						LastTransitionTime: metav1.Date(2020, time.November, 25, 7, 46, 0o4, 409158551, time.UTC),
					},
					ScaleUp: ClusterScaleUpCondition{
						Status:             ClusterAutoscalerInProgress,
						LastProbeTime:      lastProbingTime,
						LastTransitionTime: metav1.Date(2020, time.November, 25, 8, 18, 33, 613103712, time.UTC),
					},
					ScaleDown: ScaleDownCondition{
						Status:             ClusterAutoscalerCandidatesPresent,
						Candidates:         1,
						LastProbeTime:      lastProbingTime,
						LastTransitionTime: metav1.Date(2020, time.November, 25, 8, 19, 34, 73648791, time.UTC),
					},
				},
				NodeGroups: []NodeGroupStatus{
					{
						Name: "foo",
						Health: NodeGroupHealthCondition{
							Status: ClusterAutoscalerHealthy,
							NodeCounts: NodeCount{
								Registered: RegisteredNodeCount{
									Total: 5,
									Ready: 1,
									Unready: RegisteredUnreadyNodeCount{
										Total:           2,
										ResourceUnready: 0,
									},
									NotStarted:   3,
									BeingDeleted: 0,
								},
								LongUnregistered: 6,
								Unregistered:     7,
							},
							CloudProviderTarget: 2,
							MinSize:             1,
							MaxSize:             3,
							LastProbeTime:       lastProbingTime,
							LastTransitionTime:  metav1.Date(2020, time.November, 25, 7, 46, 4, 409158551, time.UTC),
						},
						ScaleUp: NodeGroupScaleUpCondition{
							Status:             ClusterAutoscalerInProgress,
							BackoffInfo:        BackoffInfo{},
							LastProbeTime:      lastProbingTime,
							LastTransitionTime: metav1.Date(2020, time.November, 25, 8, 18, 33, 613103712, time.UTC),
						},
						ScaleDown: ScaleDownCondition{
							Status:             ClusterAutoscalerCandidatesPresent,
							Candidates:         1,
							LastProbeTime:      lastProbingTime,
							LastTransitionTime: metav1.Date(2020, time.November, 25, 8, 19, 34, 73648791, time.UTC),
						},
					},
					{
						Name: "bar",
						Health: NodeGroupHealthCondition{
							Status: ClusterAutoscalerHealthy,
							NodeCounts: NodeCount{
								Registered: RegisteredNodeCount{
									Total: 2,
									Ready: 2,
									Unready: RegisteredUnreadyNodeCount{
										Total:           1,
										ResourceUnready: 0,
									},
									NotStarted:   2,
									BeingDeleted: 0,
								},
								LongUnregistered: 4,
								Unregistered:     0,
							},
							CloudProviderTarget: 2,
							MinSize:             0,
							MaxSize:             3,
							LastProbeTime:       lastProbingTime,
							LastTransitionTime:  metav1.Time{},
						},
						ScaleUp: NodeGroupScaleUpCondition{
							Status:             ClusterAutoscalerNoActivity,
							BackoffInfo:        BackoffInfo{},
							LastProbeTime:      lastProbingTime,
							LastTransitionTime: metav1.Date(2020, time.November, 25, 8, 14, 42, 467240558, time.UTC),
						},
						ScaleDown: ScaleDownCondition{
							Status:             ClusterAutoscalerNoCandidates,
							Candidates:         0,
							LastProbeTime:      lastProbingTime,
							LastTransitionTime: metav1.Date(2020, time.November, 25, 8, 14, 52, 480583803, time.UTC),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseYamlStatus(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseYamlStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

const readableStatus = `
Cluster-autoscaler status at 2020-11-25 08:19:44.090873082 +0000 UTC:
Cluster-wide:
	Health:      Healthy (ready=4 unready=2 notStarted=1 longNotStarted=0 registered=5 longUnregistered=5)
								LastProbeTime:      2020-11-25 08:19:44.088071148 +0000 UTC m=+2030.020714775
								LastTransitionTime: 2020-11-25 07:46:04.409158551 +0000 UTC m=+10.341802256
	ScaleUp:     InProgress (ready=4 registered=5)
								LastProbeTime:      2020-11-25 08:19:44.088071148 +0000 UTC m=+2030.020714775
								LastTransitionTime: 2020-11-25 08:18:33.613103712 +0000 UTC m=+1959.545747280
	ScaleDown:   CandidatesPresent (candidates=1)
								LastProbeTime:      2020-11-25 08:19:44.088071148 +0000 UTC m=+2030.020714775
								LastTransitionTime: 2020-11-25 08:19:34.073648791 +0000 UTC m=+2020.006292413

NodeGroups:
	Name:        foo
	Health:      Healthy (ready=1 unready=2 notStarted=3 longNotStarted=0 registered=5 longUnregistered=6 cloudProviderTarget=2 (minSize=1, maxSize=3))
								LastProbeTime:      2020-11-25 08:19:44.088071148 +0000 UTC m=+2030.020714775
								LastTransitionTime: 2020-11-25 07:46:04.409158551 +0000 UTC m=+10.341802256
	ScaleUp:     InProgress (ready=1 cloudProviderTarget=2)
								LastProbeTime:      2020-11-25 08:19:44.088071148 +0000 UTC m=+2030.020714775
								LastTransitionTime: 2020-11-25 08:18:33.613103712 +0000 UTC m=+1959.545747280
	ScaleDown:   CandidatesPresent (candidates=1)
								LastProbeTime:      2020-11-25 08:19:44.088071148 +0000 UTC m=+2030.020714775
								LastTransitionTime: 2020-11-25 08:19:34.073648791 +0000 UTC m=+2020.006292413

	Name:        bar
	Health:      Healthy (ready=2 unready=1 notStarted=2 longNotStarted=0 registered=2 longUnregistered=4 cloudProviderTarget=2 (minSize=0, maxSize=3))
								LastProbeTime:      2020-11-25 08:19:44.088071148 +0000 UTC m=+2030.020714775
								LastTransitionTime: 0001-01-01 00:00:00 +0000 UTC
	ScaleUp:     NoActivity (ready=2 cloudProviderTarget=2)
								LastProbeTime:      2020-11-25 08:19:44.088071148 +0000 UTC m=+2030.020714775
								LastTransitionTime: 2020-11-25 08:14:42.467240558 +0000 UTC m=+1728.399884162
	ScaleDown:   NoCandidates (candidates=0)
								LastProbeTime:      2020-11-25 08:19:44.088071148 +0000 UTC m=+2030.020714775
								LastTransitionTime: 2020-11-25 08:14:52.480583803 +0000 UTC m=+1738.413227454
`

func TestParseReadableStatus(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want *ClusterAutoscalerStatus
	}{
		{
			name: "a fully functional status",
			args: args{
				s: readableStatus,
			},
			want: &ClusterAutoscalerStatus{
				Time:             "2020-11-25 08:19:44.090873082 +0000 UTC",
				AutoscalerStatus: "", // Present in readable status but not parsed
				ClusterWide: ClusterWideStatus{
					Health: ClusterHealthCondition{
						Status: ClusterAutoscalerHealthy,
						NodeCounts: NodeCount{
							Registered: RegisteredNodeCount{
								Total: 5,
								Ready: 4,
								Unready: RegisteredUnreadyNodeCount{
									Total:           2,
									ResourceUnready: 0,
								},
								NotStarted:   1,
								BeingDeleted: 0,
							},
							LongUnregistered: 5,
							Unregistered:     0, // Not present in readable status
						},
						LastProbeTime:      lastProbingTime,
						LastTransitionTime: metav1.NewTime(time.Date(2020, time.November, 25, 7, 46, 0o4, 409158551, time.UTC)),
					},
					ScaleUp: ClusterScaleUpCondition{
						Status:             ClusterAutoscalerInProgress,
						LastProbeTime:      lastProbingTime,
						LastTransitionTime: metav1.NewTime(time.Date(2020, time.November, 25, 8, 18, 33, 613103712, time.UTC)),
					},
					ScaleDown: ScaleDownCondition{
						Status:             ClusterAutoscalerCandidatesPresent,
						Candidates:         1,
						LastProbeTime:      lastProbingTime,
						LastTransitionTime: metav1.NewTime(time.Date(2020, time.November, 25, 8, 19, 34, 73648791, time.UTC)),
					},
				},
				NodeGroups: []NodeGroupStatus{
					{
						Name: "foo",
						Health: NodeGroupHealthCondition{
							Status: ClusterAutoscalerHealthy,
							NodeCounts: NodeCount{
								Registered: RegisteredNodeCount{
									Total: 5,
									Ready: 1,
									Unready: RegisteredUnreadyNodeCount{
										Total:           2,
										ResourceUnready: 0,
									},
									NotStarted:   3,
									BeingDeleted: 0,
								},
								LongUnregistered: 6,
								Unregistered:     0, // Not present in readable status
							},
							CloudProviderTarget: 2,
							MinSize:             1,
							MaxSize:             3,
							LastProbeTime:       lastProbingTime,
							LastTransitionTime:  metav1.NewTime(time.Date(2020, time.November, 25, 7, 46, 4, 409158551, time.UTC)),
						},
						ScaleUp: NodeGroupScaleUpCondition{
							Status:             ClusterAutoscalerInProgress,
							BackoffInfo:        BackoffInfo{},
							LastProbeTime:      lastProbingTime,
							LastTransitionTime: metav1.NewTime(time.Date(2020, time.November, 25, 8, 18, 33, 613103712, time.UTC)),
						},
						ScaleDown: ScaleDownCondition{
							Status:             ClusterAutoscalerCandidatesPresent,
							Candidates:         1,
							LastProbeTime:      lastProbingTime,
							LastTransitionTime: metav1.NewTime(time.Date(2020, time.November, 25, 8, 19, 34, 73648791, time.UTC)),
						},
					},
					{
						Name: "bar",
						Health: NodeGroupHealthCondition{
							Status: ClusterAutoscalerHealthy,
							NodeCounts: NodeCount{
								Registered: RegisteredNodeCount{
									Total: 2,
									Ready: 2,
									Unready: RegisteredUnreadyNodeCount{
										Total:           1,
										ResourceUnready: 0,
									},
									NotStarted:   2,
									BeingDeleted: 0,
								},
								LongUnregistered: 4,
								Unregistered:     0, // Not present in readable status
							},
							CloudProviderTarget: 2,
							MinSize:             0,
							MaxSize:             3,
							LastProbeTime:       lastProbingTime,
							LastTransitionTime:  metav1.NewTime(time.Time{}),
						},
						ScaleUp: NodeGroupScaleUpCondition{
							Status:             ClusterAutoscalerNoActivity,
							BackoffInfo:        BackoffInfo{},
							LastProbeTime:      lastProbingTime,
							LastTransitionTime: metav1.NewTime(time.Date(2020, time.November, 25, 8, 14, 42, 467240558, time.UTC)),
						},
						ScaleDown: ScaleDownCondition{
							Status:             ClusterAutoscalerNoCandidates,
							Candidates:         0,
							LastProbeTime:      lastProbingTime,
							LastTransitionTime: metav1.NewTime(time.Date(2020, time.November, 25, 8, 14, 52, 480583803, time.UTC)),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseReadableStatus(tt.args.s)
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}
