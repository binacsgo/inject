# inject
Dependency Injection framework written in Go.



## 0 Quick Start

```go
    inject.Regist("A", &AImpl{})
    inject.Regist("B", &BImpl{})
    // ... 
    err := inject.DoInject()
    if err != nil {
        panic(err.Error())
    }
```



## 1 What & Why

[如何用最简单的方式解释依赖注入?依赖注入是如何实现解耦的?——知乎](https://www.zhihu.com/question/32108444)

简言之：

> 依赖注入，全称是“依赖注入到容器”， 容器（IOC容器）是一个设计模式，它也是个对象，你把某个类（不管有多少依赖关系）放入这个容器中，可以“解析”出这个类的实例。



## 2 How

Go 如何实现依赖注入？

首先需要熟悉[反射](https://godoc.org/reflect)，在 Go 中：

> 通过 reflect.Type 获取结构体成员信息 reflect.StructField 结构中的 Tag 被称为结构体标签（Struct Tag）。结构体标签是对结构体字段的额外信息标签。



前提：

1. 接口：定义(指针类型)
2. 结构：接口功能实现

从结构体标签中，获取其字段属性键值对。对于：

```go
// A interface B
type A interface{}

// AImpl implement
type AImpl struct{}

// B interface B
type B interface{}

// BImpl implement
type BImpl struct {
    Ba A `inject-name:"A"`
}
```

1. 可以通过实例的 `reflect.TypeOf` 获取该对象实例的 `Type`；
2. 如果 `Type` 是结构体，可以通过 `NumField()` 和 `Field()` 方法获得结构体成员的详细信息；

>对于 `BImpl`
>
>>  `NumField() 值为 1` 因为其有一个结构体成员
>
>> `Field(0)`  将返回 A 的 `StructField` ，而 `StructField` 又包含一个名为 `Tag` 类型值为`StructTag` 的结构。可以依靠 `Lookup()` 寻找字段标签值，形如：
>>
>> `tag.Lookup("inject-name") 值为 "A"`

3. 分析每一个 `对象1` 和 `对象1的结构体成员` 并创建相应的数据结构，依据 tag 值在容器中找到 `对象1的结构体成员所对应的真实对象2` 并依靠反射将 `对象1的结构体成员` 的值设置为其对应的 `真实对象2` ；
4. 完成依赖注入。



## 3 Let's do it!

### 3.1 Container

负责保存实例化的对象，并**依据不同的策略执行注入**。

对外暴露三个接口：

1. Regist
2. DoInject
3. Report

#### 3.1.1 Regist

首先通过 `Regist()` 将所有对象加入容器（内部实现为校验对象类型[Only ptr]、创建对象信息结构体[objInfo]、加入队列[list]）。

在这一步会对每一个对象的所有结构体成员进行解析 [objInfo, injectList] ，详见3.2。

#### 3.1.2 DoInject

遍历容器存储对象的队列，为每一个对象执行 `injectFields()` 操作。该操作依赖3.1.1中对对象所有结构体成员的解析结果，实现为：

1. 遍历 [对象1] 信息结构体[objInfo]的 `injectList` 
2. 依据不同的策略找到结构体成员所对应的真实 [对象2]
3. 为 [对象1] 和其结构体成员 [对象2] 执行注入 `doInject` ，注入本质为 `reflectValue.Set()`

### 3.2 Strategy

3.1.2 中寻找结构体成员所对应的真实对象时所依据的策略，模块分为三种：

1. 匹配 `name` (结构体成员 tag 中对应的值)

   > 对于本模块`inject-name:"xxx"` 则该成员的 name 即为 "xxx"

1. 匹配 `type` (reflect.Type)
2. 优先 `name` ，不存在则使用 `type`

### 3.2 ObjInfo

保存每一个加入容器的对象信息。

```go
type ObjInfo struct {
    name           string
    order          int32
    objDefination  ObjDefination
    injectComplete bool
    instance       interface{}
}

type ObjDefination struct {
    reflectType  reflect.Type
    reflectValue reflect.Value
    injectList   []InjectFieldInfo
}
```

`name` 保存其 tag 对应的值；

`order` 保存其加入容器的顺序；

`objDefination` 保存对象相关属性，其中 `injectList` 保存其所有结构体成员变量的相关信息（reflectValue、reflectType、fieldName）

> 该信息 [InjectFieldInfo] 保存了该成员的 `tag`  `objInfo指针` 等关键内容 ；而 `objInfo指针` 正是注入完成后找到成员对象的关键。

`injectComplete` 表明其是否已经注册完成；

`instance` 保存对象实例。



## 4 TODO

注入优先级