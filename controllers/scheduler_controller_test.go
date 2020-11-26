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

var ng = clusterautoscaler.NodeGroup{
	Health: clusterautoscaler.NodeGroupHealth{
		Health: clusterautoscaler.Health{
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
					From: corev1alpha1.IntOrField{
						IntVal: 1,
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					To: corev1alpha1.IntOrField{
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
							From: corev1alpha1.IntOrField{
								IntVal: 1,
							},
							Operator: corev1alpha1.ComparisonOperatorEqual,
							To: corev1alpha1.IntOrField{
								IntVal: 2,
							},
							PeriodSeconds: 60,
						},
					},
					{
						Policy: corev1alpha1.SchedulerPolicy{
							From: corev1alpha1.IntOrField{
								FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
							},
							Operator: corev1alpha1.ComparisonOperatorGreaterThan,
							To: corev1alpha1.IntOrField{
								FieldVal: fieldPointer(corev1alpha1.FieldReady),
							},
							PeriodSeconds: 120,
						},
					},
				},
				p: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						IntVal: 2,
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					To: corev1alpha1.IntOrField{
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
							From: corev1alpha1.IntOrField{
								IntVal: 1,
							},
							Operator: corev1alpha1.ComparisonOperatorEqual,
							To: corev1alpha1.IntOrField{
								IntVal: 2,
							},
							PeriodSeconds: 60,
						},
					},
					{
						Policy: corev1alpha1.SchedulerPolicy{
							From: corev1alpha1.IntOrField{
								FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
							},
							Operator: corev1alpha1.ComparisonOperatorGreaterThan,
							To: corev1alpha1.IntOrField{
								FieldVal: fieldPointer(corev1alpha1.FieldReady),
							},
							PeriodSeconds: 120,
						},
						Match: time.Date(2020, time.November, 25, 7, 46, 04, 409158551, time.UTC),
					},
				},
				p: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
					PeriodSeconds: 120,
				},
			},
			want: &matchedPolicy{
				Policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
					PeriodSeconds: 120,
				},
				Match: time.Date(2020, time.November, 25, 7, 46, 04, 409158551, time.UTC),
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
		ng  clusterautoscaler.NodeGroup
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
				ng:  ng,
				iof: corev1alpha1.IntOrField{},
			},
			want: 0,
		},
		{
			name: "an int no field should return the int value",
			args: args{
				ng: ng,
				iof: corev1alpha1.IntOrField{
					IntVal: 2,
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
					FieldVal: fieldPointer(corev1alpha1.FieldReady),
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
					FieldVal: fieldPointer(corev1alpha1.FieldUnready),
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
					FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
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
					FieldVal: fieldPointer(corev1alpha1.FieldLongNotStarted),
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
					FieldVal: fieldPointer(corev1alpha1.FieldRegistered),
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
					FieldVal: fieldPointer(corev1alpha1.FieldLongUnregistered),
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
					FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
				},
			},
			want: 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nodeGroupIntOrFieldValue(tt.args.ng, tt.args.iof); got != tt.want {
				t.Errorf("nodeGroupIntOrFieldValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_matchPolicy(t *testing.T) {
	type args struct {
		ng     clusterautoscaler.NodeGroup
		policy corev1alpha1.SchedulerPolicy
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "invalid operator should fail",
			args: args{
				ng: ng,
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						IntVal: 1,
					},
					Operator: corev1alpha1.ComparisonOperator("foo"),
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: false,
		},
		{
			name: "from 1 / operator = / to field ready (1) should succeed",
			args: args{
				ng: ng,
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						IntVal: 1,
					},
					Operator: corev1alpha1.ComparisonOperatorEqual,
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "from field ready / operator >= / to field ready should succeed",
			args: args{
				ng: ng,
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThanOrEqual,
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "from field ready (1) / operator = / to field unready (2) should fail",
			args: args{
				ng: ng,
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThanOrEqual,
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldUnready),
					},
				},
			},
			want: false,
		},
		{
			name: "from field unready (2) / operator > / to field notstarted (3) should fail",
			args: args{
				ng: ng,
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldUnready),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
					},
				},
			},
			want: false,
		},
		{
			name: "from field notstarted (3) / operator > / to field longnotstarted (4) should fail",
			args: args{
				ng: ng,
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldNotStarted),
					},
				},
			},
			want: false,
		},
		{
			name: "from field cloudProviderTarget (7) / operator <= / to field ready (1) should fail",
			args: args{
				ng: ng,
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorLowerThanOrEqual,
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: false,
		},
		{
			name: "from field cloudProviderTarget (7) / operator > / to field ready (1) should succeed",
			args: args{
				ng: ng,
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorGreaterThan,
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "from field cloudProviderTarget (7) / operator != / to field ready (1) should succeed",
			args: args{
				ng: ng,
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
					Operator: corev1alpha1.ComparisonOperatorNotEqual,
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldReady),
					},
				},
			},
			want: true,
		},
		{
			name: "from field cloudProviderTarget (7) / operator < / to field ready (1) should succeed",
			args: args{
				ng: ng,
				policy: corev1alpha1.SchedulerPolicy{
					From: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldLongUnregistered),
					},
					Operator: corev1alpha1.ComparisonOperatorLowerThan,
					To: corev1alpha1.IntOrField{
						FieldVal: fieldPointer(corev1alpha1.FieldCloudProviderTarget),
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchPolicy(tt.args.ng, tt.args.policy); got != tt.want {
				t.Errorf("matchPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nodeGroupReplicas(t *testing.T) {
	type args struct {
		ng        clusterautoscaler.NodeGroup
		operation corev1alpha1.IntOrArithmeticOperation
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "no operation should return int value",
			args: args{
				ng: ng,
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal:       3,
					OperationVal: nil,
				},
			},
			want: 3,
		},
		{
			name: "mixed operands / plus operation should work",
			args: args{
				ng: ng,
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
			name: "mixed operands / minus operation should work",
			args: args{
				ng: ng,
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
			name: "mixed operands / multiply operation should work",
			args: args{
				ng: ng,
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
			name: "mixed operands / multiply operation should work",
			args: args{
				ng: ng,
				operation: corev1alpha1.IntOrArithmeticOperation{
					IntVal: 0,
					OperationVal: &corev1alpha1.ArithmeticOperation{
						LeftOperand: corev1alpha1.IntOrField{
							IntVal: 12,
						},
						Operator: corev1alpha1.ArithmeticOperatorDivide,
						RightOperand: corev1alpha1.IntOrField{
							FieldVal: fieldPointer(corev1alpha1.FieldLongNotStarted),
						},
					},
				},
			},
			want: 3,
		},
		{
			name: "negative result should return zero",
			args: args{
				ng: ng,
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
			name: "zero division should return zero",
			args: args{
				ng: ng,
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
			if got := nodeGroupReplicas(tt.args.ng, tt.args.operation); got != tt.want {
				t.Errorf("nodeGroupReplicas() = %v, want %v", got, tt.want)
			}
		})
	}
}
