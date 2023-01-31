package ff

import "github.com/keptn-contrib/dynatrace-service/internal/env"

type GetSLIFeatureFlags struct {
	skipLowercaseSLINames                bool
	skipIncludeSLODisplayNames           bool
	skipCheckDuplicateSLIAndDisplayNames bool
}

func (g GetSLIFeatureFlags) SkipLowercaseSLINames() bool {
	return g.skipLowercaseSLINames
}

func (g GetSLIFeatureFlags) SkipIncludeSLODisplayNames() bool {
	return g.skipIncludeSLODisplayNames
}

func (g GetSLIFeatureFlags) SkipCheckDuplicateSLIAndDisplayNames() bool {
	return g.skipCheckDuplicateSLIAndDisplayNames
}

func LoadGetSLIFeatureFlags() GetSLIFeatureFlags {
	return GetSLIFeatureFlags{
		skipLowercaseSLINames:                env.GetSkipLowercaseSLINames(),
		skipIncludeSLODisplayNames:           env.GetSkipIncludeSLODisplayNames(),
		skipCheckDuplicateSLIAndDisplayNames: env.GetSkipCheckDuplicateSLIAndDisplayNames(),
	}
}
