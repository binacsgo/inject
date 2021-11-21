package inject

import "fmt"

const (
	FieldMatchStrategy_NameOnly     = "NameOnly"
	FieldMatchStrategy_TypeOnly     = "TypeOnly"
	FieldMatchStrategy_Com_NameType = "ComNameType"
	DEFAULT_STRATEGY                = FieldMatchStrategy_NameOnly
)

// FieldMatchStrategy Interface
type FieldMatchStrategy interface {
	name() string
	findObjByTFieldInfo(nameObjMap map[string]*ObjInfo, fieldInfo *InjectFieldInfo) (*ObjInfo, error)
}

// NameMatchStrategy using name
type NameMatchStrategy struct{}

func (n *NameMatchStrategy) name() string {
	return FieldMatchStrategy_NameOnly
}

func (n *NameMatchStrategy) findObjByTFieldInfo(nameObjMap map[string]*ObjInfo, fieldInfo *InjectFieldInfo) (*ObjInfo, error) {
	if n == nil {
		return nil, nil
	}
	return nameObjMap[fieldInfo.tag.injectName], nil
}

// TypeMatchStrategy using type
type TypeMatchStrategy struct{}

func (t *TypeMatchStrategy) name() string {
	return FieldMatchStrategy_TypeOnly
}

func (t *TypeMatchStrategy) findObjByTFieldInfo(nameObjMap map[string]*ObjInfo, fieldInfo *InjectFieldInfo) (*ObjInfo, error) {
	var ret *ObjInfo
	for _, v := range nameObjMap {
		if v.objDefination.reflectType.AssignableTo(fieldInfo.reflectType) {
			if ret != nil {
				return nil, fmt.Errorf("Multiple match %v %v", ret.injectName, v.injectName)
			}
			ret = v
		}
	}
	return ret, nil
}

// MatchStrategyComNameType first using name, then using type
type MatchStrategyComNameType struct{}

func (m *MatchStrategyComNameType) name() string {
	return FieldMatchStrategy_Com_NameType
}

func (m *MatchStrategyComNameType) findObjByTFieldInfo(nameObjMap map[string]*ObjInfo, fieldInfo *InjectFieldInfo) (*ObjInfo, error) {
	sty, err := strategyMap.getStrategy(FieldMatchStrategy_NameOnly).findObjByTFieldInfo(nameObjMap, fieldInfo)
	if err != nil {
		return nil, err
	}
	if sty != nil {
		return sty, nil
	}
	return strategyMap.getStrategy(FieldMatchStrategy_TypeOnly).findObjByTFieldInfo(nameObjMap, fieldInfo)
}

// FieldMatchStrategyMap map
type FieldMatchStrategyMap map[string]FieldMatchStrategy

func (f FieldMatchStrategyMap) regist(fms FieldMatchStrategy) error {
	if _, ok := f[fms.name()]; ok {
		return fmt.Errorf("Strategy named [%v] already exists", fms.name())
	}
	f[fms.name()] = fms
	return nil
}

func (f FieldMatchStrategyMap) getStrategy(name string) FieldMatchStrategy {
	ret, ok := f[name]
	if !ok {
		return f[DEFAULT_STRATEGY]
	}
	return ret
}

func (f FieldMatchStrategyMap) hasStrategy(name string) bool {
	_, ok := f[name]
	return ok
}
