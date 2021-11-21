package inject

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/binacsgo/graph"
)

// AfterVisitor after visitor
type AfterVisitor interface {
	AfterInject() error
}

// Container container of objects
type Container struct {
	nameObjMap  map[string]*ObjInfo
	orderObjMap map[int64]*ObjInfo
	registList  *list.List

	graph *graph.Graph
	order int64
	conns int64
	mutex sync.Mutex
}

// NewContainer return a container that store all the objects
func NewContainer() *Container {
	return &Container{
		nameObjMap:  make(map[string]*ObjInfo, 16),
		orderObjMap: make(map[int64]*ObjInfo, 16),
		registList:  list.New(),
	}
}

// Regist regist to container
func (ic *Container) Regist(injectName string, obj interface{}) {
	if len(injectName) <= 0 || obj == nil {
		panic(fmt.Errorf("Invalid param: obj=[%v]", obj))
	}
	if err := canRegisterCheck(obj); err != nil {
		panic(fmt.Errorf("Can not pass the canRegisterCheck: err=[%v]", err))
	}

	if err := ic.regist(injectName, obj); err != nil {
		panic(fmt.Errorf("Regist failed for [%v]: err=[%v]", injectName, err))
	}
}

// DoInject inject to container
func (ic *Container) DoInject() {
	if err := ic.inject(); err != nil {
		panic(fmt.Errorf("DoInject failed: err=[%v]", err))
	}
}

func (ic *Container) regist(injectName string, obj interface{}) error {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	// 1. Check for duplicate registrations
	if _, ok := ic.nameObjMap[injectName]; ok {
		return fmt.Errorf("Already exists in nameObjMap")
	}
	// 2. Build the ObjInfo
	newObj, err := newObjInfo(injectName, ic.order, obj)
	if err != nil {
		return fmt.Errorf("Can not pass the newObjInfo: err=[%v]", err)
	}
	// 3. Update the container
	{
		ic.registList.PushBack(newObj)
		ic.nameObjMap[injectName] = newObj
		ic.orderObjMap[ic.order] = newObj
		// [0, order)
		ic.order += 1
		ic.conns += int64(len(newObj.objDefination.injectList))
	}

	return nil
}

func (ic *Container) inject() error {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	ic.graph = graph.NewGraph(ic.order, ic.conns)
	for elem := ic.registList.Front(); elem != nil; elem = elem.Next() {
		objInfo := elem.Value.(*ObjInfo)
		objInfoName := objInfo.injectName
		for i := range objInfo.objDefination.injectList {
			depObjInfoName := objInfo.objDefination.injectList[i].injectName
			ic.graph.AddEdge(ic.nameObjMap[objInfoName].order, ic.nameObjMap[depObjInfoName].order, 1)
		}
	}
	topo, ok := graph.Topology(ic.graph)
	if !ok {
		return fmt.Errorf("Find circle in graph")
	}

	pedding := make([]*ObjInfo, 0)
	for i := len(topo) - 1; i >= 0; i-- {
		order := topo[i]
		objInfo := ic.orderObjMap[order]
		if err := ic.injectFields(objInfo); err != nil {
			return fmt.Errorf("Inject objInfo.injectName[%v] got err=[%v]", objInfo.injectName, err)
		}
		pedding = append(pedding, objInfo)
	}
	return ic.invokeAfterInject(pedding)
}

func (ic *Container) invokeAfterInject(objList []*ObjInfo) error {
	for _, obj := range objList {
		if visitor, ok := obj.instance.(AfterVisitor); ok {
			err := visitor.AfterInject()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ic *Container) injectFields(objInfo *ObjInfo) error {
	if objInfo.instance == nil {
		return nil
	}
	injectListLen := len(objInfo.objDefination.injectList)
	for i := 0; i < injectListLen; i++ {
		injectObj, err := ic.findObjByTFieldInfo(objInfo.objDefination.injectList[i])
		if err != nil {
			return err
		}
		if injectObj == nil {
			return fmt.Errorf("Object [%v] Inject fail field[%v] -> not found", objInfo.injectName, objInfo.objDefination.injectList[i].fieldName)
		}

		fmt.Printf("Object [%v] Inject field[%v] -> [%v]\n", objInfo.injectName, objInfo.objDefination.injectList[i].fieldName, injectObj.injectName)
		doInject(objInfo.objDefination.injectList[i], injectObj)
	}
	return nil
}

func (ic *Container) findObjByTFieldInfo(fieldinfo *InjectFieldInfo) (*ObjInfo, error) {
	depObj := ic.nameObjMap[fieldinfo.injectName]
	if depObj == nil {
		// DO NOT return nil, nil
		return nil, fmt.Errorf("depObj [%v]=nil, this should not happen if we perform injection by topology", fieldinfo.injectName)
	}
	if !canInjectCheck(fieldinfo.reflectType, depObj.objDefination.reflectType) {
		return nil, fmt.Errorf("canInjectCheck fail [%v] can not inject to [%v]", depObj.objDefination.reflectType, fieldinfo.reflectType)
	}
	return depObj, nil
}

func doInject(fieldinfo *InjectFieldInfo, objInfo *ObjInfo) {
	fmt.Printf("doInject: fieldName=[%v] CanSet=[%v] Kind=[%v] AssignableTo=[%v]\n",
		fieldinfo.fieldName,
		fieldinfo.reflectValue.CanSet(),
		fieldinfo.reflectValue.Kind(),
		objInfo.objDefination.reflectType.AssignableTo(fieldinfo.reflectType),
	)
	fieldinfo.reflectValue.Set(objInfo.objDefination.reflectValue)
	fieldinfo.objInfo = objInfo
}
