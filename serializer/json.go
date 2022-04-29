package serializer

import (
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// ProtobufToJson 将protobuf文件转换为json
func ProtobufToJson(message proto.Message) ([]byte, error) {
	marshaler := protojson.MarshalOptions{
		UseEnumNumbers:  true, //枚举值使用数字
		EmitUnpopulated: true, //未填充字段使用默认值
	}
	bin, err := marshaler.Marshal(proto.MessageV2(message))
	if err != nil {
		return nil, err
	}
	return bin, nil
}
