package service

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

type ImageStore interface {
	// Save 保存图片返回ID
	Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error)
	// MaxSize 最大字节树
	MaxSize() uint64
}

type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string                //保存路径
	images      map[string]*ImageInfo //图像ID对应信息
	maxSize     uint64                //大小限制
}

// ImageInfo 图像信息
type ImageInfo struct {
	LaptopID string
	Type     string
	Path     string
}

func NewDiskImageStore(imageFolder string, maxSize uint64) *DiskImageStore {
	return &DiskImageStore{imageFolder: imageFolder, images: map[string]*ImageInfo{}, maxSize: maxSize}
}

func (d *DiskImageStore) Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error) {
	imageID, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("cannot generate uuid:%v", err)
	}
	imagePath := fmt.Sprintf("%s/%s%s", d.imageFolder, imageID, imageType)
	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create disk image:%v", err)
	}
	if _, err := imageData.WriteTo(file); err != nil { //写入文件
		return "", fmt.Errorf("cannot write disk image:%v", err)
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.images[imageID.String()] = &ImageInfo{ //保存映射
		LaptopID: laptopID,
		Type:     imageType,
		Path:     imagePath,
	}
	return imageID.String(), nil
}

func (d *DiskImageStore) MaxSize() uint64 {
	return d.maxSize
}
