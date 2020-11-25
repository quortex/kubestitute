package clusterautoscaler

import (
	"bufio"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	// configMapLastUpdateFormat it the timestamp format used for last update annotation in status ConfigMap
	configMapLastUpdateFormat = "2006-01-02 15:04:05.999999999 -0700 MST"
)

// Some regex to extract data from readable string.
var (
	regexKindName               = regexp.MustCompile(`\s*Name:`)
	regexKindHealth             = regexp.MustCompile(`\s*Health:`)
	regexKindScaleUp            = regexp.MustCompile(`\s*ScaleUp:`)
	regexKindScaleDown          = regexp.MustCompile(`\s*ScaleDown:`)
	regexKindLastProbeTime      = regexp.MustCompile(`\s*LastProbeTime:`)
	regexKindLastTransitionTime = regexp.MustCompile(`\s*LastTransitionTime:`)
	regexName                   = regexp.MustCompile(`\s*Name:\s*(\w*)`)
	regexHealthStatus           = regexp.MustCompile(`(Healthy|Unhealthy)`)
	regexScaleUpStatus          = regexp.MustCompile(`(Needed|NotNeeded|InProgress|NoActivity|Backoff)`)
	regexScaleDownStatus        = regexp.MustCompile(`(CandidatesPresent|NoCandidates)`)
	regexDate                   = regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(.\d*)? \+\d* [A-Z]*)`)
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
			s := parseHealthStatus(line)
			switch reflect.TypeOf(currentMajor) {
			case reflect.TypeOf(&ClusterWide{}):
				h := currentMajor.(*ClusterWide)
				h.Health.Status = s
				currentMinor = &h.Health
			case reflect.TypeOf(&NodeGroup{}):
				h := currentMajor.(*NodeGroup)
				h.Health.Status = s
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
			s := parseScaleDownStatus(line)
			switch reflect.TypeOf(currentMajor) {
			case reflect.TypeOf(&ClusterWide{}):
				h := currentMajor.(*ClusterWide)
				h.ScaleDown.Status = s
				currentMinor = &h.ScaleDown
			case reflect.TypeOf(&NodeGroup{}):
				h := currentMajor.(*NodeGroup)
				h.ScaleDown.Status = s
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

// parseScaleUpStatus extract ScaleUpStatus from readable string
func parseScaleUpStatus(s string) ScaleUpStatus {
	return ScaleUpStatus(regexScaleUpStatus.FindStringSubmatch(s)[1])
}

// parseScaleDownStatus extract ScaleDownStatus from readable string
func parseScaleDownStatus(s string) ScaleDownStatus {
	return ScaleDownStatus(regexScaleDownStatus.FindStringSubmatch(s)[1])
}

// parseScaleDownStatus extract date from readable string
func parseDate(s string) time.Time {
	t, err := time.Parse(configMapLastUpdateFormat, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
