package main

func main() {

	err := UnmarshalDomainConfigYaml("domain.yaml")
	if err != nil {
		panic(err)
	}

}
