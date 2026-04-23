package tokenwatch

import (
	"fmt"

	"github.com/vaultwatch/internal/alert"
)

// CohortScanner scans each token, adds it to the appropriate cohort group,
// and emits an info alert whenever a group grows beyond a configured size
// threshold — useful for detecting unexpected token proliferation per label.
type CohortScanner struct {
	registry  *Registry
	cohort    *Cohort
	lookup    func(tokenID string) (TokenInfo, error)
	threshold int
}

// NewCohortScanner constructs a CohortScanner. threshold is the group-size
// that triggers an alert (0 disables alerts).
func NewCohortScanner(registry *Registry, cohort *Cohort, lookup func(string) (TokenInfo, error), threshold int) *CohortScanner {
	if registry == nil {
		panic("cohort scanner: registry must not be nil")
	}
	if cohort == nil {
		panic("cohort scanner: cohort must not be nil")
	}
	if lookup == nil {
		panic("cohort scanner: lookup must not be nil")
	}
	return &CohortScanner{
		registry:  registry,
		cohort:    cohort,
		lookup:    lookup,
		threshold: threshold,
	}
}

// Scan iterates over all registered tokens, updates cohort membership, and
// returns alerts for groups that exceed the size threshold.
func (s *CohortScanner) Scan() []alert.Alert {
	tokens := s.registry.List()
	for _, id := range tokens {
		info, err := s.lookup(id)
		if err != nil {
			continue
		}
		groupVal := info.Labels[s.cohort.cfg.GroupKey]
		if groupVal == "" {
			groupVal = "__default__"
		}
		s.cohort.Add(groupVal, id)
	}

	if s.threshold <= 0 {
		return nil
	}

	var alerts []alert.Alert
	for _, grp := range s.cohort.Groups() {
		members := s.cohort.Members(grp)
		if len(members) >= s.threshold {
			alerts = append(alerts, alert.Alert{
				LeaseID: fmt.Sprintf("cohort:%s", grp),
				Message: fmt.Sprintf("cohort group %q has %d members (threshold %d)", grp, len(members), s.threshold),
				Level:   alert.LevelWarning,
				Labels:  map[string]string{s.cohort.cfg.GroupKey: grp},
			})
		}
	}
	return alerts
}
