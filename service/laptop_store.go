package service

import (
	"errors"
	"fmt"
	"github.com/0RAJA/RPC/pb"
	"github.com/jinzhu/copier"
	"sync"
)

var (
	ErrLaptopAlreadyExists = errors.New("laptop has already exists")
)

type LaptopStore interface {
	// Save 保存Laptop
	Save(laptop *pb.Laptop) error
	Find(id string) (*pb.Laptop, error)
}

type InMemoryLaptopStore struct {
	data  map[string]*pb.Laptop
	mutex sync.RWMutex
}

func (i *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if _, ok := i.data[laptop.Id]; ok {
		return ErrLaptopAlreadyExists
	}
	other := &pb.Laptop{}
	if err := copier.Copy(other, laptop); err != nil { //为了安全复制一个新的
		return fmt.Errorf("cannot copy: %v", err)
	}
	i.data[laptop.Id] = other
	return nil
}
func (i *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	if v, ok := i.data[id]; ok {
		res := &pb.Laptop{}
		if err := copier.Copy(res, v); err != nil {
			return nil, fmt.Errorf("cannot copy: %v", err)
		}
		return res, nil
	}
	return nil, nil
}
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{data: map[string]*pb.Laptop{}, mutex: sync.RWMutex{}}
}

type DBLaptopStore struct {
}
