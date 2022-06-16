package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/0RAJA/RPC/pb"
	"github.com/jinzhu/copier"
	"golang.org/x/net/context"
)

var (
	ErrAlreadyExists = errors.New("already exists")
)

type LaptopStore interface {
	// Save 保存Laptop
	Save(laptop *pb.Laptop) error
	// Find 通过id查找对应laptop
	Find(id string) (*pb.Laptop, error)
	// Search 通过给定条件插叙满足的laptop
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

type InMemoryLaptopStore struct {
	data  map[string]*pb.Laptop
	mutex sync.RWMutex
}

func (i *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if _, ok := i.data[laptop.Id]; ok {
		return ErrAlreadyExists
	}
	other, err := deepCopy(laptop)
	if err != nil {
		return err
	}
	i.data[laptop.Id] = other
	return nil
}

func (i *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	if v, ok := i.data[id]; ok {
		other, err := deepCopy(v)
		if err != nil {
			return nil, err
		}
		return other, nil
	}
	return nil, nil
}

func (i *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	for _, laptop := range i.data {
		if isQuality(filter, laptop) {
			if err := contextErr(ctx); err != nil {
				return err
			}
			res, err := deepCopy(laptop)
			if err != nil {
				return err
			}
			if err := found(res); err != nil {
				return err
			}
		}
	}
	return nil
}

func isQuality(filter *pb.Filter, laptop *pb.Laptop) bool {
	return filter.GetMaxPriceUsd() >= laptop.GetPriceUsd() && filter.GetMinCpuCores() <= laptop.GetCpu().GetNumberCores() && filter.GetMinCpuGhz() <= laptop.GetCpu().GetMinGhz() && toBit(filter.GetMinRam()) <= toBit(laptop.GetRam())
}

//转哈不同单位到bit
func toBit(memory *pb.Memory) (ret uint64) {
	ret = memory.Value
	switch memory.Unit {
	case pb.Memory_BIT:
		return ret
	case pb.Memory_BYTE:
		return ret << 3
	case pb.Memory_KILOBYTE:
		return ret << 13
	case pb.Memory_MEGABYTE:
		return ret << 23
	case pb.Memory_GIGABYTE:
		return ret << 33
	case pb.Memory_TERABYTE:
		return ret << 43
	default:
		return 0
	}
}
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{data: map[string]*pb.Laptop{}, mutex: sync.RWMutex{}}
}

type DBLaptopStore struct {
}

func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	other := &pb.Laptop{}
	if err := copier.Copy(other, laptop); err != nil { //为了安全复制一个新的
		return nil, fmt.Errorf("cannot copy: %v", err)
	}
	return other, nil
}
