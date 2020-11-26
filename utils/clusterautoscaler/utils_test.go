package clusterautoscaler

import (
	"testing"
	"time"

	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
)

var ng = NodeGroup{
	Health: NodeGroupHealth{
		Health: Health{
			Ready:            1,
			Unready:          2,
			NotStarted:       3,
			LongNotStarted:   4,
			Registered:       5,
			LongUnregistered: 6,
		},
		CloudProviderTarget: 7,
	},
}

func TestNodeGroupIntOrFieldValue(t *testing.T) {
	type args struct {
		ng  NodeGroup
		iof corev1alpha1.IntOrField
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "no int no field should return zero",
			args: args{
				ng: ng,
				iof: corev1alpha1.IntOrField{
					IntVal:   0,
					FieldVal: "",
				},
			},
			want: 0,
		},
		{
			name: "an int no field should return the int value",
			args: args{
				ng: ng,
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: "",
				},
			},
			want: 2,
		},
		{
			name: "an int and an invalid field should return the int value",
			args: args{
				ng: ng,
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: "foo",
				},
			},
			want: 2,
		},
		{
			name: "field Ready should return the desired value",
			args: args{
				ng: ng,
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: corev1alpha1.FieldReady,
				},
			},
			want: 1,
		},
		{
			name: "field Unready should return the desired value",
			args: args{
				ng: ng,
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: corev1alpha1.FieldUnready,
				},
			},
			want: 2,
		},
		{
			name: "field NotStarted should return the desired value",
			args: args{
				ng: ng,
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: corev1alpha1.FieldNotStarted,
				},
			},
			want: 3,
		},
		{
			name: "field LongNotStarted should return the desired value",
			args: args{
				ng: ng,
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: corev1alpha1.FieldLongNotStarted,
				},
			},
			want: 4,
		},
		{
			name: "field Registered should return the desired value",
			args: args{
				ng: ng,
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: corev1alpha1.FieldRegistered,
				},
			},
			want: 5,
		},
		{
			name: "field LongUnregistered should return the desired value",
			args: args{
				ng: ng,
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: corev1alpha1.FieldLongUnregistered,
				},
			},
			want: 6,
		},
		{
			name: "field CloudProviderTarget should return the desired value",
			args: args{
				ng: ng,
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: corev1alpha1.FieldCloudProviderTarget,
				},
			},
			want: 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NodeGroupIntOrFieldValue(tt.args.ng, tt.args.iof); got != tt.want {
				t.Errorf("NodeGroupIntOrFieldValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_matchPolicy(t *testing.T) {
	type fields struct {
		Name      string
		Health    NodeGroupHealth
		ScaleDown ScaleDown
		ScaleUp   ScaleUp
	}
	type args struct {
		policy corev1alpha1.SchedulerPolicy
		now    now
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "from 1 / operator = / to field ready (1) / period < now should fail",
			fields: fields{
				Health: ng.Health,
			},
			args: args{
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						IntVal: 1,
					},
					ArithmeticOperator: corev1alpha1.ArithmeticOperatorEqual,
					To: corev1alpha1.IntOrField{
						FieldVal: corev1alpha1.FieldReady,
					},
					PeriodSeconds: 20,
				},
				now: func() time.Time {
					return ng.Health.LastTransitionTime.Add(time.Second * 21)
				},
			},
			want: false,
		},
		{
			name: "from 1 / operator = / to field ready (1) / period <= now should succeed",
			fields: fields{
				Health: ng.Health,
			},
			args: args{
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						IntVal: 1,
					},
					ArithmeticOperator: corev1alpha1.ArithmeticOperatorEqual,
					To: corev1alpha1.IntOrField{
						FieldVal: corev1alpha1.FieldReady,
					},
					PeriodSeconds: 20,
				},
				now: func() time.Time {
					return ng.Health.LastTransitionTime.Add(time.Second * 20)
				},
			},
			want: true,
		},
		{
			name: "from field cloudProviderTarget (7) / operator > / to field ready (1) / period <= now should succeed",
			fields: fields{
				Health: ng.Health,
			},
			args: args{
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						FieldVal: corev1alpha1.FieldCloudProviderTarget,
					},
					ArithmeticOperator: corev1alpha1.ArithmeticOperatorGreaterThan,
					To: corev1alpha1.IntOrField{
						FieldVal: corev1alpha1.FieldReady,
					},
					PeriodSeconds: 20,
				},
				now: func() time.Time {
					return ng.Health.LastTransitionTime.Add(time.Second * 20)
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NodeGroup{
				Name:      tt.fields.Name,
				Health:    tt.fields.Health,
				ScaleDown: tt.fields.ScaleDown,
				ScaleUp:   tt.fields.ScaleUp,
			}
			if got := n.matchPolicy(tt.args.policy, tt.args.now); got != tt.want {
				t.Errorf("NodeGroup.matchPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}
