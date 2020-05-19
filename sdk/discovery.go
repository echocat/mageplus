package sdk

func Discover(predicates ...Predicate) (Sdk, error) {
	return DiscoverUsing(DefaultDiscoveries, predicates...)
}

func DiscoverUsing(discoveries []Discovery, predicates ...Predicate) (Sdk, error) {
	for _, discovery := range discoveries {
		candidates, err := discovery.Discover()
		if err == ErrNoGoSdk {
			continue
		} else if err != nil {
			return Sdk{}, err
		}
		for _, candidate := range candidates {
			allMatches := true
			for _, predicate := range predicates {
				if match, err := predicate.Matches(candidate); err != nil {
					return Sdk{}, err
				} else if !match {
					allMatches = false
				}
			}
			if allMatches {
				return candidate, nil
			}
		}
	}
	return Sdk{}, ErrNoGoSdk
}

var DefaultDiscoveries = []Discovery{
	DiscoveryFromPath(),
	DiscoveryFromGoroot(),
	NewDefaultDownloadDiscovery(),
}

type Discovery interface {
	Discover() ([]Sdk, error)
}

func DiscoveryFromPath() Discovery {
	return DiscoveryFunc(func() ([]Sdk, error) {
		sdk, err := EvalFromPath()
		if err != nil {
			return nil, err
		}
		return []Sdk{sdk}, nil
	})
}

func DiscoveryFromGoroot() Discovery {
	return DiscoveryFunc(func() ([]Sdk, error) {
		sdk, err := EvalFromGoroot()
		if err != nil {
			return nil, err
		}
		return []Sdk{sdk}, nil
	})
}

type DiscoveryFunc func() ([]Sdk, error)

func (instance DiscoveryFunc) Discover() ([]Sdk, error) {
	return instance()
}
