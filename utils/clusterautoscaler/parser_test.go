package clusterautoscaler

import (
	"testing"
	"time"

	"github.com/go-test/deep"
)

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
      unregistered: 0
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
      unregistered: 0
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
		want    *Status
		wantErr bool
	}{
		{
			name: "a fully functional status",
			args: args{
				s: yamlStatus,
			},
			want: &Status{
				Time: time.Date(2020, time.November, 25, 8, 19, 44, 90873082, time.UTC),
				ClusterWide: ClusterWide{
					Health: Health{
						Status:             HealthStatusHealthy,
						Ready:              4,
						Unready:            2,
						NotStarted:         1,
						Registered:         5,
						LongUnregistered:   5,
						LastProbeTime:      lpt,
						LastTransitionTime: time.Date(2020, time.November, 25, 7, 46, 04, 409158551, time.UTC),
					},
					ScaleDown: ScaleDown{
						Status:             ScaleDownCandidatesPresent,
						Candidates:         1,
						LastProbeTime:      lpt,
						LastTransitionTime: time.Date(2020, time.November, 25, 8, 19, 34, 73648791, time.UTC),
					},
					ScaleUp: ScaleUp{
						Status:             ScaleUpInProgress,
						LastProbeTime:      lpt,
						LastTransitionTime: time.Date(2020, time.November, 25, 8, 18, 33, 613103712, time.UTC),
					},
				},
				NodeGroups: []NodeGroup{
					{
						Name: "foo",
						Health: NodeGroupHealth{
							Health: Health{
								Status:             HealthStatusHealthy,
								Ready:              1,
								Unready:            2,
								NotStarted:         3,
								Registered:         5,
								LongUnregistered:   6,
								LastProbeTime:      lpt,
								LastTransitionTime: time.Date(2020, time.November, 25, 7, 46, 4, 409158551, time.UTC),
							},
							CloudProviderTarget: 2,
							MinSize:             1,
							MaxSize:             3,
						},
						ScaleDown: ScaleDown{
							Status:             ScaleDownCandidatesPresent,
							Candidates:         1,
							LastProbeTime:      lpt,
							LastTransitionTime: time.Date(2020, time.November, 25, 8, 19, 34, 73648791, time.UTC),
						},
						ScaleUp: ScaleUp{
							Status:             ScaleUpInProgress,
							LastProbeTime:      lpt,
							LastTransitionTime: time.Date(2020, time.November, 25, 8, 18, 33, 613103712, time.UTC),
						},
					},
					{
						Name: "bar",
						Health: NodeGroupHealth{
							Health: Health{
								Status:             HealthStatusHealthy,
								Ready:              2,
								Unready:            1,
								NotStarted:         2,
								Registered:         2,
								LongUnregistered:   4,
								LastProbeTime:      lpt,
								LastTransitionTime: time.Time{}},
							CloudProviderTarget: 2,
							MinSize:             0,
							MaxSize:             3,
						},
						ScaleDown: ScaleDown{
							Status:             ScaleDownNoCandidates,
							LastProbeTime:      lpt,
							LastTransitionTime: time.Date(2020, time.November, 25, 8, 14, 52, 480583803, time.UTC),
						},
						ScaleUp: ScaleUp{
							Status:             ScaleUpNoActivity,
							LastProbeTime:      lpt,
							LastTransitionTime: time.Date(2020, time.November, 25, 8, 14, 42, 467240558, time.UTC),
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

var lpt = time.Date(2020, time.November, 25, 8, 19, 44, 88071148, time.UTC)

func TestParseReadableStatus(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want *Status
	}{
		{
			name: "a fully functional status",
			args: args{
				s: readableStatus,
			},
			want: &Status{
				Time: time.Date(2020, time.November, 25, 8, 19, 44, 90873082, time.UTC),
				ClusterWide: ClusterWide{
					Health: Health{
						Status:             HealthStatusHealthy,
						Ready:              4,
						Unready:            2,
						NotStarted:         1,
						Registered:         5,
						LongUnregistered:   5,
						LastProbeTime:      lpt,
						LastTransitionTime: time.Date(2020, time.November, 25, 7, 46, 04, 409158551, time.UTC),
					},
					ScaleDown: ScaleDown{
						Status:             ScaleDownCandidatesPresent,
						Candidates:         1,
						LastProbeTime:      lpt,
						LastTransitionTime: time.Date(2020, time.November, 25, 8, 19, 34, 73648791, time.UTC),
					},
					ScaleUp: ScaleUp{
						Status:             ScaleUpInProgress,
						LastProbeTime:      lpt,
						LastTransitionTime: time.Date(2020, time.November, 25, 8, 18, 33, 613103712, time.UTC),
					},
				},
				NodeGroups: []NodeGroup{
					{
						Name: "foo",
						Health: NodeGroupHealth{
							Health: Health{
								Status:             HealthStatusHealthy,
								Ready:              1,
								Unready:            2,
								NotStarted:         3,
								Registered:         5,
								LongUnregistered:   6,
								LastProbeTime:      lpt,
								LastTransitionTime: time.Date(2020, time.November, 25, 7, 46, 4, 409158551, time.UTC),
							},
							CloudProviderTarget: 2,
							MinSize:             1,
							MaxSize:             3,
						},
						ScaleDown: ScaleDown{
							Status:             ScaleDownCandidatesPresent,
							Candidates:         1,
							LastProbeTime:      lpt,
							LastTransitionTime: time.Date(2020, time.November, 25, 8, 19, 34, 73648791, time.UTC),
						},
						ScaleUp: ScaleUp{
							Status:             ScaleUpInProgress,
							LastProbeTime:      lpt,
							LastTransitionTime: time.Date(2020, time.November, 25, 8, 18, 33, 613103712, time.UTC),
						},
					},
					{
						Name: "bar",
						Health: NodeGroupHealth{
							Health: Health{
								Status:             HealthStatusHealthy,
								Ready:              2,
								Unready:            1,
								NotStarted:         2,
								Registered:         2,
								LongUnregistered:   4,
								LastProbeTime:      lpt,
								LastTransitionTime: time.Time{}},
							CloudProviderTarget: 2,
							MinSize:             0,
							MaxSize:             3,
						},
						ScaleDown: ScaleDown{
							Status:             ScaleDownNoCandidates,
							LastProbeTime:      lpt,
							LastTransitionTime: time.Date(2020, time.November, 25, 8, 14, 52, 480583803, time.UTC),
						},
						ScaleUp: ScaleUp{
							Status:             ScaleUpNoActivity,
							LastProbeTime:      lpt,
							LastTransitionTime: time.Date(2020, time.November, 25, 8, 14, 42, 467240558, time.UTC),
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
