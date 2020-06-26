package inject

import (
	"fmt"
	"reflect"
)

func isPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr
}

func isStructPtr(t reflect.Type) bool {
	return isPtr(t) && t.Elem().Kind() == reflect.Struct
}

func realType(t reflect.Type) reflect.Type {
	if isStructPtr(t) {
		return t.Elem()
	}
	return t
}

const (
	InjectTag_Name      = "inject-name"
	InjectTag_Required  = "inject-required"
	InjectTag_Strategy  = "inject-strategy"
	InjectTagFlag_True  = "true"
	InjectTagFlag_False = "flase"
)

// InjectTag tag
type InjectTag struct {
	name     string
	required bool
	strategy string
}

func newDefaultTag() InjectTag {
	return InjectTag{
		name:     "",
		required: true,
		strategy: FieldMatchStrategy_Com_NameType,
	}
}

func tagParse(tag reflect.StructTag) (InjectTag, bool) {
	ok := false
	var foundtag bool
	ret := newDefaultTag()
	if ret.name, foundtag = tag.Lookup(InjectTag_Name); foundtag {
		ok = true
	} else {
		return ret, false
	}
	if tagValue, foundtag := tag.Lookup(InjectTag_Required); foundtag {
		if tagValue == InjectTagFlag_False {
			ret.required = false
		}
	}
	if tagValue, foundtag := tag.Lookup(InjectTag_Strategy); foundtag {
		ret.strategy = tagValue
	}
	return ret, ok
}

// InjectFieldInfo field info
type InjectFieldInfo struct {
	fieldName    string
	tag          InjectTag
	reflectType  reflect.Type
	reflectValue reflect.Value
	objInfo      *ObjInfo
}

func newInjectFieldInfo(tag reflect.StructTag) (ret InjectFieldInfo, ok bool) {
	injectTag, ok := tagParse(tag)
	ret.tag = injectTag
	return ret, ok
}

// ObjDefination obj def
type ObjDefination struct {
	reflectType  reflect.Type
	reflectValue reflect.Value
	injectList   []InjectFieldInfo
}

func (def *ObjDefination) defination(obj interface{}) error {
	reflectType := reflect.TypeOf(obj)
	if !isStructPtr(reflectType) {
		return fmt.Errorf("Just support struct ptr")
	}
	def.reflectType = reflectType
	def.reflectValue = reflect.ValueOf(obj)
	def.injectList = make([]InjectFieldInfo, 0)

	fieldNum := def.reflectType.Elem().NumField()
	for i := 0; i < fieldNum; i++ {
		fieldtag := def.reflectType.Elem().Field(i).Tag
		if fieldInfo, ok := newInjectFieldInfo(fieldtag); ok {
			fieldInfo.reflectValue = def.reflectValue.Elem().Field(i)
			fieldInfo.reflectType = fieldInfo.reflectValue.Type()
			fieldInfo.fieldName = def.reflectType.Elem().Field(i).Name
			def.injectList = append(def.injectList, fieldInfo)
		}
	}
	return nil
}

// ObjInfo obj info
type ObjInfo struct {
	name           string
	order          int32
	objDefination  ObjDefination
	injectComplete bool
	instance       interface{}
}

func newObjInfo(name string, order int32, obj interface{}) (*ObjInfo, error) {
	newObjInfo := ObjInfo{name: name, instance: obj, order: order}
	if err := newObjInfo.objDefination.defination(obj); err != nil {
		return nil, err
	}
	return &newObjInfo, nil
}
