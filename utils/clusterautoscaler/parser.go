package clusterautoscaler

import (
	"bufio"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// ParseReadableString parses the cluster autoscaler status
// in readable format into a ClusterAutoscaler Status struct.
func ParseYamlStatus(s string) (*Status, error) {
	var clusterAutoscalerStatus ClusterAutoscalerStatus
	if err := yaml.Unmarshal([]byte(s), &clusterAutoscalerStatus); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status: %v", err)
	}

	status := Status{
		Time:        parseDate(clusterAutoscalerStatus.Time),
		ClusterWide: convertClusterWideStatus(clusterAutoscalerStatus.ClusterWide),
		NodeGroups:  make([]NodeGroup, len(clusterAutoscalerStatus.NodeGroups)),
	}

	for i := range clusterAutoscalerStatus.NodeGroups {
		status.NodeGroups[i] = convertNodeGroupStatus(clusterAutoscalerStatus.NodeGroups[i])
	}

	return &status, nil
}

func convertClusterWideStatus(status ClusterWideStatus) ClusterWide {
	return ClusterWide{
		Health: Health{
			Status:             HealthStatus(status.Health.Status),
			Ready:              int32(status.Health.NodeCounts.Registered.Ready),
			Unready:            int32(status.Health.NodeCounts.Registered.Unready.Total),
			NotStarted:         int32(status.Health.NodeCounts.Registered.NotStarted),
			LongNotStarted:     0, // FIXME: field does not exist anymore
			Registered:         int32(status.Health.NodeCounts.Registered.Total),
			LongUnregistered:   int32(status.Health.NodeCounts.LongUnregistered),
			LastProbeTime:      status.Health.LastProbeTime.Time,
			LastTransitionTime: status.Health.LastTransitionTime.Time,
		},
		ScaleDown: ScaleDown{
			Status:             ScaleDownStatus(status.ScaleDown.Status),
			Candidates:         int32(status.ScaleDown.Candidates),
			LastProbeTime:      status.ScaleDown.LastProbeTime.Time,
			LastTransitionTime: status.ScaleDown.LastTransitionTime.Time,
		},
		ScaleUp: ScaleUp{
			Status:             ScaleUpStatus(status.ScaleUp.Status),
			LastProbeTime:      status.ScaleUp.LastProbeTime.Time,
			LastTransitionTime: status.ScaleUp.LastTransitionTime.Time,
		},
	}
}

func convertNodeGroupStatus(status NodeGroupStatus) NodeGroup {
	return NodeGroup{
		Name: status.Name,
		Health: NodeGroupHealth{
			Health: Health{
				Status:             HealthStatus(status.Health.Status),
				Ready:              int32(status.Health.NodeCounts.Registered.Ready),
				Unready:            int32(status.Health.NodeCounts.Registered.Unready.Total),
				NotStarted:         int32(status.Health.NodeCounts.Registered.NotStarted),
				LongNotStarted:     0, // FIXME: field does not exist anymore
				Registered:         int32(status.Health.NodeCounts.Registered.Total),
				LongUnregistered:   int32(status.Health.NodeCounts.LongUnregistered),
				LastProbeTime:      status.Health.LastProbeTime.Time,
				LastTransitionTime: status.Health.LastTransitionTime.Time,
			},
			CloudProviderTarget: int32(status.Health.CloudProviderTarget),
			MinSize:             int32(status.Health.MinSize),
			MaxSize:             int32(status.Health.MaxSize),
		},
		ScaleDown: ScaleDown{
			Status:             ScaleDownStatus(status.ScaleDown.Status),
			Candidates:         int32(status.ScaleDown.Candidates),
			LastProbeTime:      status.ScaleDown.LastProbeTime.Time,
			LastTransitionTime: status.ScaleDown.LastTransitionTime.Time,
		},
		ScaleUp: ScaleUp{
			Status:             ScaleUpStatus(status.ScaleUp.Status),
			LastProbeTime:      status.ScaleUp.LastProbeTime.Time,
			LastTransitionTime: status.ScaleUp.LastTransitionTime.Time,
		},
	}
}

const (
	// configMapLastUpdateFormat it the timestamp format used for last update annotation in status ConfigMap
	configMapLastUpdateFormat = "2006-01-02 15:04:05.999999999 -0700 MST"
)

