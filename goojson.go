package goojson

import (
	"errors"
	"encoding/json"
	"strings"
	"io/ioutil"
)


var (
	// ErrOutOfBounds - Index out of bounds.
	ErrOutOfBounds = errors.New("out of bounds")

	// ErrNotObjOrArray - The target is not an object or array type.
	ErrNotObjOrArray = errors.New("not an object or array")

	// ErrNotObj - The target is not an object type.
	ErrNotObj = errors.New("not an object")

	// ErrNotArray - The target is not an array type.
	ErrNotArray = errors.New("not an array")

	// ErrPathCollision - Creating a path failed
	ErrPathCollision = errors.New("encountered value collision whilst building path")

	// ErrInvalidPath - The filepath was not valid.
	ErrInvalidPath = errors.New("invalid file path")
)

type Container struct {
	object interface{}
}

func (g *Container) Data() interface{}  {
	if g == nil{
		return nil
	}
	return g.object
}
//------------------------------

func (g *Container) Path(path string) *Container {
	return g.Search(strings.Split(path,".")... )
}
func (g *Container) Search(hierarchy ...string) *Container {
	var object interface{}

	object = g.Data()
	for target := 0; target < len(hierarchy) ; target++ {
		if mmap,ok := object.(map[string]interface{});ok{
			object ,ok = mmap[ hierarchy[target] ]
			if !ok{
				return nil
			}
		}else if marray ,ok := object.([]interface{});ok {
			tmpArray := []interface{}{}
			for _,val := range marray{
				tmpGabs := &Container{val}
				res := tmpGabs.Search( hierarchy[target:]... )
				if res != nil{
					tmpArray = append(tmpArray,res.Data())
				}
			}
			if len(tmpArray) == 0{
				return nil
			}
			return &Container{tmpArray}
		}else{
			return nil
		}
	}
	return &Container{object}
}
// S - Shorthand method, does the same thing as Search.
func (g *Container) S(hierarchy ...string) *Container {
	return g.Search(hierarchy...)
}
//------------------------------
func (g *Container) Exists(hierarchy ...string)bool {
	return g.Search(hierarchy...) != nil
}
func (g *Container) ExistsP(path string)bool {
	return g.Exists(strings.Split(path,",")...)
}

//------------------------------
func (g *Container) Index(index int) *Container {
	if array , ok := g.Data().([]interface{});ok{
		if index >= len(array){
			return &Container{nil}
		}
		return &Container{array[index]}
	}
	return &Container{nil}
}

//------------------------------
func (g *Container) Set(value interface{},path ...string) (*Container,error) {
	if len(path) == 0{
		g.object = value
		return g,nil
	}
	var object interface{}
	if g.object == nil{
		g.object = map[string]interface{}{}
	}
	object = g.object
	for target := 0; target < len(path); target++ {
		if mmap, ok := object.(map[string]interface{}); ok {
			if target == len(path)-1{
				mmap[path[target]] = value
			}else if mmap[path[target]] == nil{
				mmap[path[target]] = map[string]interface{}{}
			}
			object = mmap[path[target]]
		}else{
			return &Container{nil}, ErrPathCollision
		}
	}
	return &Container{object}, nil
}
func (g *Container) SetP(value interface{}, path string) (*Container, error) {
	return g.Set(value, strings.Split(path, ".")...)
}

func (g *Container) SetIndex(value interface{},index int)(*Container,error)  {
	if array ,ok := g.Data().([]interface{});ok{
		if index >= len(array){
			return &Container{nil}, ErrOutOfBounds
		}
		array[index] = value
		return &Container{array[index]},nil
	}
	return &Container{nil}, ErrNotArray
}
//-----------------------------------
func (g *Container) Array(path ...string)(*Container,error)  {
	return g.Set([]interface{}{},path...)
}
func (g *Container) ArrayP(path string) (*Container, error) {
	return g.Array(strings.Split(path, ".")...)
}

