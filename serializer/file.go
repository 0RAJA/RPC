package serializer

import (
	"io/ioutil"

	"github.com/golang/protobuf/proto"
)

//序列化对象

// WriteProtobufToBinaryFile 将message对象序列化并写入二进制文件中
func WriteProtobufToBinaryFile(message proto.Message, filename string) error {
	data, err := proto.Marshal(message) //序列化
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(filename, data, 0644); err != nil {
		return err
	}
	return nil
}

func ReadProtobufFromBinaryFile(filename string, message proto.Message) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(data, message); err != nil {
		return err
	}
	return nil
}

func WriteProtobufToJsonFile(message proto.Message, filename string) error {
	data, err := ProtobufToJson(message)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, data, 0664); err != nil {
		return err
	}
	return nil
}