// Some regex to extract data from readable string.
var (
	regexKindName                  = regexp.MustCompile(`\s*Name:`)
	regexKindHealth                = regexp.MustCompile(`\s*Health:`)
	regexKindScaleUp               = regexp.MustCompile(`\s*ScaleUp:`)
	regexKindScaleDown             = regexp.MustCompile(`\s*ScaleDown:`)
	regexKindLastProbeTime         = regexp.MustCompile(`\s*LastProbeTime:`)
	regexKindLastTransitionTime    = regexp.MustCompile(`\s*LastTransitionTime:`)
	regexName                      = regexp.MustCompile(`\s*Name:\s*(\S*)`)
	regexHealthStatus              = regexp.MustCompile(`(Healthy|Unhealthy)`)
	regexHealthReady               = regexp.MustCompile(`[\( ]ready=(\d*)`)
	regexHealthUnready             = regexp.MustCompile(`[\( ]unready=(\d*)`)
	regexHealthNotStarted          = regexp.MustCompile(`[\( ]notStarted=(\d*)`)
	regexHealthLongNotStarted      = regexp.MustCompile(`[\( ]longNotStarted=(\d*)`)
	regexHealthRegistered          = regexp.MustCompile(`[\( ]registered=(\d*)`)
	regexHealthLongUnregistered    = regexp.MustCompile(`[\( ]longUnregistered=(\d*)`)
	regexHealthCloudProviderTarget = regexp.MustCompile(`[\( ]cloudProviderTarget=(\d*)`)
	regexHealthMinSize             = regexp.MustCompile(`[\( ]minSize=(\d*)`)
	regexHealthMaxSize             = regexp.MustCompile(`[\( ]maxSize=(\d*)`)
	regexScaleUpStatus             = regexp.MustCompile(`(Needed|NotNeeded|InProgress|NoActivity|Backoff)`)
	regexScaleDownStatus           = regexp.MustCompile(`(CandidatesPresent|NoCandidates)`)
	regexScaleDownCandidates       = regexp.MustCompile(`[\( ]candidates=(\d*)`)
	regexDate                      = regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(.\d*)? \+\d* [A-Z]*)`)
)

// ParseReadableStatus parses the cluster autoscaler status
// in readable format into a ClusterAutoscaler Status struct.
func ParseReadableStatus(s string) *Status {
	var currentMajor interface{}
	var currentMinor interface{}

	res := &Status{}
	scanner := bufio.NewScanner(strings.NewReader(s))

	for scanner.Scan() {
		line := scanner.Text()
		// ClusterAutoscaler parsing
		if strings.HasPrefix(line, "Cluster-autoscaler status") {
			res.Time = parseDate(regexDate.FindStringSubmatch(line)[1])
			continue
		}

		// ClusterWide parsing
		if strings.HasPrefix(line, "Cluster-wide:") {
			currentMajor = &res.ClusterWide
			continue
		}

		// NodeGroup name parsing
		if regexKindName.MatchString(line) {
			res.NodeGroups = append(res.NodeGroups, NodeGroup{
				Name: regexName.FindStringSubmatch(line)[1],
			})
			currentMajor = &res.NodeGroups[len(res.NodeGroups)-1]
			continue
		}

		// Health line parsing
		if regexKindHealth.MatchString(line) {
			switch reflect.TypeOf(currentMajor) {
			case reflect.TypeOf(&ClusterWide{}):
				h := currentMajor.(*ClusterWide)
				h.Health = parseHealth(line)
				currentMinor = &h.Health
			case reflect.TypeOf(&NodeGroup{}):
				h := currentMajor.(*NodeGroup)
				h.Health = parseNodeGroupHealth(line)
				currentMinor = &h.Health
			}
			continue
		}

		// ScaleUp status parsing
		if regexKindScaleUp.MatchString(line) {
			s := parseScaleUpStatus(line)
			switch reflect.TypeOf(currentMajor) {
			case reflect.TypeOf(&ClusterWide{}):
				h := currentMajor.(*ClusterWide)
				h.ScaleUp.Status = s
				currentMinor = &h.ScaleUp
			case reflect.TypeOf(&NodeGroup{}):
				h := currentMajor.(*NodeGroup)
				h.ScaleUp.Status = s
				currentMinor = &h.ScaleUp
			}
			continue
		}

		// ScaleDown status parsing
		if regexKindScaleDown.MatchString(line) {
			s := parseScaleDown(line)
			switch reflect.TypeOf(currentMajor) {
			case reflect.TypeOf(&ClusterWide{}):
				h := currentMajor.(*ClusterWide)
				h.ScaleDown = s
				currentMinor = &h.ScaleDown
			case reflect.TypeOf(&NodeGroup{}):
				h := currentMajor.(*NodeGroup)
				h.ScaleDown = s
				currentMinor = &h.ScaleDown
			}
			continue
		}

		// LastProbeTime parsing
		if regexKindLastProbeTime.MatchString(line) {
			switch reflect.TypeOf(currentMinor) {
			case reflect.TypeOf(&Health{}):
				h := currentMinor.(*Health)
				h.LastProbeTime = parseDate(regexDate.FindStringSubmatch(line)[1])
			case reflect.TypeOf(&NodeGroupHealth{}):
				h := currentMinor.(*NodeGroupHealth)
				h.LastProbeTime = parseDate(regexDate.FindStringSubmatch(line)[1])
			case reflect.TypeOf(&ScaleUp{}):
				h := currentMinor.(*ScaleUp)
				h.LastProbeTime = parseDate(regexDate.FindStringSubmatch(line)[1])
			case reflect.TypeOf(&ScaleDown{}):
				h := currentMinor.(*ScaleDown)
				h.LastProbeTime = parseDate(regexDate.FindStringSubmatch(line)[1])
			}
			continue
		}

		// LastTransitionTime parsing
		if regexKindLastTransitionTime.MatchString(line) {
			switch reflect.TypeOf(currentMinor) {
			case reflect.TypeOf(&Health{}):
				h := currentMinor.(*Health)
				h.LastTransitionTime = parseDate(regexDate.FindStringSubmatch(line)[1])
			case reflect.TypeOf(&NodeGroupHealth{}):
				h := currentMinor.(*NodeGroupHealth)
				h.LastTransitionTime = parseDate(regexDate.FindStringSubmatch(line)[1])
			case reflect.TypeOf(&ScaleUp{}):
				h := currentMinor.(*ScaleUp)
				h.LastTransitionTime = parseDate(regexDate.FindStringSubmatch(line)[1])
			case reflect.TypeOf(&ScaleDown{}):
				h := currentMinor.(*ScaleDown)
				h.LastTransitionTime = parseDate(regexDate.FindStringSubmatch(line)[1])
			}
			continue
		}
	}

	return res
}

// parseHealthStatus extract HealthStatus from readable string
func parseHealthStatus(s string) HealthStatus {
	return HealthStatus(regexHealthStatus.FindStringSubmatch(s)[1])
}

// parseToInt32 parse a string with given regex and returns submatch
// converted to int.
func parseToInt32(r *regexp.Regexp, s string) int32 {
	i, _ := strconv.Atoi(r.FindStringSubmatch(s)[1])
	return int32(i)
}

// parseScaleDownStatus extract Health data from Health readable string
func parseHealth(s string) Health {
	return Health{
		Status:           parseHealthStatus(s),
		Ready:            parseToInt32(regexHealthReady, s),
		Unready:          parseToInt32(regexHealthUnready, s),
		NotStarted:       parseToInt32(regexHealthNotStarted, s),
		LongNotStarted:   parseToInt32(regexHealthLongNotStarted, s),
		Registered:       parseToInt32(regexHealthRegistered, s),
		LongUnregistered: parseToInt32(regexHealthLongUnregistered, s),
	}
}

// parseNodeGroupHealth extract NodeGroupHealth data from Health readable string
func parseNodeGroupHealth(s string) NodeGroupHealth {
	return NodeGroupHealth{
		Health:              parseHealth(s),
		CloudProviderTarget: parseToInt32(regexHealthCloudProviderTarget, s),
		MinSize:             parseToInt32(regexHealthMinSize, s),
		MaxSize:             parseToInt32(regexHealthMaxSize, s),
	}
}

// parseScaleUpStatus extract ScaleUpStatus from readable string
func parseScaleUpStatus(s string) ScaleUpStatus {
	return ScaleUpStatus(regexScaleUpStatus.FindStringSubmatch(s)[1])
}

// parseScaleDownStatus extract ScaleDownStatus from readable string
func parseScaleDownStatus(s string) ScaleDownStatus {
	return ScaleDownStatus(regexScaleDownStatus.FindStringSubmatch(s)[1])
}

// parseScaleDown extract ScaleDown data from ScaleDown readable string
func parseScaleDown(s string) ScaleDown {
	return ScaleDown{
		Status:     parseScaleDownStatus(s),
		Candidates: parseToInt32(regexScaleDownCandidates, s),
	}
}

// parseScaleDownStatus extract date from readable string
func parseDate(s string) time.Time {
	t, err := time.Parse(configMapLastUpdateFormat, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
