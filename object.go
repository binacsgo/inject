package inject

import (
	"fmt"
	"reflect"
)

const (
	InjectTag_Name      = "inject-name"
	InjectTag_Required  = "inject-required"
	InjectTag_Strategy  = "inject-strategy"
	InjectTagFlag_True  = "true"
	InjectTagFlag_False = "flase"
)

// InjectTag tag
type InjectTag struct {
	injectName string
	required   bool
	strategy   string
}

func parseInjectTag(tag reflect.StructTag) (*InjectTag, bool) {
	ret := &InjectTag{
		injectName: "",
		required:   true,
		strategy:   FieldMatchStrategy_Com_NameType,
	}

	if injectName, foundtag := tag.Lookup(InjectTag_Name); foundtag {
		ret.injectName = injectName
	} else {
		return nil, false
	}
	if tagValue, foundtag := tag.Lookup(InjectTag_Required); foundtag {
		// tag.required is true by default
		if tagValue == InjectTagFlag_False {
			ret.required = false
		}
	}
	if tagValue, foundtag := tag.Lookup(InjectTag_Strategy); foundtag {
		ret.strategy = tagValue
	}
	return ret, true
}

// InjectFieldInfo field info
type InjectFieldInfo struct {
	tag *InjectTag

	fieldName    string
	reflectType  reflect.Type
	reflectValue reflect.Value
	objInfo      *ObjInfo
}

func parseInjectFieldInfo(tag reflect.StructTag) *InjectFieldInfo {
	injectTag, ok := parseInjectTag(tag)
	if !ok {
		return nil
	}
	return &InjectFieldInfo{
		tag: injectTag,
	}
}

// ObjDefination obj def
type ObjDefination struct {
	reflectType  reflect.Type
	reflectValue reflect.Value
	injectList   []*InjectFieldInfo
}

// ObjInfo obj info
type ObjInfo struct {
	injectName string
	order      int64
	instance   interface{}

	objDefination  ObjDefination
	injectComplete bool
}

func newObjInfo(injectName string, order int64, obj interface{}) (*ObjInfo, error) {
	newObjInfo := &ObjInfo{injectName: injectName, instance: obj, order: order}
	if err := newObjInfo.parseObjDefination(); err != nil {
		return nil, err
	}
	return newObjInfo, nil
}

func (obj *ObjInfo) parseObjDefination() error {
	// 1. Get reflectType and reflectValue
	reflectType := reflect.TypeOf(obj.instance)
	if !isStructPtr(reflectType) {
		return fmt.Errorf("Just support struct ptr")
	}
	reflectValue := reflect.ValueOf(obj.instance)

	// 2. Get injectList
	injectList := make([]*InjectFieldInfo, 0)
	fieldNum := reflectType.Elem().NumField()
	for i := 0; i < fieldNum; i++ {
		fielldTag := reflectType.Elem().Field(i).Tag
		if fieldInfo := parseInjectFieldInfo(fielldTag); fieldInfo != nil {
			fieldInfo.reflectValue = reflectValue.Elem().Field(i)
			fieldInfo.reflectType = reflectValue.Elem().Field(i).Type()
			fieldInfo.fieldName = reflectType.Elem().Field(i).Name
			injectList = append(injectList, fieldInfo)
		}
	}

	// Set the objDefination
	obj.objDefination = ObjDefination{
		reflectType:  reflectType,
		reflectValue: reflectValue,
		injectList:   injectList,
	}
	return nil
}
