package dataflow

import "sort"

type Registers []Register

type RegisterType int

const (
	TextRegister RegisterType = iota
	NumberRegister
	EnumRegister
)

type Register interface {
	Category() string
	Name() string
	Description() string
	RegisterType() RegisterType
	Unit() string
	Sort() int
}

type RegisterStruct struct {
	category     string
	name         string
	description  string
	registerType RegisterType
	unit         string
	sort         int
}

func CreateRegisterStruct(
	category, name, description string,
	registerType RegisterType,
	unit string,
	sort int,
) RegisterStruct {
	return RegisterStruct{
		category:     category,
		name:         name,
		description:  description,
		registerType: registerType,
		unit:         unit,
		sort:         sort,
	}
}

func (r RegisterStruct) Category() string {
	return r.category
}

func (r RegisterStruct) Name() string {
	return r.name
}

func (r RegisterStruct) Description() string {
	return r.description
}

func (r RegisterStruct) RegisterType() RegisterType {
	return r.registerType
}

func (r RegisterStruct) Unit() string {
	return r.unit
}

func (r RegisterStruct) Sort() int {
	return r.sort
}

func FilterRegisters(input Registers, excludeFields []string, excludeCategories []string) (output Registers) {
	output = make(Registers, 0, len(input))
	for _, r := range input {
		if RegisterNameExcluded(excludeFields, r) {
			continue
		}
		if RegisterCategoryExcluded(excludeCategories, r) {
			continue
		}
		output = append(output, r)
	}
	return
}

func SortRegisters(input Registers) Registers {
	sort.SliceStable(input, func(i, j int) bool { return input[i].Sort() < input[j].Sort() })
	return input
}

func RegisterNameExcluded(exclude []string, r Register) bool {
	for _, e := range exclude {
		if e == r.Name() {
			return true
		}
	}
	return false
}

func RegisterCategoryExcluded(exclude []string, r Register) bool {
	for _, e := range exclude {
		if e == r.Category() {
			return true
		}
	}
	return false
}
