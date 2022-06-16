package serializer

import (
	"testing"

	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/sample"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
)

var (
	binaryFile = "../tmp/laptop.bin"
	jsonFile   = "../tmp/laptop.json"
)

func TestWriteProtobufToBinaryFile(t *testing.T) {
	t.Parallel() //允许并发
	type args struct {
		message  proto.Message
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test1", args: args{message: sample.NewLaptop(), filename: binaryFile}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WriteProtobufToBinaryFile(tt.args.message, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("WriteProtobufToBinaryFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			err := WriteProtobufToJsonFile(tt.args.message, jsonFile)
			require.NoError(t, err)
		})
	}
}

func TestReadProtobufFromBinaryFile(t *testing.T) {
	t.Parallel() //允许并发
	laptop := sample.NewLaptop()
	err := WriteProtobufToBinaryFile(laptop, binaryFile)
	require.NoError(t, err)
	msg := new(pb.Laptop)
	type args struct {
		filename string
		message  *pb.Laptop
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test1", args: args{filename: binaryFile, message: msg}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ReadProtobufFromBinaryFile(tt.args.filename, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("ReadProtobufFromBinaryFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.args.message.Id, laptop.Id)
		})
	}
}
