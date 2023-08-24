package dataflow

import "sort"

type Register interface {
	Category() string
	Name() string
	Description() string
	RegisterType() RegisterType
	Enum() map[int]string
	Unit() string
	Sort() int
	Controllable() bool
}

type RegisterStruct struct {
	category     string
	name         string
	description  string
	registerType RegisterType
	enum         map[int]string
	unit         string
	sort         int
	controllable bool
}

func CreateRegisterStruct(
	category, name, description string,
	registerType RegisterType,
	enum map[int]string,
	unit string,
	sort int,
	controllable bool,
) RegisterStruct {
	return RegisterStruct{
		category:     category,
		name:         name,
		description:  description,
		registerType: registerType,
		enum:         enum,
		unit:         unit,
		sort:         sort,
		controllable: controllable,
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

func (r RegisterStruct) Enum() map[int]string {
	return r.enum
}

func (r RegisterStruct) Unit() string {
	return r.unit
}

func (r RegisterStruct) Sort() int {
	return r.sort
}

func (r RegisterStruct) Controllable() bool {
	return r.controllable
}

func FilterRegisters(input []Register, excludeFields []string, excludeCategories []string) (output []Register) {
	output = make([]Register, 0, len(input))
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

func SortRegisters(input []Register) []Register {
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
