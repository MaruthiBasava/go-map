package main

type B struct {
	position int
}

type A struct {
	position int
	b        *B
}

func (a *A) B() *B {
	return a.b
}

func main() {

	// err := UnmarshalDomainConfigYaml("domain.yaml")
	// if err != nil {
	// 	panic(err)
	// }

}
