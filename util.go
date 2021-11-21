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

func canRegisterCheck(obj interface{}) error {
	if !isPtr(reflect.TypeOf(obj)) {
		return fmt.Errorf("Regist just support type ptr")
	}
	return nil
}

func canInjectCheck(targetType, instanceType reflect.Type) bool {
	if targetType.Kind() == reflect.Interface {
		if instanceType.AssignableTo(targetType) {
			return true
		}
		fmt.Printf("[%v] not AssignableTo [%v]\n", instanceType.Name(), targetType.Name())
		return false
	}
	if isStructPtr(targetType) {
		targetRealType := realType(targetType)
		instanceRealType := realType(instanceType)
		if instanceRealType.AssignableTo(targetRealType) {
			return true
		}
		fmt.Printf("[%v] not AssignableTo [%v] (isStructPtr)\n", instanceType.Name(), targetType.Name())
		return false
	}
	fmt.Printf("canInjectCheck unknown\n")
	return false
}

func parseInjectTag(tag reflect.StructTag) string {
	if injectName, foundtag := tag.Lookup(InjectTag_Name); foundtag {
		return injectName
	}
	return ""
}
