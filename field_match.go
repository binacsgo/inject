package inject

import (
	"fmt"
)

const (
	FieldMatchStrategy_NameOnly     = "NameOnly"
	FieldMatchStrategy_TypeOnly     = "TypeOnly"
	FieldMatchStrategy_Com_NameType = "ComNameType"
	DEFAULT_STRATEGY                = FieldMatchStrategy_NameOnly
)

// FieldMatchStrategy Interface
type FieldMatchStrategy interface {
	name() string
	findObjByTFieldInfo(objMap map[string]*ObjInfo, fieldInfo *InjectFieldInfo) (*ObjInfo, error)
}

// NameMatchStrategy using name
type NameMatchStrategy struct{}

func (n *NameMatchStrategy) name() string {
	return FieldMatchStrategy_NameOnly
}

func (n *NameMatchStrategy) findObjByTFieldInfo(objMap map[string]*ObjInfo, fieldInfo *InjectFieldInfo) (*ObjInfo, error) {
	if n == nil {
		return nil, nil
	}
	return objMap[fieldInfo.tag.name], nil
}

// TypeMatchStrategy using type
type TypeMatchStrategy struct{}

func (t *TypeMatchStrategy) name() string {
	return FieldMatchStrategy_TypeOnly
}

func (t *TypeMatchStrategy) findObjByTFieldInfo(objMap map[string]*ObjInfo, fieldInfo *InjectFieldInfo) (*ObjInfo, error) {
	var ret *ObjInfo
	for _, v := range objMap {
		if v.objDefination.reflectType.AssignableTo(fieldInfo.reflectType) {
			if ret != nil {
				return nil, fmt.Errorf("Multiple match %v %v", ret.name, v.name)
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

func (m *MatchStrategyComNameType) findObjByTFieldInfo(objMap map[string]*ObjInfo, fieldInfo *InjectFieldInfo) (*ObjInfo, error) {
	sty, err := strategyMap.getStrategy(FieldMatchStrategy_NameOnly).findObjByTFieldInfo(objMap, fieldInfo)
	if err != nil {
		return nil, err
	}
	if sty != nil {
		return sty, nil
	}
	sty, err = strategyMap.getStrategy(FieldMatchStrategy_TypeOnly).findObjByTFieldInfo(objMap, fieldInfo)
	if err != nil {
		return nil, err
	}
	return sty, err
}

// FieldMatchStrategyMap map
type FieldMatchStrategyMap map[string]FieldMatchStrategy

func (f FieldMatchStrategyMap) regist(fms FieldMatchStrategy) error {
	if _, ok := f[fms.name()]; ok {
		return fmt.Errorf("Strategy named [%v] already exist", fms.name())
	}
	f[fms.name()] = fms
	return nil
}

func (f FieldMatchStrategyMap) getStrategy(name string) FieldMatchStrategy {
	ret := f[name]
	if ret == nil {
		return f[DEFAULT_STRATEGY]
	}
	return ret
}

func (f FieldMatchStrategyMap) hasStrategy(name string) bool {
	_, ok := f[name]
	return ok
}
