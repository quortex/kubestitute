package clusterautoscaler

import (
	"bufio"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

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
	regexName                      = regexp.MustCompile(`\s*Name:\s*(\w*)`)
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

// ParseReadableString parses the cluster autoscaler status
// in readable format into a ClusterAutoscaler Status struct.
func ParseReadableString(s string) *Status {

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
