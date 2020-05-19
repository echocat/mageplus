package sdk

import "github.com/blang/semver"

func IsVersion(version string) Predicate {
	return PredicateFunc(func(sdk Sdk) (bool, error) {
		parsed, err := semver.Parse(version)
		if err != nil {
			return false, err
		}
		return parsed.Equals(sdk.Version), nil
	})
}

func IsMinVersion(version string) Predicate {
	return PredicateFunc(func(sdk Sdk) (bool, error) {
		parsed, err := semver.Parse(version)
		if err != nil {
			return false, err
		}
		return parsed.GE(sdk.Version), nil
	})
}

func IsMaxVersion(version string) Predicate {
	return PredicateFunc(func(sdk Sdk) (bool, error) {
		parsed, err := semver.Parse(version)
		if err != nil {
			return false, err
		}
		return parsed.LE(sdk.Version), nil
	})
}

type Predicate interface {
	Matches(Sdk) (bool, error)
}

type PredicateFunc func(Sdk) (bool, error)

func (instance PredicateFunc) Matches(sdk Sdk) (bool, error) {
	return instance(sdk)
}
