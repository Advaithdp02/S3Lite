//Package storage which contains structure of storage and new function to intialize 
package storage

type Storage struct{
	Root string
}
//New is a constructer
func New(root string) *Storage{
	return &Storage{
		Root: root,
	}
}
