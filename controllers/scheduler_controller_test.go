/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"reflect"
	"testing"
	"time"

	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
	"quortex.io/kubestitute/utils/clusterautoscaler"
)

var ng = clusterautoscaler.NodeGroupStatus{
	Health: clusterautoscaler.NodeGroupHealthCondition{
		NodeCounts: clusterautoscaler.NodeCount{
			Registered: clusterautoscaler.RegisteredNodeCount{
				Total:      5,
				Ready:      1,
				NotStarted: 3,
				Unready: clusterautoscaler.RegisteredUnreadyNodeCount{
					Total: 2,
				},
			},
			LongUnregistered: 6,
		},
		CloudProviderTarget: 7,
	},
}

func fieldPointer(f corev1alpha1.Field) *corev1alpha1.Field {
	return &f
}

func Test_getMatchedPolicy(t *testing.T) {
	type args struct {
		m []matchedPolicy
		p corev1alpha1.SchedulerPolicy
	}
	tests := []struct {
		name string
		args args
		want *matchedPolicy
	}{
		{
			name: "empty matchedPolicy slice should return nil",
			args: args{
				m: nil,
				p: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						IntVal: 1,
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					RightOperand: corev1alpha1.IntOrField{
						IntVal: 2,
					},
					PeriodSeconds: 120,
				},
			},
			want: nil,
		},
		{
			name: "no matching policy should return nil",
			args: args{
				m: []matchedPolicy{
					{
						Policy: corev1alpha1.SchedulerPolicy{
							LeftOperand: corev1alpha1.IntOrField{
								IntVal: 1,
							},
							Operator: corev1alpha1.ComparisonOperatorEqual,
							RightOperand: corev1alpha1.IntOrField{
								IntVal: 2,
							},
							PeriodSeconds: 60,
						},
					},
					{
						Policy: corev1alpha1.SchedulerPolicy{
							LeftOperand: corev1alpha1.IntOrField{
								FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
							},
							Operator: corev1alpha1.ComparisonOperatorGreaterThan,
							RightOperand: corev1alpha1.IntOrField{
								FieldVal: fieldPointer(corev1alpha1.FieldReady),
							},
							PeriodSeconds: 120,
						},
					},
				},
				p: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						IntVal: 2,
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
					PeriodSeconds: 120,
				},
			},
			want: nil,
		},
		{
			name: "matching policy should be returned",
			args: args{
				m: []matchedPolicy{
					{
						Policy: corev1alpha1.SchedulerPolicy{
							LeftOperand: corev1alpha1.IntOrField{
								IntVal: 1,
							},
							Operator: corev1alpha1.ComparisonOperatorEqual,
							RightOperand: corev1alpha1.IntOrField{
								IntVal: 2,
							},
							PeriodSeconds: 60,
						},
					},
					{
						Policy: corev1alpha1.SchedulerPolicy{
							LeftOperand: corev1alpha1.IntOrField{
								FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
							},
							Operator: corev1alpha1.ComparisonOperatorGreaterThan,
							RightOperand: corev1alpha1.IntOrField{
								FieldVal: fieldPointer(corev1alpha1.FieldReady),
							},
							PeriodSeconds: 120,
						},
						Match: time.Date(2020, time.November, 25, 7, 46, 0o4, 409158551, time.UTC),
					},
				},
				p: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
					PeriodSeconds: 120,
				},
			},
			want: &matchedPolicy{
				Policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
					PeriodSeconds: 120,
				},
				Match: time.Date(2020, time.November, 25, 7, 46, 0o4, 409158551, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMatchedPolicy(tt.args.m, tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMatchedPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nodeGroupIntOrFieldValue(t *testing.T) {
	type args struct {
		ngs []clusterautoscaler.NodeGroupStatus
		iof corev1alpha1.IntOrField
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "with 1 nodegroup, no int no field should return zero",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				iof: corev1alpha1.IntOrField{},
			},
			want: 0,
		},
		{
			name: "with 2 nodegroups, no int no field should return zero",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				iof: corev1alpha1.IntOrField{},
			},
			want: 0,
		},

		{
			name: "with 1 nodegroup, an int no field should return the int value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				iof: corev1alpha1.IntOrField{
					IntVal: 2,
				},
			},
			want: 2,
		},
		{
			name: "with 2 nodegroups, an int no field should return the int value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				iof: corev1alpha1.IntOrField{
					IntVal: 2,
				},
			},
			want: 2,
		},
		{
			name: "with 1 nodegroup, field Ready should return the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldReady),
				},
			},
			want: 1,
		},
		{
			name: "with 2 nodegroups, field Ready should return twice the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldReady),
				},
			},
			want: 2,
		},
		{
			name: "with 1 nodegroup, field Unready should return the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldUnready),
				},
			},
			want: 2,
		},
		{
			name: "with 2 nodegroups, field Unready should return twice the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldUnready),
				},
			},
			want: 4,
		},
		{
			name: "with 1 nodegroup, field NotStarted should return the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
				},
			},
			want: 3,
		},
		{
			name: "with 2 nodegroups, field NotStarted should return twice the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
				},
			},
			want: 6,
		},
		{
			name: "with 1 nodegroup, field LongNotStarted should return zero",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldLongNotStarted),
				},
			},
			want: 0,
		},
		{
			name: "with 2 nodegroups, field LongNotStarted should return zero",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldLongNotStarted),
				},
			},
			want: 0,
		},
		{
			name: "with 1 nodegroup, field Registered should return the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldRegistered),
				},
			},
			want: 5,
		},
		{
			name: "with 2 nodegroups, field Registered should return twice the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldRegistered),
				},
			},
			want: 10,
		},
		{
			name: "with 1 nodegroup, field LongUnregistered should return the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldLongUnregistered),
				},
			},
			want: 6,
		},
		{
			name: "with 2 nodegroups, field LongUnregistered should return twice the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldLongUnregistered),
				},
			},
			want: 12,
		},
		{
			name: "with 1 nodegroup, field CloudProviderTarget should return the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
				},
			},
			want: 7,
		},
		{
			name: "with 2 nodegroups, field CloudProviderTarget should return twice the desired value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				iof: corev1alpha1.IntOrField{
					IntVal:   2,
					FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
				},
			},
			want: 14,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nodeGroupIntOrFieldValue(tt.args.ngs, tt.args.iof); got != tt.want {
				t.Errorf("nodeGroupIntOrFieldValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_matchPolicy(t *testing.T) {
	type args struct {
		ngs    []clusterautoscaler.NodeGroupStatus
		policy corev1alpha1.SchedulerPolicy
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "with 1 nodegroup, invalid operator should fail",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						IntVal: 1,
					},
					Operator: corev1alpha1.ComparisonOperator("foo"),
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: false,
		},
		{
			name: "with 2 nodegroups, invalid operator should fail",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						IntVal: 1,
					},
					Operator: corev1alpha1.ComparisonOperator("foo"),
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: false,
		},
		{
			name: "with 1 nodegroup, from 1 / operator = / to field ready (1) should succeed",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						IntVal: 1,
					},
					Operator: corev1alpha1.ComparisonOperatorEqual,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "with 2 nodegroups, from 2 / operator = / to field ready (1 * 2) should succeed",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						IntVal: 2,
					},
					Operator: corev1alpha1.ComparisonOperatorEqual,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "with 1 nodegroup, from field ready / operator >= / to field ready should succeed",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThanOrEqual,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "with 2 nodegroups, from field ready / operator >= / to field ready should succeed",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThanOrEqual,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "with 1 nodegroup, from field ready (1) / operator = / to field unready (2) should fail",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThanOrEqual,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldUnready),
					},
				},
			},
			want: false,
		},
		{
			name: "with 2 nodegroups, from field ready (1) / operator = / to field unready (2 * 2) should fail",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThanOrEqual,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldUnready),
					},
				},
			},
			want: false,
		},
		{
			name: "with 1 nodegroup, from field unready (2) / operator > / to field notstarted (3) should fail",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldUnready),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
					},
				},
			},
			want: false,
		},
		{
			name: "with 2 nodegroups, from field unready (2 * 2) / operator > / to field notstarted (3 * 2) should fail",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldUnready),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
					},
				},
			},
			want: false,
		},
		{
			name: "with 1 nodegroup, from field notstarted (3) / operator > / to field longnotstarted (0) should fail",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
					},
				},
			},
			want: false,
		},
		{
			name: "with 2 nodegroups, from field notstarted (3 * 2) / operator > / to field longnotstarted (0 * 2) should fail",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
					},
				},
			},
			want: false,
		},
		{
			name: "with 1 nodegroup, from field cloudProviderTarget (7) / operator <= / to field ready (1) should fail",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorLowerThanOrEqual,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: false,
		},
		{
			name: "with 2 nodegroups, from field cloudProviderTarget (7 * 2) / operator <= / to field ready (1 * 2) should fail",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorLowerThanOrEqual,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: false,
		},
		{
			name: "with 1 nodegroup, from field cloudProviderTarget (7) / operator > / to field ready (1) should succeed",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "with 2 nodegroups, from field cloudProviderTarget (7 * 2) / operator > / to field ready (1 * 2) should succeed",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "with 1 nodegroup, from field cloudProviderTarget (7) / operator != / to field ready (1) should succeed",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorNotEqual,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "with 2 nodegroups, from field cloudProviderTarget (7 * 2) / operator != / to field ready (1 * 2) should succeed",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorNotEqual,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "with 1 nodegroup, from field cloudProviderTarget (7) / operator < / to field ready (1) should succeed",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldLongUnregistered),
					},
					Operator: corev1alpha1.ComparisonOperatorLowerThan,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
				},
			},
			want: true,
		},
		{
			name: "with 2 nodegroups, from field cloudProviderTarget (7 * 2) / operator < / to field ready (1 * 2) should succeed",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				policy: corev1alpha1.SchedulerPolicy{
					LeftOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldLongUnregistered),
					},
					Operator: corev1alpha1.ComparisonOperatorLowerThan,
					RightOperand: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchPolicy(tt.args.ngs, tt.args.policy); got != tt.want {
				t.Errorf("matchPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nodeGroupReplicas(t *testing.T) {
	type args struct {
		ngs       []clusterautoscaler.NodeGroupStatus
		operation corev1alpha1.IntOrArithmeticOperation
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "with 1 nodegroup, no operation should return int value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal:       3,
					OperationVal: nil,
				},
			},
			want: 3,
		},
		{
			name: "with 2 nodegroups, no operation should return int value",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal:       3,
					OperationVal: nil,
				},
			},
			want: 3,
		},
		{
			name: "with 1 nodegroup, mixed operands / plus operation should work",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					// Operation has higher priority than int value
					IntVal: 12,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							FieldVal: fieldPointer(corev1alpha1.FieldUnready),
						},
						Operator: corev1alpha1.ArithmeticOperatorPlus,
						RightOperand: corev1alpha1.IntOrField{
							IntVal: 2,
						},
					},
				},
			},
			want: 4,
		},
		{
			name: "with 2 nodegroups, mixed operands / plus operation should work",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					// Operation has higher priority than int value
					IntVal: 12,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							FieldVal: fieldPointer(corev1alpha1.FieldUnready),
						},
						Operator: corev1alpha1.ArithmeticOperatorPlus,
						RightOperand: corev1alpha1.IntOrField{
							IntVal: 2,
						},
					},
				},
			},
			want: 6,
		},
		{
			name: "with 1 nodegroup, mixed operands / minus operation should work",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal: 0,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							IntVal: 4,
						},
						Operator: corev1alpha1.ArithmeticOperatorMinus,
						RightOperand: corev1alpha1.IntOrField{
							FieldVal: fieldPointer(corev1alpha1.FieldReady),
						},
					},
				},
			},
			want: 3,
		},
		{
			name: "with 2 nodegroups, mixed operands / minus operation should work",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal: 0,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							IntVal: 4,
						},
						Operator: corev1alpha1.ArithmeticOperatorMinus,
						RightOperand: corev1alpha1.IntOrField{
							FieldVal: fieldPointer(corev1alpha1.FieldReady),
						},
					},
				},
			},
			want: 2,
		},
		{
			name: "with 1 nodegroup, mixed operands / multiply operation should work",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal: 0,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							FieldVal: fieldPointer(corev1alpha1.FieldLongUnregistered),
						},
						Operator: corev1alpha1.ArithmeticOperatorMultiply,
						RightOperand: corev1alpha1.IntOrField{
							IntVal: 2,
						},
					},
				},
			},
			want: 12,
		},
		{
			name: "with 2 nodegroups, mixed operands / multiply operation should work",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal: 0,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							FieldVal: fieldPointer(corev1alpha1.FieldLongUnregistered),
						},
						Operator: corev1alpha1.ArithmeticOperatorMultiply,
						RightOperand: corev1alpha1.IntOrField{
							IntVal: 2,
						},
					},
				},
			},
			want: 24,
		},
		{
			name: "with 1 nodegroup, mixed operands / divide operation should work",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal: 0,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							IntVal: 12,
						},
						Operator: corev1alpha1.ArithmeticOperatorDivide,
						RightOperand: corev1alpha1.IntOrField{
							FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
						},
					},
				},
			},
			want: 4,
		},
		{
			name: "with 2 nodegroups, mixed operands / divide operation should work",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal: 0,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							IntVal: 12,
						},
						Operator: corev1alpha1.ArithmeticOperatorDivide,
						RightOperand: corev1alpha1.IntOrField{
							FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
						},
					},
				},
			},
			want: 2,
		},
		{
			name: "with 1 nodegroup, negative result should return zero",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal: 0,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							IntVal: 1,
						},
						Operator: corev1alpha1.ArithmeticOperatorMinus,
						RightOperand: corev1alpha1.IntOrField{
							IntVal: 2,
						},
					},
				},
			},
			want: 0,
		},
		{
			name: "with 2 nodegroups, negative result should return zero",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal: 0,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							IntVal: 1,
						},
						Operator: corev1alpha1.ArithmeticOperatorMinus,
						RightOperand: corev1alpha1.IntOrField{
							IntVal: 2,
						},
					},
				},
			},
			want: 0,
		},
		{
			name: "with 1 nodegroup, zero division should return zero",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal: 0,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							IntVal: 1,
						},
						Operator: corev1alpha1.ArithmeticOperatorDivide,
						RightOperand: corev1alpha1.IntOrField{
							IntVal: 0,
						},
					},
				},
			},
			want: 0,
		},
		{
			name: "with 2 nodegroups, zero division should return zero",
			args: args{
				ngs: []clusterautoscaler.NodeGroupStatus{ng, ng},
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal: 0,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							IntVal: 1,
						},
						Operator: corev1alpha1.ArithmeticOperatorDivide,
						RightOperand: corev1alpha1.IntOrField{
							IntVal: 0,
						},
					},
				},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nodeGroupReplicas(tt.args.ngs, tt.args.operation); got != tt.want {
				t.Errorf("nodeGroupReplicas() = %v, want %v", got, tt.want)
			}
		})
	}
}
