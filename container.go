package inject

import (
	"container/list"
	"fmt"
	"reflect"
	"sync"
)

// BeforeVisitor before visitor
type BeforeVisitor interface {
	BeforeInject()
}

// AfterVisitor after visitor
type AfterVisitor interface {
	AfterInject() error
}

// Properties todo
type Properties interface {
	getValue(key string) string
}

// ObjFactory todo
type ObjFactory interface {
	createObj() interface{}
}

// Container container of objects
type Container struct {
	objMap                map[string]*ObjInfo
	definationMap         map[reflect.Type]*ObjDefination
	fieldMatchStrategyMap FieldMatchStrategyMap
	registList            *list.List
	order                 int32
	mutex                 sync.Mutex
}

// Regist regist to container
func (ic *Container) Regist(name string, obj interface{}) {
	ic.regist(name, obj)
}

// DoInject inject to container
func (ic *Container) DoInject() error {
	return ic.inject()
}

// Report report the container
func (ic *Container) Report() string {
	return ""
}

func (ic *Container) regist(name string, obj interface{}) error {
	if len(name) <= 0 || obj == nil {
		return fmt.Errorf("invalid Param")
	}
	if err := ic.canRegisterCheck(obj); err != nil {
		fmt.Printf("Object [%s] Regist fail: %s\n", name, err.Error())
		return err
	}
	ic.mutex.Lock()
	defer ic.mutex.Unlock()
	if _, ok := ic.objMap[name]; ok {
		return fmt.Errorf("Object name %s already exist", name)
	}
	ic.order++
	newObj, err := newObjInfo(name, ic.order, obj)
	if err != nil {
		return err
	}
	ic.registList.PushBack(newObj)
	ic.objMap[name] = newObj
	return nil
}

func (ic *Container) canRegisterCheck(obj interface{}) error {
	if !isPtr(reflect.TypeOf(obj)) {
		return fmt.Errorf("Regist just support type ptr")
	}
	return nil
}

func (ic *Container) inject() error {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()
	pedding := make([]*ObjInfo, 0)
	elem := ic.registList.Front()
	for elem != nil {
		objInfo := elem.Value.(*ObjInfo)
		if objInfo.injectComplete {
			continue
		}
		ic.invokeInjectBefore(objInfo)
		if err := ic.injectFields(objInfo); err != nil {
			fmt.Printf("Inject fail, objInfo.name=%v\n", objInfo.name)
			return err
		}
		pedding = append(pedding, objInfo)
		elem = elem.Next()
	}
	return ic.invokeInjectAfter(pedding)
}

func (ic *Container) invokeInjectBefore(obj *ObjInfo) {
	if visitor, ok := obj.instance.(BeforeVisitor); ok {
		visitor.BeforeInject()
	}
}

func (ic *Container) invokeInjectAfter(objList []*ObjInfo) error {
	for _, obj := range objList {
		if visitor, ok := obj.instance.(AfterVisitor); ok {
			err := visitor.AfterInject()
			if err != nil {
				return err
			}
		}
		obj.injectComplete = true
	}
	return nil
}

func (ic *Container) injectFields(objInfo *ObjInfo) error {
	if objInfo.instance == nil {
		return nil
	}
	listlen := len(objInfo.objDefination.injectList)
	for i := 0; i < listlen; i++ {
		injectObj, err := ic.findObjByTFieldInfo(&(objInfo.objDefination.injectList[i]))
		if err != nil {
			return err
		}
		if injectObj != nil {
			fmt.Printf("Object [%s] Inject field[%s] -> %s\n", objInfo.name, objInfo.objDefination.injectList[i].fieldName, injectObj.name)
			doInject(&(objInfo.objDefination.injectList[i]), injectObj)
		} else {
			if objInfo.objDefination.injectList[i].tag.required {
				fmt.Printf("Object [%s] Inject fail field[%s] -> not found\n", objInfo.name, objInfo.objDefination.injectList[i].fieldName)
				return fmt.Errorf("Object [%s] Inject fail field[%s] -> not found", objInfo.name, objInfo.objDefination.injectList[i].fieldName)
			}
			fmt.Printf("Object [%s] Inject fail field[%s] -> not found (not required)\n", objInfo.name, objInfo.objDefination.injectList[i].fieldName)
		}
	}
	return nil
}

func (ic *Container) findObjByTFieldInfo(fieldinfo *InjectFieldInfo) (*ObjInfo, error) {
	depObj, err := ic.fieldMatchStrategyMap.getStrategy(fieldinfo.tag.strategy).findObjByTFieldInfo(ic.objMap, fieldinfo)
	if err != nil {
		return nil, err
	}
	if depObj == nil {
		return nil, nil
	}
	if !injectFieldCheck(fieldinfo.reflectType, depObj.objDefination.reflectType) {
		fmt.Printf("injectFieldCheck fail %v can not inject to %v\n", depObj.objDefination.reflectType, fieldinfo.reflectType)
		return nil, fmt.Errorf("injectFieldCheck fail %v can not inject to %v", depObj.objDefination.reflectType, fieldinfo.reflectType)
	}
	return depObj, nil
}

func injectFieldCheck(targetType, instanceType reflect.Type) bool {
	if targetType.Kind() == reflect.Interface {
		if instanceType.AssignableTo(targetType) {
			return true
		}
		fmt.Printf("%v not AssignableTo %v\n", instanceType.Name(), targetType.Name())
		return false
	}
	if isStructPtr(targetType) {
		targetRealType := realType(targetType)
		instanceRealType := realType(instanceType)
		if instanceRealType.AssignableTo(targetRealType) {
			return true
		}
		fmt.Printf("%v not AssignableTo %v (isStructPtr)\n", instanceType.Name(), targetType.Name())
		return false
	}
	fmt.Printf("injectFieldCheck unknown\n")
	return false
}

func doInject(fieldinfo *InjectFieldInfo, objInfo *ObjInfo) {
	fmt.Printf("doInject fieldName = %s %v %v %v\n", fieldinfo.fieldName, fieldinfo.reflectValue.CanSet(), fieldinfo.reflectValue.Kind(), objInfo.objDefination.reflectType.AssignableTo(fieldinfo.reflectType))
	fieldinfo.reflectValue.Set(objInfo.objDefination.reflectValue)
	fieldinfo.objInfo = objInfo
}
