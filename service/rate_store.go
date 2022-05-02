package service

import "sync"

type RateStore interface {
	Add(laptopID string, score float64) (*Rating, error)
}

type Rating struct {
	Count uint32  // 评级次数
	Sum   float64 //评分总和
}

type InMemoryRateStoreStore struct {
	mutex   sync.RWMutex
	ratings map[string]*Rating //电脑ID -> 评级对象
}

func NewInMemoryRateStoreStore() *InMemoryRateStoreStore {
	return &InMemoryRateStoreStore{mutex: sync.RWMutex{}, ratings: map[string]*Rating{}}
}

func (store *InMemoryRateStoreStore) Add(laptopID string, score float64) (*Rating, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	raing := store.ratings[laptopID]
	if raing == nil {
		raing = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		raing.Count++
		raing.Sum += score
	}
	store.ratings[laptopID] = raing
	return raing, nil
}
