package service_test

import (
	"testing"

	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/sample"
	"github.com/0RAJA/RPC/service"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//测试
func TestLaptopServer_CreateLaptop(t *testing.T) {
	t.Parallel() //允许并行
	type fields struct {
		Store service.LaptopStore
	}
	type args struct {
		ctx context.Context
		req *pb.CreateLaptopRequest
	}
	store := service.NewInMemoryLaptopStore() //共享一个store

	defaultContext := context.Background()

	laptopNoID := sample.NewLaptop()
	laptopNoID.Id = ""

	laptopInvalidID := sample.NewLaptop()
	laptopInvalidID.Id = "invalid-uuid"

	alreadyExistsLaptop := sample.NewLaptop()
	res, err := service.NewLaptopServer(store, nil, nil).CreateLaptop(defaultContext, &pb.CreateLaptopRequest{Laptop: alreadyExistsLaptop})
	require.NoError(t, err)
	require.Equal(t, res.Id, alreadyExistsLaptop.Id)

	tests := []struct {
		name    string
		fields  fields
		args    args
		code    codes.Code
		wantErr bool
	}{
		{
			name:   "success_with_id",
			fields: fields{Store: store},
			args:   args{ctx: defaultContext, req: &pb.CreateLaptopRequest{Laptop: sample.NewLaptop()}},
			code:   codes.OK,
		},
		{
			name:   "success_with_no_id",
			fields: fields{Store: store},
			args:   args{ctx: defaultContext, req: &pb.CreateLaptopRequest{Laptop: laptopNoID}},
			code:   codes.OK,
		},
		{
			name:    "failure_invalid_id",
			fields:  fields{Store: store},
			args:    args{ctx: defaultContext, req: &pb.CreateLaptopRequest{Laptop: laptopInvalidID}},
			code:    codes.InvalidArgument,
			wantErr: true,
		},
		{
			name:    "failure_id_already_exist",
			fields:  fields{Store: store},
			args:    args{ctx: defaultContext, req: &pb.CreateLaptopRequest{Laptop: alreadyExistsLaptop}},
			code:    codes.AlreadyExists,
			wantErr: true,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			laptop := service.NewLaptopServer(tt.fields.Store, nil, nil)
			got, err := laptop.CreateLaptop(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, st.Code(), tt.code)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, got)
				if id := tt.args.req.Laptop.Id; len(id) > 0 {
					require.Equal(t, id, got.Id)
				}
			}
		})
	}
}
