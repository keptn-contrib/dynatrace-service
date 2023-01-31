package result

import keptn "github.com/keptn/go-utils/pkg/lib"

type SLOCriteria struct {
	Criteria []string
}

func (c *SLOCriteria) toKeptnDomain() *keptn.SLOCriteria {
	return &keptn.SLOCriteria{Criteria: c.Criteria}
}

func sloCriteriaFromKeptnDomain(criteria *keptn.SLOCriteria) *SLOCriteria {
	if criteria == nil {
		return nil
	}

	return &SLOCriteria{
		Criteria: criteria.Criteria,
	}
}

type SLOCriteriaList []*SLOCriteria

func (s SLOCriteriaList) hasActualSLOCriteria() bool {
	for _, c := range s {
		if c == nil {
			continue
		}

		if len(c.Criteria) > 0 {
			return true
		}
	}
	return false
}

func (s SLOCriteriaList) toKeptnDomain() []*keptn.SLOCriteria {
	var criteria []*keptn.SLOCriteria
	for _, c := range s {
		criteria = append(criteria, c.toKeptnDomain())
	}

	return criteria
}

func sloCriteriaListFromKeptnDomain(criteria []*keptn.SLOCriteria) SLOCriteriaList {
	var sloCriteria SLOCriteriaList
	for _, c := range criteria {
		criteria := sloCriteriaFromKeptnDomain(c)
		if criteria != nil {
			sloCriteria = append(sloCriteria, criteria)
		}
	}

	return sloCriteria
}

type SLO struct {
	SLI         string
	DisplayName string
	Pass        SLOCriteriaList
	Warning     SLOCriteriaList
	Weight      int
	KeySLI      bool
}

func (s SLO) ToKeptnDomain() *keptn.SLO {
	return &keptn.SLO{
		SLI:         s.SLI,
		DisplayName: s.DisplayName,
		Pass:        s.Pass.toKeptnDomain(),
		Warning:     s.Warning.toKeptnDomain(),
		Weight:      s.Weight,
		KeySLI:      s.KeySLI,
	}
}

func CreateInformationalSLO(sliName string) SLO {
	return SLO{
		SLI:    sliName,
		Weight: 1,
	}
}

func sloFromKeptnDomain(slo *keptn.SLO) SLO {
	if slo == nil {
		return SLO{}
	}

	return SLO{
		SLI:         slo.SLI,
		DisplayName: slo.DisplayName,
		Pass:        sloCriteriaListFromKeptnDomain(slo.Pass),
		Warning:     sloCriteriaListFromKeptnDomain(slo.Warning),
		Weight:      slo.Weight,
		KeySLI:      slo.KeySLI,
	}
}

func (s SLO) IsNotInformational() bool {
	return s.Pass.hasActualSLOCriteria() || s.Warning.hasActualSLOCriteria()
}

type SLOs []SLO

func (s SLOs) GetAndRemoveFirstSLOWithName(name string) (*SLO, SLOs) {
	for i, o := range s {
		if o.SLI == name {
			return &o, append(s[:i], s[i+1:]...)
		}
	}
	return nil, s
}

func SLOsFromKeptnDomain(slos []*keptn.SLO) SLOs {
	var sloList SLOs
	for _, slo := range slos {
		sloList = append(sloList, sloFromKeptnDomain(slo))
	}

	return sloList
}
