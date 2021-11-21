# inject (Sugar)
Dependency Injection framework written in Go.

>   Sugar means a minimalist but efficient framework that only supports name matching strategies



## 1 What

Ref: [如何用最简单的方式解释依赖注入?依赖注入是如何实现解耦的?——知乎](https://www.zhihu.com/question/32108444)

简言之：把有依赖关系的类放到容器中，通过某种方式解析出这些类的实例，就是依赖注入。



## 2  Why

1.   自动处理依赖关系：依赖注入可以自动实现依赖关系的管理和对象属性的赋值。

     Case：

     ```go
     
     ```

2.   不同对象间解耦：将对象的创建与其依赖对象的创建分离，减少耦合。

     Case：

     ```go
     
     ```

3.   Clean Code：降低项目维护成本，增强可拓展性。

     Case：

     ```go
     
     ```

4.   对 Go 来说还有一个好处：方便写单测。

     Case：

     ```go
     // 为当前对象编写单测时 涉及到的其他对象及其功能函数可以直接Mock（只要实现相同的 interface 即可）
     ```



一些已有的依赖注入框架及其对比：

| framework \ difference   | 注入原理                    | 优点                                                       | 缺点                                                         |
| ------------------------ | --------------------------- | ---------------------------------------------------------- | ------------------------------------------------------------ |
| Google / wire            | 基于代码生成 与手写效果相同 | 生成代码 无运行时开销<br>注入类型丰富<br/>无需关注注入顺序 | 产生大量生成代码<br>无法实现注入后的其他初始化操作<br>使用繁琐 |
| Uber / dig               | 基于 reflect 实现运行时注入 | 注入类型丰富                                               | reflect 的运行时开销<br>无法实现注入后的其他初始化操作<br/>需要关注注入顺序<br/>使用繁琐 |
| Facebook(Meta?) / inject | 基于 reflect 实现运行时注入 | 简单可用                                                   | reflect 的运行时开销<br>无法实现注入后的其他初始化操作<br/>需要关注注入顺序 |

>   -   **所有基于 reflect 的框架，需要注入的字段都需要是 public 的** 
>
>   -   **以上框架都只能进行属性赋值，不能或很难执行后续的初始化操作**



## 3 How

### 3.1 明确需求

-   只**关心启动流程中**对象的初始化：反射虽有开销，但只在启动时使用并无大碍

-   为**接口**与**结构体指针**注入：初始化流程中使用到的对象往往都是接口或结构体指针，不需要关心各种复杂类型

-   **低门槛**和**易用性**：

    -   使用方法简单直观

    -   无需关心注入顺序
    -   注入后可执行其它复杂初始化操作

### 3.2 明确思路

-   基于反射实现注入

-   注入类型只支持接口与结构体指针

-   易用：

    -   只提供添加对象和执行注入两个方法

    -   通过拓扑排序自动调整注入顺序

    -   需要执行的对象通过实现统一接口来实现复杂初始化操作

### 3.3 大体流程

1.   提供一个 Container 用以保存所有对象的相关信息
2.   依次为所有对象执行注册流程，将其加入 Container 并通过 reflect 分析并保存其相关信息
3.   执行注入：根据拓扑序依次为对象填充其依赖对象
4.   完成注入

### 3.4 有些枯燥的实现细节

首先需要熟悉[反射](https://godoc.org/reflect)，在 Go 中：

> 通过 reflect.Type 获取结构体成员信息 reflect.StructField 结构中的 Tag 被称为结构体标签（Struct Tag）。结构体标签是对结构体字段的额外信息标签。

#### 3.4.1 Container

对外暴露两个接口：

1. Regist

    通过 `Regist()` 将所有对象加入容器（内部实现为校验对象类型[Only ptr]、创建对象信息结构体[objInfo]、加入队列[list]）。

2. DoInject

    根据对象间的依赖关系构造拓扑序，依据拓扑序遍历对象并为每一个对象执行注入操作，注入本质为 `reflectValue.Set()`。

#### 3.4.2 ObjInfo

保存每一个加入 Container 的对象信息。

```go
type ObjInfo struct {
    injectName string
    order      int64
    instance   interface{}

    objDefination  ObjDefination
}

type ObjDefination struct {
    reflectType  reflect.Type
    reflectValue reflect.Value
    injectList   []*InjectFieldInfo
}
```

`name` 保存其 tag 对应的值；

`order` 保存其加入容器的顺序；

`objDefination` 保存对象相关属性，其中 `injectList` 保存其所有结构体成员变量的相关信息（reflectValue、reflectType、fieldName）

> 该信息 [InjectFieldInfo] 保存了该成员的 `tag`  `objInfo指针` 等关键内容 ；而 `objInfo指针` 正是注入完成后找到成员对象的关键。

`instance` 保存对象实例。

#### 3.4.3 Topology

对象被添加进 Container 时，我们会使用反射对对象的所有字段进行分析。

在所有对象都被添加进容器之后，即可通过分析之前获取的**对象间的依赖关系**构造有向图。

-   如果该有向图存在环，说明在代码实现中存在循环依赖的问题，这样显然是不合法的。
-   否则，该有向图即为**有向无环图**（DAG），我们可以根据拓扑序的逆序依次为每个对象注入其依赖的对象。该次序可以保证对象初始化的顺序合法。



### 3.5 注入流程示例

从结构体标签中，获取其字段属性键值对。对于：

```go
// A: interface A
type A interface{}

// AImpl: implement the A interface
type AImpl struct{}

// B: interface B
type B interface{}

// BImpl: implement the B interface
type BImpl struct {
    Ba A `inject-name:"A"`
}
```

1. 可以通过实例的 `reflect.TypeOf` 获取该对象实例的 `Type`；
2. 如果 `Type` 是结构体，可以通过 `NumField()` 和 `Field()` 方法获得结构体成员的详细信息；

>对于 `BImpl`
>
>>  `NumField() 值为 1 (假定忽略隐式字段)` 因为其有一个结构体成员
>
>> `Field(0)`  将返回 A 的 `StructField` ，而 `StructField` 又包含一个名为 `Tag` 类型值为`StructTag` 的结构。可以依靠 `Lookup()` 寻找字段标签值，形如：
>>
>> `tag.Lookup("inject-name") 值为 "A"`

3. 分析每一个 `对象1` 和 `对象1的结构体成员` 并创建相应的数据结构，依据 tag 值在容器中找到 `对象1的结构体成员所对应的真实对象2` 并依靠反射将 `对象1的结构体成员` 的值设置为其对应的 `真实对象2` ；
4. 完成依赖注入。



## 4 到底怎么用？到底好在哪？

### 4.1 简洁地处理多层级依赖关系

使用依赖注入前，需要依次构造对象并作为参数来构造其他对象：

```go
package main

type A struct{}

type B struct {
    Ba *A
}

type C struct {
    Cb *B
}

func main() {
    a := A{}
    b := B{Ba: &a}
    c := C{Cb: &b}
    // use c ...
}
```

使用依赖注入后，无需关心依赖关系并使用复杂参数构造其他对象：

```go
type A struct{}

type B struct {
    Ba *A `inject-name:"A"`
}

type C struct {
    Cb *B `inject-name:"B"`
}

func main() {
    inject.Regist("A", &A{})
    inject.Regist("B", &B{})
    inject.Regist("C", &C{})
    inject.DoInject()
    // use C ...
}
```

以某处项目代码为例：

TODO

### 4.2 减少耦合

使用依赖注入前，低层对象依赖的参数将会被层层传递，而高层对象中将不得不出现大量无用参数：

>   如构造 C 的过程中出现对 C 本身无用的 valueA valueB

```go
package main

type A struct {
    valueA int
}

func NewA(valueA int) *A {
    return &A{
        valueA: valueA,
    }
}

type B struct {
    Ba     *A
    valueB int
}

func NewB(valueA, valueB int) *B {
    return &B{
        Ba:     NewA(valueA),
        valueB: valueB,
    }
}

type C struct {
    Cb     *B
    valueC int
}

func NewC(valueA, valueB, valueC int) *C {
    return &C{
        Cb:     NewB(valueA, valueB),
        valueC: valueC,
    }
}

func main() {
    c := NewC(1, 2, 3)
    // use c ...
}
```

使用依赖注入后，将初始化对象所需的参数统一：

```go
package main

import (
    "github.com/binacsgo/inject"
)

type Config struct {
    valueA int
    valueB int
    valueC int
}

func NewConfig(valueA, valueB, valueC int) *Config {
    return &Config{
        valueA: valueA,
        valueB: valueB,
        valueC: valueC,
    }
}

type A struct {
    Config *Config
    valueA int
}

func (A *A) AfterInject() error {
    A.valueA = A.Config.valueA
    return nil
}

type B struct {
    Config *Config
    Ba     *A `inject-name:"A"`
    valueB int
}

func (B *B) AfterInject() error {
    B.valueB = B.Config.valueB
    return nil
}

type C struct {
    Config *Config
    Cb     *B `inject-name:"B"`
    valueC int
}

func (C *C) AfterInject() error {
    C.valueC = C.Config.valueC
    return nil
}

func main() {
    inject.Regist("config", NewConfig(1, 2, 3))
    inject.Regist("A", &A{})
    inject.Regist("B", &B{})
    inject.Regist("C", &C{})
    inject.DoInject()
    // use C ...
}
```

以某处项目代码为例：

TODO

### 4.3 易拓展

使用依赖注入前，假定 C 新增成员 Ca ：

```go
package main

type A struct{}

type B struct {
    Ba *A
}

type C struct {
    Ca *A
    Cb *B
}

func main() {
    a := A{}
    b := B{Ba: &a}
  c := C{Ca: &a, Cb: &b}
    // use c ...
}
```

使用依赖注入后，无需改动构造参数，只需新增结构体成员及其相应注解即可：

```go
type A struct{}

type B struct {
    Ba *A `inject-name:"A"`
}

type C struct {
    Ba *A `inject-name:"A"`
    Cb *B `inject-name:"B"`
}

func main() {
    inject.Regist("A", &A{})
    inject.Regist("B", &B{})
    inject.Regist("C", &C{})
    inject.DoInject()
    // use C ...
}
```

以某处项目代码为例：

TODO

### 4.4 方便单测

使用依赖注入前，对某对象编写单测函数时构造成员和 Mock 函数较复杂。

使用依赖注入后，可以通过实现统一接口的形式来简单 Mock ，极大拓展可覆盖场景，达到极高的测试覆盖率。

```go
package main

import (
    "fmt"

    "github.com/binacsgo/inject"
)

// A interface B
type A interface {
    Hello()
}

// AImpl implement
type AImpl struct{}

// Work implement the B interface
func (a *AImpl) Hello() {
    fmt.Printf("Hello I'm A\n")
}

// B interface B
type B interface {
    Hello()
}

// BImpl implement
type BImpl struct {
    Ba A `inject-name:"A"`
}

// Work implement the B interface
func (b *BImpl) Hello() {
    b.Ba.Hello()
}

func main() {
    b := initService()
    b.Hello()
}

func initService() B {
    b := BImpl{}
    inject.Regist("B", &b)
    inject.Regist("A", &AImpl{})
    inject.DoInject()
    return &b
}

```

```go
// inplement the A interface
type FakeA struct{}

func newFakeA() A {
    return &FakeA{}
}

func (fakeA *FakeA) Hello() {
    fmt.Printf("Hello I'm FakeA used for unittest!")
}

func TestBImpl_Work(t *testing.T) {
    type fields struct {
        Ba A
    }
    tests := []struct {
        name   string
        fields fields
    }{
        {
            name: "test",
            fields: fields{
                Ba: newFakeA(),
            },
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            b := &BImpl{
                Ba: tt.fields.Ba,
            }
            b.Hello()
        })
    }
}
```

以某处项目代码为例：

TODO