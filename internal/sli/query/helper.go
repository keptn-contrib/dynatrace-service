package query

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

// IsMatchingMetricID checks whether a query is matching
// When passing a query to dynatrace using filter expressions - the dimension names in a filter will be escaped with
// special characters, e.g: filter(dt.entity.browser,IE) becomes filter(dt~entity~browser,ie).
// This function here tries to come up with a better matching algorithm
// WHILE NOT PERFECT - HERE IS THE FIRST IMPLEMENTATION
func IsMatchingMetricID(singleResultMetricID string, queryMetricID string) bool {
	if strings.Compare(singleResultMetricID, queryMetricID) == 0 {
		return true
	}

	// lets do some basic fuzzy matching
	if strings.Contains(singleResultMetricID, "~") {
		log.WithFields(
			log.Fields{
				"singleResultMetricID": singleResultMetricID,
				"queryMetricID":        queryMetricID,
			}).Debug("Need fuzzy matching")

		//
		// lets just see whether everything until the first : matches
		if strings.Contains(singleResultMetricID, ":") {
			log.Debug("Just compare before first")

			fuzzyResultMetricID := strings.Split(singleResultMetricID, ":")[0]
			fuzzyQueryMetricID := strings.Split(queryMetricID, ":")[0]
			if strings.Compare(fuzzyResultMetricID, fuzzyQueryMetricID) == 0 {
				log.Debug("FUZZY MATCH")
				return true
			}
		}

		// TODO - more fuzzy checks
	}

	return false
}
