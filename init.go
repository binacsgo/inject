package inject

var globalContainer *Container

func init() {
	globalContainer = NewContainer()
}

// Regist regist to globalContainer
func Regist(name string, obj interface{}) {
	globalContainer.Regist(name, obj)
}

// DoInject inject to globalContainer
func DoInject() {
	globalContainer.DoInject()
}
