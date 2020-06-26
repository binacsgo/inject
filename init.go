package inject

import "container/list"

var (
	strategyMap FieldMatchStrategyMap

	globalContainer *Container
)

func init() {
	strategyMap = getStrategy()
	strategyMap[DEFAULT_STRATEGY] = &NameMatchStrategy{}

	globalContainer = &Container{
		objMap:                make(map[string]*ObjInfo, 16),
		registList:            list.New(),
		fieldMatchStrategyMap: strategyMap,
	}
}

func getStrategy() FieldMatchStrategyMap {
	ret := make(FieldMatchStrategyMap)
	ret.regist(&NameMatchStrategy{})
	ret.regist(&TypeMatchStrategy{})
	ret.regist(&MatchStrategyComNameType{})
	return ret
}

// Regist regist to globalContainer
func Regist(name string, obj interface{}) {
	globalContainer.Regist(name, obj)
}

// DoInject inject to globalContainer
func DoInject() error {
	return globalContainer.DoInject()
}

// Report report the globalContainer
func Report() string {
	return globalContainer.Report()
}