func (g *Container) ArrayOfSize(size int,path ...string)(*Container,error){
	a := make([]interface{},size)
	return g.Set(a,path...)
}
func (g *Container) ArrayOfSizeP(size int, path string) (*Container, error) {
	return g.ArrayOfSize(size, strings.Split(path, ".")...)
}
//------------------------------
func (g *Container) Delete(path ...string) error {
	var object interface{}

	if g.object == nil{
		return ErrNotObj
	}
	object = g.object
	for target := 0; target < len(path); target++ {
		if mmap ,ok := object.(map[string]interface{});ok{
			if target == len(path) -1{
				if _,ok := mmap[path[target]]; ok{
					delete(mmap,path[target])
				}else{
					return ErrNotObj
				}
			}
			object = mmap[path[target]]
		}else{
			return ErrNotObj
		}
	}
	return nil
}
func (g *Container) DeleteP(path string) error {
	return g.Delete(strings.Split(path, ".")...)
}
//------------------------------
func (g *Container) ArrayAppend(value interface{}, path ...string) error {
	if array ,ok := g.Search(path...).Data().([]interface{});ok{
		array = append(array,value)
		_,err := g.Set(array,path...)
		return err
	}
	newArray := []interface{}{}
	if v := g.Search(path...).Data();v != nil{
		newArray = append(newArray,v)
	}
	newArray = append(newArray,value)
	_, err := g.Set(newArray, path...)
	return err
}
func (g *Container) ArrayAppendP(value interface{}, path string) error {
	return g.ArrayAppend(value, strings.Split(path, ".")...)
}
func (g *Container) ArrayRemove(index int, path ...string) error {
	if index <0{
		return ErrOutOfBounds
	}
	array ,ok:= g.Search(path...).Data().([]interface{})
	if !ok{
		return ErrNotArray
	}
	if index < len(array){
		array = append(array[:index],array[index+1:]...)
	}else{
		return ErrOutOfBounds
	}
	_,err := g.Set(array,path...)
	return err
}
func (g *Container) ArrayRemoveP(index int, path string) error {
	return g.ArrayRemove(index, strings.Split(path, ".")...)
}

//------------------------------
func (g *Container) ArrayElement(index int,path ...string)(*Container,error){
	if index < 0 {
		return &Container{nil}, ErrOutOfBounds
	}
	array, ok := g.Search(path...).Data().([]interface{})
	if !ok {
		return &Container{nil}, ErrNotArray
	}
	if index < len(array) {
		return &Container{array[index]}, nil
	}
	return &Container{nil}, ErrOutOfBounds
}
func (g *Container) ArrayElementP(index int, path string) (*Container, error) {
	return g.ArrayElement(index, strings.Split(path, ".")...)
}

//------------------------------
func (g *Container) ArrayCount(path ...string) (int, error) {
	if array,ok := g.Search(path...).Data().([]interface{});ok{
		return len(array),nil
	}
	return 0,ErrNotArray
}
func (g *Container) ArrayCountP(path string) (int, error) {
	return g.ArrayCount(strings.Split(path, ".")...)
}

//------------------------------
func (g *Container) Bytes() []byte {
	if g.Data() != nil{
		if bytes ,err := json.Marshal(g.object);err == nil{
			return bytes
		}
	}
	return []byte("{}")
}

//带格式，缩进
func (g *Container) BytesIndent(prefix string, indent string) []byte {
	if g.object != nil {
		if bytes, err := json.MarshalIndent(g.object, prefix, indent); err == nil {
			return bytes
		}
	}
	return []byte("{}")
}
func (g *Container) String() string {
	return string(g.Bytes())
}
func (g *Container) StringIndent(prefix string, indent string) string {
	return string(g.BytesIndent(prefix, indent))
}
//------------------------------

func New() *Container  {
	return &Container{map[string]interface{}{} }
}
func Consume(root interface{}) (*Container, error) {
	return &Container{root}, nil
}
func ParseJSON(sample []byte)(*Container,error){
	var c Container
	if err := json.Unmarshal(sample,c.object);err != nil{
		return nil,err
	}
	return &c, nil
}
func ParseJSONFile(path string) (*Container, error) {
	if len(path) > 0 {
		cBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		container, err := ParseJSON(cBytes)
		if err != nil {
			return nil, err
		}

		return container, nil
	}
	return nil, ErrInvalidPath
}
