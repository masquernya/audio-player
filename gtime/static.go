package gtime

var singleton = New()

func Start(label string) {
	singleton.Start(label)
}

func End(label string) {
	singleton.End(label)
}
